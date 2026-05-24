package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/almc/cocktails/internal/handler"
	"github.com/almc/cocktails/internal/model"
	"github.com/almc/cocktails/internal/store"
)

// stubFavoriteStore implements store.FavoriteStore for tests.
type stubFavoriteStore struct {
	isFav     bool
	addErr    error
	removeErr error
	listErr   error
	favorites []*model.Favorite
}

func (s *stubFavoriteStore) Add(userID, recipeID string) error        { return s.addErr }
func (s *stubFavoriteStore) Remove(userID, recipeID string) error     { return s.removeErr }
func (s *stubFavoriteStore) IsFavorite(userID, recipeID string) (bool, error) {
	return s.isFav, nil
}
func (s *stubFavoriteStore) ListByUser(userID string) ([]*model.Favorite, error) {
	return s.favorites, s.listErr
}
func (s *stubFavoriteStore) CountByRecipe(recipeID string) (int, error) { return 0, nil }

var _ store.FavoriteStore = (*stubFavoriteStore)(nil)

func testFavMux(fs store.FavoriteStore, rs store.RecipeStore) *http.ServeMux {
	favH := handler.NewFavoriteHandler(fs, rs)
	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/recipes/favorites", handler.RequireAuth(http.HandlerFunc(favH.List)))
	mux.Handle("PUT /api/v1/recipes/{id}/favorite", handler.RequireAuth(http.HandlerFunc(favH.Add)))
	mux.Handle("DELETE /api/v1/recipes/{id}/favorite", handler.RequireAuth(http.HandlerFunc(favH.Remove)))
	mux.Handle("GET /api/v1/recipes/{id}/favorite", handler.RequireAuth(http.HandlerFunc(favH.Check)))
	return mux
}

// (1) PUT with valid auth returns 204
func TestFavoriteAdd_OK(t *testing.T) {
	recipe := sampleRecipe("r1", "Mojito", "other-user")
	rs := newStubRecipeStore(recipe)
	fs := &stubFavoriteStore{}
	mux := testFavMux(fs, rs)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/r1/favorite", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t, "user-1", "alice", false))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("got %d want 204", rec.Code)
	}
}

// (2) PUT with no token returns 401
func TestFavoriteAdd_NoToken(t *testing.T) {
	rs := newStubRecipeStore()
	fs := &stubFavoriteStore{}
	mux := testFavMux(fs, rs)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/r1/favorite", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got %d want 401", rec.Code)
	}
}

// (3) PUT when recipe.creator_id == claims.UserID returns 403 with code FORBIDDEN
func TestFavoriteAdd_OwnRecipe(t *testing.T) {
	recipe := sampleRecipe("r1", "Mojito", "user-1")
	rs := newStubRecipeStore(recipe)
	fs := &stubFavoriteStore{}
	mux := testFavMux(fs, rs)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/r1/favorite", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t, "user-1", "alice", false))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("got %d want 403", rec.Code)
	}
	var body map[string]any
	json.NewDecoder(rec.Body).Decode(&body) //nolint:errcheck
	errObj, _ := body["error"].(map[string]any)
	if errObj["code"] != "FORBIDDEN" {
		t.Errorf("expected code FORBIDDEN, got %v", errObj["code"])
	}
}

// (4) PUT when recipe not found returns 404
func TestFavoriteAdd_NotFound(t *testing.T) {
	rs := newStubRecipeStore()
	fs := &stubFavoriteStore{}
	mux := testFavMux(fs, rs)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/missing/favorite", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t, "user-1", "alice", false))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got %d want 404", rec.Code)
	}
}

// (5) DELETE with valid auth returns 204
func TestFavoriteRemove_OK(t *testing.T) {
	rs := newStubRecipeStore()
	fs := &stubFavoriteStore{}
	mux := testFavMux(fs, rs)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/recipes/r1/favorite", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t, "user-1", "alice", false))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("got %d want 204", rec.Code)
	}
}

// (6) GET check returns {"is_favorite":true} when IsFavorite returns true
func TestFavoriteCheck_True(t *testing.T) {
	rs := newStubRecipeStore()
	fs := &stubFavoriteStore{isFav: true}
	mux := testFavMux(fs, rs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/r1/favorite", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t, "user-1", "alice", false))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d want 200", rec.Code)
	}
	var body map[string]any
	json.NewDecoder(rec.Body).Decode(&body) //nolint:errcheck
	if body["is_favorite"] != true {
		t.Errorf("expected is_favorite=true, got %v", body["is_favorite"])
	}
}

// (7) GET check returns {"is_favorite":false} when not favorited
func TestFavoriteCheck_False(t *testing.T) {
	rs := newStubRecipeStore()
	fs := &stubFavoriteStore{isFav: false}
	mux := testFavMux(fs, rs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/r1/favorite", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t, "user-1", "alice", false))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d want 200", rec.Code)
	}
	var body map[string]any
	json.NewDecoder(rec.Body).Decode(&body) //nolint:errcheck
	if body["is_favorite"] != false {
		t.Errorf("expected is_favorite=false, got %v", body["is_favorite"])
	}
}

// (8) GET /recipes/favorites returns list with is_favorite:true on each item
func TestFavoriteList_OK(t *testing.T) {
	recipe := sampleRecipe("r1", "Mojito", "other-user")
	rs := newStubRecipeStore(recipe)
	fs := &stubFavoriteStore{
		favorites: []*model.Favorite{
			{UserID: "user-1", RecipeID: "r1", CreatedAt: time.Now()},
		},
	}
	mux := testFavMux(fs, rs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/favorites", nil)
	req.Header.Set("Authorization", "Bearer "+validToken(t, "user-1", "alice", false))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d want 200", rec.Code)
	}
	var body map[string]any
	json.NewDecoder(rec.Body).Decode(&body) //nolint:errcheck
	data, _ := body["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("expected 1 item, got %d", len(data))
	}
	item, _ := data[0].(map[string]any)
	if item["is_favorite"] != true {
		t.Errorf("expected is_favorite=true on list item, got %v", item["is_favorite"])
	}
}
