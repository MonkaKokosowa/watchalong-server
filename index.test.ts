import { test, expect, describe, beforeEach } from "bun:test";
import { Server } from "bun";

// Mock the server to avoid listening on a real port during tests
const mockServer = {
  publish: (topic: string, data: string) => {
    // This is where we can check if the broadcast happened
  },
  upgrade: () => false,
} as unknown as Server;

// Import the server logic
import { fetch, websocket, db } from "./index";

describe("Watchalong Server", () => {
  beforeEach(() => {
    db.run("DELETE FROM movies");
  });

  test("POST /add - should add a movie", async () => {
    const movie = {
      name: "Test Movie",
      watched: false,
      type: "Movie",
      proposed_by: "Tester",
      ratings: { user: 5 },
    };

    const req = new Request("http://localhost:3000/add", {
      method: "POST",
      body: JSON.stringify(movie),
    });

    const res = await fetch(req, mockServer);
    expect(res).toBeInstanceOf(Response);
    if (res) {
      expect(res.status).toBe(201);
    }

    const movies = db.query("SELECT * FROM movies").all();
    expect(movies.length).toBe(1);
    expect(movies[0].name).toBe("Test Movie");
  });

  test("GET /movies - should return all movies", async () => {
    db.run(
      "INSERT INTO movies (name, watched, type, proposed_by, ratings) VALUES ('Test Movie', 0, 'Movie', 'Tester', '{}')",
    );
    const req = new Request("http://localhost:3000/movies");
    const res = await fetch(req, mockServer);
    expect(res).toBeInstanceOf(Response);
    if (res) {
      expect(res.status).toBe(200);
      const body = await res.json();
      expect(body.length).toBe(1);
      expect(body[0].name).toBe("Test Movie");
    }
  });

  test("POST /queue/add - should add a movie to the queue", async () => {
    const insertedMovie = db
      .query(
        "INSERT INTO movies (name, watched, type, proposed_by, ratings) VALUES ('Test Movie', 0, 'Movie', 'Tester', '{}') RETURNING id",
      )
      .get();
    const req = new Request("http://localhost:3000/queue/add", {
      method: "POST",
      body: JSON.stringify({ id: insertedMovie.id }),
    });

    const res = await fetch(req, mockServer);
    expect(res).toBeInstanceOf(Response);
    if (res) {
      expect(res.status).toBe(200);
    }

    const movie = db
      .query(`SELECT * FROM movies WHERE id = ${insertedMovie.id}`)
      .get();
    expect(movie).not.toBeNull();
    if (movie) {
      expect(movie.queue_position).toBe(1);
    }
  });

  test("GET /queue - should return the queue", async () => {
    db.run(
      "INSERT INTO movies (name, watched, type, proposed_by, ratings, queue_position) VALUES ('Test Movie', 0, 'Movie', 'Tester', '{}', 1)",
    );
    const req = new Request("http://localhost:3000/queue");
    const res = await fetch(req, mockServer);
    expect(res).toBeInstanceOf(Response);
    if (res) {
      expect(res.status).toBe(200);
      const body = await res.json();
      expect(body.length).toBe(1);
      expect(body[0].name).toBe("Test Movie");
    }
  });

  test("POST /queue/remove - should remove a movie from the queue", async () => {
    const insertedMovie = db
      .query(
        "INSERT INTO movies (name, watched, type, proposed_by, ratings, queue_position) VALUES ('Test Movie', 0, 'Movie', 'Tester', '{}', 1) RETURNING id",
      )
      .get();
    const req = new Request("http://localhost:3000/queue/remove", {
      method: "POST",
    });

    const res = await fetch(req, mockServer);
    expect(res).toBeInstanceOf(Response);
    if (res) {
      expect(res.status).toBe(200);
    }

    const movie = db
      .query(`SELECT * FROM movies WHERE id = ${insertedMovie.id}`)
      .get();
    expect(movie).not.toBeNull();
    if (movie) {
      expect(movie.queue_position).toBeNull();
    }
  });

  test("WebSocket - should send initial data on open", () => {
    db.run(
      "INSERT INTO movies (name, watched, type, proposed_by, ratings) VALUES ('Test Movie', 0, 'Movie', 'Tester', '{}')",
    );
    const ws = {
      subscribe: () => {},
      send: (data: string) => {
        const { movies, queue } = JSON.parse(data);
        expect(movies.length).toBe(1);
        expect(queue.length).toBe(0);
      },
      unsubscribe: () => {},
    } as any;

    websocket.open(ws);
  });

  test("WebSocket - should broadcast update on add", async () => {
    const promise = new Promise((resolve) => {
      mockServer.publish = (topic: string, data: string) => {
        const { movies, queue } = JSON.parse(data);
        expect(topic).toBe("watchalong");
        expect(movies.length).toBe(1);
        expect(movies[0].name).toBe("Test Movie");
        resolve(undefined);
      };
    });

    const movie = {
      name: "Test Movie",
      watched: false,
      type: "Movie",
      proposed_by: "Tester",
      ratings: { user: 5 },
    };

    const req = new Request("http://localhost:3000/add", {
      method: "POST",
      body: JSON.stringify(movie),
    });

    await fetch(req, mockServer);

    await promise;
  });
});
