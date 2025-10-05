package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func StartServer() {
	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	http.ListenAndServe(":8080", router)
}

func AddMovie(w http.ResponseWriter, r *http.Request) {
	// Add movie logic here
	w.Write([]byte("Movie added"))
}
