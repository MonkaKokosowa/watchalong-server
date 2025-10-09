package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/MonkaKokosowa/watchalong-server/api"
	"github.com/MonkaKokosowa/watchalong-server/logger"
	"github.com/gorilla/websocket"
)

var WsManager = NewManager()

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
		logger.Error("Failed to upgrade websocket ", err)
		return
	}
	defer conn.Close()

	m.clients[conn] = true

	for {
		_, message, err := conn.ReadMessage()

		// HACKY FIX FIND A BETTER WAY TO HANDLE CLOSE
		if err != nil {
			if err.Error() == "websocket: close 1000 (normal)" {
				delete(m.clients, conn)
				break
			}
			logger.Error("Failed to read message", err)
			delete(m.clients, conn)
			break
		}
		logger.Info(fmt.Sprint("Received message: ", string(message)))

		// Handle incoming requests
		m.handleRequest(conn, message)
	}
}

func (m *Manager) handleRequest(conn *websocket.Conn, message []byte) {
	var request struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(message, &request); err != nil {
		logger.Error("Failed to unmarshal request", err)
		return
	}

	switch request.Type {
	case "movies":
		m.handleMoviesRequest(conn)
	case "queue":
		m.handleQueueRequest(conn)
	case "alias":
		m.handleAliasRequest(conn)
	default:
		logger.Info(fmt.Sprint("Unknown request type: ", request.Type))
	}
}

func (m *Manager) handleMoviesRequest(conn *websocket.Conn) {
	movies, err := api.GetMovies()
	if err != nil {
		logger.Error("Failed to get movies", err)
		movies = []api.Movie{}
	}

	response := struct {
		Type   string      `json:"type"`
		Movies []api.Movie `json:"movies"`
	}{
		Type:   "movies",
		Movies: movies,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal movies response", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
		logger.Error("Failed to send movies response", err)
	}
}

func (m *Manager) handleQueueRequest(conn *websocket.Conn) {
	queue, err := api.GetQueue()
	if err != nil {
		logger.Error("Failed to get queue", err)
		queue = []api.Movie{}
	}

	response := struct {
		Type  string      `json:"type"`
		Queue []api.Movie `json:"queue"`
	}{
		Type:  "queue",
		Queue: queue,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal queue response", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
		logger.Error("Failed to send queue response", err)
	}
}

func (m *Manager) handleAliasRequest(conn *websocket.Conn) {
	aliases, err := api.GetAliases()
	if err != nil {
		logger.Error("Failed to get aliases", err)
		aliases = make(map[string]string)
	}

	response := struct {
		Type    string            `json:"type"`
		Aliases map[string]string `json:"aliases"`
	}{
		Type:    "alias",
		Aliases: aliases,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal alias response", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
		logger.Error("Failed to send alias response", err)
	}
}

func (m *Manager) BroadcastUpdates(movies []api.Movie, queue []api.Movie) {
	response := struct {
		Movies []api.Movie `json:"movies"`
		Queue  []api.Movie `json:"queue"`
	}{
		Movies: movies,
		Queue:  queue,
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
