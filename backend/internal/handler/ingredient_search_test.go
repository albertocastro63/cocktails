package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/almc/cocktails/internal/handler"
	"github.com/almc/cocktails/internal/model"
)

// recipes used across ingredient search tests
var (
	recipeGinFizz  = &model.Recipe{ID: "ginFizz", Name: "Gin Fizz", Ingredients: []model.Ingredient{{Name: "Gin"}, {Name: "Lemon Juice"}, {Name: "Sugar"}}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	recipeBeesKnee = &model.Recipe{ID: "beesKnee", Name: "Bee's Knees", Ingredients: []model.Ingredient{{Name: "Gin"}, {Name: "Lemon Juice"}, {Name: "Honey"}}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	recipeDaiquiri = &model.Recipe{ID: "daiquiri", Name: "Daiquiri", Ingredients: []model.Ingredient{{Name: "Rum"}, {Name: "Lime Juice"}, {Name: "Sugar"}}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
)

func doIngSearch(t *testing.T, query string, recipes ...*model.Recipe) (int, []any) {
	t.Helper()
	rs := newStubRecipeStore(recipes...)
	h := handler.NewRecipeHandler(rs)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes?"+query, nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	data := resp["data"].([]any)
	total := int(resp["total"].(float64))
	return total, data
}

// T004a: "gin and lemon juice" → SearchByIngredients → only gin+lemon recipes
func TestIngredientSearch_NaturalLanguage_And(t *testing.T) {
	total, data := doIngSearch(t, "q=gin+and+lemon+juice",
		recipeGinFizz, recipeBeesKnee, recipeDaiquiri)
	if total != 2 || len(data) != 2 {
		t.Errorf("expected 2 results, got total=%d len=%d", total, len(data))
	}
	for _, item := range data {
		m := item.(map[string]any)
		if m["name"] == "Daiquiri" {
			t.Error("Daiquiri should not match gin+lemon juice search")
		}
	}
}

// T004b: "gin+lemon+juice" (no spaces around +) same result
func TestIngredientSearch_PlusNoSpaces(t *testing.T) {
	total, data := doIngSearch(t, "q=gin%2Blemon+juice",
		recipeGinFizz, recipeBeesKnee, recipeDaiquiri)
	if total != 2 || len(data) != 2 {
		t.Errorf("expected 2 results, got total=%d len=%d", total, len(data))
	}
}

// T004c: "gin  +  lemon juice" (extra whitespace around +)
func TestIngredientSearch_PlusWithSpaces(t *testing.T) {
	total, data := doIngSearch(t, "q=gin+%2B+lemon+juice",
		recipeGinFizz, recipeBeesKnee, recipeDaiquiri)
	if total != 2 || len(data) != 2 {
		t.Errorf("expected 2 results, got total=%d len=%d", total, len(data))
	}
}

// T004d: single term "gin" stays on existing Search path
func TestIngredientSearch_SingleTerm_FallsToSearch(t *testing.T) {
	rs := newStubRecipeStore(recipeGinFizz, recipeBeesKnee, recipeDaiquiri)
	h := handler.NewRecipeHandler(rs)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes?q=gin", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d", rec.Code)
	}
	// stub Search matches on name; "gin" → Gin Fizz + Bee's Knees (name contains "gin")
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp["data"].([]any)
	if len(data) == 0 {
		t.Error("single-term search should return results via existing Search path")
	}
	// should NOT call SearchByIngredients — stub would return all 3 if it did
	if len(data) == 3 {
		t.Error("single term should use Search (name match), not SearchByIngredients")
	}
}

// T004e: empty ?q= → returns full list
func TestIngredientSearch_EmptyQuery_ReturnsList(t *testing.T) {
	total, data := doIngSearch(t, "q=",
		recipeGinFizz, recipeBeesKnee, recipeDaiquiri)
	if total != 3 || len(data) != 3 {
		t.Errorf("empty query should return all 3, got total=%d len=%d", total, len(data))
	}
}

// T004f: "gin + + lemon juice" (empty middle token) treated as 2-token search
func TestIngredientSearch_EmptyMiddleToken_Ignored(t *testing.T) {
	total, data := doIngSearch(t, "q=gin+%2B++%2B+lemon+juice",
		recipeGinFizz, recipeBeesKnee, recipeDaiquiri)
	if total != 2 || len(data) != 2 {
		t.Errorf("empty token should be discarded, expected 2, got total=%d len=%d", total, len(data))
	}
}

// T004g: pagination ?page=2 returns correct offset (FR-007)
func TestIngredientSearch_Pagination(t *testing.T) {
	// Need 3 gin+lemon recipes to test page 2
	r1 := &model.Recipe{ID: "r1", Name: "A", Ingredients: []model.Ingredient{{Name: "Gin"}, {Name: "Lemon Juice"}}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	r2 := &model.Recipe{ID: "r2", Name: "B", Ingredients: []model.Ingredient{{Name: "Gin"}, {Name: "Lemon Juice"}}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	r3 := &model.Recipe{ID: "r3", Name: "C", Ingredients: []model.Ingredient{{Name: "Gin"}, {Name: "Lemon Juice"}}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}

	rs := newStubRecipeStore(r1, r2, r3)
	h := handler.NewRecipeHandler(rs)

	// page 1 limit 2
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes?q=gin+and+lemon+juice&page=1&limit=2", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)
	var resp1 map[string]any
	json.NewDecoder(rec.Body).Decode(&resp1)
	if int(resp1["total"].(float64)) != 3 {
		t.Errorf("total should be 3, got %v", resp1["total"])
	}
	data1 := resp1["data"].([]any)
	if len(data1) != 2 {
		t.Errorf("page 1 len: got %d want 2", len(data1))
	}

	// page 2 limit 2 → 1 remaining
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/recipes?q=gin+and+lemon+juice&page=2&limit=2", nil)
	rec2 := httptest.NewRecorder()
	h.List(rec2, req2)
	var resp2 map[string]any
	json.NewDecoder(rec2.Body).Decode(&resp2)
	data2 := resp2["data"].([]any)
	if len(data2) != 1 {
		t.Errorf("page 2 len: got %d want 1", len(data2))
	}
	// pages must not overlap
	name1 := data1[0].(map[string]any)["name"]
	name2 := data2[0].(map[string]any)["name"]
	if name1 == name2 {
		t.Error("page 1 and page 2 share same item")
	}
}

// parseIngredientQuery is tested indirectly via handler — also test edge cases
func TestIngredientSearch_UppercaseAnd(t *testing.T) {
	total, data := doIngSearch(t, "q=Gin+AND+Lemon+Juice",
		recipeGinFizz, recipeBeesKnee, recipeDaiquiri)
	if total != 2 || len(data) != 2 {
		t.Errorf("uppercase AND should parse the same, got total=%d len=%d", total, len(data))
	}
}

// Test that stub's SearchByIngredients is called and returns ingredient-specific matches
func TestIngredientSearch_IngredientsOnlyMatch(t *testing.T) {
	// Recipe whose NAME contains "rum" but ingredient does not contain "gin"
	rumName := &model.Recipe{
		ID: "rumname", Name: "Gin-Inspired Daiquiri",
		Ingredients: []model.Ingredient{{Name: "Rum"}, {Name: "Lime Juice"}},
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	total, _ := doIngSearch(t, "q=gin+and+lime+juice", rumName)
	// "gin" should NOT match the recipe name — ingredient-only search
	// stub SearchByIngredients checks ingredient names
	if total != 0 {
		t.Errorf("multi-ingredient search should NOT match recipe name, got %d results", total)
	}
}

// Verify stub satisfies updated interface (compile-time check moved here for clarity)
var _ interface {
	SearchByIngredients(ingredients []string, page, limit int) ([]*model.Recipe, int, error)
} = (*stubRecipeStore)(nil)

// helper: names of returned recipes
func recipeNames(data []any) []string {
	var names []string
	for _, item := range data {
		m := item.(map[string]any)
		names = append(names, m["name"].(string))
	}
	return names
}

func containsName(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

func TestIngredientSearch_ThreeIngredients(t *testing.T) {
	total, data := doIngSearch(t, "q=gin+and+lemon+juice+and+sugar",
		recipeGinFizz, recipeBeesKnee, recipeDaiquiri)
	// Only Gin Fizz has all three: Gin, Lemon Juice, Sugar
	if total != 1 {
		t.Errorf("expected 1 result (only Gin Fizz has all three), got total=%d len=%d names=%v",
			total, len(data), recipeNames(data))
	}
	if len(data) == 1 {
		m := data[0].(map[string]any)
		if m["name"] != "Gin Fizz" {
			t.Errorf("expected Gin Fizz, got %v", m["name"])
		}
	}
}

// mustContainOnlyIngredients reused by TestIngredientSearch_IngredientsOnlyMatch via strings
var _ = strings.Contains // suppress unused import if removed
