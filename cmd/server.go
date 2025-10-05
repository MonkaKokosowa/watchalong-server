package main

import (
	"log"

	"github.com/MonkaKokosowa/watchalong-server/api"
)

func main() {
	// Initialize the database
	if err := api.InitializeDatabase(); err != nil {
		log.Fatal(err)
	}
	defer api.CloseDatabase()

	// Start the server
	if err := api.StartServer(); err != nil {
		log.Fatal(err)
	}
}
