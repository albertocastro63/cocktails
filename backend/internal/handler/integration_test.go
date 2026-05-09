package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/almc/cocktails/internal/handler"
	"github.com/almc/cocktails/internal/model"
)

func buildTestRouter(t *testing.T) http.Handler {
	t.Helper()
	recipe := sampleRecipe("r1", "Mojito", "u1")
	recipe.Properties = map[string]string{"style": "refreshing", "base_spirit": "rum"}
	rs := newStubRecipeStore(recipe)
	us := newStubUserStore()

	recipes := handler.NewRecipeHandler(rs)
	authH := handler.NewAuthHandler(us)
	adminH := handler.NewAdminHandler(us)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/recipes", recipes.List)
	mux.HandleFunc("GET /api/v1/recipes/random", recipes.Random)
	mux.HandleFunc("GET /api/v1/recipes/{id}", recipes.GetByID)
	mux.Handle("POST /api/v1/recipes", handler.RequireAuth(http.HandlerFunc(recipes.Create)))
	mux.Handle("PUT /api/v1/recipes/{id}", handler.RequireAuth(http.HandlerFunc(recipes.Update)))
	mux.Handle("DELETE /api/v1/recipes/{id}", handler.RequireAuth(http.HandlerFunc(recipes.Delete)))
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.Handle("POST /api/v1/admin/users",
		handler.RequireAuth(handler.RequireAdmin(http.HandlerFunc(adminH.CreateUser))))
	return mux
}

// T071: all public endpoints return correct JSON structure without auth token
func TestIntegration_PublicListEndpoint(t *testing.T) {
	h := buildTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/recipes: got %d want 200", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"data", "total", "page", "limit"} {
		if _, ok := resp[key]; !ok {
			t.Errorf("missing key %q in response", key)
		}
	}
}

func TestIntegration_PublicRandomEndpoint(t *testing.T) {
	h := buildTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/random", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/recipes/random: got %d want 200", rec.Code)
	}
	var recipe model.Recipe
	if err := json.NewDecoder(rec.Body).Decode(&recipe); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if recipe.Name == "" {
		t.Error("expected non-empty recipe name")
	}
}

func TestIntegration_PublicGetByIDEndpoint(t *testing.T) {
	h := buildTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/r1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/recipes/r1: got %d want 200", rec.Code)
	}
}

func TestIntegration_PublicSearchEndpoint(t *testing.T) {
	h := buildTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes?q=mojito", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/recipes?q=: got %d want 200", rec.Code)
	}
}

// T072: flexible properties fully included in API responses
func TestIntegration_FlexiblePropertiesInResponse(t *testing.T) {
	h := buildTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/r1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var recipe map[string]any
	json.NewDecoder(rec.Body).Decode(&recipe)
	props, ok := recipe["properties"].(map[string]any)
	if !ok || len(props) == 0 {
		t.Error("expected non-empty properties in recipe response")
	}
	if props["style"] != "refreshing" {
		t.Errorf("expected style=refreshing, got %v", props["style"])
	}
	if props["base_spirit"] != "rum" {
		t.Errorf("expected base_spirit=rum, got %v", props["base_spirit"])
	}
}
