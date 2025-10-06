package tests

import (
	"testing"

	"github.com/MonkaKokosowa/watchalong-server/api/movie"
	"github.com/MonkaKokosowa/watchalong-server/api/queue"
	_ "modernc.org/sqlite"
)

func TestAddMovie(t *testing.T) {
	PrepareDB()
	newMovie := movie.Movie{
		Name:         "Test Movie",
		IsMovie:      true,
		ProposedBy:   "test",
		TmdbID:       123,
		TmdbImageUrl: "http://example.com/image.jpg",
	}

	id, err := newMovie.AddMovie()
	if err != nil {
		t.Fatalf("AddMovie() error = %v", err)
	}

	if id == 0 {
		t.Fatalf("AddMovie() returned id 0")
	}

	retrievedMovie, err := movie.GetMovie(id)
	if err != nil {
		t.Fatalf("GetMovie() error = %v", err)
	}

	if retrievedMovie.Name != newMovie.Name {
		t.Errorf("got %s, want %s", retrievedMovie.Name, newMovie.Name)
	}
}

func TestGetMovies(t *testing.T) {
	PrepareDB()
	movie1 := movie.Movie{
		Name:         "Test Movie 1",
		IsMovie:      true,
		ProposedBy:   "test",
		TmdbID:       123,
		TmdbImageUrl: "http://example.com/image.jpg",
	}

	movie2 := movie.Movie{
		Name:         "Test Movie 2",
		IsMovie:      true,
		ProposedBy:   "test",
		TmdbID:       456,
		TmdbImageUrl: "http://example.com/image2.jpg",
	}

	_, err := movie1.AddMovie()
	if err != nil {
		t.Fatalf("AddMovie() error = %v", err)
	}

	_, err = movie2.AddMovie()
	if err != nil {
		t.Fatalf("AddMovie() error = %v", err)
	}

	movies, err := movie.GetMovies()
	if err != nil {
		t.Fatalf("GetMovies() error = %v", err)
	}

	if len(movies) != 2 {
		t.Errorf("got %d movies, want 2", len(movies))
	}
}

func TestDeleteMovie(t *testing.T) {
	PrepareDB()
	newMovie := movie.Movie{
		Name:         "Test Movie",
		IsMovie:      true,
		ProposedBy:   "test",
		TmdbID:       123,
		TmdbImageUrl: "http://example.com/image.jpg",
	}

	id, err := newMovie.AddMovie()
	if err != nil {
		t.Fatalf("AddMovie() error = %v", err)
	}

	newMovie.ID = id
	if err := newMovie.DeleteMovie(); err != nil {
		t.Fatalf("DeleteMovie() error = %v", err)
	}

	_, err = movie.GetMovie(id)
	if err == nil {
		t.Fatalf("GetMovie() should have failed, but it didn't")
	}
}

func TestAddMovieToQueue(t *testing.T) {
	PrepareDB()
	newMovie := movie.Movie{
		Name:         "Test Movie",
		IsMovie:      true,
		ProposedBy:   "test",
		TmdbID:       123,
		TmdbImageUrl: "http://example.com/image.jpg",
	}

	id, err := newMovie.AddMovie()
	if err != nil {
		t.Fatalf("AddMovie() error = %v", err)
	}

	newMovie.ID = id
	if err := newMovie.AddMovieToQueue(); err != nil {
		t.Fatalf("AddMovieToQueue() error = %v", err)
	}

	retrievedMovie, err := movie.GetMovie(id)
	if err != nil {
		t.Fatalf("GetMovie() error = %v", err)
	}

	if retrievedMovie.QueuePosition.Int64 != 1 {
		t.Errorf("got queue position %d, want 1", retrievedMovie.QueuePosition.Int64)
	}
}

func TestFinishMovie(t *testing.T) {
	PrepareDB()
	newMovie := movie.Movie{
		Name:         "Test Movie",
		IsMovie:      true,
		ProposedBy:   "test",
		TmdbID:       123,
		TmdbImageUrl: "http://example.com/image.jpg",
	}

	id, err := newMovie.AddMovie()
	if err != nil {
		t.Fatalf("AddMovie() error = %v", err)
	}

	newMovie.ID = id
	if err := newMovie.AddMovieToQueue(); err != nil {
		t.Fatalf("AddMovieToQueue() error = %v", err)
	}

	if err := newMovie.FinishMovie(); err != nil {
		t.Fatalf("FinishMovie() error = %v", err)
	}

	retrievedMovie, err := movie.GetMovie(id)
	if err != nil {
		t.Fatalf("GetMovie() error = %v", err)
	}

	if !retrievedMovie.Watched {
		t.Errorf("got watched %t, want true", retrievedMovie.Watched)
	}

	if retrievedMovie.QueuePosition.Valid {
		t.Errorf("got queue position %d, want null", retrievedMovie.QueuePosition.Int64)
	}
}

func TestRemoveMovieFromQueue(t *testing.T) {
	PrepareDB()
	movie1 := movie.Movie{
		Name:         "Test Movie 1",
		IsMovie:      true,
		ProposedBy:   "test",
		TmdbID:       123,
		TmdbImageUrl: "http://example.com/image.jpg",
	}

	movie2 := movie.Movie{
		Name:         "Test Movie 2",
		IsMovie:      true,
		ProposedBy:   "test",
		TmdbID:       456,
		TmdbImageUrl: "http://example.com/image2.jpg",
	}

	id1, err := movie1.AddMovie()
	if err != nil {
		t.Fatalf("AddMovie() error = %v", err)
	}

	id2, err := movie2.AddMovie()
	if err != nil {
		t.Fatalf("AddMovie() error = %v", err)
	}

	movie1.ID = id1
	movie2.ID = id2

	if err := movie1.AddMovieToQueue(); err != nil {
		t.Fatalf("AddMovieToQueue() error = %v", err)
	}

	if err := movie2.AddMovieToQueue(); err != nil {
		t.Fatalf("AddMovieToQueue() error = %v", err)
	}

	if err := queue.RemoveMovieFromQueue(id1); err != nil {
		t.Fatalf("RemoveMovieFromQueue() error = %v", err)
	}

	retrievedMovie2, err := movie.GetMovie(id2)
	if err != nil {
		t.Fatalf("GetMovie() error = %v", err)
	}

	if retrievedMovie2.QueuePosition.Int64 != 1 {
		t.Errorf("got queue position %d, want 1", retrievedMovie2.QueuePosition.Int64)
	}
}

func TestRateMovie(t *testing.T) {
	PrepareDB()

	newMovie := movie.Movie{
		Name:    "Test Movie",
		IsMovie: true,
	}
	id, err := newMovie.AddMovie()
	if err != nil {
		t.Fatal(err)
	}

	if err := movie.RateMovie(id, "test", 5); err != nil {
		t.Fatal(err)
	}

	retrievedMovie, err := movie.GetMovie(id)
	if err != nil {
		t.Fatal(err)
	}

	if retrievedMovie.Ratings != "{\"test\":5}" {
		t.Errorf("got ratings %s, want {\"test\":5}", retrievedMovie.Ratings)
	}
}
