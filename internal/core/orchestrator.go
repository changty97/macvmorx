package core

import (
	"errors"
	"log"
	"sync"

	"github.com/changty97/macvmorx/internal/models"
	"github.com/changty97/macvmorx/internal/store"
)

// Orchestrator is the central component responsible for managing Mac Mini nodes and VMs.
type Orchestrator struct {
	nodeStore *store.NodeStore
	// Add other components like a scheduler, VM manager, etc.
	mu sync.Mutex // Mutex for protecting orchestrator state if needed
}

// NewOrchestrator creates a new Orchestrator instance.
func NewOrchestrator(ns *store.NodeStore) *Orchestrator {
	return &Orchestrator{
		nodeStore: ns,
	}
}

// ScheduleVM attempts to schedule a VM based on the request.
// It prioritizes nodes with the image cached, then nodes that can download it.
func (o *Orchestrator) ScheduleVM(vmRequest models.VMRequest) (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	log.Printf("Attempting to schedule VM request for image: %s", vmRequest.ImageName)

	var (
		nodeWithImageAvailable *models.NodeStatus = nil
		nodeForDownload        *models.NodeStatus = nil
	)

	nodes := o.nodeStore.GetAllNodeStatuses()

	// First pass: Look for nodes that are online, have less than 2 VMs, and already have the image cached.
	for _, node := range nodes {
		if node.IsOnline && node.VMCount < 2 {
			hasImage := false
			for _, cachedImage := range node.CachedImages {
				if cachedImage == vmRequest.ImageName {
					hasImage = true
					break
				}
			}
			if hasImage {
				nodeWithImageAvailable = node
				log.Printf("Found suitable node %s with image %s cached.", node.NodeID, vmRequest.ImageName)
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
				log.Printf("Found available node %s that can download image %s.", node.NodeID, vmRequest.ImageName)
				break // Found a node that can download, consider this
			}
		}
	}

	if nodeWithImageAvailable != nil {
		// Found a node with the image ready
		// In a real implementation, you'd send a command to this node
		// to provision the VM, and update its VMCount.
		log.Printf("Scheduling VM on node %s (image already cached).", nodeWithImageAvailable.NodeID)
		// TODO: Trigger VM provisioning on nodeWithImageAvailable.NodeID
		return nodeWithImageAvailable.NodeID, nil
	} else if nodeForDownload != nil {
		// Found a node that can download the image
		log.Printf("Scheduling VM on node %s (will download image %s).", nodeForDownload.NodeID, vmRequest.ImageName)
		// TODO: Send command to nodeForDownload.NodeID to download vmRequest.ImageName in background
		// TODO: Then, trigger VM provisioning once download is confirmed.
		return nodeForDownload.NodeID, nil
	}

	log.Println("No suitable node found for VM scheduling at this time.")
	return "", errors.New("no suitable node found for VM scheduling")
}

// InstructNodeToDownloadImage is a placeholder for instructing a Mac Mini agent
// to download a specific VM image from GCP.
// This would typically involve sending a command to the Mac Mini agent.
func (o *Orchestrator) InstructNodeToDownloadImage(nodeID string, imageName string) error {
	log.Printf("Instructing node %s to download image: %s from GCP.", nodeID, imageName)
	// TODO: Implement actual communication with the Mac Mini agent to initiate image download.
	// This might involve gRPC, SSH, or another secure channel.
	// The agent would then handle the LRU cache logic locally.
	return nil
}

// ProvisionVM instructs a specific node to provision a VM.
// This would typically involve sending a command to the Mac Mini agent.
func (o *Orchestrator) ProvisionVM(nodeID string, vmConfig interface{}) error {
	log.Printf("Instructing node %s to provision VM with config: %+v", nodeID, vmConfig)
	// TODO: Implement actual communication with the Mac Mini agent to start a VM.
	// This might involve gRPC, SSH, or another secure channel.
	return nil
}
