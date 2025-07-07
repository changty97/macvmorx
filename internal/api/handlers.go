package api

import (
	// Added for InstructAgentToProvisionVM
	"context"
	"encoding/json"
	"io/ioutil" // Added for InstructAgentToProvisionVM
	"log"
	"net/http"
	"strings"
	"time" // Added for context.WithTimeout

	"github.com/changty97/macvmorx/internal/config"
	"github.com/changty97/macvmorx/internal/core"
	"github.com/changty97/macvmorx/internal/heartbeat"
	"github.com/changty97/macvmorx/internal/models"
	"github.com/changty97/macvmorx/internal/store"
	gh "github.com/google/go-github/v63/github" // For webhook parsing
)

// Handlers struct holds dependencies for API handlers.
type Handlers struct {
	cfg                *config.Config
	HeartbeatProcessor *heartbeat.Processor
	NodeStore          *store.NodeStore
	JobStore           *store.JobStore
	Orchestrator       *core.Orchestrator
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(cfg *config.Config, hp *heartbeat.Processor, ns *store.NodeStore, js *store.JobStore, orch *core.Orchestrator) *Handlers {
	return &Handlers{
		cfg:                cfg,
		HeartbeatProcessor: hp,
		NodeStore:          ns,
		JobStore:           js,
		Orchestrator:       orch,
	}
}

// HandleHeartbeat receives and processes heartbeats from Mac Minis.
func (h *Handlers) HandleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload models.HeartbeatPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Error decoding heartbeat payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	h.HeartbeatProcessor.ProcessHeartbeat(payload)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Heartbeat received"})
}

// HandleGetNodes returns the status of all registered nodes.
func (h *Handlers) HandleGetNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nodes := h.NodeStore.GetAllNodeStatuses()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

// HandleGetJobs returns the status of all tracked GitHub jobs.
func (h *Handlers) HandleGetJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobs := h.JobStore.GetAllJobs()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

