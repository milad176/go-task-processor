package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/milad176/go-task-processor/internal/job"
	"github.com/milad176/go-task-processor/internal/worker"
)

func main() {
	fmt.Println("Task Processor Starting...")

	// Create a channel (queue)
	jobQueue := make(chan job.Job, 100)

	var wg sync.WaitGroup
	jobStore := job.NewJobStore()

	// Start workers (goroutines)
	for i := 1; i <= 3; i++ {

		go worker.StartWorker(i, jobQueue, jobStore, &wg)
	}

	// Add jobs
	jobList := []job.Job{
		{
			ID:   "1",
			Type: "print",
			Payload: map[string]interface{}{
				"message": "Hello from Job 1",
			},
		},
		{
			ID:   "2",
			Type: "sleep",
			Payload: map[string]interface{}{
				"duration": 2,
			},
		},
		{
			ID:   "3",
			Type: "print",
			Payload: map[string]interface{}{
				"message": "Hello from Job 3",
			},
		},
	}

	wg.Add(len(jobList))

	// Send jobs into the queue
	go func() {
		for _, job := range jobList {
			jobStore.Add(job)
			jobQueue <- job
		}
		close(jobQueue) // Close the channel after adding all jobs
	}()

	wg.Wait() // Wait for all jobs to be processed

	for _, job := range jobList {
		result := jobStore.Get(job.ID)
		fmt.Println("Job", result.ID, "status:", result.Status)
	}

	fmt.Println("All jobs processed. Shutting down.")
	time.Sleep(1 * time.Second) // Give workers time to finish
}
