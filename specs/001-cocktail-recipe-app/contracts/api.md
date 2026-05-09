# API Contract: Cocktail Recipe App

**Version**: v1 | **Date**: 2026-05-08 | **Branch**: `001-cocktail-recipe-app`

All endpoints are prefixed with `/api/v1`. Requests and responses use `application/json`.  
Authentication uses `Authorization: Bearer <jwt>` where required.

---

## Authentication

### POST `/api/v1/auth/login`

Authenticate a user and receive a JWT.

**Request**:
```json
{
  "username": "string (required)",
  "password": "string (required)"
}
```

**Response `200 OK`**:
```json
{
  "token": "string (JWT)",
  "expires_at": "ISO 8601 timestamp"
}
```

**Errors**:
- `400 Bad Request` — missing or empty fields
- `401 Unauthorized` — invalid credentials

---

## Recipes — Public (no auth required)

### GET `/api/v1/recipes`

List all recipes, optionally filtered by a search query.

**Query parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `q` | string | No | Free-text search across all fields (name, ingredients, steps, properties) |
| `page` | integer | No | Page number, 1-based. Default: 1 |
| `limit` | integer | No | Results per page. Default: 20, max: 100 |

**Response `200 OK`**:
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "string",
      "ingredients": [
        { "name": "string", "quantity": "string", "unit": "string" }
      ],
      "steps": ["string"],
      "properties": { "key": "value" },
      "creator_id": "uuid",
      "created_at": "ISO 8601",
      "updated_at": "ISO 8601"
    }
  ],
  "total": "integer",
  "page": "integer",
  "limit": "integer"
}
```

**Errors**:
- `400 Bad Request` — invalid query parameters

---

### GET `/api/v1/recipes/random`

Return one randomly selected recipe.

**Response `200 OK`**: Single recipe object (same shape as items in the list response).

**Response `204 No Content`**: No recipes exist in the system (FR-011).

---

### GET `/api/v1/recipes/{id}`

Return a single recipe by ID.

**Response `200 OK`**: Single recipe object.

**Errors**:
- `404 Not Found` — recipe does not exist

---

## Recipes — Authenticated (JWT required)

### POST `/api/v1/recipes`

Create a new recipe. `creator_id` is set from the authenticated user's identity.

**Request**:
```json
{
  "name": "string (required, non-empty)",
  "ingredients": [
    { "name": "string (required)", "quantity": "string (required)", "unit": "string (optional)" }
  ],
  "steps": ["string"],
  "properties": { "key": "value" }
}
```

**Response `201 Created`**:
```json
{
  "data": { /* full recipe object */ },
  "warnings": ["string"]
}
```

The `warnings` array is non-empty when a recipe with the same name already exists (FR-017). The recipe is saved regardless.

**Errors**:
- `400 Bad Request` — validation failure (missing name, malformed body)
- `401 Unauthorized` — missing or invalid JWT

---

### PUT `/api/v1/recipes/{id}`

Update an existing recipe. Partial updates are supported — omitted fields retain their current values. The `creator_id`, `id`, and `created_at` fields are immutable and ignored if included.

**Request**: Same shape as POST; all fields optional.

**Response `200 OK`**: Updated recipe object.

**Errors**:
- `400 Bad Request` — validation failure
- `401 Unauthorized` — missing or invalid JWT
- `403 Forbidden` — authenticated user is not the recipe's creator
- `404 Not Found` — recipe does not exist

---

### DELETE `/api/v1/recipes/{id}`

Delete a recipe. Only the recipe's creator may delete it (FR-014).

**Response `204 No Content`**: Recipe deleted.

**Errors**:
- `401 Unauthorized` — missing or invalid JWT
- `403 Forbidden` — authenticated user is not the recipe's creator
- `404 Not Found` — recipe does not exist

---

## Admin — Admin JWT required (`is_admin: true` claim)

### POST `/api/v1/admin/users`

Create a new user account. Only callable by an admin (FR-013).

**Request**:
```json
{
  "username": "string (required, 3–50 chars, alphanumeric/hyphen/underscore)",
  "password": "string (required, ≥ 8 chars)",
  "is_admin": "boolean (optional, default false)"
}
```

**Response `201 Created`**:
```json
{
  "id": "uuid",
  "username": "string",
  "is_admin": "boolean",
  "created_at": "ISO 8601"
}
```

`password_hash` is never returned.

**Errors**:
- `400 Bad Request` — validation failure
- `401 Unauthorized` — missing or invalid JWT
- `403 Forbidden` — authenticated user is not an admin
- `409 Conflict` — username already exists

---

## Error Response Shape

All error responses follow this format:
```json
{
  "error": {
    "code": "string (machine-readable, e.g. NOT_FOUND, UNAUTHORIZED)",
    "message": "string (human-readable)"
  }
}
```

Stack traces and internal error details are never included in responses (Constitution Principle III).

---

## Pagination Notes

- `page` and `limit` apply to `/api/v1/recipes` only.
- `GET /api/v1/recipes/random` always returns exactly one recipe (or 204).
- Results are ordered by `created_at` descending unless a search query is active, in which case relevance ordering is used.

---

## Versioning

The `/v1` prefix reserves the right to introduce `/v2` endpoints without breaking existing consumers (FR-010 specifies the external interface must remain stable).

---

## curl Usage Examples

```bash
BASE=http://localhost:8080

# List all recipes (paginated)
curl "$BASE/api/v1/recipes"

# Search recipes by any field
curl "$BASE/api/v1/recipes?q=lime"
curl "$BASE/api/v1/recipes?q=tequila&page=1&limit=10"

# Get a random recipe (204 when DB is empty)
curl "$BASE/api/v1/recipes/random"

# Get a recipe by ID
curl "$BASE/api/v1/recipes/<id>"

# Log in and capture JWT
TOKEN=$(curl -s -X POST "$BASE/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"yourpassword"}' | jq -r .token)

# Create a recipe (requires auth)
curl -X POST "$BASE/api/v1/recipes" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Mojito",
    "ingredients": [{"name":"rum","quantity":"50","unit":"ml"},{"name":"mint","quantity":"10","unit":"leaves"}],
    "steps": ["Muddle mint with sugar","Add rum and lime juice","Top with soda water"],
    "properties": {"style":"refreshing","base_spirit":"rum","garnish":"lime wedge"}
  }'

# Update a recipe (partial — only fields provided are changed)
curl -X PUT "$BASE/api/v1/recipes/<id>" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"properties":{"occasion":"Summer party"}}'

# Delete a recipe
curl -X DELETE "$BASE/api/v1/recipes/<id>" \
  -H "Authorization: Bearer $TOKEN"

# Create a user account (admin only)
curl -X POST "$BASE/api/v1/admin/users" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"username":"alice","password":"s3cur3pass"}'
```
