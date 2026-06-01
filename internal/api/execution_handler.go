package api

import (
	"encoding/json"
	"strings"
	"time"

	"net/http"

	"github.com/google/uuid"
	"github.com/siddhesh241o/code-execution-engine/internal/domain"
)

type ExecutionHandler struct {
	Dispatcher domain.SubmissionDispatcher
	ResultStore      domain.JobStateStore
	InfoStore 		 domain.JobInfoStore
	FetchSecret		 string 
	CallbackSecret   string
}

func NewExecutionHandler(d domain.SubmissionDispatcher, s domain.JobStateStore, i domain.JobInfoStore, f, c string) *ExecutionHandler {
	return &ExecutionHandler{
		Dispatcher: d,
		ResultStore: s,
		InfoStore: i,
		FetchSecret: f,
		CallbackSecret: c,
	}
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
	err := h.InfoStore.Set(r.Context(), executionJob)
	if err != nil {
		http.Error(w, "Something went wrong: "+ err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.Dispatcher.Dispatch(r.Context(), executionJob)
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
	result, err := h.ResultStore.Get(r.Context(), jobID)
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

func (h *ExecutionHandler) HandleGetJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	secret := helperGetSecret(r.Header.Get("Authorization"))
	if h.FetchSecret != secret {
		helperWriteJSONError(w, http.StatusUnauthorized, "unauthorized action")
		return
	}
	req, err := h.InfoStore.Get(r.Context(), jobID)
	if err != nil {
		helperWriteJSONError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	if req == nil {
		helperWriteJSONError(w, http.StatusNotFound, "job unavailable")
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(req)
}

func (h *ExecutionHandler) HandlePostResult(w http.ResponseWriter, r *http.Request) {
    jobID := r.PathValue("id")

    secret := helperGetSecret(r.Header.Get("Authorization"))
    if h.CallbackSecret != secret {
        helperWriteJSONError(w, http.StatusUnauthorized, "unauthorized")
        return
    }

    var resp domain.ExecutionResponse
    if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
        helperWriteJSONError(w, http.StatusBadRequest, "invalid request")
        return
    }

    resp.ID = jobID

    if err := h.ResultStore.Set(r.Context(), resp); err != nil {
        helperWriteJSONError(w, http.StatusInternalServerError, "failed to save result")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "job_id": jobID,
        "status": "stored",
    })
}

func helperGetSecret(s string) string {
	s = strings.TrimSpace(s)
	return strings.TrimSpace(strings.TrimPrefix(s, "Bearer"))
}

func helperWriteJSONError(w http.ResponseWriter, status int, msg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]string{"error": msg})
}