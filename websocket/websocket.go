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
		if err.Error() == "websocket: close 1000 (normal)" {
			delete(m.clients, conn)
			break
		} else if err != nil {
			logger.Error("Failed to read message", err)
			delete(m.clients, conn)
			break
		}
		logger.Info(fmt.Sprint("Received message: ", string(message)))

	}
}

func (m *Manager) BroadcastUpdates(movies []api.Movie, queue []api.Movie, aliases []api.Alias, vote []api.Movie) {
	response := struct {
		Movies  []api.Movie `json:"movies"`
		Queue   []api.Movie `json:"queue"`
		Aliases []api.Alias `json:"aliases"`
		Vote    []api.Movie `json:"vote"`
	}{
		Movies:  movies,
		Queue:   queue,
		Aliases: aliases,
		Vote:    vote,
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
