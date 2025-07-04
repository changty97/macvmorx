package models

import (
	"time"
)

// VMInfo represents details about a single VM running on a Mac Mini.
type VMInfo struct {
	VMID           string `json:"vmId"`           // Unique ID of the VM
	ImageName      string `json:"imageName"`      // Name of the image used for this VM
	RuntimeSeconds int64  `json:"runtimeSeconds"` // How long the VM has been running in seconds
	// Add more VM-specific details as needed (e.g., CPU, Memory allocation)
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

// NodeStatus represents the current state of a Mac Mini node,
// maintained by the orchestrator.
type NodeStatus struct {
	HeartbeatPayload
	LastSeen time.Time `json:"lastSeen"` // Timestamp of the last received heartbeat
	IsOnline bool      `json:"isOnline"` // True if the node is considered online
}

// VMRequest defines the structure for requesting a new VM.
type VMRequest struct {
	ImageName string `json:"imageName"` // The name of the VM image required
	// Add other VM configuration parameters here (e.g., CPU, Memory, Disk size)
}
