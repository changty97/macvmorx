package heartbeat

import (
	"log"
	"time"

	"github.com/changty97/macvmorx/internal/models"
	"github.com/changty97/macvmorx/internal/store"
)

// Processor handles the incoming heartbeats and updates the node store.
type Processor struct {
	nodeStore *store.NodeStore
}

// NewProcessor creates a new Heartbeat Processor.
func NewProcessor(ns *store.NodeStore) *Processor {
	return &Processor{
		nodeStore: ns,
	}
}

// ProcessHeartbeat receives a heartbeat payload and updates the node store.
func (p *Processor) ProcessHeartbeat(payload models.HeartbeatPayload) {
	log.Printf("Received heartbeat from NodeID: %s, VMs: %d, CPU: %.2f%%, Cached Images: %v",
		payload.NodeID, payload.VMCount, payload.CPUUsagePercent, payload.CachedImages)
	p.nodeStore.UpdateNodeStatus(payload)
}

// StartOfflineMonitor periodically checks for offline nodes and updates their status.
func (p *Processor) StartOfflineMonitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		nodes := p.nodeStore.GetAllNodeStatuses() // This call already updates IsOnline
		for _, node := range nodes {
			if !node.IsOnline {
				log.Printf("Node %s is now considered OFFLINE (last seen %s ago)", node.NodeID, time.Since(node.LastSeen).Round(time.Second))
			}
		}
	}
}
