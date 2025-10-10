package api

import (
	"database/sql"
	"encoding/json"

	"github.com/MonkaKokosowa/watchalong-server/database"
	_ "modernc.org/sqlite"
)

type Alias struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Alias    string `json:"alias"`
}

func (alias *Alias) AddAlias() error {
	var existingAlias Alias
	row := database.DB.QueryRow(`SELECT * FROM aliases WHERE username = ?`, alias.Username)
	if err := row.Scan(&existingAlias.ID, &existingAlias.Username, &existingAlias.Alias); err != nil {
		if err == sql.ErrNoRows {
			if _, err := database.DB.Exec(`INSERT INTO aliases (username, alias) VALUES (?, ?)`, alias.Username, alias.Alias); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if _, err := database.DB.Exec(`UPDATE aliases SET alias = ? WHERE username = ?`, alias.Alias, alias.Username); err != nil {
			return err
		}
	}
	return nil
}

func GetAliases() (map[string]string, error) {
	aliases := make(map[string]string)
	rows, err := database.DB.Query(`SELECT * FROM aliases`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var alias Alias
		if err := rows.Scan(&alias.ID, &alias.Username, &alias.Alias); err != nil {
			return nil, err
		}
		aliases[alias.Username] = alias.Alias
	}

	return aliases, nil
}

func ClearAliases() error {
	if _, err := database.DB.Exec(`DELETE FROM aliases`); err != nil {
		return err
	}
	return nil
}

type Movie struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	Watched       bool          `json:"watched"`
	IsMovie       bool          `json:"is_movie"`
	ProposedBy    string        `json:"proposed_by"`
	Ratings       string        `json:"ratings"`
	QueuePosition sql.NullInt64 `json:"queue_position"`
	TmdbID        int           `json:"tmdb_id"`
	TmdbImageUrl  string        `json:"tmdb_image_url"`
}

func RateMovie(movieID int, username string, rating float64) error {
	movie, err := GetMovie(movieID)
	if err != nil {
		return err
	}

	ratings := make(map[string]float64)
	if movie.Ratings != "[]" {
		if err := json.Unmarshal([]byte(movie.Ratings), &ratings); err != nil {
			return err
		}
	}

	ratings[username] = rating

	jsonBytes, err := json.Marshal(ratings)
	if err != nil {
		return err
	}

	if _, err := database.DB.Exec(`UPDATE movies SET ratings = ? WHERE id = ?`, string(jsonBytes), movieID); err != nil {
		return err
	}

	return nil
}

func (movie *Movie) ToJSON() string {
	jsonBytes, err := json.Marshal(movie)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

func (movie *Movie) AddMovie() (int, error) {
	if _, err := database.DB.Exec(`INSERT INTO movies (
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
	row := database.DB.QueryRow(`SELECT id FROM movies WHERE name = ?`, movie.Name)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func GetMovies() ([]Movie, error) {
	var movies []Movie
	rows, err := database.DB.Query(`SELECT * FROM movies`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var movie Movie
		if err := rows.Scan(&movie.ID,
			&movie.Name,
			&movie.Watched,
			&movie.IsMovie,
			&movie.ProposedBy,
			&movie.Ratings,
			&movie.QueuePosition,
			&movie.TmdbID,
			&movie.TmdbImageUrl); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	return movies, nil
}

func GetMovie(id int) (Movie, error) {
	var movie Movie
	row := database.DB.QueryRow(`SELECT * FROM movies WHERE id = ?`, id)
	if err := row.Scan(&movie.ID,
		&movie.Name,
		&movie.Watched,
		&movie.IsMovie,
		&movie.ProposedBy,
		&movie.Ratings,
		&movie.QueuePosition,
		&movie.TmdbID,
		&movie.TmdbImageUrl); err != nil {
		return movie, err
	}

	return movie, nil
}

func (movie *Movie) DeleteMovie() error {
	if _, err := database.DB.Exec(`DELETE FROM movies WHERE id = ?`, movie.ID); err != nil {
		return err
	}
	return nil
}

func ClearMovies() error {
	if _, err := database.DB.Exec(`DELETE FROM movies`); err != nil {
		return err
	}
	return nil
}

func (movie *Movie) AddMovieToQueue() error {
	// get highest queue position
	var highestQueuePosition sql.NullInt64
	row := database.DB.QueryRow(`SELECT MAX(queue_position) FROM movies WHERE queue_position IS NOT NULL`)
	if err := row.Scan(&highestQueuePosition); err != nil {
		return err
	}

	if highestQueuePosition.Valid {
		return database.DB.QueryRow("UPDATE movies SET queue_position = ? WHERE id = ? RETURNING id", highestQueuePosition.Int64+1, movie.ID).Scan(&movie.ID)
	} else {
		return database.DB.QueryRow("UPDATE movies SET queue_position = ? WHERE id = ? RETURNING id", 1, movie.ID).Scan(&movie.ID)
	}
}

func (movie *Movie) FinishMovie() error {
	if _, err := database.DB.Exec(`UPDATE movies SET watched = true WHERE id = ?`, movie.ID); err != nil {
		return err
	}

	if _, err := database.DB.Exec(`UPDATE movies SET queue_position = NULL WHERE id = ?`, movie.ID); err != nil {
		return err
	}
	return nil
}

func (movie *Movie) RemoveMovieFromQueue() error {

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`UPDATE movies SET queue_position = NULL WHERE id = ?`, movie.ID); err != nil {
		tx.Rollback()
		return err
	}

	rows, err := tx.Query(`SELECT * FROM movies WHERE queue_position > ?`, movie.QueuePosition.Int64)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var m Movie

		if err := rows.Scan(&m.ID, &m.Name, &m.Watched, &m.IsMovie, &m.ProposedBy, &m.Ratings, &m.QueuePosition, &m.TmdbID, &m.TmdbImageUrl); err != nil {
			tx.Rollback()
			return err
		}

		if m.QueuePosition.Int64 > movie.QueuePosition.Int64 {
			if _, err := tx.Exec("UPDATE movies SET queue_position = ? WHERE id = ?", m.QueuePosition.Int64-1, m.ID); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

func GetQueue() ([]Movie, error) {
	var movies []Movie
	rows, err := database.DB.Query(`SELECT * FROM movies WHERE queue_position IS NOT NULL ORDER BY queue_position ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var movie Movie
		if err := rows.Scan(&movie.ID, &movie.Name, &movie.Watched, &movie.IsMovie, &movie.ProposedBy, &movie.Ratings, &movie.QueuePosition, &movie.TmdbID, &movie.TmdbImageUrl); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	return movies, nil
}
