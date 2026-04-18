package job

import (
	"database/sql"
	"encoding/json"
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
		`INSERT INTO jobs (id, type, payload, status, retries, max_retries) VALUES (?, ?, ?, ?, ?, ?)`,
		job.ID, job.Type, string(payloadBytes), job.Status, job.Retries, job.MaxRetries,
	)

	return err
}

func (r *Repository) Get(id string) (Job, error) {
	var job Job
	var payloadStr string

	err := r.db.QueryRow(`SELECT id, type, payload, status, retries, max_retries FROM jobs WHERE id = ?`, id).
		Scan(&job.ID, &job.Type, &payloadStr, &job.Status, &job.Retries, &job.MaxRetries)

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

func (r *Repository) ClaimJob(id string) (bool, error) {
	result, err := r.db.Exec(`UPDATE jobs SET status = 'processing' WHERE id = ? AND status = 'pending'`, id)

	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected == 1, nil
}
