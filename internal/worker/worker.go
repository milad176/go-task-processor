package worker

import (
	"fmt"
	"time"

	"github.com/milad176/go-task-processor/internal/job"
)

func StartWorker(id int, jobs chan job.Job, jobStore *job.JobStore) {
	fmt.Printf("Worker %d started\n", id)

	for job := range jobs {
		current := jobStore.Get(job.ID)
		fmt.Printf("Worker %d picked job %s of type %s with status %s\n", id, current.ID, current.Type, current.Status)

		updated := jobStore.UpdateStatus(job.ID, "processing")
		fmt.Printf("Job %s moved to %s\n", job.ID, updated.Status)

		err := processJob(current)

		if err != nil {
			job.Retries++
			if job.Retries <= job.MaxRetries {
				jobStore.UpdateStatus(job.ID, "pending")
				fmt.Printf("Job %s failed (attempt %d/%d). Retrying...\n", job.ID, job.Retries, job.MaxRetries)

				jobStore.Update(job) // update retries

				time.Sleep(1 * time.Second)

				jobs <- job // requeue job

			} else {
				jobStore.UpdateStatus(job.ID, "failed")
				fmt.Printf("Job %s failed permanently after %d attempts\n", job.ID, job.Retries)
			}

		} else {
			jobStore.UpdateStatus(job.ID, "done")
		}

		final := jobStore.Get(job.ID)
		fmt.Printf("Worker %d finished job %s with status %s\n", id, final.ID, final.Status)
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
