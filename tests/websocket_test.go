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
}
