package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/almc/cocktails/internal/model"
)

type RecipeStore struct{ db *sql.DB }

func (s *RecipeStore) Create(r *model.Recipe) error {
	ing, err := json.Marshal(r.Ingredients)
	if err != nil {
		return err
	}
	steps, err := json.Marshal(r.Steps)
	if err != nil {
		return err
	}
	props, err := json.Marshal(r.Properties)
	if err != nil {
		return err
	}
	creatorID := sql.NullString{String: r.CreatorID, Valid: r.CreatorID != ""}
	_, err = s.db.Exec(
		`INSERT INTO recipes (id, name, ingredients, steps, properties, notes, creator_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.Name, ing, steps, props, r.Notes, creatorID,
		r.CreatedAt.UTC().Format(time.RFC3339Nano),
		r.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return err
	}
	return s.upsertFTS(r)
}

func (s *RecipeStore) GetByID(id string) (*model.Recipe, error) {
	row := s.db.QueryRow(
		`SELECT id, name, ingredients, steps, properties, notes, creator_id, created_at, updated_at
		 FROM recipes WHERE id = ?`, id)
	r, err := scanRecipe(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("recipe %q not found", id)
	}
	return r, err
}

func (s *RecipeStore) List(page, limit int) ([]*model.Recipe, int, error) {
	offset := (page - 1) * limit
	var total int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM recipes`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.Query(
		`SELECT id, name, ingredients, steps, properties, notes, creator_id, created_at, updated_at
		 FROM recipes ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return scanRecipes(rows, total)
}

func (s *RecipeStore) Search(query string, page, limit int) ([]*model.Recipe, int, error) {
	if strings.TrimSpace(query) == "" {
		return s.List(page, limit)
	}
	offset := (page - 1) * limit
	var total int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM recipes_fts WHERE search_text MATCH ?`, query+"*",
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.Query(
		`SELECT r.id, r.name, r.ingredients, r.steps, r.properties, r.notes, r.creator_id, r.created_at, r.updated_at
		 FROM recipes r
		 JOIN recipes_fts fts ON fts.recipe_id = r.id
		 WHERE fts.search_text MATCH ?
		 ORDER BY rank
		 LIMIT ? OFFSET ?`, query+"*", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return scanRecipes(rows, total)
}

func (s *RecipeStore) Random() (*model.Recipe, error) {
	row := s.db.QueryRow(
		`SELECT id, name, ingredients, steps, properties, notes, creator_id, created_at, updated_at
		 FROM recipes ORDER BY RANDOM() LIMIT 1`)
	r, err := scanRecipe(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return r, err
}

func (s *RecipeStore) Update(r *model.Recipe) error {
	ing, err := json.Marshal(r.Ingredients)
	if err != nil {
		return err
	}
	steps, err := json.Marshal(r.Steps)
	if err != nil {
		return err
	}
	props, err := json.Marshal(r.Properties)
	if err != nil {
		return err
	}
	r.UpdatedAt = time.Now().UTC()
	res, err := s.db.Exec(
		`UPDATE recipes SET name=?, ingredients=?, steps=?, properties=?, notes=?, updated_at=? WHERE id=?`,
		r.Name, ing, steps, props, r.Notes, r.UpdatedAt.UTC().Format(time.RFC3339Nano), r.ID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("recipe %q not found", r.ID)
	}
	return s.upsertFTS(r)
}

func (s *RecipeStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM recipes_fts WHERE recipe_id = ?`, id)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`DELETE FROM recipes WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("recipe %q not found", id)
	}
	return nil
}

func (s *RecipeStore) ListByCreator(creatorID string, page, limit int) ([]*model.Recipe, int, error) {
	offset := (page - 1) * limit
	var total int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM recipes WHERE creator_id = ? AND creator_id != ''`, creatorID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.Query(
		`SELECT id, name, ingredients, steps, properties, notes, creator_id, created_at, updated_at
		 FROM recipes WHERE creator_id = ? AND creator_id != ''
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`, creatorID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return scanRecipes(rows, total)
}

func (s *RecipeStore) ExistsByName(name string) (bool, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM recipes WHERE name = ?`, name).Scan(&n)
	return n > 0, err
}

func (s *RecipeStore) ListAll() ([]*model.Recipe, error) {
	rows, err := s.db.Query(
		`SELECT id, name, ingredients, steps, properties, notes, creator_id, created_at, updated_at
		 FROM recipes ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	recipes, _, err := scanRecipes(rows, 0)
	return recipes, err
}

func (s *RecipeStore) ImportBatch(recipes []*model.Recipe, _ string) (int, int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback() //nolint:errcheck

	created, skipped := 0, 0
	for _, r := range recipes {
		var n int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM recipes WHERE name = ?`, r.Name).Scan(&n); err != nil {
			return 0, 0, err
		}
		if n > 0 {
			skipped++
			continue
		}
		ing, _ := json.Marshal(r.Ingredients)
		steps, _ := json.Marshal(r.Steps)
		props, _ := json.Marshal(r.Properties)
		cid := sql.NullString{String: r.CreatorID, Valid: r.CreatorID != ""}
		if _, err := tx.Exec(
			`INSERT INTO recipes (id, name, ingredients, steps, properties, notes, creator_id, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			r.ID, r.Name, ing, steps, props, r.Notes, cid,
			r.CreatedAt.UTC().Format(time.RFC3339Nano),
			r.UpdatedAt.UTC().Format(time.RFC3339Nano),
		); err != nil {
			return 0, 0, err
		}
		created++
	}
	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}
	return created, skipped, nil
}

func (s *RecipeStore) SearchByIngredients(ingredients []string, page, limit int) ([]*model.Recipe, int, error) {
	if len(ingredients) == 0 {
		return s.List(page, limit)
	}
	var clauses []string
	var countArgs []any
	for _, ing := range ingredients {
		clauses = append(clauses, `(SELECT COUNT(*) FROM json_each(r.ingredients) WHERE LOWER(json_extract(value,'$.name')) LIKE ?) > 0`)
		countArgs = append(countArgs, "%"+strings.ToLower(ing)+"%")
	}
	where := strings.Join(clauses, " AND ")

	var total int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM recipes r WHERE `+where, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	selectArgs := append(countArgs, limit, offset)
	rows, err := s.db.Query(
		`SELECT r.id, r.name, r.ingredients, r.steps, r.properties, r.notes, r.creator_id, r.created_at, r.updated_at
		 FROM recipes r WHERE `+where+` ORDER BY r.created_at DESC LIMIT ? OFFSET ?`,
		selectArgs...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return scanRecipes(rows, total)
}

func (s *RecipeStore) upsertFTS(r *model.Recipe) error {
	var parts []string
	parts = append(parts, r.Name)
	for _, ing := range r.Ingredients {
		parts = append(parts, ing.Name)
	}
	parts = append(parts, r.Steps...)
	for _, v := range r.Properties {
		parts = append(parts, v)
	}
	searchText := strings.Join(parts, " ")
	_, err := s.db.Exec(`DELETE FROM recipes_fts WHERE recipe_id = ?`, r.ID)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO recipes_fts (recipe_id, search_text) VALUES (?, ?)`, r.ID, searchText)
	return err
}

func scanRecipe(row *sql.Row) (*model.Recipe, error) {
	var r model.Recipe
	var ing, steps, props []byte
	var creatorID sql.NullString
	var createdAt, updatedAt string
	err := row.Scan(&r.ID, &r.Name, &ing, &steps, &props, &r.Notes, &creatorID, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	r.CreatorID = creatorID.String
	if err := json.Unmarshal(ing, &r.Ingredients); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(steps, &r.Steps); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(props, &r.Properties); err != nil {
		return nil, err
	}
	r.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	r.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return &r, nil
}

func scanRecipes(rows *sql.Rows, total int) ([]*model.Recipe, int, error) {
	var results []*model.Recipe
	for rows.Next() {
		var r model.Recipe
		var ing, steps, props []byte
		var creatorID sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(&r.ID, &r.Name, &ing, &steps, &props, &r.Notes, &creatorID, &createdAt, &updatedAt); err != nil {
			return nil, 0, err
		}
		r.CreatorID = creatorID.String
		if err := json.Unmarshal(ing, &r.Ingredients); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(steps, &r.Steps); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(props, &r.Properties); err != nil {
			return nil, 0, err
		}
		r.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		r.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
		results = append(results, &r)
	}
	return results, total, rows.Err()
}
