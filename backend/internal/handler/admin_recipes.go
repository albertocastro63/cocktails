package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/almc/cocktails/internal/model"
	"github.com/almc/cocktails/internal/store"
	"github.com/google/uuid"
)

const recipeSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Recipe",
  "description": "A cocktail recipe. Use this schema to prepare import files or validate exported data.",
  "type": "object",
  "required": ["name"],
  "additionalProperties": false,
  "properties": {
    "name": {
      "type": "string",
      "description": "The name of the cocktail (required)"
    },
    "ingredients": {
      "type": "array",
      "description": "Ordered list of ingredients",
      "items": {
        "type": "object",
        "required": ["name"],
        "additionalProperties": false,
        "properties": {
          "name": { "type": "string", "description": "Ingredient name" },
          "quantity": { "type": "string", "description": "Amount (e.g. '1.5')" },
          "unit": { "type": "string", "description": "Unit of measure (e.g. 'oz')" }
        }
      }
    },
    "steps": {
      "type": "array",
      "description": "Ordered preparation steps",
      "items": { "type": "string" }
    },
    "properties": {
      "type": "object",
      "description": "Named text-value pairs (e.g. glass type, garnish)",
      "additionalProperties": { "type": "string" }
    },
    "notes": {
      "type": "string",
      "description": "Free-form notes about the recipe (markdown supported)"
    },
    "garnishes": {
      "type": "array",
      "description": "Ordered list of garnish instructions",
      "items": { "type": "string" }
    }
  }
}`

type AdminRecipeHandler struct {
	recipes store.RecipeStore
}

func NewAdminRecipeHandler(rs store.RecipeStore) *AdminRecipeHandler {
	return &AdminRecipeHandler{recipes: rs}
}

// recipeExport is the wire format for export/import (no server-generated fields).
type recipeExport struct {
	Name        string             `json:"name"`
	Ingredients []model.Ingredient `json:"ingredients,omitempty"`
	Steps       []string           `json:"steps,omitempty"`
	Properties  map[string]string  `json:"properties,omitempty"`
	Notes       string             `json:"notes,omitempty"`
	Garnishes   []string           `json:"garnishes,omitempty"`
}

func (h *AdminRecipeHandler) ExportSchema(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="recipe-schema.json"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(recipeSchema))
}

func (h *AdminRecipeHandler) ExportRecipes(w http.ResponseWriter, r *http.Request) {
	recipes, err := h.recipes.ListAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve recipes")
		return
	}

	exports := make([]recipeExport, len(recipes))
	for i, rec := range recipes {
		exports[i] = recipeExport{
			Name:        rec.Name,
			Ingredients: rec.Ingredients,
			Steps:       rec.Steps,
			Properties:  rec.Properties,
			Notes:       rec.Notes,
			Garnishes:   rec.Garnishes,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="recipes-export.json"`)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(exports)
}

func (h *AdminRecipeHandler) ImportRecipes(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	var raw []json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		msg := "import file must be a JSON array"
		if strings.Contains(err.Error(), "request body too large") {
			msg = "import file exceeds maximum allowed size"
		}
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", msg)
		return
	}

	claims := ClaimsFromContext(r.Context())

	recipes := make([]*model.Recipe, 0, len(raw))
	for i, item := range raw {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(item, &obj); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf("recipe at index %d: must be a JSON object", i))
			return
		}

		nameRaw, ok := obj["name"]
		if !ok {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf("recipe at index %d: name is required", i))
			return
		}
		var name string
		if err := json.Unmarshal(nameRaw, &name); err != nil || strings.TrimSpace(name) == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf("recipe at index %d: name must be a non-empty string", i))
			return
		}

		rec := &model.Recipe{
			ID:        uuid.NewString(),
			Name:      name,
			CreatorID: claims.UserID,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		if v, ok := obj["ingredients"]; ok {
			if err := json.Unmarshal(v, &rec.Ingredients); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf("recipe at index %d: ingredients must be an array of objects", i))
				return
			}
			for j, ing := range rec.Ingredients {
				if strings.TrimSpace(ing.Name) == "" {
					writeError(w, http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf("recipe at index %d: ingredient name at index %d is required", i, j))
					return
				}
			}
		}

		if v, ok := obj["steps"]; ok {
			if err := json.Unmarshal(v, &rec.Steps); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf("recipe at index %d: steps must be an array of strings", i))
				return
			}
		}

		if v, ok := obj["properties"]; ok {
			if err := json.Unmarshal(v, &rec.Properties); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf("recipe at index %d: properties must be an object with string values", i))
				return
			}
		}

		if v, ok := obj["notes"]; ok {
			if err := json.Unmarshal(v, &rec.Notes); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf("recipe at index %d: notes must be a string", i))
				return
			}
		}

		if v, ok := obj["garnishes"]; ok {
			if err := json.Unmarshal(v, &rec.Garnishes); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf("recipe at index %d: garnishes must be an array of strings", i))
				return
			}
		}

		recipes = append(recipes, rec)
	}

	created, skipped, err := h.recipes.ImportBatch(recipes, claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "import failed; no recipes were created")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{
		"imported": created,
		"skipped":  skipped,
	})
}
