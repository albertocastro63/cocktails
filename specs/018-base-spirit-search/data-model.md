# Data Model: Base Spirit Search Filter

**Feature**: 018-base-spirit-search | **Date**: 2026-05-25

---

## Existing Entities (no changes required)

### Ingredient

```
Ingredient {
  Name         string   -- ingredient name (used for all substring matching)
  Quantity     string
  Unit         string   (optional)
  IsBaseSpirit bool     -- true for the ingredient flagged as the base spirit
}
```

The `IsBaseSpirit` field already exists (feature 011). No schema migration is needed for either the SQLite or DynamoDB stores.

### Recipe

```
Recipe {
  ID          string
  Name        string
  Ingredients []Ingredient
  Steps       []string
  Properties  map[string]string
  Notes       string
  CreatorID   string
  CreatedAt   time.Time
  UpdatedAt   time.Time
}
```

---

## New Store Interface Methods

Two new methods are added to `store.RecipeStore`:

```go
// SearchByBaseSpirit returns recipes that have at least one ingredient with
// IsBaseSpirit=true whose Name contains baseSpirit (case-insensitive substring).
SearchByBaseSpirit(baseSpirit string, page, limit int) ([]*model.Recipe, int, error)

// SearchByBaseSpiritAndIngredients returns recipes matching both the base-spirit
// constraint and all ingredient substring constraints (intersection).
SearchByBaseSpiritAndIngredients(
    baseSpirit string,
    ingredients []string,
    page, limit int,
) ([]*model.Recipe, int, error)
```

Both methods follow the same pagination contract as the existing search methods.

---

## Query Parameter Contract

The `GET /api/v1/recipes` endpoint gains one new optional query parameter:

| Parameter    | Type   | Description |
|--------------|--------|-------------|
| `base_spirit` | string | Substring to match against the base-spirit ingredient name. Ignored if empty. |

Existing parameters (`q`, `page`, `limit`) are unchanged.
