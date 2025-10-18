package tests

import (
	"testing"

	"github.com/MonkaKokosowa/watchalong-server/api"
	_ "modernc.org/sqlite"
)

func TestAddMovie(t *testing.T) {
	PrepareDB()
	newMovie := api.Movie{
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

	retrievedMovie, err := api.GetMovie(id)
	if err != nil {
		t.Fatalf("GetMovie() error = %v", err)
	}

	if retrievedMovie.Name != newMovie.Name {
		t.Errorf("got %s, want %s", retrievedMovie.Name, newMovie.Name)
	}
}

func TestGetMovies(t *testing.T) {
	PrepareDB()
	movie1 := api.Movie{
		Name:         "Test Movie 1",
		IsMovie:      true,
		ProposedBy:   "test",
		TmdbID:       123,
		TmdbImageUrl: "http://example.com/image.jpg",
	}

	movie2 := api.Movie{
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

	movies, err := api.GetMovies()
	if err != nil {
		t.Fatalf("GetMovies() error = %v", err)
	}

	if len(movies) != 2 {
		t.Errorf("got %d movies, want 2", len(movies))
	}
}

func TestDeleteMovie(t *testing.T) {
	PrepareDB()
	newMovie := api.Movie{
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

	_, err = api.GetMovie(id)
	if err == nil {
		t.Fatalf("GetMovie() should have failed, but it didn't")
	}
}

func TestAddMovieToQueue(t *testing.T) {
	PrepareDB()
	newMovie := api.Movie{
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

	retrievedMovie, err := api.GetMovie(id)
	if err != nil {
		t.Fatalf("GetMovie() error = %v", err)
	}

	if retrievedMovie.QueuePosition.Int64 != 1 {
		t.Errorf("got queue position %d, want 1", retrievedMovie.QueuePosition.Int64)
	}
}

func TestFinishMovie(t *testing.T) {
	PrepareDB()
	newMovie := api.Movie{
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

	retrievedMovie, err := api.GetMovie(id)
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
	movie1 := api.Movie{
		Name:         "Test Movie 1",
		IsMovie:      true,
		ProposedBy:   "test",
		TmdbID:       123,
		TmdbImageUrl: "http://example.com/image.jpg",
	}

	movie2 := api.Movie{
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

	if err := movie1.RemoveMovieFromQueue(); err != nil {
		t.Fatalf("RemoveMovieFromQueue() error = %v", err)
	}

	retrievedMovie2, err := api.GetMovie(id2)
	if err != nil {
		t.Fatalf("GetMovie() error = %v", err)
	}

	if retrievedMovie2.QueuePosition.Int64 != 1 {
		t.Errorf("got queue position %d, want 1", retrievedMovie2.QueuePosition.Int64)
	}
}

func TestRateMovie(t *testing.T) {
	PrepareDB()

	newMovie := api.Movie{
		Name:    "Test Movie",
		IsMovie: true,
	}
	id, err := newMovie.AddMovie()
	if err != nil {
		t.Fatal(err)
	}

	if err := api.RateMovie(id, "test", 5); err != nil {
		t.Fatal(err)
	}

	retrievedMovie, err := api.GetMovie(id)
	if err != nil {
		t.Fatal(err)
	}

	if retrievedMovie.Ratings != "{\"test\":5}" {
		t.Errorf("got ratings %s, want {\"test\":5}", retrievedMovie.Ratings)
	}
}

func TestCreateNewVote(t *testing.T) {
	PrepareDB()
	movie1 := api.Movie{
		Name:    "Test Movie 1",
		IsMovie: true,
	}
	movie2 := api.Movie{
		Name:    "Test Movie 2",
		IsMovie: true,
	}
	id1, err := movie1.AddMovie()
	if err != nil {
		t.Fatal(err)
	}
	id2, err := movie2.AddMovie()
	if err != nil {
		t.Fatal(err)
	}

	movieIDs := []int{id1, id2}
	if err := api.CreateNewVote(movieIDs); err != nil {
		t.Fatal(err)
	}

	currentVote, err := api.GetCurrentVote()
	if err != nil {
		t.Fatal(err)
	}

	if len(currentVote) != 2 {
		t.Errorf("got %d movies in current vote, want 2", len(currentVote))
	}
}

func TestGetCurrentVote(t *testing.T) {
	PrepareDB()
	movie1 := api.Movie{
		Name:    "Test Movie 1",
		IsMovie: true,
	}
	id1, err := movie1.AddMovie()
	if err != nil {
		t.Fatal(err)
	}

	movieIDs := []int{id1}
	if err := api.CreateNewVote(movieIDs); err != nil {
		t.Fatal(err)
	}

	currentVote, err := api.GetCurrentVote()
	if err != nil {
		t.Fatal(err)
	}

	if len(currentVote) != 1 {
		t.Errorf("got %d movies in current vote, want 1", len(currentVote))
	}
	if currentVote[0].ID != id1 {
		t.Errorf("got movie id %d, want %d", currentVote[0].ID, id1)
	}
}

func TestCastVote(t *testing.T) {
	PrepareDB()
	movie1 := api.Movie{
		Name:    "Test Movie 1",
		IsMovie: true,
	}
	id1, err := movie1.AddMovie()
	if err != nil {
		t.Fatal(err)
	}

	movieIDs := []int{id1}
	if err := api.CreateNewVote(movieIDs); err != nil {
		t.Fatal(err)
	}

	if err := api.CastVote(movieIDs); err != nil {
		t.Fatal(err)
	}

	winner, err := api.GetVoteWinner()
	if err != nil {
		t.Fatal(err)
	}

	if winner.ID != id1 {
		t.Errorf("got winner id %d, want %d", winner.ID, id1)
	}
}

func reverseInts(input []int) []int {
	if len(input) == 0 {
		return input
	}
	return append(reverseInts(input[1:]), input[0])
}

func TestGetVoteResults(t *testing.T) {
	PrepareDB()
	movie1 := api.Movie{
		Name:    "Test Movie 1",
		IsMovie: true,
	}
	movie2 := api.Movie{
		Name:    "Test Movie 2",
		IsMovie: true,
	}
	id1, err := movie1.AddMovie()
	if err != nil {
		t.Fatal(err)
	}
	id2, err := movie2.AddMovie()
	if err != nil {
		t.Fatal(err)
	}

	movieIDs := []int{id1, id2}
	if err := api.CreateNewVote(movieIDs); err != nil {
		t.Fatal(err)
	}

	if err := api.CastVote(reverseInts(movieIDs)); err != nil {
		t.Fatal(err)
	}

	results, err := api.GetVoteResults()
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].ID != id2 {
		t.Errorf("got winner id %d, want %d", results[0].ID, id2)
	}

	if results[1].ID != id1 {
		t.Errorf("got second place id %d, want %d", results[1].ID, id1)
	}
}

func TestGetUnwatchedMoviesNotInQueue(t *testing.T) {
	PrepareDB()
	movie1 := api.Movie{
		Name:    "Test Movie 1",
		IsMovie: true,
	}
	movie2 := api.Movie{
		Name:    "Test Movie 2",
		IsMovie: true,
	}
	movie3 := api.Movie{
		Name:    "Test Movie 3",
		IsMovie: true,
	}
	id1, err := movie1.AddMovie()
	if err != nil {
		t.Fatal(err)
	}
	id2, err := movie2.AddMovie()
	if err != nil {
		t.Fatal(err)
	}
	movie2.ID = id2
	if err := movie2.FinishMovie(); err != nil {
		t.Fatal(err)
	}
	id3, err := movie3.AddMovie()
	if err != nil {
		t.Fatal(err)
	}
	movie1.ID = id1
	movie3.ID = id3
	if err := movie3.AddMovieToQueue(); err != nil {
		t.Fatal(err)
	}

	movies, err := api.GetUnwatchedMoviesNotInQueue()
	if err != nil {
		t.Fatal(err)
	}

	if len(movies) != 1 {
		t.Errorf("got %d movies, want 1", len(movies))
	}
	if movies[0].ID != id1 {
		t.Errorf("got movie id %d, want %d", movies[0].ID, id1)
	}
}
