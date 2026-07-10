package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/almc/cocktails/internal/logging"
	"github.com/almc/cocktails/internal/model"
	"github.com/almc/cocktails/internal/store"
	"github.com/google/uuid"
)

var (
	reAnd  = regexp.MustCompile(`(?i)\s+and\s+`)
	rePlus = regexp.MustCompile(`\s*\+\s*`)
)

func parseIngredientQuery(q string) []string {
	var tokens []string
	for _, part := range reAnd.Split(q, -1) {
		for _, tok := range rePlus.Split(part, -1) {
			tok = strings.TrimSpace(tok)
			if tok != "" {
				tokens = append(tokens, tok)
			}
		}
	}
	return tokens
}

type RecipeHandler struct {
	recipes store.RecipeStore
}

func NewRecipeHandler(rs store.RecipeStore) *RecipeHandler {
	return &RecipeHandler{recipes: rs}
}

func (h *RecipeHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	baseSpirit := strings.TrimSpace(r.URL.Query().Get("base_spirit"))
	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", 20)
	if limit > 100 {
		limit = 100
	}

	var recipes []*model.Recipe
	var total int
	var err error

	tokens := parseIngredientQuery(q)
	switch {
	case baseSpirit != "" && len(tokens) >= 1:
		recipes, total, err = h.recipes.SearchByBaseSpiritAndIngredients(baseSpirit, tokens, page, limit)
	case baseSpirit != "":
		recipes, total, err = h.recipes.SearchByBaseSpirit(baseSpirit, page, limit)
	case len(tokens) >= 2:
		recipes, total, err = h.recipes.SearchByIngredients(tokens, page, limit)
	case len(tokens) == 1:
		recipes, total, err = h.recipes.Search(tokens[0], page, limit)
	default:
		recipes, total, err = h.recipes.List(page, limit)
	}
	log := logging.FromContext(r.Context())
	if err != nil {
		log.Error("recipe list failed", "action", "recipe.list", "outcome", "failure", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve recipes")
		return
	}
	if recipes == nil {
		recipes = []*model.Recipe{}
	}
	action := "recipe.list"
	switch {
	case baseSpirit != "":
		action = "search.base_spirit"
	case len(tokens) >= 1:
		action = "search.ingredients"
	}
	log.Debug("recipes listed", "action", action, "outcome", "success",
		"count", total, "q", q, "base_spirit", baseSpirit)
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  recipes,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *RecipeHandler) Random(w http.ResponseWriter, r *http.Request) {
	log := logging.FromContext(r.Context())
	recipe, err := h.recipes.Random()
	if err != nil {
		log.Error("random recipe failed", "action", "recipe.random", "outcome", "failure", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve recipe")
		return
	}
	if recipe == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	log.Debug("random recipe served", "action", "recipe.random", "outcome", "success", "recipe_id", recipe.ID)
	writeJSON(w, http.StatusOK, recipe)
}

func (h *RecipeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	log := logging.FromContext(r.Context())
	id := r.PathValue("id")
	recipe, err := h.recipes.GetByID(id)
	if err != nil {
		log.Debug("recipe not found", "action", "recipe.get", "outcome", "failure", "recipe_id", id, "reason", "not_found")
		writeError(w, http.StatusNotFound, "NOT_FOUND", "recipe not found")
		return
	}
	log.Debug("recipe served", "action", "recipe.get", "outcome", "success", "recipe_id", id)
	writeJSON(w, http.StatusOK, recipe)
}

func (h *RecipeHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())

	var body struct {
		Name        string             `json:"name"`
		Ingredients []model.Ingredient `json:"ingredients"`
		Steps       []string           `json:"steps"`
		Properties  map[string]string  `json:"properties"`
		Notes       *string            `json:"notes"`
		Garnishes   []string           `json:"garnishes"`
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

	notes := ""
	if body.Notes != nil {
		notes = *body.Notes
	}

	var garnishes []string
	for _, g := range body.Garnishes {
		if strings.TrimSpace(g) != "" {
			garnishes = append(garnishes, g)
		}
	}
	if garnishes == nil {
		garnishes = []string{}
	}

	now := time.Now().UTC()
	recipe := &model.Recipe{
		ID:          uuid.NewString(),
		Name:        body.Name,
		Ingredients: body.Ingredients,
		Steps:       body.Steps,
		Properties:  body.Properties,
		Notes:       notes,
		Garnishes:   garnishes,
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

	log := logging.FromContext(r.Context())
	if err := h.recipes.Create(recipe); err != nil {
		log.Error("recipe create failed", "action", "recipe.create", "outcome", "failure",
			"user_id", claims.UserID, "error", err.Error())
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create recipe")
		return
	}
	log.Info("recipe created", "action", "recipe.create", "outcome", "success",
		"user_id", claims.UserID, "recipe_id", recipe.ID)
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
	if !claims.IsAdmin && existing.CreatorID != claims.UserID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only the recipe creator can edit this recipe")
		return
	}

	var body struct {
		Name        *string            `json:"name"`
		Ingredients []model.Ingredient `json:"ingredients"`
		Steps       []string           `json:"steps"`
		Properties  map[string]string  `json:"properties"`
		Notes       *string            `json:"notes"`
		Garnishes   []string           `json:"garnishes"`
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
	if body.Notes != nil {
		existing.Notes = *body.Notes
	}
	if body.Garnishes != nil {
		var garnishes []string
		for _, g := range body.Garnishes {
			if strings.TrimSpace(g) != "" {
				garnishes = append(garnishes, g)
			}
		}
		if garnishes == nil {
			garnishes = []string{}
		}
		existing.Garnishes = garnishes
	}

	log := logging.FromContext(r.Context())
	if err := h.recipes.Update(existing); err != nil {
		log.Error("recipe update failed", "action", "recipe.update", "outcome", "failure",
			"user_id", claims.UserID, "recipe_id", id, "error", err.Error())
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update recipe")
		return
	}
	log.Info("recipe updated", "action", "recipe.update", "outcome", "success",
		"user_id", claims.UserID, "recipe_id", id)
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
	if !claims.IsAdmin && existing.CreatorID != claims.UserID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only the recipe creator can delete this recipe")
		return
	}
	if err := h.recipes.Delete(id); err != nil {
		logging.FromContext(r.Context()).Error("recipe delete failed", "action", "recipe.delete",
			"outcome", "failure", "user_id", claims.UserID, "recipe_id", id, "error", err.Error())
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete recipe")
		return
	}
	logging.FromContext(r.Context()).Info("recipe deleted", "action", "recipe.delete",
		"outcome", "success", "user_id", claims.UserID, "recipe_id", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *RecipeHandler) Mine(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", 20)
	if limit > 100 {
		limit = 100
	}
	log := logging.FromContext(r.Context())
	recipes, total, err := h.recipes.ListByCreator(claims.UserID, page, limit)
	if err != nil {
		log.Error("list own recipes failed", "action", "recipe.list", "outcome", "failure",
			"user_id", claims.UserID, "error", err.Error())
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve recipes")
		return
	}
	if recipes == nil {
		recipes = []*model.Recipe{}
	}
	log.Debug("own recipes listed", "action", "recipe.list", "outcome", "success",
		"user_id", claims.UserID, "count", total)
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  recipes,
		"total": total,
		"page":  page,
		"limit": limit,
	})
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
