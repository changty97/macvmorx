package models

import (
	"time"
)

// VMInfo represents details about a single VM running on a Mac Mini.
type VMInfo struct {
	VMID           string `json:"vmId"`           // Unique ID of the VM
	ImageName      string `json:"imageName"`      // Name of the image used for this VM
	RuntimeSeconds int64  `json:"runtimeSeconds"` // How long the VM has been running in seconds
	VMHostname     string `json:"vmHostname"`     // Hostname of the VM
	VMIPAddress    string `json:"vmIpAddress"`    // IP address of the VM
}

// HeartbeatPayload represents the data sent by a Mac Mini in its heartbeat.
type HeartbeatPayload struct {
	NodeID          string   `json:"nodeId"`          // Unique identifier for the Mac Mini
	VMCount         int      `json:"vmCount"`         // Number of VMs currently running (0, 1, or 2)
	VMs             []VMInfo `json:"vms"`             // Details of running VMs
	CPUUsagePercent float64  `json:"cpuUsagePercent"` // Current CPU usage percentage
	MemoryUsageGB   float64  `json:"memoryUsageGB"`   // Current memory usage in GB
	TotalMemoryGB   float64  `json:"totalMemoryGB"`   // Total memory in GB
	DiskUsageGB     float64  `json:"diskUsageGB"`     // Current disk usage in GB
	TotalDiskGB     float64  `json:"totalDiskGB"`     // Total disk space in GB
	Status          string   `json:"status"`          // General status (e.g., "healthy", "warning", "offline")
	CachedImages    []string `json:"cachedImages"`    // List of VM image names cached on this Mac Mini
}

// NodeStatus represents the full status of a Mac Mini node, derived from heartbeats.
type NodeStatus struct {
	HeartbeatPayload           // Embeds all fields from the latest heartbeat
	LastSeen         time.Time // Timestamp of the last received heartbeat
	IsOnline         bool      // True if within offline timeout, false otherwise
}

// VMProvisionCommand represents a command from the orchestrator to provision a VM.
type VMProvisionCommand struct {
	VMID                    string   `json:"vmId"`                    // Unique ID for the new VM
	ImageName               string   `json:"imageName"`               // Image to use for the VM
	RunnerRegistrationToken string   `json:"runnerRegistrationToken"` // GitHub Actions runner registration token
	RunnerName              string   `json:"runnerName"`              // Unique name for the GitHub runner
	RunnerLabels            []string `json:"runnerLabels"`            // Labels for the GitHub runner
	// Add other VM configuration details
}

// VMDeleteCommand represents a command from the orchestrator to delete a VM.
type VMDeleteCommand struct {
	VMID string `json:"vmId"` // ID of the VM to delete
}

// JobStatus represents the status of a GitHub workflow job being managed by macvmorx.
type JobStatus struct {
	JobID                 int64      `json:"jobId"`                           // GitHub Workflow Job ID
	RunnerName            string     `json:"runnerName"`                      // Name assigned to the GitHub Actions runner
	ImageName             string     `json:"imageName"`                       // VM image requested for this job
	Status                string     `json:"status"`                          // Current status (e.g., "queued", "provisioning", "running", "completed", "failed")
	NodeID                string     `json:"nodeId,omitempty"`                // Mac Mini Node ID where the VM is/was running
	VMID                  string     `json:"vmId,omitempty"`                  // VM ID on the Mac Mini
	VMIPAddress           string     `json:"vmIpAddress,omitempty"`           // IP Address of the VM
	Labels                []string   `json:"labels"`                          // Labels from the GitHub workflow job
	QueueTime             time.Time  `json:"queueTime"`                       // When the job was queued on GitHub
	ProvisioningStartTime *time.Time `json:"provisioningStartTime,omitempty"` // When VM provisioning started
	VMStartTime           *time.Time `json:"vmStartTime,omitempty"`           // When the VM actually started running
	EndTime               *time.Time `json:"endTime,omitempty"`               // When the job completed/failed
}
