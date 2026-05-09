package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		PRAGMA journal_mode=WAL;
		PRAGMA foreign_keys=ON;

		CREATE TABLE IF NOT EXISTS users (
			id            TEXT PRIMARY KEY,
			username      TEXT NOT NULL UNIQUE COLLATE NOCASE,
			password_hash TEXT NOT NULL,
			is_admin      INTEGER NOT NULL DEFAULT 0,
			created_at    TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS recipes (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL,
			ingredients TEXT NOT NULL DEFAULT '[]',
			steps       TEXT NOT NULL DEFAULT '[]',
			properties  TEXT NOT NULL DEFAULT '{}',
			creator_id  TEXT NOT NULL REFERENCES users(id),
			created_at  TEXT NOT NULL,
			updated_at  TEXT NOT NULL
		);

		CREATE VIRTUAL TABLE IF NOT EXISTS recipes_fts USING fts5(
			recipe_id UNINDEXED,
			search_text,
			tokenize='unicode61'
		);
	`)
	return err
}

// Open returns a RecipeStore and UserStore sharing the same SQLite connection.
func Open(path string) (*RecipeStore, *UserStore, error) {
	db, err := openDB(path)
	if err != nil {
		return nil, nil, err
	}
	return &RecipeStore{db: db}, &UserStore{db: db}, nil
}
