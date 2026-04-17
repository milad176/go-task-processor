package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

func InitDB() *sql.DB {
	database, err := sql.Open("sqlite", "./jobs.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create table if not exists
	query := `
	CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		type TEXT,
		payload TEXT,
		status TEXT,
		retries INTEGER,
		max_retries INTEGER
	);
	`

	_, err = database.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

	return database
}
