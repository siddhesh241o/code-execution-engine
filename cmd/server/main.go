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
	ttl, err := time.ParseDuration(ttlStr)
	port := os.Getenv("HTTP_PORT")

	rdb, err := redis.NewRedisClient(redisAddr)
	if err != nil {
		log.Fatalf("%v", err)
	}
	queue := submission.NewQueueSubmissionDispatcher(redis.NewRedisQueue(rdb))
	store := redis.NewRedisResultStore(rdb, ttl)
	executionHandler := api.NewExecutionHandler(queue, store)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/execute", executionHandler.HandleExecuteCode)
	mux.HandleFunc("GET /api/result/{id}", executionHandler.HandleGetResult)
	log.Printf("Server started at %s", port)
	http.ListenAndServe(":"+port, enableCORS(mux))
}
