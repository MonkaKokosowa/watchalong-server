# watchalong-server

A Golang server for managing a shared movie watchlist with queue management, ratings, and real-time updates via WebSockets.

## Getting Started

### Prerequisites

- Go 1.21 or later
- SQLite (included via modernc.org/sqlite)

### Installation

```bash
# Clone the repository
git clone https://github.com/MonkaKokosowa/watchalong-server.git
cd watchalong-server

# Install dependencies
go mod download
```

### Running the Server

Using Go directly:
```bash
go run cmd/server.go
```

Using Make:
```bash
make run
```

Build the binary:
```bash
make build
# Binary will be created at build/watchalong
./build/watchalong
```

Using Docker:
```bash
docker build -t watchalong-server .
docker run -p 8080:8080 watchalong-server
```

The server will start on **port 8080** and create a SQLite database file `watchalong.sqlite` in the current directory.

## HTTP API

### Movies

#### Get All Movies
```
GET /movies
```
Returns a list of all movies in the database.

**Response:**
```json
[
  {
    "id": 1,
    "name": "Movie Name",
    "watched": false,
    "is_movie": true,
    "proposed_by": "username",
    "ratings": "{}",
    "queue_position": null,
    "tmdb_id": 12345,
    "tmdb_image_url": "https://..."
  }
]
```

#### Get Single Movie
```
GET /movies/{movie_id}
```
Returns details for a specific movie.

#### Add Movie
```
POST /add/movie
Content-Type: application/json

{
  "name": "Movie Name",
  "is_movie": true,
  "proposed_by": "username",
  "tmdb_id": 12345,
  "tmdb_image_url": "https://..."
}
```
Creates a new movie and returns its ID.

**Response:**
```json
{"id": 1}
```

#### Rate Movie
```
POST /movies/rate
Content-Type: application/json

{
  "movieID": 1,
  "rating": 5,
  "username": "user123"
}
```
Adds or updates a user's rating for a movie (0-10 scale).

### Queue Management

#### Get Queue
```
GET /queue
```
Returns all movies currently in the queue, ordered by queue position.

#### Add Movie to Queue
```
POST /queue/add
Content-Type: application/json

{
  "id": 1
}
```
Adds a movie to the end of the queue.

#### Remove Movie from Queue
```
POST /queue/remove
Content-Type: application/json

{
  "id": 1
}
```
Removes a movie from the queue and reorders remaining movies.

### User Aliases

#### Get All Aliases
```
GET /alias
```
Returns a mapping of usernames to display names.

**Response:**
```json
{
  "username1": "Display Name 1",
  "username2": "Display Name 2"
}
```

#### Add/Update Alias
```
POST /alias
Content-Type: application/json

{
  "username": "user123",
  "alias": "Cool User"
}
```
Creates or updates a display name for a username.

### OAuth Callback

#### OAuth Callback Handler
```
GET /callback
```
HTML page that handles OAuth redirects and forwards tokens to the application via custom URL scheme.

## WebSocket API

The server supports WebSocket connections at `/ws` endpoint. You can send JSON messages to request data:

### Movies Request
Send:
```json
{"type": "movies"}
```

Receive:
```json
{
  "type": "movies",
  "movies": [
    {
      "id": 1,
      "name": "Movie Name",
      "watched": false,
      "is_movie": true,
      "proposed_by": "username",
      "ratings": "{}",
      "queue_position": null,
      "tmdb_id": 12345,
      "tmdb_image_url": "https://..."
    }
  ]
}
```

### Queue Request
Send:
```json
{"type": "queue"}
```

Receive:
```json
{
  "type": "queue",
  "queue": [
    {
      "id": 1,
      "name": "Movie Name",
      "watched": false,
      "is_movie": true,
      "proposed_by": "username",
      "ratings": "{}",
      "queue_position": {"Int64": 1, "Valid": true},
      "tmdb_id": 12345,
      "tmdb_image_url": "https://..."
    }
  ]
}
```

### Alias Request
Send:
```json
{"type": "alias"}
```

Receive:
```json
{
  "type": "alias",
  "aliases": {
    "username": "DisplayName"
  }
}
```

### Broadcast Updates
The server also broadcasts updates to all connected clients when movies or queue changes via HTTP endpoints. These broadcast messages have the format:
```json
{
  "movies": [...],
  "queue": [...]
}
```

## Data Models

### Movie
```go
{
  "id": int,              // Unique identifier
  "name": string,         // Movie/show name
  "watched": bool,        // Whether it has been watched
  "is_movie": bool,       // true for movies, false for TV shows
  "proposed_by": string,  // Username of proposer
  "ratings": string,      // JSON object of username:rating pairs
  "queue_position": {     // Position in queue (null if not queued)
    "Int64": int,
    "Valid": bool
  },
  "tmdb_id": int,         // The Movie Database ID
  "tmdb_image_url": string // URL to movie poster/image
}
```

### Alias
```go
{
  "id": int,           // Unique identifier
  "username": string,  // Original username
  "alias": string      // Display name
}
```

## Database

The server uses SQLite for data persistence. The database is automatically initialized on startup with two tables:

- **movies**: Stores movie/show information, ratings, and queue positions
- **aliases**: Stores username-to-displayname mappings

Database file: `watchalong.sqlite` (created in the working directory)

## Testing

Run all tests:
```bash
make test
# or
go test -v ./tests
```

Run specific tests:
```bash
go test -v ./tests -run TestHTTPGetMovies
```

Test coverage includes:
- API layer tests (movie CRUD, queue management, ratings, aliases)
- HTTP endpoint tests
- WebSocket connection and messaging tests

## Project Structure

```
.
├── api/              # Database operations and business logic
├── cmd/              # Application entry point
├── database/         # Database initialization and connection
├── http/             # HTTP server and routing
│   └── routes/       # HTTP route handlers
├── logger/           # Logging utilities
├── tests/            # Test files
└── websocket/        # WebSocket connection management
```

## Development

### Adding New Endpoints

1. Add database operations in `api/api.go`
2. Create HTTP handler in `http/routes/routes.go`
3. Register route in `http/http.go` `AddRoutes()` function
4. Add tests in `tests/`

### WebSocket Message Handling

To add new WebSocket message types, modify `websocket/websocket.go`:
1. Add a case in `handleRequest()` switch statement
2. Implement handler function following the pattern of existing handlers

## License

See LICENSE file for details.
