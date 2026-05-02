package worker

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/milad176/go-task-processor/internal/job"
	"github.com/milad176/go-task-processor/internal/metrics"
)

func StartWorker(id int, repo *job.Repository, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("Worker %d started\n", id)

	for {
		// Graceful shutdown:
		// only stop BEFORE taking a new job,
		// never interrupt a job already being processed.
		if ctx.Err() != nil {
			fmt.Printf("Worker %d stopping...\n", id)
			return
		}

		job, err := repo.ClaimNextPendingJob() // Atomically fetch + claim next pending job
		if err != nil {

			if err == sql.ErrNoRows {
				time.Sleep(500 * time.Millisecond) // no jobs available
				continue
			}

			fmt.Printf("Worker %d DB error: %v\n", id, err)
			time.Sleep(1 * time.Second)
			continue
		}

		fmt.Printf("CLAIMED job %s at %d\n", job.ID, job.ClaimedAt)

		// Now we OWN the job
		fmt.Printf("Worker %d processing job %s [%s] with priority %d\n", id, job.ID, job.Type, job.Priority)

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
				metrics.IncrementFailed()
				fmt.Printf("Job %s failed permanently after %d attempts\n", job.ID, job.Retries)
			}

		} else {
			repo.UpdateStatus(job.ID, "done")
			metrics.IncrementProcessed()
		}

		finalJob, _ := repo.Get(job.ID)
		fmt.Printf("Worker %d finished job %s [%s] with status %s\n", id, finalJob.ID, finalJob.Type, finalJob.Status)

		fmt.Printf("METRICS DEBUG => processed=%d failed=%d recovered=%d\n",
			metrics.Snapshot().Processed,
			metrics.Snapshot().Failed,
			metrics.Snapshot().Recovered,
		)

		time.Sleep(1 * time.Second) // prevent busy loop if DB is very fast
	}
}

func processJob(job job.Job) error {
	switch job.Type {

	case "print":
		fmt.Println("Message:", job.Payload["message"])
		return nil

	case "payment":
		fmt.Printf("Processing payment for order %v...\n", job.Payload["orderId"])
		time.Sleep(7 * time.Second)
		fmt.Println("Payment completed")
		return nil

	case "report":
		fmt.Printf("Generating report %v...\n", job.Payload["reportName"])
		time.Sleep(5 * time.Second)
		fmt.Println("Report generated")
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
