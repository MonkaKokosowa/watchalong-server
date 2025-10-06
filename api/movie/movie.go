package movie

import (
	"database/sql"
	"encoding/json"

	"github.com/MonkaKokosowa/watchalong-server/database"
	_ "modernc.org/sqlite"
)

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
		database.DB.Exec("UPDATE movies SET queue_position = ? WHERE id = ?", highestQueuePosition.Int64+1, movie.ID)
	} else {
		database.DB.Exec("UPDATE movies SET queue_position = ? WHERE id = ?", 1, movie.ID)
	}
	return nil
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
