package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	logger.Info("Websocket manager initialized successfully")

	// Create HTTP server
	server := http.NewServer(websocket.WsManager)

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		logger.Info("Starting webserver on port 8080")
		serverErrors <- server.ListenAndServe()
	}()

	// Listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err.Error() != "http: Server closed" {
			logger.Error("Failed to start server", err)
		}
	case sig := <-sigChan:
		logger.Info(fmt.Sprint("Received signal, shutting down gracefully: ", sig))
		// Create a deadline to wait for.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server forced to shutdown: ", err)
		} else {
			logger.Info("Server shutdown gracefully")
		}
	}
}
