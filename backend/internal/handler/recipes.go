package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/almc/cocktails/internal/model"
	"github.com/almc/cocktails/internal/store"
	"github.com/google/uuid"
)

type RecipeHandler struct {
	recipes store.RecipeStore
}

func NewRecipeHandler(rs store.RecipeStore) *RecipeHandler {
	return &RecipeHandler{recipes: rs}
}

func (h *RecipeHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", 20)
	if limit > 100 {
		limit = 100
	}

	var recipes []*model.Recipe
	var total int
	var err error

	if strings.TrimSpace(q) != "" {
		recipes, total, err = h.recipes.Search(q, page, limit)
	} else {
		recipes, total, err = h.recipes.List(page, limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve recipes")
		return
	}
	if recipes == nil {
		recipes = []*model.Recipe{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  recipes,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *RecipeHandler) Random(w http.ResponseWriter, r *http.Request) {
	recipe, err := h.recipes.Random()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve recipe")
		return
	}
	if recipe == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeJSON(w, http.StatusOK, recipe)
}

func (h *RecipeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	recipe, err := h.recipes.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "recipe not found")
		return
	}
	writeJSON(w, http.StatusOK, recipe)
}

func (h *RecipeHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())

	var body struct {
		Name        string            `json:"name"`
		Ingredients []model.Ingredient `json:"ingredients"`
		Steps       []string          `json:"steps"`
		Properties  map[string]string `json:"properties"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "name is required")
		return
	}

	var warnings []string
	exists, err := h.recipes.ExistsByName(body.Name)
	if err == nil && exists {
		warnings = append(warnings, "a recipe with this name already exists")
	}

	now := time.Now().UTC()
	recipe := &model.Recipe{
		ID:          uuid.NewString(),
		Name:        body.Name,
		Ingredients: body.Ingredients,
		Steps:       body.Steps,
		Properties:  body.Properties,
		CreatorID:   claims.UserID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if recipe.Ingredients == nil {
		recipe.Ingredients = []model.Ingredient{}
	}
	if recipe.Steps == nil {
		recipe.Steps = []string{}
	}
	if recipe.Properties == nil {
		recipe.Properties = map[string]string{}
	}

	if err := h.recipes.Create(recipe); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create recipe")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"data":     recipe,
		"warnings": warnings,
	})
}

func (h *RecipeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	claims := ClaimsFromContext(r.Context())

	existing, err := h.recipes.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "recipe not found")
		return
	}
	if existing.CreatorID != claims.UserID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only the recipe creator can edit this recipe")
		return
	}

	var body struct {
		Name        *string            `json:"name"`
		Ingredients []model.Ingredient `json:"ingredients"`
		Steps       []string          `json:"steps"`
		Properties  map[string]string `json:"properties"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	if body.Name != nil {
		if strings.TrimSpace(*body.Name) == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "name cannot be empty")
			return
		}
		existing.Name = *body.Name
	}
	if body.Ingredients != nil {
		existing.Ingredients = body.Ingredients
	}
	if body.Steps != nil {
		existing.Steps = body.Steps
	}
	if body.Properties != nil {
		existing.Properties = body.Properties
	}

	if err := h.recipes.Update(existing); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update recipe")
		return
	}
	writeJSON(w, http.StatusOK, existing)
}

func (h *RecipeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	claims := ClaimsFromContext(r.Context())

	existing, err := h.recipes.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "recipe not found")
		return
	}
	if existing.CreatorID != claims.UserID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only the recipe creator can delete this recipe")
		return
	}
	if err := h.recipes.Delete(id); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete recipe")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return def
	}
	return n
}
