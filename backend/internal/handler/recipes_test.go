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

// T003: Create with notes returns notes in response
func TestRecipeCreate_WithNotes(t *testing.T) {
	rs := newStubRecipeStore()
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Create))

	token := validToken(t, "u1", "alice", false)
	body := `{"name":"Mojito","notes":"Try with aged rum."}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/recipes", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d want 201: %s", rec.Code, rec.Body)
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp["data"].(map[string]any)
	if data["notes"] != "Try with aged rum." {
		t.Errorf("notes: got %v want 'Try with aged rum.'", data["notes"])
	}
}

// T003: Update setting notes is reflected in response
func TestRecipeUpdate_SetsNotes(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u1"))
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Update))

	token := validToken(t, "u1", "alice", false)
	body := `{"notes":"A new note."}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/r1", strings.NewReader(body))
	req.SetPathValue("id", "r1")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200: %s", rec.Code, rec.Body)
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["notes"] != "A new note." {
		t.Errorf("notes: got %v want 'A new note.'", resp["notes"])
	}
}

// T003: Update omitting notes preserves existing notes
func TestRecipeUpdate_PreservesNotes(t *testing.T) {
	existing := sampleRecipe("r1", "Mojito", "u1")
	existing.Notes = "Original note."
	rs := newStubRecipeStore(existing)
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
		t.Fatalf("got %d want 200: %s", rec.Code, rec.Body)
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["notes"] != "Original note." {
		t.Errorf("notes: got %v want 'Original note.'", resp["notes"])
	}
}

// T003: Update notes as non-creator returns 403
func TestRecipeUpdate_NotesNonCreator_Forbidden(t *testing.T) {
	existing := sampleRecipe("r1", "Mojito", "u1")
	existing.Notes = "Secret note."
	rs := newStubRecipeStore(existing)
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Update))

	token := validToken(t, "u2", "bob", false)
	body := `{"notes":"Trying to change notes."}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/r1", strings.NewReader(body))
	req.SetPathValue("id", "r1")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("got %d want 403", rec.Code)
	}
}

// T004: legacy recipe (empty creator_id) blocks non-admin edit
func TestRecipeUpdate_LegacyRecipe_NonAdmin_Forbidden(t *testing.T) {
	legacy := sampleRecipe("r-legacy", "Old Recipe", "")
	rs := newStubRecipeStore(legacy)
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Update))

	token := validToken(t, "u2", "bob", false)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/r-legacy", strings.NewReader(`{}`))
	req.SetPathValue("id", "r-legacy")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("got %d want 403 for legacy recipe edit by non-admin", rec.Code)
	}
}

// T004: legacy recipe (empty creator_id) blocks non-admin delete
func TestRecipeDelete_LegacyRecipe_NonAdmin_Forbidden(t *testing.T) {
	legacy := sampleRecipe("r-legacy", "Old Recipe", "")
	rs := newStubRecipeStore(legacy)
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Delete))

	token := validToken(t, "u2", "bob", false)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/recipes/r-legacy", nil)
	req.SetPathValue("id", "r-legacy")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("got %d want 403 for legacy recipe delete by non-admin", rec.Code)
	}
}

// T009: admin can edit a recipe created by another user (failing until T010 implemented)
func TestRecipeUpdate_Admin_CanEditAny(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u1"))
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Update))

	token := validToken(t, "admin-1", "admin", true)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/r1", strings.NewReader(`{"name":"Admin Edit"}`))
	req.SetPathValue("id", "r1")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d want 200 for admin edit of other user's recipe", rec.Code)
	}
}

// T009: admin can delete a recipe created by another user (failing until T011 implemented)
func TestRecipeDelete_Admin_CanDeleteAny(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u1"))
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Delete))

	token := validToken(t, "admin-1", "admin", true)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/recipes/r1", nil)
	req.SetPathValue("id", "r1")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("got %d want 204 for admin delete of other user's recipe", rec.Code)
	}
}

// T009: admin can edit legacy recipe (empty creator_id)
func TestRecipeUpdate_Admin_CanEditLegacy(t *testing.T) {
	legacy := sampleRecipe("r-legacy", "Old Recipe", "")
	rs := newStubRecipeStore(legacy)
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Update))

	token := validToken(t, "admin-1", "admin", true)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/r-legacy", strings.NewReader(`{"name":"Admin Fixed"}`))
	req.SetPathValue("id", "r-legacy")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d want 200 for admin edit of legacy recipe", rec.Code)
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

func TestRecipeMine_ReturnsOwnedRecipes(t *testing.T) {
	rs := newStubRecipeStore(
		sampleRecipe("r1", "Mojito", "u1"),
		sampleRecipe("r2", "Daiquiri", "u1"),
		sampleRecipe("r3", "Negroni", "u2"),
	)
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Mine))

	token := validToken(t, "u1", "alice", false)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/mine", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rec.Code)
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp["data"].([]any)
	if len(data) != 2 {
		t.Errorf("expected 2 recipes, got %d", len(data))
	}
}

func TestRecipeMine_RequiresAuth(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u1"))
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Mine))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/mine", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got %d want 401", rec.Code)
	}
}

// T003 (011-base-spirit): is_base_spirit on one ingredient survives POST → GET round-trip
func TestRecipeCreate_BaseSpiritRoundTrip(t *testing.T) {
	rs := newStubRecipeStore()
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Create))

	token := validToken(t, "u1", "alice", false)
	body := `{"name":"Manhattan","ingredients":[` +
		`{"name":"Rye Whiskey","quantity":"60","unit":"ml","is_base_spirit":true},` +
		`{"name":"Sweet Vermouth","quantity":"30","unit":"ml"},` +
		`{"name":"Angostura Bitters","quantity":"2","unit":"dashes"}` +
		`],"steps":["stir","strain"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/recipes", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("POST got %d want 201: %s", rec.Code, rec.Body)
	}
	var createResp map[string]any
	json.NewDecoder(rec.Body).Decode(&createResp)
	id := createResp["data"].(map[string]any)["id"].(string)

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/"+id, nil)
	getReq.SetPathValue("id", id)
	getRec := httptest.NewRecorder()
	h.GetByID(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("GET got %d want 200: %s", getRec.Code, getRec.Body)
	}
	var recipe map[string]any
	json.NewDecoder(getRec.Body).Decode(&recipe)

	ings := recipe["ingredients"].([]any)
	if len(ings) != 3 {
		t.Fatalf("expected 3 ingredients, got %d", len(ings))
	}
	for _, ing := range ings {
		m := ing.(map[string]any)
		name := m["name"].(string)
		isBase, _ := m["is_base_spirit"].(bool)
		if name == "Rye Whiskey" && !isBase {
			t.Errorf("Rye Whiskey: expected is_base_spirit=true, got false/absent")
		}
		if name != "Rye Whiskey" && isBase {
			t.Errorf("%s: expected is_base_spirit absent/false, got true", name)
		}
	}
}

