package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MonkaKokosowa/watchalong-server/api"
	customhttp "github.com/MonkaKokosowa/watchalong-server/http"
	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

func setup(t *testing.T) (*httptest.Server, func()) {
	PrepareDB()
	router := mux.NewRouter()
	customhttp.AddRoutes(router)
	server := httptest.NewServer(router)

	cleanup := func() {
		server.Close()
		CleanupDB()
	}

	return server, cleanup
}

func TestHTTPGetMovies(t *testing.T) {
	server, cleanup := setup(t)
	defer cleanup()

	newMovie := api.Movie{
		Name:    "Test Movie",
		IsMovie: true,
	}
	_, err := newMovie.AddMovie()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(server.URL + "/movies")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var movies []api.Movie
	if err := json.Unmarshal(body, &movies); err != nil {
		t.Fatal(err)
	}

	if len(movies) != 1 {
		t.Fatalf("expected 1 movie, got %d", len(movies))
	}

	if movies[0].Name != newMovie.Name {
		t.Errorf("expected movie name %s, got %s", newMovie.Name, movies[0].Name)
	}
}

func TestHTTPGetMovie(t *testing.T) {
	server, cleanup := setup(t)
	defer cleanup()

	newMovie := api.Movie{
		Name:    "Test Movie",
		IsMovie: true,
	}
	id, err := newMovie.AddMovie()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(server.URL + "/movies/" + fmt.Sprintf("%d", id))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var retrievedMovie api.Movie
	if err := json.Unmarshal(body, &retrievedMovie); err != nil {
		t.Fatal(err)
	}

	if retrievedMovie.Name != newMovie.Name {
		t.Errorf("expected movie name %s, got %s", newMovie.Name, retrievedMovie.Name)
	}
}

func TestHTTPAddMovie(t *testing.T) {
	server, cleanup := setup(t)
	defer cleanup()

	newMovie := api.Movie{
		Name:    "Test Movie",
		IsMovie: true,
	}

	jsonMovie, err := json.Marshal(newMovie)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(server.URL+"/add/movie", "application/json", strings.NewReader(string(jsonMovie)))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status Created, got %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]int
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}

	if _, ok := result["id"]; !ok {
		t.Errorf("expected id in response, got %v", result)
	}
}

func TestHTTPRateMovie(t *testing.T) {
	server, cleanup := setup(t)
	defer cleanup()

	newMovie := api.Movie{
		Name:    "Test Movie",
		IsMovie: true,
	}
	id, err := newMovie.AddMovie()
	if err != nil {
		t.Fatal(err)
	}

	rating := struct {
		MovieID  int     `json:"movieID"`
		Rating   float64 `json:"rating"`
		Username string  `json:"username"`
	}{
		MovieID:  id,
		Rating:   5,
		Username: "test",
	}

	jsonRating, err := json.Marshal(rating)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(server.URL+"/movies/rate", "application/json", strings.NewReader(string(jsonRating)))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status OK, got %v. Body: %s", resp.Status, string(body))
	}
}

func TestHTTPAddAlias(t *testing.T) {
	server, cleanup := setup(t)
	defer cleanup()

	newAlias := api.Alias{
		Alias:    "Test Alias",
		Username: "test",
	}

	jsonAlias, err := json.Marshal(newAlias)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(server.URL+"/alias", "application/json", strings.NewReader(string(jsonAlias)))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %v", resp.Status)
	}
}

func TestHTTPGetAliases(t *testing.T) {
	server, cleanup := setup(t)
	defer cleanup()

	newAlias := api.Alias{
		Alias:    "Test Alias",
		Username: "test",
	}
	err := newAlias.AddAlias()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(server.URL + "/alias")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var aliases map[string]string
	if err := json.Unmarshal(body, &aliases); err != nil {
		t.Fatal(err)
	}

	if len(aliases) != 1 {
		t.Fatalf("expected 1 alias, got %d", len(aliases))
	}

	if aliases["test"] != "Test Alias" {
		t.Errorf("expected alias 'Test Alias', got %s", aliases["test"])
	}
}

func TestHTTPAddMovieToQueue(t *testing.T) {
	server, cleanup := setup(t)
	defer cleanup()

	newMovie := api.Movie{
		Name:    "Test Movie",
		IsMovie: true,
	}
	id, err := newMovie.AddMovie()
	if err != nil {
		t.Fatal(err)
	}

	body := struct {
		ID int `json:"id"`
	}{
		ID: id,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(server.URL+"/queue/add", "application/json", strings.NewReader(string(jsonBody)))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %v", resp.Status)
	}
}

// func TestHTTPRemoveMovieFromQueue(t *testing.T) {
// 	server, cleanup := setup(t)
// 	defer cleanup()

// 	newMovie := api.Movie{
// 		Name:         "Test Movie",
// 		IsMovie:      true,
// 		ProposedBy:   "Test User",
// 		Ratings:      "{}",
// 		TmdbID:       00000,
// 		TmdbImageUrl: "google.com",
// 	}
// 	id, err := newMovie.AddMovie()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	newMovie.ID = id
// 	err = newMovie.AddMovieToQueue()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	body := struct {
// 		ID      int  `json:"id"`
// 		watched bool `json:"watched"`
// 	}{
// 		ID:      id,
// 		watched: true,
// 	}

// 	jsonBody, err := json.Marshal(body)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	resp, err := http.Post(server.URL+"/queue/remove", "application/json", strings.NewReader(string(jsonBody)))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		t.Fatalf("expected status OK, got %v", resp.Status)
// 	}
// }

func TestHTTPGetQueue(t *testing.T) {
	server, cleanup := setup(t)
	defer cleanup()

	newMovie := api.Movie{
		Name:    "Test Movie",
		IsMovie: true,
	}
	id, err := newMovie.AddMovie()
	if err != nil {
		t.Fatal(err)
	}
	newMovie.ID = id

	err = newMovie.AddMovieToQueue()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(server.URL + "/queue")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var movies []api.Movie
	if err := json.Unmarshal(body, &movies); err != nil {
		t.Fatal(err)
	}

	if len(movies) != 1 {
		t.Fatalf("expected 1 movie in queue, got %d", len(movies))
	}

	if movies[0].Name != newMovie.Name {
		t.Errorf("expected movie name %s, got %s", newMovie.Name, movies[0].Name)
	}
}
