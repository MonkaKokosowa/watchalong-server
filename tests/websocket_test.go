package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MonkaKokosowa/watchalong-server/api"
	customhttp "github.com/MonkaKokosowa/watchalong-server/http"
	"github.com/MonkaKokosowa/watchalong-server/websocket"
	"github.com/gorilla/mux"
	gwebsocket "github.com/gorilla/websocket"
)

func TestWebSocketUpdate(t *testing.T) {
	PrepareDB()
	router := mux.NewRouter()
	customhttp.AddRoutes(router)
	router.HandleFunc("/ws", websocket.WsManager.WsHandler)
	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	ws, _, err := gwebsocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("could not open a ws connection on %s: %v", wsURL, err)
	}
	defer ws.Close()

	// 1. Add a movie
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

	// Read the first message
	_, _, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("could not read message: %v", err)
	}

	var addMovieResponse struct {
		ID int `json:"id"`
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(bodyBytes, &addMovieResponse); err != nil {
		t.Fatal(err)
	}

	// 2. Add movie to queue
	body := struct {
		ID int `json:"id"`
	}{
		ID: addMovieResponse.ID,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = http.Post(server.URL+"/queue/add", "application/json", strings.NewReader(string(jsonBody)))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// 3. Check for websocket message
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("could not read message: %v", err)
	}

	expected := fmt.Sprintf(`{"movies":[{"id":%d,"name":"Test Movie","watched":false,"is_movie":true,"proposed_by":"","ratings":"[]","queue_position":{"Int64":1,"Valid":true},"tmdb_id":0,"tmdb_image_url":""}],"queue":[{"id":%d,"name":"Test Movie","watched":false,"is_movie":true,"proposed_by":"","ratings":"[]","queue_position":{"Int64":1,"Valid":true},"tmdb_id":0,"tmdb_image_url":""}]}`, addMovieResponse.ID, addMovieResponse.ID)
	if string(p) != expected {
		t.Errorf("got %s, want %s", string(p), expected)
	}

	// Gracefully close the connection
	err = ws.WriteMessage(gwebsocket.CloseMessage, gwebsocket.FormatCloseMessage(gwebsocket.CloseNormalClosure, ""))
	if err != nil {
		t.Fatalf("could not write close message: %v", err)
	}
	defer ws.Close()

}

func TestWebSocketMoviesRequest(t *testing.T) {
	PrepareDB()
	router := mux.NewRouter()
	customhttp.AddRoutes(router)
	router.HandleFunc("/ws", websocket.WsManager.WsHandler)
	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	ws, _, err := gwebsocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("could not open a ws connection on %s: %v", wsURL, err)
	}
	defer ws.Close()

	// Add a test movie first
	newMovie := api.Movie{
		Name:    "Test Movie for Request",
		IsMovie: true,
	}
	id, err := newMovie.AddMovie()
	if err != nil {
		t.Fatal(err)
	}

	// Send a movies request
	request := struct {
		Type string `json:"type"`
	}{
		Type: "movies",
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	err = ws.WriteMessage(gwebsocket.TextMessage, requestBytes)
	if err != nil {
		t.Fatalf("could not send movies request: %v", err)
	}

	// Read the response
	_, message, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("could not read message: %v", err)
	}

	var response struct {
		Type   string      `json:"type"`
		Movies []api.Movie `json:"movies"`
	}

	if err := json.Unmarshal(message, &response); err != nil {
		t.Fatalf("could not unmarshal response: %v", err)
	}

	if response.Type != "movies" {
		t.Errorf("expected type 'movies', got '%s'", response.Type)
	}

	if len(response.Movies) == 0 {
		t.Error("expected at least one movie")
	}

	found := false
	for _, movie := range response.Movies {
		if movie.ID == id {
			found = true
			if movie.Name != "Test Movie for Request" {
				t.Errorf("expected movie name 'Test Movie for Request', got '%s'", movie.Name)
			}
		}
	}

	if !found {
		t.Error("expected to find the test movie in the response")
	}
}

func TestWebSocketQueueRequest(t *testing.T) {
	PrepareDB()
	router := mux.NewRouter()
	customhttp.AddRoutes(router)
	router.HandleFunc("/ws", websocket.WsManager.WsHandler)
	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	ws, _, err := gwebsocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("could not open a ws connection on %s: %v", wsURL, err)
	}
	defer ws.Close()

	// Add a test movie and add it to queue
	newMovie := api.Movie{
		Name:    "Test Movie for Queue Request",
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

	// Send a queue request
	request := struct {
		Type string `json:"type"`
	}{
		Type: "queue",
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	err = ws.WriteMessage(gwebsocket.TextMessage, requestBytes)
	if err != nil {
		t.Fatalf("could not send queue request: %v", err)
	}

	// Read the response
	_, message, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("could not read message: %v", err)
	}

	var response struct {
		Type  string      `json:"type"`
		Queue []api.Movie `json:"queue"`
	}

	if err := json.Unmarshal(message, &response); err != nil {
		t.Fatalf("could not unmarshal response: %v", err)
	}

	if response.Type != "queue" {
		t.Errorf("expected type 'queue', got '%s'", response.Type)
	}

	if len(response.Queue) == 0 {
		t.Error("expected at least one movie in queue")
	}

	found := false
	for _, movie := range response.Queue {
		if movie.ID == id {
			found = true
			if movie.Name != "Test Movie for Queue Request" {
				t.Errorf("expected movie name 'Test Movie for Queue Request', got '%s'", movie.Name)
			}
		}
	}

	if !found {
		t.Error("expected to find the test movie in the queue response")
	}
}

func TestWebSocketAliasRequest(t *testing.T) {
	PrepareDB()
	router := mux.NewRouter()
	customhttp.AddRoutes(router)
	router.HandleFunc("/ws", websocket.WsManager.WsHandler)
	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	ws, _, err := gwebsocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("could not open a ws connection on %s: %v", wsURL, err)
	}
	defer ws.Close()

	// Add a test alias
	testAlias := api.Alias{
		Username: "testuser",
		Alias:    "TestUser",
	}
	err = testAlias.AddAlias()
	if err != nil {
		t.Fatal(err)
	}

	// Send an alias request
	request := struct {
		Type string `json:"type"`
	}{
		Type: "alias",
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	err = ws.WriteMessage(gwebsocket.TextMessage, requestBytes)
	if err != nil {
		t.Fatalf("could not send alias request: %v", err)
	}

	// Read the response
	_, message, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("could not read message: %v", err)
	}

	var response struct {
		Type    string            `json:"type"`
		Aliases map[string]string `json:"aliases"`
	}

	if err := json.Unmarshal(message, &response); err != nil {
		t.Fatalf("could not unmarshal response: %v", err)
	}

	if response.Type != "alias" {
		t.Errorf("expected type 'alias', got '%s'", response.Type)
	}

	if len(response.Aliases) == 0 {
		t.Error("expected at least one alias")
	}

	if alias, ok := response.Aliases["testuser"]; !ok {
		t.Error("expected to find 'testuser' in aliases")
	} else if alias != "TestUser" {
		t.Errorf("expected alias 'TestUser', got '%s'", alias)
	}
}
