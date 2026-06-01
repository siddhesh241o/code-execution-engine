package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/siddhesh241o/code-execution-engine/internal/api"
	"github.com/siddhesh241o/code-execution-engine/internal/infrastructure/redis"
	"github.com/siddhesh241o/code-execution-engine/internal/infrastructure/submission"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corsOrigin := os.Getenv("CORS_ORIGIN")
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
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
	queue := submission.NewQueueSubmissionDispatcher(redis.NewRedisQueue(rdb))
	store := redis.NewRedisResultStore(rdb, ttl)
	jobInfo := redis.NewRedisJobInfoStore(rdb, ttl)
	executionHandler := api.NewExecutionHandler(queue, store, jobInfo, fetchSecret, callbackSecret)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/execute", executionHandler.HandleExecuteCode)
	mux.HandleFunc("GET /api/result/{id}", executionHandler.HandleGetResult)
	mux.HandleFunc("GET /api/jobs/{id}", executionHandler.HandleGetJob)
	log.Printf("Server started at %s", port)
	http.ListenAndServe(":"+port, enableCORS(mux))
}
