package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/milad176/go-task-processor/internal/api"
	"github.com/milad176/go-task-processor/internal/db"
	"github.com/milad176/go-task-processor/internal/job"
	"github.com/milad176/go-task-processor/internal/worker"
)

func main() {
	fmt.Println("Task Processor Starting...")

	jobQueue := make(chan job.Job, 100) // Create a channel (queue)
	dbConn := db.InitDB()
	repo := job.NewRepository(dbConn)

	ctx, cancel := context.WithCancel(context.Background()) // Create a context for graceful shutdown
	defer cancel()

	// Start workers (goroutines)
	for i := 1; i <= 3; i++ {

		go worker.StartWorker(i, jobQueue, repo, ctx)
	}

	server := api.NewServer(repo, jobQueue) // Start API server

	// Start HTTP server in goroutine
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Server error:", err)
		}
	}()

	// Listen for Ctrl+C BEFORE blocking
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block here (VERY IMPORTANT)
	<-sigChan

	fmt.Println("\nShutting down gracefully...")

	// Stop workers
	cancel()

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Println("Server shutdown error:", err)
	}

	fmt.Println("Shutdown complete")
}
