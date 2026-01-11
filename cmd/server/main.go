package main

import (
	"log"
	"net/http"

	"github.com/moby/moby/client"
	"github.com/siddhesh241o/code-execution-engine/internal/api"
	"github.com/siddhesh241o/code-execution-engine/internal/runner"
)

func main() {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		log.Fatalf("Docker init failed: %v", err)
	}
	fm := runner.NewFileManager()
	executor := runner.NewDockerExecutor(cli, fm)
	executionHandler := api.NewExecutionHandler(executor)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/execute", executionHandler.HandleExecuteCode)
	log.Println("Server started at 5005")
	http.ListenAndServe(":5005", mux)
}