func TestRecipeMine_EmptyForUserWithNoRecipes(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u2"))
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Mine))

	token := validToken(t, "u1", "alice", false)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/mine", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rec.Code)
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp["data"].([]any)
	if len(data) != 0 {
		t.Errorf("expected 0 recipes, got %d", len(data))
	}
}

func baseSpiritRecipe(id, name, spirit string) *model.Recipe {
	return &model.Recipe{
		ID:        id,
		Name:      name,
		Ingredients: []model.Ingredient{
			{Name: spirit, IsBaseSpirit: true, Quantity: "50", Unit: "ml"},
			{Name: "lime juice", Quantity: "20", Unit: "ml"},
		},
		Steps:     []string{"mix"},
		CreatorID: "u1",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

func TestList_BaseSpiritFilter(t *testing.T) {
	rs := newStubRecipeStore(
		baseSpiritRecipe("r1", "Gimlet", "gin"),
		baseSpiritRecipe("r2", "Daiquiri", "rum"),
	)
	h := handler.NewRecipeHandler(rs)

	t.Run("base_spirit=gin returns only gin recipe", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes?base_spirit=gin", nil)
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
		got := data[0].(map[string]any)
		if got["id"] != "r1" {
			t.Errorf("expected r1, got %v", got["id"])
		}
	})

	t.Run("q=lime and base_spirit=gin returns intersection", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes?q=lime&base_spirit=gin", nil)
		rec := httptest.NewRecorder()
		h.List(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got %d want 200", rec.Code)
		}
		var resp map[string]any
		json.NewDecoder(rec.Body).Decode(&resp)
		data := resp["data"].([]any)
		if len(data) != 1 {
			t.Errorf("expected 1 recipe (intersection), got %d", len(data))
		}
	})

	t.Run("empty base_spirit is ignored", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes?base_spirit=", nil)
		rec := httptest.NewRecorder()
		h.List(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got %d want 200", rec.Code)
		}
		var resp map[string]any
		json.NewDecoder(rec.Body).Decode(&resp)
		data := resp["data"].([]any)
		if len(data) != 2 {
			t.Errorf("expected 2 recipes (no filter), got %d", len(data))
		}
	})

	t.Run("whitespace-only base_spirit is ignored", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes?base_spirit=%20", nil)
		rec := httptest.NewRecorder()
		h.List(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got %d want 200", rec.Code)
		}
		var resp map[string]any
		json.NewDecoder(rec.Body).Decode(&resp)
		data := resp["data"].([]any)
		if len(data) != 2 {
			t.Errorf("expected 2 recipes, got %d", len(data))
		}
	})
}
