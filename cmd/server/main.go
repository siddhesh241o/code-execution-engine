package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	redisv9 "github.com/redis/go-redis/v9"
	"github.com/siddhesh241o/code-execution-engine/internal/api"
	"github.com/siddhesh241o/code-execution-engine/internal/domain"
	"github.com/siddhesh241o/code-execution-engine/internal/infrastructure/memory"
	"github.com/siddhesh241o/code-execution-engine/internal/infrastructure/redis"
	"github.com/siddhesh241o/code-execution-engine/internal/infrastructure/submission"
	"github.com/siddhesh241o/code-execution-engine/internal/observability"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corsOrigin := os.Getenv("CORS_ORIGIN")
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-Frontend-Secret, ngrok-skip-browser-warning")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func main() {
	_ = godotenv.Load()
	redisAddr := os.Getenv("REDIS_ADDR")
	ttlStr := os.Getenv("RESULT_TTL")
	fetchSecret := os.Getenv("FETCH_SECRET")
	callbackSecret := os.Getenv("CALLBACK_SECRET")
	frontendSecret := os.Getenv("FRONTEND_SHARED_SECRET")
	corsOrigin := os.Getenv("CORS_ORIGIN")
	storeMode := os.Getenv("STORE_MODE")
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		log.Fatalf("invalid RESULT_TTL: %v", err)
	}
	port := os.Getenv("HTTP_PORT")

	var resultStore domain.JobStateStore
	var jobInfoStore domain.JobInfoStore
	var rdb *redisv9.Client
	var memProvider *memory.InMemoryProvider

	if storeMode == "redis" {
		rdb, err = redis.NewRedisClient(redisAddr)
		if err != nil {
			slog.Error("redis connection failed", "error", err)
			os.Exit(1)
		}
		resultStore = redis.NewRedisResultStore(rdb, ttl)
		jobInfoStore = redis.NewRedisJobInfoStore(rdb, ttl)
	} else {
		slog.Info("Starting in MEMORY storage mode")
		memProvider = memory.NewInMemoryProvider(ttl)
		resultStore = &memory.MemoryJobStateStore{P: memProvider}
		jobInfoStore = &memory.MemoryJobInfoStore{P: memProvider}
	}

	mode := os.Getenv("EXECUTION_MODE")
	dispatcher, err := initDispatcher(mode, rdb, memProvider)
	if err != nil {
		slog.Error("failed to init dispatcher", "error", err)
		os.Exit(1)
	}

	executionHandler := api.NewExecutionHandler(dispatcher, resultStore, jobInfoStore, fetchSecret, callbackSecret)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/execute", executionHandler.HandleExecuteCode)
	mux.HandleFunc("GET /api/result/{id}", executionHandler.HandleGetResult)
	mux.HandleFunc("GET /api/jobs/{id}", executionHandler.HandleGetJob)
	mux.HandleFunc("POST /api/jobs/{id}/result", executionHandler.HandlePostResult)
	mux.Handle("GET /metrics", observability.Handler())
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      enableCORS(api.RateLimit(api.RequireFrontendAuth(mux, frontendSecret, corsOrigin))),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("Server started", "mode", mode)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func initDispatcher(mode string, rdb *redisv9.Client, mem *memory.InMemoryProvider) (domain.SubmissionDispatcher, error) {
	switch mode {
	case "gha":
		cfg := submission.GithubActionsConfig{
			Owner:          os.Getenv("GITHUB_OWNER"),
			Repo:           os.Getenv("GITHUB_REPO"),
			Workflow:       os.Getenv("GITHUB_WORKFLOW"),
			Ref:            os.Getenv("GITHUB_REF"),
			Token:          os.Getenv("GITHUB_TOKEN"),
			BaseURL:        "https://api.github.com",
			FetchURL:       os.Getenv("FETCH_URL"),
			CallbackURL:    os.Getenv("CALLBACK_URL"),
			FetchSecret:    os.Getenv("FETCH_SECRET"),
			CallbackSecret: os.Getenv("CALLBACK_SECRET"),
		}
		return submission.NewGitHubActionsSubmissionDispatcher(cfg, nil)
	case "local":
		if rdb != nil {
			return submission.NewQueueSubmissionDispatcher(redis.NewRedisQueue(rdb)), nil
		}
		if mem != nil {
			return submission.NewQueueSubmissionDispatcher(&memory.MemoryJobQueue{P: mem}), nil
		}
		return nil, fmt.Errorf("local mode requires either redis or memory provider")
	default:
		return nil, fmt.Errorf("invalid execution mode: %s", mode)
	}
}
