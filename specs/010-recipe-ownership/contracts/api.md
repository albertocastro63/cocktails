# API Contracts: Recipe Ownership

## Changed Endpoints

### PUT /api/v1/recipes/{id}

**Auth**: Bearer token required  
**Change**: Admin bypass added. If the caller's JWT has `is_admin: true`, the ownership check is skipped.

**Authorization logic**:
```
if !claims.IsAdmin && existing.CreatorID != claims.UserID && existing.CreatorID != "" {
    → 403 FORBIDDEN
}
```

**Response — 403 Forbidden** (non-owner, non-admin):
```json
{ "error": { "code": "FORBIDDEN", "message": "only the recipe creator can edit this recipe" } }
```

---

### DELETE /api/v1/recipes/{id}

**Auth**: Bearer token required  
**Change**: Admin bypass added. Same logic as PUT above.

**Response — 403 Forbidden** (non-owner, non-admin):
```json
{ "error": { "code": "FORBIDDEN", "message": "only the recipe creator can delete this recipe" } }
```

---

## New Endpoint

### GET /api/v1/recipes/mine

Returns the paginated list of recipes created by the authenticated user.

**Auth**: Bearer token required  
**Route**: Must be registered before `GET /api/v1/recipes/{id}` to avoid the `{id}` wildcard matching `"mine"`.

**Query Parameters**:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| page      | int  | 1       | Page number (1-based) |
| limit     | int  | 20      | Results per page (max 100) |

**Response — 200 OK**:
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Negroni",
      "ingredients": [],
      "steps": [],
      "properties": {},
      "notes": "",
      "creator_id": "user-uuid",
      "created_at": "2026-05-17T10:00:00Z",
      "updated_at": "2026-05-17T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

**Response — 401 Unauthorized** (missing/invalid token):
```json
{ "error": { "code": "UNAUTHORIZED", "message": "authentication required" } }
```

**Response — 200 OK, empty** (no owned recipes):
```json
{ "data": [], "total": 0, "page": 1, "limit": 20 }
```
