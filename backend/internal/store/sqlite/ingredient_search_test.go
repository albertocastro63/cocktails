package sqlite_test

import (
	"testing"
	"time"

	"github.com/almc/cocktails/internal/model"
	sqstore "github.com/almc/cocktails/internal/store/sqlite"
	"github.com/google/uuid"
)

func newTestStoreForIngredients(t *testing.T) *sqstore.RecipeStore {
	t.Helper()
	rs, _, err := sqstore.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return rs
}

func makeRecipe(name string, ingredients ...string) *model.Recipe {
	var ings []model.Ingredient
	for _, n := range ingredients {
		ings = append(ings, model.Ingredient{Name: n, Quantity: "30", Unit: "ml"})
	}
	return &model.Recipe{
		ID:          uuid.NewString(),
		Name:        name,
		Ingredients: ings,
		Steps:       []string{"mix"},
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
}

func TestSearchByIngredients_TwoIngredients(t *testing.T) {
	rs := newTestStoreForIngredients(t)
	if err := rs.Create(makeRecipe("Gin Fizz", "Gin", "Lemon Juice", "Sugar", "Soda Water")); err != nil {
		t.Fatal(err)
	}
	if err := rs.Create(makeRecipe("Daiquiri", "Rum", "Lime Juice", "Sugar")); err != nil {
		t.Fatal(err)
	}
	if err := rs.Create(makeRecipe("Bee's Knees", "Gin", "Lemon Juice", "Honey")); err != nil {
		t.Fatal(err)
	}

	got, total, err := rs.SearchByIngredients([]string{"gin", "lemon juice"}, 1, 20)
	if err != nil {
		t.Fatalf("SearchByIngredients: %v", err)
	}
	if total != 2 {
		t.Errorf("total: got %d want 2", total)
	}
	if len(got) != 2 {
		t.Errorf("len: got %d want 2", len(got))
	}
	for _, r := range got {
		if r.Name == "Daiquiri" {
			t.Error("Daiquiri should not be in results (no Gin)")
		}
	}
}

func TestSearchByIngredients_ThreeIngredients(t *testing.T) {
	rs := newTestStoreForIngredients(t)
	if err := rs.Create(makeRecipe("Gin Fizz", "Gin", "Lemon Juice", "Sugar", "Soda Water")); err != nil {
		t.Fatal(err)
	}
	if err := rs.Create(makeRecipe("Bee's Knees", "Gin", "Lemon Juice", "Honey")); err != nil {
		t.Fatal(err)
	}

	got, total, err := rs.SearchByIngredients([]string{"gin", "lemon juice", "sugar"}, 1, 20)
	if err != nil {
		t.Fatalf("SearchByIngredients: %v", err)
	}
	if total != 1 {
		t.Errorf("total: got %d want 1 (only Gin Fizz has all three)", total)
	}
	if len(got) != 1 || got[0].Name != "Gin Fizz" {
		t.Errorf("expected only Gin Fizz, got %v", got)
	}
}

func TestSearchByIngredients_NoMatch(t *testing.T) {
	rs := newTestStoreForIngredients(t)
	if err := rs.Create(makeRecipe("Daiquiri", "Rum", "Lime Juice", "Sugar")); err != nil {
		t.Fatal(err)
	}

	got, total, err := rs.SearchByIngredients([]string{"gin", "rum"}, 1, 20)
	if err != nil {
		t.Fatalf("SearchByIngredients: %v", err)
	}
	if total != 0 || len(got) != 0 {
		t.Errorf("expected no results, got %d", len(got))
	}
}

func TestSearchByIngredients_CaseInsensitive(t *testing.T) {
	rs := newTestStoreForIngredients(t)
	if err := rs.Create(makeRecipe("Gin Sour", "Gin", "Lemon Juice")); err != nil {
		t.Fatal(err)
	}

	got, _, err := rs.SearchByIngredients([]string{"GIN", "LEMON JUICE"}, 1, 20)
	if err != nil {
		t.Fatalf("SearchByIngredients: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 result for uppercase tokens, got %d", len(got))
	}
}

func TestSearchByIngredients_SubstringMatch(t *testing.T) {
	rs := newTestStoreForIngredients(t)
	if err := rs.Create(makeRecipe("Sloe Gin Fizz", "Sloe Gin", "Lemon Juice", "Soda")); err != nil {
		t.Fatal(err)
	}

	// "gin" should match "Sloe Gin" as a substring
	got, _, err := rs.SearchByIngredients([]string{"gin", "lemon juice"}, 1, 20)
	if err != nil {
		t.Fatalf("SearchByIngredients: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 result (substring match on Sloe Gin), got %d", len(got))
	}
}

func TestSearchByIngredients_Pagination(t *testing.T) {
	rs := newTestStoreForIngredients(t)
	for i := 0; i < 5; i++ {
		if err := rs.Create(makeRecipe(uuid.NewString(), "Gin", "Lemon Juice")); err != nil {
			t.Fatal(err)
		}
	}

	page1, total, err := rs.SearchByIngredients([]string{"gin", "lemon juice"}, 1, 2)
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if total != 5 {
		t.Errorf("total: got %d want 5", total)
	}
	if len(page1) != 2 {
		t.Errorf("page 1 len: got %d want 2", len(page1))
	}

	page2, _, err := rs.SearchByIngredients([]string{"gin", "lemon juice"}, 2, 2)
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}
	if len(page2) != 2 {
		t.Errorf("page 2 len: got %d want 2", len(page2))
	}
	// pages must not overlap
	if page1[0].ID == page2[0].ID {
		t.Error("page 1 and page 2 share same first item")
	}
}

func TestSearchByIngredients_EmptySlice(t *testing.T) {
	rs := newTestStoreForIngredients(t)
	if err := rs.Create(makeRecipe("Mojito", "Rum", "Lime")); err != nil {
		t.Fatal(err)
	}

	// empty slice → falls back to List
	got, total, err := rs.SearchByIngredients([]string{}, 1, 20)
	if err != nil {
		t.Fatalf("SearchByIngredients empty: %v", err)
	}
	if total != 1 || len(got) != 1 {
		t.Errorf("expected 1 recipe from fallback List, got %d", len(got))
	}
}
