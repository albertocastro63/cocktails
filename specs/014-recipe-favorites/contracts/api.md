# API Contracts: Recipe Favorites

**Feature**: 014-recipe-favorites
**Base path**: `/api/v1`
**Auth**: `Authorization: Bearer <jwt>` header (all endpoints require authentication)

---

## PUT /api/v1/recipes/{id}/favorite

Add a recipe to the authenticated user's favorites. Idempotent — adding an already-favorited recipe returns 204 without error.

**Auth**: Required  
**Path param**: `id` — recipe UUID

**Constraints**:
- User MUST NOT be the creator of the recipe (returns 403 if `recipe.creator_id == claims.UserID`)
- Recipe MUST exist (returns 404 if not found)

**Responses**:

| Status | Body | Condition |
|--------|------|-----------|
| `204 No Content` | — | Favorite added (or already existed) |
| `401 Unauthorized` | `{"error": {"code": "UNAUTHORIZED", "message": "..."}}` | Missing or invalid token |
| `403 Forbidden` | `{"error": {"code": "FORBIDDEN", "message": "cannot favorite your own recipe"}}` | User owns the recipe |
| `404 Not Found` | `{"error": {"code": "NOT_FOUND", "message": "recipe not found"}}` | Recipe does not exist |
| `500 Internal Server Error` | `{"error": {"code": "INTERNAL_ERROR", "message": "..."}}` | Store failure |

---

## DELETE /api/v1/recipes/{id}/favorite

Remove a recipe from the authenticated user's favorites. Idempotent — removing a recipe that was never favorited returns 204 without error.

**Auth**: Required  
**Path param**: `id` — recipe UUID

**Responses**:

| Status | Body | Condition |
|--------|------|-----------|
| `204 No Content` | — | Favorite removed (or did not exist) |
| `401 Unauthorized` | `{"error": {...}}` | Missing or invalid token |
| `500 Internal Server Error` | `{"error": {...}}` | Store failure |

---

## GET /api/v1/recipes/{id}/favorite

Check whether the authenticated user has favorited a specific recipe.

**Auth**: Required  
**Path param**: `id` — recipe UUID

**Response `200 OK`**:
```json
{
  "is_favorite": true
}
```

| Status | Body | Condition |
|--------|------|-----------|
| `200 OK` | `{"is_favorite": bool}` | Always (even if recipe does not exist — returns `false`) |
| `401 Unauthorized` | `{"error": {...}}` | Missing or invalid token |

---

## GET /api/v1/recipes/favorites

List all recipes the authenticated user has favorited, in reverse chronological order (most recently favorited first). Supports pagination.

**Auth**: Required  
**Query params**:

| Param | Type | Default | Max |
|-------|------|---------|-----|
| `page` | int | 1 | — |
| `limit` | int | 20 | 100 |

**Response `200 OK`**:
```json
{
  "data": [
    {
      "id": "abc123",
      "name": "Mojito",
      "ingredients": [...],
      "steps": [...],
      "properties": {},
      "notes": "",
      "creator_id": "user456",
      "created_at": "2026-01-15T10:00:00Z",
      "updated_at": "2026-01-15T10:00:00Z",
      "is_favorite": true
    }
  ],
  "total": 3,
  "page": 1,
  "limit": 20
}
```

Note: `is_favorite: true` is always present on all items in this response (they are all favorites by definition).

| Status | Body | Condition |
|--------|------|-----------|
| `200 OK` | `{data, total, page, limit}` | Success (empty `data: []` if no favorites) |
| `401 Unauthorized` | `{"error": {...}}` | Missing or invalid token |
| `500 Internal Server Error` | `{"error": {...}}` | Store failure |

---

## Route Registration Order (Go stdlib mux)

The Go 1.22+ stdlib mux resolves more specific patterns first. Register in this order:

```go
// Specific paths before wildcards
mux.Handle("GET /api/v1/recipes/favorites",        handler.RequireAuth(...favorites.List))
mux.Handle("PUT /api/v1/recipes/{id}/favorite",    handler.RequireAuth(...favorites.Add))
mux.Handle("DELETE /api/v1/recipes/{id}/favorite", handler.RequireAuth(...favorites.Remove))
mux.Handle("GET /api/v1/recipes/{id}/favorite",    handler.RequireAuth(...favorites.Check))
// Existing routes unchanged
```

`/api/v1/recipes/favorites` must be registered before `GET /api/v1/recipes/{id}` to avoid the wildcard capturing the literal "favorites" segment.
