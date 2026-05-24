package sqlite

import (
	"database/sql"
	"time"

	"github.com/almc/cocktails/internal/model"
)

// FavoriteStore implements store.FavoriteStore backed by SQLite.
type FavoriteStore struct{ db *sql.DB }

func NewFavoriteStore(db *sql.DB) *FavoriteStore {
	return &FavoriteStore{db: db}
}

func (s *FavoriteStore) Add(userID, recipeID string) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO favorites (user_id, recipe_id, created_at) VALUES (?, ?, ?)`,
		userID, recipeID, time.Now().UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *FavoriteStore) Remove(userID, recipeID string) error {
	_, err := s.db.Exec(
		`DELETE FROM favorites WHERE user_id = ? AND recipe_id = ?`,
		userID, recipeID,
	)
	return err
}

func (s *FavoriteStore) IsFavorite(userID, recipeID string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM favorites WHERE user_id = ? AND recipe_id = ?`,
		userID, recipeID,
	).Scan(&count)
	return count > 0, err
}

func (s *FavoriteStore) ListByUser(userID string) ([]*model.Favorite, error) {
	rows, err := s.db.Query(
		`SELECT user_id, recipe_id, created_at FROM favorites WHERE user_id = ? ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favs []*model.Favorite
	for rows.Next() {
		var f model.Favorite
		var createdAt string
		if err := rows.Scan(&f.UserID, &f.RecipeID, &createdAt); err != nil {
			return nil, err
		}
		f.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		favs = append(favs, &f)
	}
	if favs == nil {
		favs = []*model.Favorite{}
	}
	return favs, rows.Err()
}

// P3: replace with Query on recipe_id GSI when favorite count UI is implemented
func (s *FavoriteStore) CountByRecipe(recipeID string) (int, error) {
	return 0, nil
}
