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
			fmt.Printf("Worker %d picked job %s of type %s with status %s\n", id, job.ID, job.Type, job.Status)

			repo.UpdateStatus(job.ID, "processing")
			fmt.Printf("Job %s moved to %s\n", job.ID, "processing")

			err := processJob(job)

			if err != nil {
				job.Retries++

				if job.Retries <= job.MaxRetries {

					backoff := getBackoffDuration(job.Retries)

					repo.UpdateStatus(job.ID, "pending")

					fmt.Printf(
						"Job %s failed (attempt %d/%d). Retrying in %v...\n", job.ID, job.Retries, job.MaxRetries, backoff,
					)

					time.Sleep(backoff)

					jobs <- job // requeue

				} else {
					repo.UpdateStatus(job.ID, "failed")
					fmt.Printf("Job %s failed permanently after %d attempts\n", job.ID, job.Retries)
				}

			} else {
				repo.UpdateStatus(job.ID, "done")
			}

			finalJob, _ := repo.Get(job.ID)
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
