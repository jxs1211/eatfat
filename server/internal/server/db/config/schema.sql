CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS players (
    -- A unique identifier for the player
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    -- A reference to the user who owns this player
    user_id INTEGER NOT NULL,
    -- The name of the player
    name TEXT NOT NULL UNIQUE,
    -- The player’s best score of all time
    best_score INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id)
);