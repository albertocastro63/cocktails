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
		`INSERT INTO users (id, username, password_hash, is_admin, first_name, last_name, email, token_version, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Username, u.PasswordHash, boolToInt(u.IsAdmin),
		u.FirstName, u.LastName, u.Email, u.TokenVersion,
		u.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint") {
		return store.ErrDuplicate
	}
	return err
}

func (s *UserStore) GetByUsername(username string) (*model.User, error) {
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, is_admin, first_name, last_name, email, token_version, created_at
		 FROM users WHERE username = ? COLLATE NOCASE`,
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
		`SELECT id, username, password_hash, is_admin, first_name, last_name, email, token_version, created_at
		 FROM users WHERE id = ?`, id,
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

func (s *UserStore) List() ([]*model.User, error) {
	rows, err := s.db.Query(
		`SELECT id, username, password_hash, is_admin, first_name, last_name, email, token_version, created_at
		 FROM users WHERE is_admin = 0 ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]*model.User, 0)
	for rows.Next() {
		var u model.User
		var isAdmin int
		var createdAt string
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &isAdmin, &u.FirstName, &u.LastName, &u.Email, &u.TokenVersion, &createdAt); err != nil {
			return nil, err
		}
		u.IsAdmin = isAdmin == 1
		u.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		users = append(users, &u)
	}
	return users, rows.Err()
}

func (s *UserStore) Update(u *model.User) error {
	res, err := s.db.Exec(
		`UPDATE users SET first_name=?, last_name=?, email=?, password_hash=?, token_version=? WHERE id=?`,
		u.FirstName, u.LastName, u.Email, u.PasswordHash, u.TokenVersion, u.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return store.ErrDuplicate
		}
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user %q not found", u.ID)
	}
	return nil
}

func (s *UserStore) Delete(id string) error {
	res, err := s.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user %q not found", id)
	}
	return nil
}

func (s *UserStore) GetByEmail(email string) (*model.User, error) {
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, is_admin, first_name, last_name, email, token_version, created_at
		 FROM users WHERE email = ?`,
		email,
	)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("user with email %q not found", email)
	}
	return u, err
}

func scanUser(row *sql.Row) (*model.User, error) {
	var u model.User
	var isAdmin int
	var createdAt string
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &isAdmin, &u.FirstName, &u.LastName, &u.Email, &u.TokenVersion, &createdAt)
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
