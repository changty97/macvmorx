package heartbeat

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/changty97/macvmorx/internal/models"
	"github.com/changty97/macvmorx/internal/store"
)

// Processor handles the incoming heartbeats and updates the node store.
type Processor struct {
	nodeStore *store.NodeStore
	jobStore  *store.JobStore // Added JobStore
}

// NewProcessor creates a new Heartbeat Processor.
func NewProcessor(ns *store.NodeStore, js *store.JobStore) *Processor {
	return &Processor{
		nodeStore: ns,
		jobStore:  js,
	}
}

// ProcessHeartbeat receives a heartbeat payload and updates the node store.
func (p *Processor) ProcessHeartbeat(payload models.HeartbeatPayload) {
	log.Printf("Received heartbeat from NodeID: %s, VMs: %d, CPU: %.2f%%, Cached Images: %v",
		payload.NodeID, payload.VMCount, payload.CPUUsagePercent, payload.CachedImages)
	p.nodeStore.UpdateNodeStatus(payload)

	// Update job statuses based on VM info in heartbeat
	for _, vmInfo := range payload.VMs {
		// Assuming VMID is the JobID for simplicity, or you have a mapping
		// Convert VMID string to int64 if VMID is "vm-JOBID"
		var parsedJobID int64
		if strings.HasPrefix(vmInfo.VMID, "vm-") {
			var err error
			parsedJobID, err = strconv.ParseInt(strings.TrimPrefix(vmInfo.VMID, "vm-"), 10, 64)
			if err != nil {
				log.Printf("Error parsing JobID from VMID '%s': %v", vmInfo.VMID, err)
				continue
			}
		} else {
			// If VMID is not in "vm-JOBID" format, assume it's already the JobID
			var err error
			parsedJobID, err = strconv.ParseInt(vmInfo.VMID, 10, 64)
			if err != nil {
				log.Printf("Error parsing JobID from VMID '%s': %v", vmInfo.VMID, err)
				continue
			}
		}

		p.jobStore.UpdateJobStatus(parsedJobID, func(job *models.JobStatus) {
			job.Status = "running"
			job.NodeID = payload.NodeID
			job.VMID = vmInfo.VMID
			job.VMIPAddress = vmInfo.VMIPAddress
			if job.VMStartTime == nil {
				now := time.Now()
				job.VMStartTime = &now
			}
			// Update runtimeSeconds in VMInfo if needed (though orchestrator's view is derived)
			// vmInfo.RuntimeSeconds = job.RuntimeSeconds // This line seems incorrect, as vmInfo is input, not output
		})
	}
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
				// Optionally, update jobs running on this node to a "node offline" status
				// This would require iterating through jobs and checking their NodeID
			}
		}
	}
}
