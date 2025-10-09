# watchalong-server
Basic Golang server for handling the logic

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
