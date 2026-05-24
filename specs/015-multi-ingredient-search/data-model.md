# Data Model: Multi-Ingredient Search

**Feature**: 015-multi-ingredient-search
**Date**: 2026-05-24

## No Schema Changes Required

This feature adds no new tables, DynamoDB tables, or fields to existing records.
The `recipes` table and DynamoDB `cocktails-recipes` table are used as-is.

## Runtime Structures (in-memory only)

### IngredientQuery

Derived from the raw `q` URL parameter by the handler's parse function. Not persisted.

| Field | Type | Description |
|-------|------|-------------|
| `tokens` | `[]string` | Lowercased, trimmed ingredient name substrings. Length ≥ 2 for multi-ingredient path. |
| `isMulti` | `bool` | True when `len(tokens) >= 2`. Routes to `SearchByIngredients`. |

## RecipeStore Interface Addition

```go
// New method added to store.RecipeStore:
SearchByIngredients(ingredients []string, page, limit int) ([]*model.Recipe, int, error)
```

This method is a pure extension — all existing `RecipeStore` methods are unchanged.

## SQLite Query Pattern

For a query with N ingredient tokens, the dynamically-built SQL is:

```sql
SELECT COUNT(*) FROM recipes r
WHERE
  (SELECT COUNT(*) FROM json_each(r.ingredients)
   WHERE LOWER(json_extract(value,'$.name')) LIKE ?) > 0
  AND
  (SELECT COUNT(*) FROM json_each(r.ingredients)
   WHERE LOWER(json_extract(value,'$.name')) LIKE ?) > 0
  -- one block per token
```

Args: `%token_1%`, `%token_2%`, ..., `%token_1%`, `%token_2%`, ... (repeated for COUNT + SELECT).

The `ingredients` column is a JSON array of objects: `[{"name":"Gin","quantity":"60","unit":"ml",...},...]`.
`json_extract(value, '$.name')` retrieves the `name` field of each element.
