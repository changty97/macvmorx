package store

import (
	"sync"
	"time"

	"github.com/changty97/macvmorx/internal/models"
)

// NodeStore manages the state of all registered Mac Mini nodes.
type NodeStore struct {
	nodes          sync.Map      // Stores map[string]*models.NodeStatus, using sync.Map for concurrent access
	offlineTimeout time.Duration // Duration after which a node is considered offline if no heartbeat is received
}

// NewNodeStore creates and initializes a new NodeStore.
func NewNodeStore(offlineTimeout time.Duration) *NodeStore {
	return &NodeStore{
		offlineTimeout: offlineTimeout,
	}
}

// UpdateNodeStatus updates or adds a node's status based on a heartbeat.
func (s *NodeStore) UpdateNodeStatus(payload models.HeartbeatPayload) {
	nodeStatus := models.NodeStatus{
		HeartbeatPayload: payload,
		LastSeen:         time.Now(),
		IsOnline:         true, // Mark as online upon receiving a heartbeat
	}
	s.nodes.Store(payload.NodeID, &nodeStatus)
}

// GetNodeStatus retrieves the status of a specific node.
func (s *NodeStore) GetNodeStatus(nodeID string) (*models.NodeStatus, bool) {
	if val, ok := s.nodes.Load(nodeID); ok {
		return val.(*models.NodeStatus), true
	}
	return nil, false
}

// GetAllNodeStatuses retrieves the statuses of all nodes.
func (s *NodeStore) GetAllNodeStatuses() []*models.NodeStatus {
	var allNodes []*models.NodeStatus
	s.nodes.Range(func(key, value interface{}) bool {
		node := value.(*models.NodeStatus)
		// Check if the node should be marked as offline
		if time.Since(node.LastSeen) > s.offlineTimeout {
			node.IsOnline = false
		} else {
			node.IsOnline = true
		}
		allNodes = append(allNodes, node)
		return true
	})
	return allNodes
}
