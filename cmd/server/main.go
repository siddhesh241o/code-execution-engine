package main

import (
	"log"
	"net/http"
	"time"

	"github.com/siddhesh241o/code-execution-engine/internal/api"
	"github.com/siddhesh241o/code-execution-engine/internal/infrastructure/redis"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
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
	rdb, err := redis.NewRedisClient("localhost:6379")
	if err != nil {
		log.Fatalf("%v", err)
	}
	queue := redis.NewRedisQueue(rdb)
	store := redis.NewRedisResultStore(rdb, time.Minute*5)
	executionHandler := api.NewExecutionHandler(queue, store)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/execute", executionHandler.HandleExecuteCode)
	log.Println("Server started at 5005")
	http.ListenAndServe(":5005", enableCORS(mux))
}
