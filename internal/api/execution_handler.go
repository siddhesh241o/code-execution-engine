package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/siddhesh241o/code-execution-engine/internal/domain"
	"github.com/siddhesh241o/code-execution-engine/internal/runner"
)

type ExecutionHandler struct {
	executor *runner.DockerExecutor
}

func NewExecutionHandler(e *runner.DockerExecutor) *ExecutionHandler {
	return &ExecutionHandler{executor: e}
}

func (h *ExecutionHandler) HandleExecuteCode(w http.ResponseWriter, r *http.Request) {
	var req domain.ExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	result, err := h.executor.Execute(r.Context(), req)
	if err != nil {
		http.Error(w, "Execution failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		log.Printf("Error streaming response to client %v", err)
	}
}
