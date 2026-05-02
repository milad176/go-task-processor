package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/milad176/go-task-processor/internal/api"
	"github.com/milad176/go-task-processor/internal/db"
	"github.com/milad176/go-task-processor/internal/job"
	"github.com/milad176/go-task-processor/internal/worker"
)

func main() {
	fmt.Println("Task Processor Starting...")

	dbConn := db.InitDB()
	repo := job.NewRepository(dbConn)

	ctx, cancel := context.WithCancel(context.Background()) // Create a context for graceful shutdown
	defer cancel()

	var wg sync.WaitGroup // WaitGroup to wait for all workers to finish

	// Start workers (goroutines)
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go worker.StartWorker(i, repo, ctx, &wg)
		time.Sleep(150 * time.Millisecond)
	}

	go func() {
		time.Sleep(10 * time.Second) // allow app to stabilize first

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Recovery service stopping...")
				return
			default:
				err := repo.RecoverStuckJobs()
				if err != nil {
					fmt.Println("Recovery service error:", err)
				}
				time.Sleep(10 * time.Second)
			}
		}
	}()

	server := api.NewServer(repo) // Start API server

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
	wg.Wait() // Wait for all workers to finish

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Println("Server shutdown error:", err)
	}

	fmt.Println("Shutdown complete")
}
