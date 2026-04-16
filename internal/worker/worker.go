package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/milad176/go-task-processor/internal/job"
)

func StartWorker(id int, jobs chan job.Job, jobStore *job.JobStore, ctx context.Context) {
	fmt.Printf("Worker %d started\n", id)

	for {
		select {

		// Shutdown signal
		case <-ctx.Done():
			fmt.Printf("Worker %d stopping...\n", id)
			return

		// New job
		case job := <-jobs:
			current := jobStore.Get(job.ID)

			fmt.Printf("Worker %d picked job %s of type %s with status %s\n",
				id, current.ID, current.Type, current.Status)

			updated := jobStore.UpdateStatus(job.ID, "processing")
			fmt.Printf("Job %s moved to %s\n", job.ID, updated.Status)

			err := processJob(current)

			if err != nil {
				job.Retries++

				if job.Retries <= job.MaxRetries {

					backoff := getBackoffDuration(job.Retries)

					jobStore.UpdateStatus(job.ID, "pending")

					fmt.Printf(
						"Job %s failed (attempt %d/%d). Retrying in %v...\n",
						job.ID,
						job.Retries,
						job.MaxRetries,
						backoff,
					)

					jobStore.Update(job)

					time.Sleep(backoff)

					jobs <- job // requeue

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

func getBackoffDuration(retries int) time.Duration {
	max := 10 * time.Second
	backoff := time.Duration(1<<retries) * time.Second

	if backoff > max {
		return max
	}
	return backoff
}
