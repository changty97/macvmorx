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
	Orchestrator       *core.Orchestrator
	// Removed GitHubClient dependency
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(cfg *config.Config, hp *heartbeat.Processor, ns *store.NodeStore, orch *core.Orchestrator) *Handlers {
	return &Handlers{
		cfg:                cfg,
		HeartbeatProcessor: hp,
		NodeStore:          ns,
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

// HandleGitHubWebhook receives and processes GitHub webhook events.
func (h *Handlers) HandleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payloadBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading webhook payload: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Verify webhook signature for security
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		log.Println("Webhook signature missing.")
		http.Error(w, "Unauthorized: Signature missing", http.StatusUnauthorized)
		return
	}

	// Use the webhook secret from config for validation
	_, err = gh.ValidatePayload(r, []byte(h.cfg.GitHubWebhookSecret))
	if err != nil {
		log.Printf("Webhook signature validation failed: %v", err)
		http.Error(w, "Unauthorized: Invalid signature", http.StatusUnauthorized)
		return
	}

	eventType := r.Header.Get("X-GitHub-Event")
	log.Printf("Received GitHub webhook event: %s", eventType)

	switch eventType {
	case "workflow_job":
		var event gh.WorkflowJobEvent
		if err := json.Unmarshal(payloadBytes, &event); err != nil {
			log.Printf("Error unmarshalling workflow_job event: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if event.Action != nil && *event.Action == "queued" {
			// Ensure the job is for a self-hosted runner
			isSelfHosted := false
			for _, label := range event.WorkflowJob.Labels {
				if label == "self-hosted" { // Standard label for self-hosted runners
					isSelfHosted = true
					break
				}
			}

			if !isSelfHosted {
				log.Printf("Workflow job '%s' (%d) is queued but not for a self-hosted runner. Skipping.",
					event.WorkflowJob.GetName(), event.WorkflowJob.GetID())
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"message": "Job not for self-hosted runner, skipping"})
				return
			}

			// Check if a runner registration token is configured
			if h.cfg.GitHubRunnerRegistrationToken == "" {
				log.Printf("Error: GitHub runner registration token is not configured. Cannot provision VM for job %d.", event.WorkflowJob.GetID())
				http.Error(w, "GitHub runner registration token not configured", http.StatusInternalServerError)
				return
			}

			log.Printf("Self-hosted workflow job '%s' (%d) queued for repository %s. Labels: %v",
				event.WorkflowJob.GetName(), event.Repo.GetFullName(), event.WorkflowJob.Labels)

			// Extract required image name from labels or a predefined mapping
			// Example: a label like "macos-image-sonoma-runner"
			imageName := "default-macos-image" // Fallback
			for _, label := range event.WorkflowJob.Labels {
				if strings.HasPrefix(label, "macos-image-") {
					imageName = strings.TrimPrefix(label, "macos-image-")
					break
				}
			}
			if imageName == "default-macos-image" {
				log.Printf("Warning: Could not determine image name from labels for job %d. Using default '%s'.", event.WorkflowJob.GetID(), imageName)
			}

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
					log.Printf("Failed to schedule VM for job %d: %v", event.WorkflowJob.GetID(), err)
					// TODO: Potentially update GitHub Check Run status to indicate failure
					// You would need the Check Run ID from the webhook payload to do this.
				} else {
					log.Printf("VM for job %d successfully scheduled on node %s", event.WorkflowJob.GetID(), nodeID)
					// TODO: Potentially update GitHub Check Run status to indicate provisioning started
				}
			}()
		} else if event.Action != nil && (*event.Action == "completed" || *event.Action == "cancelled" || *event.Action == "skipped") {
			log.Printf("Workflow job '%s' (%d) %s. Runner ID: %d",
				event.WorkflowJob.GetName(), event.WorkflowJob.GetID(), *event.Action, event.WorkflowJob.GetRunnerID())
			// TODO: Trigger VM deletion on the agent for this runner ID
			// This would involve mapping runner ID to VMID (which is jobID in our case)
			// and then calling agent.DeleteVM via the orchestrator.
			// For ephemeral runners, GitHub will eventually remove them, but immediate cleanup is better.
		}

	default:
		log.Printf("Unhandled GitHub event type: %s", eventType)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Webhook received and processed"})
}
