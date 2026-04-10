package worker

import (
	"fmt"
	"sync"
	"time"

	"github.com/milad176/go-task-processor/internal/job"
)

func StartWorker(id int, jobs <-chan job.Job, jobStore *job.JobStore, wg *sync.WaitGroup) {
	fmt.Printf("Worker %d started\n", id)

	for job := range jobs {
		current := jobStore.Get(job.ID)
		fmt.Printf("Worker %d picked job %s of type %s with status %s\n", id, current.ID, current.Type, current.Status)

		updated := jobStore.UpdateStatus(job.ID, "processing")
		fmt.Printf("Job %s moved to %s\n", job.ID, updated.Status)

		err := processJob(current)

		if err != nil {
			jobStore.UpdateStatus(job.ID, "failed")
			fmt.Printf("Job %s failed: %v\n", job.ID, err)
		} else {
			jobStore.UpdateStatus(job.ID, "done")
		}

		final := jobStore.Get(job.ID)
		fmt.Printf("Worker %d finished job %s with status %s\n", id, final.ID, final.Status)

		wg.Done()
	}
}

func processJob(job job.Job) error {
	switch job.Type {
	case "print":
		fmt.Println("Message:", job.Payload["message"])
		return nil

	case "sleep":
		fmt.Println("Sleeping...")
		time.Sleep(2 * time.Second)
		return nil

	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
}
