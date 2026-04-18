package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/milad176/go-task-processor/internal/job"
)

func StartWorker(id int, jobs chan job.Job, repo *job.Repository, ctx context.Context) {
	fmt.Printf("Worker %d started\n", id)

	for {
		select {

		// Shutdown signal
		case <-ctx.Done():
			fmt.Printf("Worker %d stopping...\n", id)
			return

		// New job
		case job := <-jobs:

			// Try to claim job
			claimed, err := repo.ClaimJob(job.ID)
			if err != nil {
				fmt.Printf("Worker %d failed to claim job %s: %v\n", id, job.ID, err)
				continue
			}

			if !claimed {
				// another worker already took it
				fmt.Printf("Worker %d skipped job %s (already claimed)\n", id, job.ID)
				continue
			}

			// Now we OWN the job
			fmt.Printf("Worker %d processing job %s\n", id, job.ID)

			// Get latest data (optional but good)
			current, _ := repo.Get(job.ID)

			err = processJob(current)

			if err != nil {
				current.Retries++

				if current.Retries <= current.MaxRetries {

					backoff := getBackoffDuration(current.Retries)

					repo.UpdateStatus(current.ID, "pending")
					repo.UpdateRetries(current.ID, current.Retries)

					fmt.Printf(
						"Job %s failed (attempt %d/%d). Retrying in %v...\n",
						current.ID,
						current.Retries,
						current.MaxRetries,
						backoff,
					)

					time.Sleep(backoff)

					jobs <- current // requeue

				} else {
					repo.UpdateStatus(current.ID, "failed")
					fmt.Printf("Job %s failed permanently after %d attempts\n", current.ID, current.Retries)
				}

			} else {
				repo.UpdateStatus(current.ID, "done")
			}

			finalJob, _ := repo.Get(current.ID)
			fmt.Printf("Worker %d finished job %s with status %s\n", id, finalJob.ID, finalJob.Status)
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
