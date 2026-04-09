package main

import (
	"fmt"
	"time"

	"github.com/milad176/go-task-processor/internal/job"
	"github.com/milad176/go-task-processor/internal/worker"
)

func main() {
	fmt.Println("Task Processor Starting...")

	// 1. Create a channel (queue)
	jobQueue := make(chan job.Job, 100)

	// 2. Start workers (goroutines)
	for i := 1; i <= 3; i++ {

		go worker.StartWorker(i, jobQueue)
	}

	// 3. Send jobs into the queue
	jobQueue <- job.Job{
		ID:   "1",
		Type: "print",
		Payload: map[string]interface{}{
			"message": "Hello from Job 1",
		},
	}

	jobQueue <- job.Job{
		ID:   "2",
		Type: "sleep",
		Payload: map[string]interface{}{
			"duration": 2,
		},
	}

	jobQueue <- job.Job{
		ID:   "3",
		Type: "email",
		Payload: map[string]interface{}{
			"message": "Hello from Job 3",
		},
	}

	// 4. Keep the program running
	time.Sleep(5 * time.Second)
}
