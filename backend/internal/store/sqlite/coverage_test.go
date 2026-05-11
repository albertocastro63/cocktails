package sqlite_test

import (
	"testing"
	"time"

	"github.com/almc/cocktails/internal/model"
	"github.com/google/uuid"
)

func TestDelete_NotFound(t *testing.T) {
	rs, _ := newTestStores(t)
	if err := rs.Delete("nonexistent"); err == nil {
		t.Fatal("expected error deleting nonexistent recipe")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	rs, _ := newTestStores(t)
	r := &model.Recipe{
		ID:          uuid.NewString(),
		Name:        "Ghost",
		Ingredients: []model.Ingredient{},
		Steps:       []string{},
		Properties:  map[string]string{},
		CreatorID:   "u1",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := rs.Update(r); err == nil {
		t.Fatal("expected error updating nonexistent recipe")
	}
}

func TestList_Pagination(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	for i := 0; i < 5; i++ {
		r := sampleRecipe(uid)
		r.ID = uuid.NewString()
		r.Name = "Recipe " + string(rune('A'+i))
		if err := rs.Create(r); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}
	recipes, total, err := rs.List(2, 3)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 5 {
		t.Errorf("total: got %d want 5", total)
	}
	if len(recipes) != 2 {
		t.Errorf("page 2 len: got %d want 2", len(recipes))
	}
}

func TestRandom_WithData(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	if err := rs.Create(sampleRecipe(uid)); err != nil {
		t.Fatalf("Create: %v", err)
	}
	r, err := rs.Random()
	if err != nil {
		t.Fatalf("Random: %v", err)
	}
	if r == nil {
		t.Fatal("expected a recipe from Random, got nil")
	}
}

func TestSearch_NoResults(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	if err := rs.Create(sampleRecipe(uid)); err != nil {
		t.Fatalf("Create: %v", err)
	}
	results, total, err := rs.Search("zzznomatch", 1, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if total != 0 || len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestCreate_WithProperties(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	r := sampleRecipe(uid)
	r.Properties["custom_key"] = "custom_value"
	r.Properties["another"] = "42"
	if err := rs.Create(r); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := rs.GetByID(r.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Properties["custom_key"] != "custom_value" {
		t.Errorf("custom_key: got %q", got.Properties["custom_key"])
	}
}

// T004: notes persisted and returned by GetByID
func TestCreate_WithNotes(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	r := sampleRecipe(uid)
	r.Notes = "Great with aged rum.\nAlso good with mint."
	if err := rs.Create(r); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := rs.GetByID(r.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Notes != r.Notes {
		t.Errorf("Notes: got %q want %q", got.Notes, r.Notes)
	}
}

// T004: search query matching only notes returns no results
func TestSearch_NotesExcludedFromFTS(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	r := sampleRecipe(uid)
	r.Notes = "secretkeyword"
	if err := rs.Create(r); err != nil {
		t.Fatalf("Create: %v", err)
	}
	results, total, err := rs.Search("secretkeyword", 1, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if total != 0 || len(results) != 0 {
		t.Errorf("expected 0 results for notes-only keyword, got %d", len(results))
	}
}

func TestUserCreate_Duplicate(t *testing.T) {
	_, us := newTestStores(t)
	u := &model.User{
		ID:           uuid.NewString(),
		Username:     "alice",
		PasswordHash: "$2a$12$placeholder",
		CreatedAt:    time.Now().UTC(),
	}
	if err := us.Create(u); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	u2 := *u
	u2.ID = uuid.NewString()
	if err := us.Create(&u2); err == nil {
		t.Fatal("expected duplicate error on second Create with same username")
	}
}
