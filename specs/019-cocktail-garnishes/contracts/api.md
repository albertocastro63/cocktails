# API Contract Changes: Cocktail Recipe Garnishes

**Feature**: 019-cocktail-garnishes  
**Date**: 2026-05-26

## Changed Endpoints

All recipe endpoints already exist. This feature extends the recipe JSON shape by adding the optional `garnishes` field. No new endpoints are introduced.

---

### GET /recipes, GET /recipes/:id, GET /recipes/mine

**Change**: `garnishes` added to response body (omitted when empty).

Response body (partial):
```json
{
  "garnishes": ["Express orange oil over the cocktail", "Use orange peel to garnish"]
}
```

---

### POST /recipes (create)

**Change**: `garnishes` accepted in request body (optional).

Request body (partial):
```json
{
  "garnishes": ["Express orange oil over the cocktail"]
}
```

- `null` or absent: recipe created with empty garnish list.
- Empty strings in the array: filtered out before storage.

---

### PUT /recipes/:id (update)

**Change**: `garnishes` accepted in request body (optional, full-replace semantics).

Request body (partial):
```json
{
  "garnishes": ["Use orange peel to garnish"]
}
```

- `null` or absent: existing garnishes are left unchanged.
- Empty array `[]`: clears all garnishes.
- Non-null value: replaces the entire garnish list.

---

### GET /admin/recipes/export

**Change**: `garnishes` included in each exported recipe object when non-empty.

---

### POST /admin/recipes/import

**Change**: `garnishes` accepted as an optional array of strings in each recipe object.  
- Import validates: if present, must be a JSON array of strings.
- Absent: recipe imported with empty garnish list.

---

### GET /admin/recipes/schema

**Change**: `recipeSchema` updated to include `garnishes` as an optional `array of strings`:

```json
"garnishes": {
  "type": "array",
  "description": "Ordered list of garnish instructions",
  "items": { "type": "string" }
}
```

---

## Backward Compatibility

All changes are additive and optional. Existing API clients that do not send `garnishes` continue to work without modification. Existing recipe responses gain the `garnishes` field only when the recipe has garnishes; otherwise the field is omitted.
