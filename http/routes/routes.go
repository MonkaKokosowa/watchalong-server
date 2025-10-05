package routes

import (
	"net/http"

	"github.com/gorilla/mux"
)

func GetMovies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

}
