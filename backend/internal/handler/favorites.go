package handler

import (
	"net/http"

	"github.com/almc/cocktails/internal/store"
)

type FavoriteHandler struct {
	favorites store.FavoriteStore
	recipes   store.RecipeStore
}

func NewFavoriteHandler(fs store.FavoriteStore, rs store.RecipeStore) *FavoriteHandler {
	return &FavoriteHandler{favorites: fs, recipes: rs}
}

func (h *FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	id := r.PathValue("id")

	recipe, err := h.recipes.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "recipe not found")
		return
	}
	if recipe.CreatorID == claims.UserID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "cannot favorite your own recipe")
		return
	}
	if err := h.favorites.Add(claims.UserID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to save favorite")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	id := r.PathValue("id")

	if err := h.favorites.Remove(claims.UserID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to remove favorite")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FavoriteHandler) Check(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	id := r.PathValue("id")

	ok, err := h.favorites.IsFavorite(claims.UserID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to check favorite")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"is_favorite": ok})
}

func (h *FavoriteHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())

	favs, err := h.favorites.ListByUser(claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list favorites")
		return
	}

	// N+1: fetch each recipe by ID — acceptable for current scale; optimize with batch fetch later
	type recipeWithFav struct {
		ID          any    `json:"-"`
		IsFavorite  bool   `json:"is_favorite"`
	}
	items := make([]map[string]any, 0, len(favs))
	for _, fav := range favs {
		recipe, err := h.recipes.GetByID(fav.RecipeID)
		if err != nil {
			continue
		}
		item := map[string]any{
			"id":          recipe.ID,
			"name":        recipe.Name,
			"ingredients": recipe.Ingredients,
			"steps":       recipe.Steps,
			"properties":  recipe.Properties,
			"notes":       recipe.Notes,
			"creator_id":  recipe.CreatorID,
			"created_at":  recipe.CreatedAt,
			"updated_at":  recipe.UpdatedAt,
			"is_favorite": true,
		}
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  items,
		"total": len(items),
		"page":  1,
		"limit": len(items),
	})
}
