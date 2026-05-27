package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

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
			creator_id  TEXT REFERENCES users(id) ON DELETE SET NULL,
			created_at  TEXT NOT NULL,
			updated_at  TEXT NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	// Idempotent: add notes column to existing databases.
	_, err = db.Exec(`ALTER TABLE recipes ADD COLUMN notes TEXT NOT NULL DEFAULT ''`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}

	// Idempotent: add garnishes column to existing databases.
	_, err = db.Exec(`ALTER TABLE recipes ADD COLUMN garnishes TEXT NOT NULL DEFAULT '[]'`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}

	_, err = db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS recipes_fts USING fts5(
			recipe_id UNINDEXED,
			search_text,
			tokenize='unicode61'
		);
	`)
	if err != nil {
		return err
	}

	// Idempotent: add user profile and token_version columns.
	for _, stmt := range []string{
		`ALTER TABLE users ADD COLUMN first_name TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN last_name TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN email TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE users ADD COLUMN token_version INTEGER NOT NULL DEFAULT 0`,
	} {
		if _, err = db.Exec(stmt); err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}
	}

	// Idempotent: partial unique index on non-empty emails.
	_, err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique ON users(email) WHERE email != ''`)
	if err != nil {
		return err
	}

	// Idempotent: favorites table for user-saved recipes.
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS favorites (
			user_id    TEXT NOT NULL,
			recipe_id  TEXT NOT NULL,
			created_at TEXT NOT NULL,
			PRIMARY KEY (user_id, recipe_id)
		)
	`)
	if err != nil {
		return err
	}

	// Make recipes.creator_id nullable if the existing table has it as NOT NULL.
	var notnull int
	if err = db.QueryRow(`SELECT "notnull" FROM pragma_table_info('recipes') WHERE name='creator_id'`).Scan(&notnull); err != nil {
		return err
	}
	if notnull == 1 {
		if err = rebuildRecipesTable(db); err != nil {
			return err
		}
	}

	return nil
}

func rebuildRecipesTable(db *sql.DB) error {
	tx, txErr := db.Begin()
	if txErr != nil {
		return txErr
	}
	for _, stmt := range []string{
		`CREATE TABLE recipes_new (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL,
			ingredients TEXT NOT NULL DEFAULT '[]',
			steps       TEXT NOT NULL DEFAULT '[]',
			properties  TEXT NOT NULL DEFAULT '{}',
			notes       TEXT NOT NULL DEFAULT '',
			creator_id  TEXT REFERENCES users(id) ON DELETE SET NULL,
			created_at  TEXT NOT NULL,
			updated_at  TEXT NOT NULL
		)`,
		`INSERT INTO recipes_new (id, name, ingredients, steps, properties, notes, creator_id, created_at, updated_at)
		 SELECT id, name, ingredients, steps, properties, notes, creator_id, created_at, updated_at FROM recipes`,
		`DROP TABLE recipes`,
		`ALTER TABLE recipes_new RENAME TO recipes`,
	} {
		if _, err := tx.Exec(stmt); err != nil {
			tx.Rollback() //nolint:errcheck
			return err
		}
	}
	return tx.Commit()
}

// Open returns a RecipeStore and UserStore sharing the same SQLite connection.
func Open(path string) (*RecipeStore, *UserStore, error) {
	db, err := openDB(path)
	if err != nil {
		return nil, nil, err
	}
	return &RecipeStore{db: db}, &UserStore{db: db}, nil
}

// OpenAll returns all three stores sharing the same SQLite connection.
func OpenAll(path string) (*RecipeStore, *UserStore, *FavoriteStore, error) {
	db, err := openDB(path)
	if err != nil {
		return nil, nil, nil, err
	}
	return &RecipeStore{db: db}, &UserStore{db: db}, &FavoriteStore{db: db}, nil
}
