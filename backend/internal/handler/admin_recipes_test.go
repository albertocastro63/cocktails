package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/almc/cocktails/internal/handler"
	"github.com/almc/cocktails/internal/model"
)

// --- helpers ---

func newAdminRecipeHandler(rs ...*model.Recipe) *handler.AdminRecipeHandler {
	store := newStubRecipeStore(rs...)
	return handler.NewAdminRecipeHandler(store)
}

// --- US2: ExportRecipes ---

func TestExportRecipes_Empty(t *testing.T) {
	h := newAdminRecipeHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/recipes/export", nil)
	rec := httptest.NewRecorder()
	h.ExportRecipes(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rec.Code)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "recipes-export.json") {
		t.Errorf("Content-Disposition missing filename: %q", cd)
	}
	var result []interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d items", len(result))
	}
}

func TestExportRecipes_WithData(t *testing.T) {
	r1 := &model.Recipe{
		ID:   "id-1",
		Name: "Margarita",
		Ingredients: []model.Ingredient{
			{Name: "tequila", Quantity: "50", Unit: "ml"},
		},
		Steps:     []string{"Shake"},
		CreatorID: "user-1",
	}
	r2 := &model.Recipe{ID: "id-2", Name: "Negroni", CreatorID: "user-1"}
	h := newAdminRecipeHandler(r1, r2)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/recipes/export", nil)
	rec := httptest.NewRecorder()
	h.ExportRecipes(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rec.Code)
	}
	var result []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 recipes, got %d", len(result))
	}
	// Server-generated fields must not appear in export
	for _, item := range result {
		if _, ok := item["id"]; ok {
			t.Error("export must not include 'id' field")
		}
		if _, ok := item["creator_id"]; ok {
			t.Error("export must not include 'creator_id' field")
		}
	}
}

// --- US3: ImportRecipes ---

func importRequest(t *testing.T, h *handler.AdminRecipeHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/recipes/import", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+validToken(t, "user-1", "admin", true))
	wrapped := handler.RequireAuth(http.HandlerFunc(h.ImportRecipes))
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)
	return rec
}

func TestImportRecipes_ValidArray(t *testing.T) {
	h := newAdminRecipeHandler()
	rec := importRequest(t, h, `[{"name":"Mojito"},{"name":"Daiquiri","steps":["Blend"]}]`)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200: %s", rec.Code, rec.Body.String())
	}
	var result map[string]int
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["imported"] != 2 {
		t.Errorf("imported: got %d want 2", result["imported"])
	}
	if result["skipped"] != 0 {
		t.Errorf("skipped: got %d want 0", result["skipped"])
	}
}

func TestImportRecipes_DuplicateSkipped(t *testing.T) {
	existing := &model.Recipe{ID: "id-1", Name: "Mojito", CreatorID: "user-1"}
	h := newAdminRecipeHandler(existing)
	rec := importRequest(t, h, `[{"name":"Mojito"},{"name":"Negroni"}]`)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200: %s", rec.Code, rec.Body.String())
	}
	var result map[string]int
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["imported"] != 1 {
		t.Errorf("imported: got %d want 1", result["imported"])
	}
	if result["skipped"] != 1 {
		t.Errorf("skipped: got %d want 1", result["skipped"])
	}
}

func TestImportRecipes_NotAnArray(t *testing.T) {
	h := newAdminRecipeHandler()
	rec := importRequest(t, h, `{"name":"Mojito"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d want 400", rec.Code)
	}
}

func TestImportRecipes_MissingName(t *testing.T) {
	h := newAdminRecipeHandler()
	rec := importRequest(t, h, `[{"steps":["Blend"]}]`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d want 400", rec.Code)
	}
}

// --- Benchmarks (T029) ---

func BenchmarkExportRecipes(b *testing.B) {
	recipes := make([]*model.Recipe, 500)
	for i := range recipes {
		recipes[i] = &model.Recipe{
			ID:        fmt.Sprintf("id-%d", i),
			Name:      fmt.Sprintf("Cocktail %d", i),
			CreatorID: "user-1",
		}
	}
	h := newAdminRecipeHandler(recipes...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/recipes/export", nil)
		rec := httptest.NewRecorder()
		h.ExportRecipes(rec, req)
	}
}

func BenchmarkImportBatch(b *testing.B) {
	body := buildImportBody(500)
	token := validBenchToken(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := newAdminRecipeHandler()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/recipes/import", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		wrapped := handler.RequireAuth(http.HandlerFunc(h.ImportRecipes))
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
	}
}

func buildImportBody(n int) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i := 0; i < n; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		fmt.Fprintf(&sb, `{"name":"Cocktail %d"}`, i)
	}
	sb.WriteString("]")
	return sb.String()
}

func validBenchToken(b *testing.B) string {
	b.Helper()
	t := &testing.T{}
	return validToken(t, "bench-user", "admin", true)
}

// --- US1: ExportSchema ---

func TestExportSchema(t *testing.T) {
	h := newAdminRecipeHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/schema", nil)
	rec := httptest.NewRecorder()
	h.ExportSchema(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rec.Code)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "recipe-schema.json") {
		t.Errorf("Content-Disposition missing filename: %q", cd)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"required"`) {
		t.Errorf("schema missing 'required' key")
	}
	if !strings.Contains(body, `"name"`) {
		t.Errorf("schema missing 'name' field")
	}
}
