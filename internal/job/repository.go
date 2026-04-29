package job

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(job Job) error {
	payloadBytes, err := json.Marshal(job.Payload)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO jobs (id, type, payload, status, retries, max_retries, priority, claimed_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		job.ID, job.Type, string(payloadBytes), job.Status, job.Retries, job.MaxRetries, int64(job.Priority), nil,
	)

	return err
}

func (r *Repository) Get(id string) (Job, error) {
	var job Job
	var payloadStr string

	err := r.db.QueryRow(`SELECT id, type, payload, status, retries, max_retries, priority FROM jobs WHERE id = ?`, id).
		Scan(&job.ID, &job.Type, &payloadStr, &job.Status, &job.Retries, &job.MaxRetries, &job.Priority)

	if err != nil {
		return Job{}, err
	}

	err = json.Unmarshal([]byte(payloadStr), &job.Payload)
	if err != nil {
		return Job{}, err
	}

	return job, nil
}

func (r *Repository) UpdateStatus(id string, status string) error {
	_, err := r.db.Exec(`UPDATE jobs SET status = ? WHERE id = ?`, status, id)
	return err
}

func (r *Repository) UpdateRetries(id string, retries int) error {
	_, err := r.db.Exec(`UPDATE jobs SET retries = ? WHERE id = ?`, retries, id)
	return err
}

func (r *Repository) ClaimNextPendingJob() (Job, error) {
	const maxRetries = 5

	for i := 0; i < maxRetries; i++ {

		tx, err := r.db.Begin()
		if err != nil {
			return Job{}, err
		}

		var job Job
		var payload string

		// 1. pick candidate
		err = tx.QueryRow(`
			SELECT id, type, payload, status, retries, max_retries, priority
			FROM jobs
			WHERE status = 'pending'
			ORDER BY priority DESC, id ASC
			LIMIT 1
		`).Scan(
			&job.ID,
			&job.Type,
			&payload,
			&job.Status,
			&job.Retries,
			&job.MaxRetries,
			&job.Priority,
		)

		if err != nil {
			tx.Rollback()

			if err == sql.ErrNoRows {
				return Job{}, sql.ErrNoRows
			}

			return Job{}, err
		}

		// 2. atomic claim
		res, err := tx.Exec(`
			UPDATE jobs
			SET status = 'processing',
			    claimed_at = ?
			WHERE id = ? AND status = 'pending'
		`, time.Now().Unix(), job.ID)

		if err != nil {
			tx.Rollback()
			continue
		}

		rows, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			continue
		}

		// someone else stole it
		if rows == 0 {
			tx.Rollback()
			continue
		}

		// 3. commit
		err = tx.Commit()
		if err != nil {
			continue
		}

		// 4. hydrate payload
		_ = json.Unmarshal([]byte(payload), &job.Payload)
		job.Status = "processing"
		job.ClaimedAt = int64(time.Now().Unix())

		return job, nil
	}

	return Job{}, fmt.Errorf("failed to claim job after retries")
}

func (r *Repository) RecoverStuckJobs() error {
	threshold := time.Now().Add(-60 * time.Second).Unix()

	result, err := r.db.Exec(`
		UPDATE jobs
		SET status = 'pending',
		claimed_at = NULL
		WHERE status = 'processing'
        AND claimed_at IS NOT NULL
        AND claimed_at < ?
	`, threshold)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err == nil && rows > 0 {
		fmt.Printf("Recovered %d stuck job(s)\n", rows)
	}

	return nil
}
