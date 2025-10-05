package api

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type Movie struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Watched       bool   `json:"watched"`
	IsMovie       bool   `json:"is_movie"`
	ProposedBy    string `json:"proposed_by"`
	Ratings       string `json:"ratings"`
	QueuePosition int    `json:"queue_position"`
	TmdbID        int    `json:"tmdb_id"`
	TmdbImageUrl  string `json:"tmdb_image_url"`
}

type Alias struct {
	id       int    `json:"id"`
	username string `json:"username"`
	alias    string `json:"alias"`
}

var db *sql.DB

func InitializeDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "./watchalong.db")
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS movies (
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
		return err
	}
	return nil
}

func CloseDatabase() error {
	if err := db.Close(); err != nil {
		return err
	}
	return nil
}

func (movie *Movie) AddMovie() (int, error) {
	if _, err := db.Exec(`INSERT INTO movies (
		name,
		is_movie,
		proposed_by,
		tmdb_id,
		tmdb_image_url
	) VALUES (?, ?, ?, ?, ?)`,
		movie.Name,
		movie.IsMovie,
		movie.ProposedBy,
		movie.TmdbID,
		movie.TmdbImageUrl); err != nil {
		return 0, err
	}

	id := 0
	row := db.QueryRow(`SELECT id FROM movies WHERE name = ?`, movie.Name)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func GetMovie(id int) (Movie, error) {
	var movie Movie
	row := db.QueryRow(`SELECT * FROM movies WHERE id = ?`, id)
	if err := row.Scan(&movie.ID, &movie.Name, &movie.Watched, &movie.IsMovie, &movie.ProposedBy, &movie.Ratings, &movie.QueuePosition, &movie.TmdbID, &movie.TmdbImageUrl); err != nil {
		return movie, err
	}

	return movie, nil
}

func (movie *Movie) DeleteMovie() error {
	if _, err := db.Exec(`DELETE FROM movies WHERE id = ?`, movie.ID); err != nil {
		return err
	}
	return nil
}

func (movie *Movie) AddMovieToQueue() error {
	// get highest queue position
	var highestQueuePosition int
	row := db.QueryRow(`SELECT MAX(queue_position) FROM movies WHERE queue_position IS NOT NULL`)
	if err := row.Scan(&highestQueuePosition); err != nil {
		return err
	}

	db.Exec("UPDATE movies SET queue_position = ? WHERE id = ?", highestQueuePosition+1, movie.ID)
	return nil
}

func (movie *Movie) FinishMovie() error {
	if _, err := db.Exec(`UPDATE movies SET watched = true WHERE id = ?`, movie.ID); err != nil {
		return err
	}

	if _, err := db.Exec(`UPDATE movies SET queue_position = NULL WHERE id = ?`, movie.ID); err != nil {
		return err
	}
	return nil
}

func (movie Movie) RemoveMovieFromQueue() error {

	if _, err := db.Exec(`UPDATE movies SET queue_position = NULL WHERE id = ?`, movie.ID); err != nil {
		return err
	}

	movies, err := db.Query(`SELECT * FROM movies WHERE queue_position > ?`, movie.QueuePosition)
	if err != nil {
		return err
	}
	defer movies.Close()

	for movies.Next() {
		var id int
		var name string
		var watched bool
		var isMovie bool
		var proposedBy string
		var ratings string
		var queuePosition int
		var tmdbID int
		var tmdbImageUrl string

		if err := movies.Scan(&id, &name, &watched, &isMovie, &proposedBy, &ratings, &queuePosition, &tmdbID, &tmdbImageUrl); err != nil {
			return err
		}

		if queuePosition > movie.QueuePosition {
			db.Exec("UPDATE movies SET queue_position = ? WHERE id = ?", queuePosition-1, id)
		}
	}
	return nil
}
