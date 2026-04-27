package job

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
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
		`INSERT INTO jobs (id, type, payload, status, retries, max_retries, priority) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		job.ID, job.Type, string(payloadBytes), job.Status, job.Retries, job.MaxRetries, int64(job.Priority),
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

func (r *Repository) GetNextPendingJob() (Job, error) {
	var job Job
	var payload string

	err := r.db.QueryRow(`
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
		return job, err
	}

	json.Unmarshal([]byte(payload), &job.Payload)

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

func (r *Repository) ClaimJob(id string) (bool, error) {

	const maxRetries = 3

	for i := 0; i < maxRetries; i++ {

		result, err := r.db.Exec(`UPDATE jobs SET status = 'processing' WHERE id = ? AND status = 'pending'`, id)

		if err != nil {

			// Retry if DB is locked
			if strings.Contains(err.Error(), "database is locked") {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Real error
			return false, err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return false, err
		}

		return rowsAffected == 1, nil
	}

	// If all retries failed
	return false, fmt.Errorf("failed to claim job after retries (db locked)")
}
