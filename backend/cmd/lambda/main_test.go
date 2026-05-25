package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/almc/cocktails/internal/model"
)

// noopRecipeStore satisfies store.RecipeStore with no-op implementations.
type noopRecipeStore struct{}

func (noopRecipeStore) Create(*model.Recipe) error                                            { return nil }
func (noopRecipeStore) GetByID(string) (*model.Recipe, error)                                { return nil, errors.New("not found") }
func (noopRecipeStore) List(int, int) ([]*model.Recipe, int, error)                          { return nil, 0, nil }
func (noopRecipeStore) Search(string, int, int) ([]*model.Recipe, int, error)                { return nil, 0, nil }
func (noopRecipeStore) SearchByIngredients([]string, int, int) ([]*model.Recipe, int, error)           { return nil, 0, nil }
func (noopRecipeStore) SearchByBaseSpirit(string, int, int) ([]*model.Recipe, int, error)              { return nil, 0, nil }
func (noopRecipeStore) SearchByBaseSpiritAndIngredients(string, []string, int, int) ([]*model.Recipe, int, error) {
	return nil, 0, nil
}
func (noopRecipeStore) Random() (*model.Recipe, error)                                                 { return nil, nil }
func (noopRecipeStore) Update(*model.Recipe) error                                           { return nil }
func (noopRecipeStore) Delete(string) error                                                  { return nil }
func (noopRecipeStore) ExistsByName(string) (bool, error)                                    { return false, nil }
func (noopRecipeStore) ListAll() ([]*model.Recipe, error)                                    { return nil, nil }
func (noopRecipeStore) ImportBatch([]*model.Recipe, string) (int, int, error)                { return 0, 0, nil }
func (noopRecipeStore) ListByCreator(string, int, int) ([]*model.Recipe, int, error)         { return nil, 0, nil }

// noopUserStore satisfies store.UserStore with no-op implementations.
type noopUserStore struct{}

func (noopUserStore) Create(*model.User) error                   { return nil }
func (noopUserStore) GetByID(string) (*model.User, error)        { return nil, errors.New("not found") }
func (noopUserStore) GetByUsername(string) (*model.User, error)  { return nil, errors.New("not found") }
func (noopUserStore) Count() (int, error)                        { return 0, nil }
func (noopUserStore) List() ([]*model.User, error)               { return nil, nil }
func (noopUserStore) Update(*model.User) error                   { return nil }
func (noopUserStore) Delete(string) error                        { return nil }
func (noopUserStore) GetByEmail(string) (*model.User, error)     { return nil, errors.New("not found") }

// noopFavoriteStore satisfies store.FavoriteStore with no-op implementations.
type noopFavoriteStore struct{}

func (noopFavoriteStore) Add(string, string) error                       { return nil }
func (noopFavoriteStore) Remove(string, string) error                    { return nil }
func (noopFavoriteStore) IsFavorite(string, string) (bool, error)        { return false, nil }
func (noopFavoriteStore) ListByUser(string) ([]*model.Favorite, error)   { return nil, nil }
func (noopFavoriteStore) CountByRecipe(string) (int, error)              { return 0, nil }

// TestBuildHandler_NoMissingRoutes ensures every expected route is registered
// in the Lambda handler (returns anything except 405 Method Not Allowed).
// This test would have caught the bug where GET /api/v1/admin/users was present
// in cmd/server/main.go but absent from cmd/lambda/main.go.
func TestBuildHandler_NoMissingRoutes(t *testing.T) {
	h := buildHandler(noopRecipeStore{}, noopUserStore{}, noopFavoriteStore{})

	routes := []struct{ method, path string }{
		// Public recipe routes
		{"GET", "/api/v1/recipes"},
		{"GET", "/api/v1/recipes/random"},
		{"GET", "/api/v1/recipes/abc"},
		// Auth-protected recipe routes (expect 401, not 405)
		{"GET", "/api/v1/recipes/mine"},
		{"POST", "/api/v1/recipes"},
		{"PUT", "/api/v1/recipes/abc"},
		{"DELETE", "/api/v1/recipes/abc"},
		// Favorite routes (expect 401, not 405)
		{"GET", "/api/v1/recipes/favorites"},
		{"GET", "/api/v1/recipes/abc/favorite"},
		{"PUT", "/api/v1/recipes/abc/favorite"},
		{"DELETE", "/api/v1/recipes/abc/favorite"},
		// Auth route
		{"POST", "/api/v1/auth/login"},
		// Admin user routes (expect 401, not 405)
		{"GET", "/api/v1/admin/users"},
		{"POST", "/api/v1/admin/users"},
		{"GET", "/api/v1/admin/users/abc"},
		{"PUT", "/api/v1/admin/users/abc"},
		{"DELETE", "/api/v1/admin/users/abc"},
		// Admin recipe management routes (expect 401, not 405)
		{"GET", "/api/v1/admin/schema"},
		{"GET", "/api/v1/admin/recipes/export"},
		{"POST", "/api/v1/admin/recipes/import"},
	}

	for _, tt := range routes {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code == http.StatusMethodNotAllowed {
				t.Errorf("got 405 — route %s %s is not registered in the Lambda handler", tt.method, tt.path)
			}
		})
	}
}
