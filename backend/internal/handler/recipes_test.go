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

func sampleRecipe(id, name, creatorID string) *model.Recipe {
	return &model.Recipe{
		ID:          id,
		Name:        name,
		Ingredients: []model.Ingredient{{Name: "rum", Quantity: "50", Unit: "ml"}},
		Steps:       []string{"mix"},
		Properties:  map[string]string{"style": "tropical"},
		CreatorID:   creatorID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
}

// T022: GET /api/v1/recipes returns list
func TestRecipeList_OK(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u1"))
	h := handler.NewRecipeHandler(rs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rec.Code)
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Errorf("expected 1 recipe, got %d", len(data))
	}
}

// T022: empty store returns empty array
func TestRecipeList_Empty(t *testing.T) {
	rs := newStubRecipeStore()
	h := handler.NewRecipeHandler(rs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rec.Code)
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp["data"].([]any)
	if len(data) != 0 {
		t.Errorf("expected empty array, got %d items", len(data))
	}
}

// T023: GET /api/v1/recipes/random returns one recipe
func TestRecipeRandom_OK(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u1"))
	h := handler.NewRecipeHandler(rs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/random", nil)
	rec := httptest.NewRecorder()
	h.Random(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d want 200", rec.Code)
	}
}

// T023: 204 when store empty
func TestRecipeRandom_Empty(t *testing.T) {
	rs := newStubRecipeStore()
	h := handler.NewRecipeHandler(rs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/random", nil)
	rec := httptest.NewRecorder()
	h.Random(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("got %d want 204", rec.Code)
	}
}

// T036: search with q returns matching recipes only
func TestRecipeList_Search(t *testing.T) {
	rs := newStubRecipeStore(
		sampleRecipe("r1", "Mojito", "u1"),
		sampleRecipe("r2", "Margarita", "u1"),
	)
	h := handler.NewRecipeHandler(rs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes?q=mojito", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rec.Code)
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp["data"].([]any)
	if len(data) != 1 {
		t.Errorf("expected 1 search result, got %d", len(data))
	}
}

// T044: GET /api/v1/recipes/{id} returns recipe
func TestRecipeGetByID_OK(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u1"))
	h := handler.NewRecipeHandler(rs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/r1", nil)
	req.SetPathValue("id", "r1")
	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d want 200", rec.Code)
	}
}

// T044: unknown ID returns 404
func TestRecipeGetByID_NotFound(t *testing.T) {
	rs := newStubRecipeStore()
	h := handler.NewRecipeHandler(rs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/missing", nil)
	req.SetPathValue("id", "missing")
	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got %d want 404", rec.Code)
	}
}

// T056: POST /api/v1/recipes creates recipe with creator_id from JWT
func TestRecipeCreate_OK(t *testing.T) {
	rs := newStubRecipeStore()
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Create))

	token := validToken(t, "u1", "alice", false)
	body := `{"name":"Mojito","ingredients":[],"steps":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/recipes", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got %d want 201: %s", rec.Code, rec.Body)
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp["data"].(map[string]any)
	if data["creator_id"] != "u1" {
		t.Errorf("creator_id: got %v want u1", data["creator_id"])
	}
}

// T056: duplicate name warning
func TestRecipeCreate_DuplicateNameWarning(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u1"))
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Create))

	token := validToken(t, "u2", "bob", false)
	body := `{"name":"Mojito","ingredients":[],"steps":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/recipes", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got %d want 201", rec.Code)
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	warnings := resp["warnings"].([]any)
	if len(warnings) == 0 {
		t.Error("expected duplicate name warning")
	}
}

// T057: PUT /api/v1/recipes/{id} updates recipe
func TestRecipeUpdate_OK(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u1"))
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Update))

	token := validToken(t, "u1", "alice", false)
	body := `{"name":"Mojito Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/r1", strings.NewReader(body))
	req.SetPathValue("id", "r1")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d want 200: %s", rec.Code, rec.Body)
	}
}

// T057: 401 if no JWT
func TestRecipeUpdate_NoAuth(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u1"))
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Update))

	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/r1", strings.NewReader(`{}`))
	req.SetPathValue("id", "r1")
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got %d want 401", rec.Code)
	}
}

// T057: 403 if not creator
func TestRecipeUpdate_NotCreator(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u1"))
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Update))

	token := validToken(t, "u2", "bob", false)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/r1", strings.NewReader(`{}`))
	req.SetPathValue("id", "r1")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("got %d want 403", rec.Code)
	}
}

// T058: DELETE /api/v1/recipes/{id} deletes recipe
func TestRecipeDelete_OK(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u1"))
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Delete))

	token := validToken(t, "u1", "alice", false)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/recipes/r1", nil)
	req.SetPathValue("id", "r1")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("got %d want 204", rec.Code)
	}
}

// T058: 403 if not creator
func TestRecipeDelete_NotCreator(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u1"))
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Delete))

	token := validToken(t, "u2", "bob", false)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/recipes/r1", nil)
	req.SetPathValue("id", "r1")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("got %d want 403", rec.Code)
	}
}
