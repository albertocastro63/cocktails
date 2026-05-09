package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/almc/cocktails/internal/model"
	"github.com/almc/cocktails/internal/store"
)

type UserStore struct{ db *sql.DB }

func (s *UserStore) Create(u *model.User) error {
	_, err := s.db.Exec(
		`INSERT INTO users (id, username, password_hash, is_admin, created_at) VALUES (?, ?, ?, ?, ?)`,
		u.ID, u.Username, u.PasswordHash, boolToInt(u.IsAdmin),
		u.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint") {
		return store.ErrDuplicate
	}
	return err
}

func (s *UserStore) GetByUsername(username string) (*model.User, error) {
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, is_admin, created_at FROM users WHERE username = ? COLLATE NOCASE`,
		username,
	)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("user %q not found", username)
	}
	return u, err
}

func (s *UserStore) GetByID(id string) (*model.User, error) {
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, is_admin, created_at FROM users WHERE id = ?`, id,
	)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("user %q not found", id)
	}
	return u, err
}

func (s *UserStore) Count() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

func scanUser(row *sql.Row) (*model.User, error) {
	var u model.User
	var isAdmin int
	var createdAt string
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &isAdmin, &createdAt)
	if err != nil {
		return nil, err
	}
	u.IsAdmin = isAdmin == 1
	u.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return &u, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