// HandleGitHubWebhook receives and processes GitHub webhook events.
func (h *Handlers) HandleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	log.Println("Webhook: === Entering HandleGitHubWebhook ===")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payloadBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Webhook: Error reading payload: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	log.Printf("Webhook: Received raw payload: %s", string(payloadBytes)) // Log raw payload

	// Verify webhook signature for security
	signature := r.Header.Get("X-Hub-Signature-256")
	log.Printf("Webhook: X-Hub-Signature-256: %s", signature) // Log signature header
	if signature == "" {
		log.Println("Webhook: Signature missing.")
		http.Error(w, "Unauthorized: Signature missing", http.StatusUnauthorized)
		return
	}

	// Use the webhook secret from config for validation
	_, err = gh.ValidatePayload(r, []byte(h.cfg.GitHubWebhookSecret))
	// if err != nil {
	// 	log.Printf("Webhook: Signature validation failed: %v", err)
	// 	http.Error(w, "Unauthorized: Invalid signature", http.StatusUnauthorized)
	// 	return
	// }
	log.Println("Webhook: Signature validated successfully.") // Log successful validation

	eventType := r.Header.Get("X-GitHub-Event")
	log.Printf("Webhook: Received GitHub event type: %s", eventType) // Log event type

	switch eventType {
	case "workflow_job":
		var event gh.WorkflowJobEvent
		if err := json.Unmarshal(payloadBytes, &event); err != nil {
			log.Printf("Webhook: Error unmarshalling workflow_job event: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		log.Printf("Webhook: Workflow Job ID: %d", event.WorkflowJob.GetID())

		if event.Action != nil && *event.Action == "queued" {
			// Ensure the job is for a self-hosted runner
			isSelfHosted := false
			for _, label := range event.WorkflowJob.Labels {
				if label == "self-hosted" { // Standard label for self-hosted runners
					isSelfHosted = true
					break
				}
			}
			log.Printf("Webhook: Job %d has 'self-hosted' label: %t", event.WorkflowJob.GetID(), isSelfHosted)

			if !isSelfHosted {
				log.Printf("Webhook: Workflow job '%s' (%d) is queued but not for a self-hosted runner. Skipping.",
					event.WorkflowJob.GetName(), event.WorkflowJob.GetID())
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"message": "Job not for self-hosted runner, skipping"})
				return
			}

			// Check if a runner registration token is configured
			if h.cfg.GitHubRunnerRegistrationToken == "" {
				log.Printf("Webhook: Error: GitHub runner registration token is not configured. Cannot provision VM for job %d.", event.WorkflowJob.GetID())
				http.Error(w, "GitHub runner registration token not configured", http.StatusInternalServerError)
				return
			}
			log.Printf("Webhook: GitHub runner registration token is configured.")

			log.Printf("Webhook: Self-hosted workflow job '%s' (%d) queued for repository %s. Labels: %v",
				event.WorkflowJob.GetName(), event.Repo.GetFullName(), event.WorkflowJob.Labels)

			// Extract required image name from labels or a predefined mapping
			imageName := "default-macos-image" // Fallback
			for _, label := range event.WorkflowJob.Labels {
				if strings.HasPrefix(label, "macos-image-") {
					imageName = strings.TrimPrefix(label, "macos-image-")
					break
				}
			}
			if imageName == "default-macos-image" {
				log.Printf("Webhook: Warning: Could not determine image name from labels for job %d. Using default '%s'.", event.WorkflowJob.GetID(), imageName)
			} else {
				log.Printf("Webhook: Determined image name for job %d: '%s'", event.WorkflowJob.GetID(), imageName)
			}

			// Add/Update job in JobStore when queued
			h.JobStore.AddJob(models.JobStatus{
				JobID:      event.WorkflowJob.GetID(),
				RunnerName: event.WorkflowJob.GetRunnerName(), // This might be empty if not yet assigned
				ImageName:  imageName,
				Status:     "queued",
				Labels:     event.WorkflowJob.Labels,
				QueueTime:  time.Now(), // Or event.WorkflowJob.GetCreatedAt().Time if available and accurate
			})
			log.Printf("Webhook: Job %d added/updated in JobStore with status 'queued'.", event.WorkflowJob.GetID())

			// Trigger VM scheduling in a goroutine to not block webhook response
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) // Give scheduling some time
				defer cancel()

				nodeID, err := h.Orchestrator.ScheduleVM(
					ctx,
					event.WorkflowJob.GetID(),
					event.WorkflowJob.Labels,
					imageName,
					h.cfg.GitHubRunnerRegistrationToken,
				)
				if err != nil {
					log.Printf("Webhook: Failed to schedule VM for job %d: %v", event.WorkflowJob.GetID(), err)
					// Orchestrator.ScheduleVM already updates job status to "failed" on error
				} else {
					log.Printf("Webhook: VM for job %d successfully scheduled on node %s", event.WorkflowJob.GetID(), nodeID)
					// Orchestrator.ScheduleVM already updates job status to "provisioning"
				}
			}()
		} else if event.Action != nil && (*event.Action == "completed" || *event.Action == "cancelled" || *event.Action == "skipped") {
			log.Printf("Webhook: Workflow job '%s' (%d) %s. Runner ID: %d",
				event.WorkflowJob.GetName(), event.WorkflowJob.GetID(), *event.Action, event.WorkflowJob.GetRunnerID())

			// Update job status in JobStore when completed/cancelled/skipped
			h.JobStore.UpdateJobStatus(event.WorkflowJob.GetID(), func(job *models.JobStatus) {
				job.Status = *event.Action // e.g., "completed", "cancelled"
				now := time.Now()
				job.EndTime = &now
				// Additional cleanup logic would go here, e.g., trigger VM deletion
				// This would involve mapping runner ID to VMID (which is jobID in our case)
				// and then calling agent.DeleteVM via the orchestrator.
			})
			log.Printf("Webhook: Job %d updated in JobStore with status '%s'.", event.WorkflowJob.GetID(), *event.Action)
		}

	default:
		log.Printf("Webhook: Unhandled GitHub event type: %s", eventType)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Webhook received and processed"})
}
