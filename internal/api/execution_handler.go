package api

import (
	"encoding/json"
	"time"

	"net/http"

	"github.com/google/uuid"
	"github.com/siddhesh241o/code-execution-engine/internal/domain"
)

type ExecutionHandler struct {
	Queue domain.QueueProvider
	Store domain.ResultStore
}

func NewExecutionHandler(q domain.QueueProvider, s domain.ResultStore) *ExecutionHandler {
	return &ExecutionHandler{Queue: q, Store: s}
}

func (h *ExecutionHandler) HandleExecuteCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code     string `json:"code"`
		Language string `json:"language"`
		Input    string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	jobID := uuid.New().String()
	executionJob := domain.ExecutionRequest{
		ID:        jobID,
		Language:  req.Language,
		Code:      req.Code,
		Input:     req.Input,
		CreatedAt: time.Now(),
	}

	err := h.Queue.Push(r.Context(), executionJob)
	if err != nil {
		http.Error(w, "failed to queue jo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"job_id": jobID,
		"status": "Queued",
	})
}
