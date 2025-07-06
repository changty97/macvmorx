package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/changty97/macvmorx/internal/config"
	"github.com/changty97/macvmorx/internal/models"
	"github.com/changty97/macvmorx/internal/store"
)

// Orchestrator is the central component responsible for managing Mac Mini nodes and VMs.
type Orchestrator struct {
	nodeStore  *store.NodeStore
	cfg        *config.Config // Add config to access config values
	httpClient *http.Client   // HTTP client for agent communication
	mu         sync.Mutex     // Mutex for protecting orchestrator state if needed
}

// NewOrchestrator creates a new Orchestrator instance.
func NewOrchestrator(ns *store.NodeStore, cfg *config.Config) (*Orchestrator, error) {
	// Revert to simple HTTP client
	httpClient := &http.Client{
		Timeout: 10 * time.Second, // Timeout for agent communication
	}

	return &Orchestrator{
		nodeStore:  ns,
		cfg:        cfg,
		httpClient: httpClient,
	}, nil
}

// ScheduleVM attempts to schedule a VM based on the request.
// It prioritizes nodes with the image cached, then nodes that can download it.
// This is now called by the GitHub webhook handler.
func (o *Orchestrator) ScheduleVM(
	ctx context.Context,
	jobID int64,
	runnerLabels []string,
	imageName string, // Image name derived from labels or other config
	runnerRegistrationToken string, // Static token provided by config
) (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	log.Printf("Attempting to schedule VM for GitHub job %d, image: %s, labels: %v", jobID, imageName, runnerLabels)

	var (
		nodeWithImageAvailable *models.NodeStatus = nil
		nodeForDownload        *models.NodeStatus = nil
	)

	nodes := o.nodeStore.GetAllNodeStatuses()

	// First pass: Look for nodes that are online, have less than 2 VMs, and already have the image cached.
	for _, node := range nodes {
		if node.IsOnline && node.VMCount < 2 { // Mac Minis support up to 2 parallel VMs
			hasImage := false
			for _, cachedImage := range node.CachedImages {
				if cachedImage == imageName {
					hasImage = true
					break
				}
			}
			if hasImage {
				nodeWithImageAvailable = node
				log.Printf("Found suitable node %s with image %s cached.", node.NodeID, imageName)
				break // Found a perfect match, prioritize this
			}
		}
	}

	// Second pass (if no perfect match): Look for nodes that are online, have less than 2 VMs,
	// and can download the image (i.e., don't already have 2 VMs).
	if nodeWithImageAvailable == nil {
		for _, node := range nodes {
			if node.IsOnline && node.VMCount < 2 {
				// This node can accept a VM and potentially download the image
				nodeForDownload = node
				log.Printf("Found available node %s that can download image %s.", node.NodeID, imageName)
				break // Found a node that can download, consider this
			}
		}
	}

	var targetNode *models.NodeStatus
	if nodeWithImageAvailable != nil {
		targetNode = nodeWithImageAvailable
	} else if nodeForDownload != nil {
		targetNode = nodeForDownload
	} else {
		log.Println("No suitable node found for VM scheduling at this time.")
		return "", errors.New("no suitable node found for VM scheduling")
	}

	// Generate a unique runner name to avoid collisions
	// Combine node ID, job ID, and timestamp for high uniqueness
	runnerName := fmt.Sprintf("macvmorx-runner-%s-%d-%s", targetNode.NodeID, jobID, time.Now().Format("060102150405"))

	// Instruct the agent to provision the VM
	provisionCmd := models.VMProvisionCommand{
		VMID:                    fmt.Sprintf("vm-%d", jobID), // Use job ID for VM ID
		ImageName:               imageName,
		RunnerRegistrationToken: runnerRegistrationToken, // Use the static token
		RunnerName:              runnerName,
		RunnerLabels:            runnerLabels,
	}

	err := o.InstructAgentToProvisionVM(targetNode.NodeID, provisionCmd)
	if err != nil {
		return "", fmt.Errorf("failed to instruct agent %s to provision VM: %w", targetNode.NodeID, err)
	}

	log.Printf("VM for job %d scheduled on node %s. Runner name: %s", jobID, targetNode.NodeID, runnerName)
	return targetNode.NodeID, nil
}

// InstructAgentToProvisionVM sends a command to the specified agent to provision a VM.
func (o *Orchestrator) InstructAgentToProvisionVM(nodeID string, cmd models.VMProvisionCommand) error {
	// Use HTTP for communication with agents
	agentURL := fmt.Sprintf("http://%s:8081/provision-vm", nodeID) // Changed to HTTP

	jsonPayload, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal VM provision command: %w", err)
	}

	req, err := http.NewRequest("POST", agentURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request to agent: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req) // Use the simple HTTP client
	if err != nil {
		return fmt.Errorf("failed to send provision command to agent %s: %w", nodeID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("agent %s returned non-202 status for provision command: %s, body: %s", nodeID, resp.Status, string(bodyBytes))
	}

	log.Printf("Provision command sent to agent %s for VM %s", nodeID, cmd.VMID)
	return nil
}
