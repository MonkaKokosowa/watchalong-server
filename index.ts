import { Database } from "bun:sqlite";
import type { Server, ServerWebSocket, WebSocketHandler } from "bun";

export let db = new Database("watchalong.sqlite");

if (process.env.NODE_ENV === "test") {
  db = new Database(":memory:");
}

// Create table if it doesn't exist
db.run(`
  CREATE TABLE IF NOT EXISTS movies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    watched BOOLEAN NOT NULL,
    type TEXT NOT NULL,
    proposed_by TEXT NOT NULL,
    ratings TEXT NOT NULL,
    queue_position INTEGER
  )
`);
db.run(`
  CREATE TABLE IF NOT EXISTS aliases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    alias TEXT NOT NULL
  )
`);

function getMoviesAndQueue() {
  const movies = db.query("SELECT * FROM movies").all();
  const queue = db
    .query(
      "SELECT * FROM movies WHERE queue_position IS NOT NULL ORDER BY queue_position ASC",
    )
    .all();
  return { movies, queue };
}

function broadcastUpdates(server: Server) {
  const { movies, queue } = getMoviesAndQueue();
  server.publish("watchalong", JSON.stringify({ movies, queue }));
}

export async function fetch(req: Request, server: Server) {
  const url = new URL(req.url);


  // Handle CORS preflight
  if (req.method === "OPTIONS") {
    return new Response(null, {
      status: 204,
      headers: {
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
        "Access-Control-Allow-Headers": "Content-Type",
      },
    });
  }

  if (server.upgrade(req)) {
    return; // do not return a Response object
  }

  if (req.method === "POST" && url.pathname === "/alias") {
    const { username, alias } = await req.json();
    
    if (!username || !alias) {
      return new Response("Missing required fields", { status: 400, headers: { "Access-Control-Allow-Origin": "*" } });
    }
    const existing = db.query("SELECT * FROM aliases WHERE username = ?").get(username);
    if (existing) {
      const stmt = db.prepare("UPDATE aliases SET alias = ? WHERE username = ?");
      stmt.run(alias, username);
    } else {
      const stmt = db.prepare("INSERT INTO aliases (username, alias) VALUES (?, ?)");
      stmt.run(username, alias);
    }
    return new Response("Alias set successfully", { status: 200, headers: { "Access-Control-Allow-Origin": "*" } });
  } 
  if (req.method === "GET" && url.pathname === "/alias") {
    
    const stmt = db.prepare("SELECT * FROM aliases");
    const aliases = stmt.all();
    const aliasMap: Record<string, string> = {};
    for (const row of aliases) {
      aliasMap[(row as any).username] = (row as any).alias;
    }
    return new Response(JSON.stringify(aliasMap), { status: 200, headers: { "Content-Type": "application/json", "Access-Control-Allow-Origin": "*" } });
  }
  if (req.method === "POST" && url.pathname === "/add") {
    const { name, watched, type, proposed_by } = await req.json();

    if (
      !name ||
      typeof watched !== "boolean" ||
      !type ||
      !proposed_by

    ) {
      return new Response("Missing required fields", { status: 400, headers: { "Access-Control-Allow-Origin": "*" } });
    }

    const stmt = db.prepare(
      "INSERT INTO movies (name, watched, type, proposed_by, ratings) VALUES (?, ?, ?, ?, ?)",
    );
    stmt.run(name, watched, type, proposed_by, "{}");

    broadcastUpdates(server);
    return new Response("Movie added successfully", { status: 201, headers: { "Access-Control-Allow-Origin": "*" } });
  }

  if (req.method === "GET" && url.pathname === "/movies") {
    const movies = db.query("SELECT * FROM movies").all();
    return new Response(JSON.stringify(movies), {
      headers: { "Content-Type": "application/json", "Access-Control-Allow-Origin": "*" },
    });
  }

  if (req.method === "POST" && url.pathname.startsWith("/movies/rate")) {
    const { movieID, rating, username } = await req.json();
    
    if (!movieID || rating === undefined || !username) {
      return new Response("Missing required fields", { status: 400, headers: { "Access-Control-Allow-Origin": "*" } });
    }
    const movie = db.query("SELECT * FROM movies WHERE id = ?").get(movieID);
    if (!movie) {
      return new Response("Movie not found", { status: 404, headers: { "Access-Control-Allow-Origin": "*" } });
    }
    var ratings;
    console.log((movie as any).ratings);
    if ((movie as any).ratings !== "{}") {
      ratings = JSON.parse((movie as any).ratings);
    } else {
      ratings = {};
    }
  ratings[username] = rating;
    const stmt = db.prepare("UPDATE movies SET ratings = ? WHERE id = ?");
    stmt.run(JSON.stringify(ratings), movieID);
    
    broadcastUpdates(server);
    return new Response("Rating added successfully", { status: 200, headers: { "Access-Control-Allow-Origin": "*" } });

  }
  if (req.method === "POST" && url.pathname === "/queue/add") {
    const { id } = await req.json();

    if (!id) {
      return new Response("Missing movie id", { status: 400, headers: { "Access-Control-Allow-Origin": "*" } });
    }

    const maxQueuePosition = db
      .query("SELECT MAX(queue_position) as max_position FROM movies")
      .get() as { max_position: number | null };
    const newQueuePosition = (maxQueuePosition.max_position || 0) + 1;

    const stmt = db.prepare(
      "UPDATE movies SET queue_position = ? WHERE id = ?",
    );
    stmt.run(newQueuePosition, id);

    broadcastUpdates(server);
    return new Response("Movie added to queue successfully", { status: 200, headers: { "Access-Control-Allow-Origin": "*" } });
  }

  if (req.method === "POST" && url.pathname === "/queue/remove") {
    const movieToRemove = db
      .query(
        "SELECT * FROM movies WHERE queue_position IS NOT NULL ORDER BY queue_position ASC LIMIT 1",
      )
      .get();

    if (!movieToRemove) {
      return new Response("Queue is empty", { status: 400, headers: { "Access-Control-Allow-Origin": "*" } });
    }

    const stmt = db.prepare(
      "UPDATE movies SET queue_position = NULL, watched = true WHERE id = ?",
    );
    
    stmt.run((movieToRemove as any).id);

    broadcastUpdates(server);
    return new Response("Movie removed from queue successfully", {
      status: 200,
      headers: { "Access-Control-Allow-Origin": "*" }
    });
  }

  if (req.method === "GET" && url.pathname === "/queue") {
    const queue = db
      .query(
        "SELECT * FROM movies WHERE queue_position IS NOT NULL ORDER BY queue_position ASC",
      )
      .all();
    return new Response(JSON.stringify(queue), {
      headers: { "Content-Type": "application/json", "Access-Control-Allow-Origin": "*" },
    });
  }

  return new Response("Not Found", { status: 404, headers: { "Access-Control-Allow-Origin": "*" } });
}

export const websocket: WebSocketHandler = {
  open(ws: ServerWebSocket) {
    ws.subscribe("watchalong");
    const { movies, queue } = getMoviesAndQueue();
    ws.send(JSON.stringify({ movies, queue }));
  },
  message(ws, message) {},
  close(ws, code, message) {
    ws.unsubscribe("watchalong");
  },
};

if (process.env.NODE_ENV !== "test") {
  console.log("Database initialized.");
  Bun.serve({
    port: 3000,
    fetch: fetch,
    websocket: websocket,
  });
  console.log("Server listening on http://localhost:3000");
}
