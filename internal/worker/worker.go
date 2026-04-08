package worker

import (
	"fmt"
	"time"

	"github.com/milad176/go-task-processor/internal/job"
)

func StartWorker(id int, jobs <-chan job.Job) {
	fmt.Printf("Worker %d started\n", id)

	for job := range jobs {
		fmt.Printf("Worker %d picked job %s of type %s\n", id, job.ID, job.Type)

		processJob(job)
		fmt.Printf("Worker %d finished job: %s\n", id, job.ID)
	}
}

func processJob(job job.Job) {
	switch job.Type {
	case "print":
		fmt.Println("Message:", job.Payload["message"])

	case "sleep":
		fmt.Println("Sleeping...")
		time.Sleep(2 * time.Second)

	default:
		fmt.Printf("Unknown job type: %s\n", job.Type)
	}
}
