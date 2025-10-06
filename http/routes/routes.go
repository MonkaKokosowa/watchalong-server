package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MonkaKokosowa/watchalong-server/api/alias"
	"github.com/MonkaKokosowa/watchalong-server/api/movie"
	"github.com/MonkaKokosowa/watchalong-server/api/queue"
	"github.com/MonkaKokosowa/watchalong-server/logger"
	"github.com/gorilla/mux"
)

func GetMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := movie.GetMovies()
	if err != nil {
		logger.Error("Failed to get movies", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonBytes, err := json.Marshal(movies)
	if err != nil {
		logger.Error("Failed to marshal movies", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func GetMovie(w http.ResponseWriter, r *http.Request) {
	movie_id, err := strconv.Atoi(mux.Vars(r)["movie_id"])

	if err != nil {
		logger.Error("Failed to parse movie ID", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	retrievedMovie, err := movie.GetMovie(movie_id)
	if err != nil {
		logger.Error("Failed to get movie", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(retrievedMovie.ToJSON()))
}

func AddMovie(w http.ResponseWriter, r *http.Request) {
	var newMovie movie.Movie
	if err := json.NewDecoder(r.Body).Decode(&newMovie); err != nil {
		logger.Error("Failed to decode movie", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := newMovie.AddMovie()
	if err != nil {
		logger.Error("Failed to add movie", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"id": %d}`, id)))
}

func RateMovie(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MovieID  int     `json:"movieID"`
		Rating   float64 `json:"rating"`
		Username string  `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.Error("Failed to decode movie rating", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := movie.RateMovie(body.MovieID, body.Username, body.Rating); err != nil {
		logger.Error("Failed to rate movie", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func AddAlias(w http.ResponseWriter, r *http.Request) {
	var newAlias alias.Alias
	if err := json.NewDecoder(r.Body).Decode(&newAlias); err != nil {
		logger.Error("Failed to decode alias", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := newAlias.AddAlias(); err != nil {
		logger.Error("Failed to add alias", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetAliases(w http.ResponseWriter, r *http.Request) {
	aliases, err := alias.GetAliases()
	if err != nil {
		logger.Error("Failed to get aliases", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonBytes, err := json.Marshal(aliases)
	if err != nil {
		logger.Error("Failed to marshal aliases", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func AddMovieToQueue(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.Error("Failed to decode movie ID", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	retrievedMovie, err := movie.GetMovie(body.ID)
	if err != nil {
		logger.Error("Failed to get movie", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := retrievedMovie.AddMovieToQueue(); err != nil {
		logger.Error("Failed to add movie to queue", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func RemoveMovieFromQueue(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.Error("Failed to decode movie ID", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := queue.RemoveMovieFromQueue(body.ID); err != nil {
		logger.Error("Failed to remove movie from queue", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetQueue(w http.ResponseWriter, r *http.Request) {
	queue, err := queue.GetQueue()
	if err != nil {
		logger.Error("Failed to get queue", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonBytes, err := json.Marshal(queue)
	if err != nil {
		logger.Error("Failed to marshal queue", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func Callback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Redirecting...</title>
    <script>
        // Extract the access_token parameter from the URL
        const urlParams = new URLSearchParams(window.location.hash);
        const access_token = urlParams.get('access_token');

        // Redirect to the custom scheme with the access_token
        if (access_token) {
            window.location.href = 'watchalong://callback#access_token=' + access_token;
        } else {
            document.body.innerHTML = '<h1>Error: No code parameter found in URL</h1>';
        }
    </script>
</head>
<body>
    <h1>Redirecting...</h1>
</body>
</html>`))
}
