package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// SetRelated sets recipeID's related set and reconciles the symmetric reverse
// relation on each counterpart. Normalizes the request (dedupe, drop self, drop
// non-existent), diffs against the current set, and writes both sides.
func (s *RecipeStore) SetRelated(recipeID string, requested []string) error {
	old, err := s.relatedIDs(recipeID)
	if err != nil {
		return err
	}
	next, err := s.normalizeRelated(recipeID, requested)
	if err != nil {
		return err
	}
	if err := s.writeRelatedIDs(recipeID, next); err != nil {
		return err
	}
	// Counterparts: add recipeID to newly related, remove from no-longer related.
	for _, b := range subtract(next, old) {
		if err := s.addCounterpart(b, recipeID); err != nil {
			return err
		}
	}
	for _, b := range subtract(old, next) {
		if err := s.removeCounterpart(b, recipeID); err != nil {
			return err
		}
	}
	return nil
}

// relatedIDs returns the stored related set for a recipe (error if not found).
func (s *RecipeStore) relatedIDs(id string) ([]string, error) {
	var raw []byte
	err := s.db.QueryRow(`SELECT related_ids FROM recipes WHERE id = ?`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("recipe %q not found", id)
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &ids); err != nil {
			return nil, err
		}
	}
	return ids, nil
}

func (s *RecipeStore) writeRelatedIDs(id string, ids []string) error {
	if ids == nil {
		ids = []string{}
	}
	raw, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE recipes SET related_ids = ? WHERE id = ?`, raw, id)
	return err
}

// normalizeRelated dedupes, drops self, and drops ids that don't resolve to an
// existing recipe.
func (s *RecipeStore) normalizeRelated(self string, requested []string) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	for _, id := range requested {
		if id == "" || id == self || seen[id] {
			continue
		}
		var n int
		if err := s.db.QueryRow(`SELECT COUNT(*) FROM recipes WHERE id = ?`, id).Scan(&n); err != nil {
			return nil, err
		}
		if n == 0 {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out, nil
}

func (s *RecipeStore) addCounterpart(id, other string) error {
	ids, err := s.relatedIDs(id)
	if err != nil {
		return err
	}
	for _, x := range ids {
		if x == other {
			return nil // already present
		}
	}
	return s.writeRelatedIDs(id, append(ids, other))
}

func (s *RecipeStore) removeCounterpart(id, other string) error {
	ids, err := s.relatedIDs(id)
	if err != nil {
		return err
	}
	out := ids[:0:0]
	for _, x := range ids {
		if x != other {
			out = append(out, x)
		}
	}
	return s.writeRelatedIDs(id, out)
}

// subtract returns items in a that are not in b.
func subtract(a, b []string) []string {
	inB := map[string]bool{}
	for _, x := range b {
		inB[x] = true
	}
	var out []string
	for _, x := range a {
		if !inB[x] {
			out = append(out, x)
		}
	}
	return out
}
