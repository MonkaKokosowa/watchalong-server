package http

import (
	"encoding/json"
	"net/http"

	"github.com/MonkaKokosowa/watchalong-server/api"
	"github.com/MonkaKokosowa/watchalong-server/logger"
)

type Vote struct {
	MovieID int `json:"movie_id"`
}

func GetCurrentVote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	movies, err := api.GetCurrentVote()
	if err != nil {
		logger.Error("Error getting current vote: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(movies)
}

func CastVote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var v Vote
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := api.CastVote(v.MovieID); err != nil {
		logger.Error("Error casting vote: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
