package api

import (
	"encoding/json"
	"time"

	"net/http"

	"github.com/google/uuid"
	"github.com/siddhesh241o/code-execution-engine/internal/domain"
)

type ExecutionHandler struct {
	Dispatcher domain.SubmissionDispatcher
	Store      domain.JobStateStore
}

func NewExecutionHandler(d domain.SubmissionDispatcher, s domain.JobStateStore) *ExecutionHandler {
	return &ExecutionHandler{Dispatcher: d, Store: s}
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
	err := h.Dispatcher.Dispatch(r.Context(), executionJob)
	if err != nil {
		http.Error(w, "Submission failed to dispatch: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"job_id": jobID,
		"status": "Queued",
	})
}

func (h *ExecutionHandler) HandleGetResult(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	result, err := h.Store.Get(r.Context(), jobID)
	if err != nil {
		http.Error(w, "result store failed "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if result == nil {
		json.NewEncoder(w).Encode(map[string]string{
			"job_id": jobID,
			"status": "Processing",
		})
		return
	}
	json.NewEncoder(w).Encode(result)
}
