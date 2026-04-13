package main

import (
	"fmt"

	"github.com/milad176/go-task-processor/internal/api"
	"github.com/milad176/go-task-processor/internal/job"
	"github.com/milad176/go-task-processor/internal/worker"
)

func main() {
	fmt.Println("Task Processor Starting...")

	jobQueue := make(chan job.Job, 100) // Create a channel (queue)

	jobStore := job.NewJobStore() // Create a job store

	// Start workers (goroutines)
	for i := 1; i <= 3; i++ {

		go worker.StartWorker(i, jobQueue, jobStore)
	}

	server := api.NewServer(jobStore, jobQueue) // Start API server
	server.Start()
}
