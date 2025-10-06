package http

import (
	"net/http"

	"github.com/MonkaKokosowa/watchalong-server/http/routes"
	"github.com/MonkaKokosowa/watchalong-server/websocket"
	"github.com/gorilla/mux"
)

func StartServer(wsManager *websocket.Manager) error {
	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	AddRoutes(router)
	router.HandleFunc("/ws", wsManager.WsHandler)

	error := http.ListenAndServe(":8080", router)

	return error
}

func AddRoutes(router *mux.Router) {
	router.HandleFunc("/movies", routes.GetMovies)
	router.HandleFunc("/movies/rate", routes.RateMovie).Methods("POST")
	router.HandleFunc("/movies/{movie_id}", routes.GetMovie)
	router.HandleFunc("/add/movie", routes.AddMovie).Methods("POST")
	router.HandleFunc("/alias", routes.AddAlias).Methods("POST")
	router.HandleFunc("/alias", routes.GetAliases).Methods("GET")
	router.HandleFunc("/queue/add", routes.AddMovieToQueue).Methods("POST")
	router.HandleFunc("/queue/remove", routes.RemoveMovieFromQueue).Methods("POST")
	router.HandleFunc("/queue", routes.GetQueue).Methods("GET")
	router.HandleFunc("/callback", routes.Callback).Methods("GET")
}
