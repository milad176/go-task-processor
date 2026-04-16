package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/milad176/go-task-processor/internal/api"
	"github.com/milad176/go-task-processor/internal/job"
	"github.com/milad176/go-task-processor/internal/worker"
)

func main() {
	fmt.Println("Task Processor Starting...")

	jobQueue := make(chan job.Job, 100) // Create a channel (queue)
	jobStore := job.NewJobStore()       // Create a job store

	ctx, cancel := context.WithCancel(context.Background()) // Create a context for graceful shutdown
	defer cancel()

	// Start workers (goroutines)
	for i := 1; i <= 3; i++ {

		go worker.StartWorker(i, jobQueue, jobStore, ctx)
	}

	server := api.NewServer(jobStore, jobQueue) // Start API server

	// ✅ Run server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			fmt.Println("Server error:", err)
		}
	}()

	// ✅ Setup signal handling BEFORE blocking
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// ✅ Block here (VERY IMPORTANT)
	<-sigChan

	fmt.Println("\nShutting down gracefully...")

	// Stop workers
	cancel()

	fmt.Println("Shutdown complete")
}
