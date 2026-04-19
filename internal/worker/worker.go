package worker

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/milad176/go-task-processor/internal/job"
)

func StartWorker(id int, repo *job.Repository, ctx context.Context) {
	fmt.Printf("Worker %d started\n", id)

	for {
		select {

		// Shutdown signal
		case <-ctx.Done():
			fmt.Printf("Worker %d stopping...\n", id)
			return

		default:
			// Try to get a pending job
			job, err := repo.GetNextPendingJob()
			if err != nil {

				if err == sql.ErrNoRows {
					// No jobs → normal situation
					time.Sleep(500 * time.Millisecond)
					continue
				}

				// Real error → log it
				fmt.Printf("Worker %d DB error: %v\n", id, err)
				time.Sleep(1 * time.Second) // optional backoff
				continue
			}

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

			err = processJob(job)

			if err != nil {
				job.Retries++

				if job.Retries <= job.MaxRetries {

					backoff := getBackoffDuration(job.Retries)
					repo.UpdateStatus(job.ID, "pending")
					repo.UpdateRetries(job.ID, job.Retries)

					fmt.Printf(
						"Job %s failed (attempt %d/%d). Retrying in %v...\n",
						job.ID, job.Retries, job.MaxRetries, backoff,
					)

					time.Sleep(backoff)
				} else {
					repo.UpdateStatus(job.ID, "failed")
					fmt.Printf("Job %s failed permanently after %d attempts\n", job.ID, job.Retries)
				}

			} else {
				repo.UpdateStatus(job.ID, "done")
			}

			finalJob, _ := repo.Get(job.ID)
			fmt.Printf("Worker %d finished job %s with status %s\n", id, finalJob.ID, finalJob.Status)

			time.Sleep(20 * time.Millisecond) // prevent busy loop if DB is very fast
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
