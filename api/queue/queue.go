package queue

import (
	"github.com/MonkaKokosowa/watchalong-server/api/movie"
	"github.com/MonkaKokosowa/watchalong-server/database"
	_ "modernc.org/sqlite"
)

func RemoveMovieFromQueue(id int) error {

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}

	retrievedMovie, err := movie.GetMovie(id)
	if err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`UPDATE movies SET queue_position = NULL WHERE id = ?`, id); err != nil {
		tx.Rollback()
		return err
	}

	rows, err := tx.Query(`SELECT * FROM movies WHERE queue_position > ?`, retrievedMovie.QueuePosition.Int64)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var m movie.Movie

		if err := rows.Scan(&m.ID, &m.Name, &m.Watched, &m.IsMovie, &m.ProposedBy, &m.Ratings, &m.QueuePosition, &m.TmdbID, &m.TmdbImageUrl); err != nil {
			tx.Rollback()
			return err
		}

		if m.QueuePosition.Int64 > retrievedMovie.QueuePosition.Int64 {
			if _, err := tx.Exec("UPDATE movies SET queue_position = ? WHERE id = ?", m.QueuePosition.Int64-1, m.ID); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

func GetQueue() ([]movie.Movie, error) {
	var movies []movie.Movie
	rows, err := database.DB.Query(`SELECT * FROM movies WHERE queue_position IS NOT NULL ORDER BY queue_position ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var movie movie.Movie
		if err := rows.Scan(&movie.ID, &movie.Name, &movie.Watched, &movie.IsMovie, &movie.ProposedBy, &movie.Ratings, &movie.QueuePosition, &movie.TmdbID, &movie.TmdbImageUrl); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	return movies, nil
}
