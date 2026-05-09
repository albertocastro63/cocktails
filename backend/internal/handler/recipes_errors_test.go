package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/almc/cocktails/internal/handler"
)

func TestRecipeList_StoreError(t *testing.T) {
	rs := newStubRecipeStore()
	rs.err = errGeneric
	h := handler.NewRecipeHandler(rs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d want 500", rec.Code)
	}
}

func TestRecipeRandom_StoreError(t *testing.T) {
	rs := newStubRecipeStore()
	rs.err = errGeneric
	h := handler.NewRecipeHandler(rs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/random", nil)
	rec := httptest.NewRecorder()
	h.Random(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d want 500", rec.Code)
	}
}

func TestRecipeCreate_MissingName(t *testing.T) {
	rs := newStubRecipeStore()
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Create))

	token := validToken(t, "u1", "alice", false)
	body := `{"name":"  ","ingredients":[],"steps":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/recipes", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d want 400", rec.Code)
	}
}

func TestRecipeCreate_InvalidBody(t *testing.T) {
	rs := newStubRecipeStore()
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Create))

	token := validToken(t, "u1", "alice", false)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/recipes", strings.NewReader("not-json"))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d want 400", rec.Code)
	}
}

func TestRecipeCreate_StoreError(t *testing.T) {
	rs := newStubRecipeStore()
	rs.err = errGeneric
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Create))

	token := validToken(t, "u1", "alice", false)
	body := `{"name":"Test","ingredients":[],"steps":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/recipes", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d want 500", rec.Code)
	}
}

func TestRecipeUpdate_InvalidBody(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u1"))
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Update))

	token := validToken(t, "u1", "alice", false)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/r1", strings.NewReader("bad"))
	req.SetPathValue("id", "r1")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d want 400", rec.Code)
	}
}

func TestRecipeUpdate_EmptyName(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "u1"))
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Update))

	token := validToken(t, "u1", "alice", false)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/r1", strings.NewReader(`{"name":"  "}`))
	req.SetPathValue("id", "r1")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d want 400", rec.Code)
	}
}

func TestRecipeDelete_NotFound(t *testing.T) {
	rs := newStubRecipeStore()
	h := handler.NewRecipeHandler(rs)
	wrapped := handler.RequireAuth(http.HandlerFunc(h.Delete))

	token := validToken(t, "u1", "alice", false)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/recipes/nope", nil)
	req.SetPathValue("id", "nope")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got %d want 404", rec.Code)
	}
}

func TestRecipeList_SearchWithQ(t *testing.T) {
	rs := newStubRecipeStore(
		sampleRecipe("r1", "Mojito", "u1"),
		sampleRecipe("r2", "Martini", "u1"),
	)
	h := handler.NewRecipeHandler(rs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes?q=martini", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rec.Code)
	}
}
