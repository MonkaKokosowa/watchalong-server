package tests

import (
	"os"
	"testing"

	"github.com/MonkaKokosowa/watchalong-server/api/alias"
	"github.com/MonkaKokosowa/watchalong-server/api/movie"
	"github.com/MonkaKokosowa/watchalong-server/database"
)

func PrepareDB() {
	database.InitializeDB("testing.sqlite")
	movie.ClearMovies()
	alias.ClearAliases()
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
