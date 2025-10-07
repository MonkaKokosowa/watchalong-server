package main

import (
	"github.com/MonkaKokosowa/watchalong-server/database"
	"github.com/MonkaKokosowa/watchalong-server/http"
	"github.com/MonkaKokosowa/watchalong-server/logger"
	"github.com/MonkaKokosowa/watchalong-server/websocket"
)

func main() {
	// Initialize the database

	_, err := database.InitializeDB("watchalong.sqlite")
	if err != nil {
		logger.Error("Failed to initialize database", err)
		return
	} else {
		logger.Info("Database initialized successfully")
	}
	defer database.CloseDatabase()

	// go wsManager.BroadcastUpdates()

	logger.Info("Websocket manager initialized successfully")

	// Start the server
	logger.Info("Starting webserver on port 8080")
	if err := http.StartServer(websocket.WsManager); err != nil {
		logger.Error("Failed to start server", err)
	}
}
