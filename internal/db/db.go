package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

func InitDB() *sql.DB {
	database, err := sql.Open("sqlite", "./jobs.db?_busy_timeout=5000")
	if err != nil {
		log.Fatal(err)
	}

	// Create table if not exists
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
	    type TEXT,
	    payload TEXT,
	    status TEXT,
	    retries INTEGER,
	    max_retries INTEGER,
	    priority INTEGER
	);
	`

	_, err = database.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	// Add index on status for faster job retrieval
	createIndexQuery := `
	CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
	`

	_, err = database.Exec(createIndexQuery)
	if err != nil {
		log.Fatal(err)
	}

	return database
}
