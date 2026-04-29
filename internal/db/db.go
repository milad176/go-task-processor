package db

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

func InitDB() *sql.DB {
	database, err := sql.Open("sqlite", "./jobs.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		log.Fatal(err)
	}

	// VERY IMPORTANT: SQLite should use only one open writer connection
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	database.SetConnMaxLifetime(time.Hour)

	// Create table if not exists
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		type TEXT,
		payload TEXT,
		status TEXT,
		retries INTEGER,
		max_retries INTEGER,
		priority INTEGER,
		claimed_at INTEGER
	);
	`

	_, err = database.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	// Faster pending-job search
	createStatusIndex := `
	CREATE INDEX IF NOT EXISTS idx_jobs_status_priority 
	ON jobs(status, priority DESC);
	`

	_, err = database.Exec(createStatusIndex)
	if err != nil {
		log.Fatal(err)
	}

	// Faster stuck-job recovery
	createClaimedAtIndex := `
	CREATE INDEX IF NOT EXISTS idx_jobs_claimed_at
	ON jobs(claimed_at);
	`

	_, err = database.Exec(createClaimedAtIndex)
	if err != nil {
		log.Fatal(err)
	}

	return database
}
