package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/MonkaKokosowa/watchalong-server/api/movie"
	"github.com/MonkaKokosowa/watchalong-server/api/queue"
	"github.com/gorilla/websocket"
)

type Manager struct {
	clients   map[*websocket.Conn]bool
	upgrader  websocket.Upgrader
	broadcast chan []byte
}

func NewManager() *Manager {
	return &Manager{
		clients:   make(map[*websocket.Conn]bool),
		upgrader:  websocket.Upgrader{},
		broadcast: make(chan []byte),
	}
}

func (m *Manager) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	m.clients[conn] = true

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			delete(m.clients, conn)
			break
		}
		log.Printf("Received message: %s", message)
	}
}

func (m *Manager) BroadcastUpdates() {
	movies, err := movie.GetMovies()
	if err != nil {
		log.Println(err)
		return
	}

	movieQueue, err := queue.GetQueue()
	if err != nil {
		log.Println(err)
		return
	}

	response := struct {
		Movies []movie.Movie `json:"movies"`
		Queue  []movie.Movie `json:"queue"`
	}{
		Movies: movies,
		Queue:  movieQueue,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
		return
	}

	for client := range m.clients {
		if err := client.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
			log.Println(err)
			client.Close()
			delete(m.clients, client)
		}
	}
}
