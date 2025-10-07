package tests

import (
	"os"
	"testing"

	"github.com/MonkaKokosowa/watchalong-server/api"
	"github.com/MonkaKokosowa/watchalong-server/database"
)

func PrepareDB() {
	database.InitializeDB("testing.sqlite")
	api.ClearMovies()
	api.ClearAliases()
}

func CleanupDB() {
	database.CloseDatabase()
}

func TestMain(m *testing.M) {
	database.InitializeDB("testing.sqlite")

	code := m.Run()

	database.CloseDatabase()

	os.Remove("testing.sqlite")

	os.Exit(code)
}
