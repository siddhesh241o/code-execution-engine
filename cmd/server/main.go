package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	redisv9 "github.com/redis/go-redis/v9"
	"github.com/siddhesh241o/code-execution-engine/internal/api"
	"github.com/siddhesh241o/code-execution-engine/internal/domain"
	"github.com/siddhesh241o/code-execution-engine/internal/infrastructure/redis"
	"github.com/siddhesh241o/code-execution-engine/internal/infrastructure/submission"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corsOrigin := os.Getenv("CORS_ORIGIN")
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ngrok-skip-browser-warning")
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
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		log.Fatalf("invalid RESULT_TTL: %v", err)
	}
	port := os.Getenv("HTTP_PORT")

	rdb, err := redis.NewRedisClient(redisAddr)
	if err != nil {
		log.Fatalf("%v", err)
	}

	mode := os.Getenv("EXECUTION_MODE")
	dispatcher, err := initDispatcher(mode, rdb)
	if err != nil {
		log.Fatalf("failed to init dispatcher: %v", err)
	}

	store := redis.NewRedisResultStore(rdb, ttl)
	jobInfo := redis.NewRedisJobInfoStore(rdb, ttl)
	executionHandler := api.NewExecutionHandler(dispatcher, store, jobInfo, fetchSecret, callbackSecret)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/execute", executionHandler.HandleExecuteCode)
	mux.HandleFunc("GET /api/result/{id}", executionHandler.HandleGetResult)
	mux.HandleFunc("GET /api/jobs/{id}", executionHandler.HandleGetJob)
	mux.HandleFunc("POST /api/jobs/{id}/result", executionHandler.HandlePostResult)
	log.Printf("Server started at %s", port)
	http.ListenAndServe(":"+port, enableCORS(mux))
}

func initDispatcher(mode string, rdb *redisv9.Client) (domain.SubmissionDispatcher, error) {
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
		return submission.NewQueueSubmissionDispatcher(redis.NewRedisQueue(rdb)), nil
	default:
		return nil, fmt.Errorf("invalid execution mode: %s", mode)
	}
}
