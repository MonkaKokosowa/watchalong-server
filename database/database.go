package database

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitializeDB(filepath string) (*sql.DB, error) {
	var err error
	DB, err = sql.Open("sqlite", filepath)
	if err != nil {
		return nil, err
	}

	statement, err := DB.Prepare(`CREATE TABLE IF NOT EXISTS movies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		watched BOOLEAN NOT NULL DEFAULT 0,
		is_movie BOOLEAN NOT NULL,
		proposed_by TEXT NOT NULL,
		ratings TEXT NOT NULL DEFAULT '[]',
		queue_position INTEGER,
		tmdb_id INTEGER NOT NULL,
		tmdb_image_url TEXT NOT NULL
		)`)
	if err != nil {
		return nil, err
	}
	statement.Exec()

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS aliases (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		alias TEXT NOT NULL,
		avatar_url TEXT
	)`)
	if err != nil {
		return nil, err
	}

	return DB, nil
}

func CloseDatabase() error {
	if err := DB.Close(); err != nil {
		return err
	}
	return nil
}
