package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/changty97/macvmorx/internal/core"
	"github.com/changty97/macvmorx/internal/heartbeat"
	"github.com/changty97/macvmorx/internal/models"
	"github.com/changty97/macvmorx/internal/store"
)

// Handlers struct holds dependencies for API handlers.
type Handlers struct {
	HeartbeatProcessor *heartbeat.Processor
	NodeStore          *store.NodeStore
	Orchestrator       *core.Orchestrator
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(hp *heartbeat.Processor, ns *store.NodeStore, orch *core.Orchestrator) *Handlers {
	return &Handlers{
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

// HandleScheduleVM is an API endpoint for scheduling a VM.
func (h *Handlers) HandleScheduleVM(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var vmRequest models.VMRequest
	if err := json.NewDecoder(r.Body).Decode(&vmRequest); err != nil {
		log.Printf("Error decoding VM request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	nodeID, err := h.Orchestrator.ScheduleVM(vmRequest)
	if err != nil {
		log.Printf("Failed to schedule VM: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable) // Using 503 for no suitable node
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "VM scheduled successfully", "nodeId": nodeID})
}
