# API Contract: Multi-Ingredient Search

**Feature**: 015-multi-ingredient-search
**Date**: 2026-05-24

## Modified Endpoint

### GET /api/v1/recipes

The existing list/search endpoint is extended to support compound ingredient queries. The request and
response shapes are unchanged — only the server-side matching semantics change for compound `q` values.

#### Request

```
GET /api/v1/recipes?q={query}&page={page}&limit={limit}
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `q` | string | No | Search query. Single term → full-text search (existing). Compound (see below) → ingredient AND search. |
| `page` | int | No | Page number, default 1. |
| `limit` | int | No | Results per page, default 20, max 100. |

#### Multi-ingredient `q` Syntax

| Form | Example | Parsed tokens |
|------|---------|---------------|
| Natural language | `gin and lemon juice` | `["gin", "lemon juice"]` |
| Shorthand (spaces) | `gin + lemon + sugar` | `["gin", "lemon", "sugar"]` |
| Shorthand (no spaces) | `gin+lemon+sugar` | `["gin", "lemon", "sugar"]` |
| Mixed case | `Gin AND Lemon` | `["gin", "lemon"]` |
| Single term (unchanged) | `mojito` | existing full-text search |
| Empty | `` | all recipes (existing) |

Splitting rules (applied in order):
1. Split on `\s+and\s+` (case-insensitive) if present.
2. Split results on `\s*\+\s*`.
3. Trim whitespace from each token; discard empty tokens.
4. If ≤ 1 token remains → existing `Search` path. If ≥ 2 tokens → `SearchByIngredients` path.

#### Matching Semantics (multi-ingredient)

A recipe is included in results when **every** token appears as a case-insensitive substring
of **at least one** of its ingredient names. Token `"gin"` matches ingredient names `"Gin"`,
`"Sloe Gin"`, `"Gin Fizz Base"` but not recipe steps or the recipe name.

#### Response (unchanged)

```json
{
  "data": [
    {
      "id": "string",
      "name": "string",
      "ingredients": [
        { "name": "string", "quantity": "string", "unit": "string", "is_base_spirit": false }
      ],
      "steps": ["string"],
      "properties": { "key": "value" },
      "notes": "string",
      "creator_id": "string",
      "created_at": "2026-05-24T00:00:00Z",
      "updated_at": "2026-05-24T00:00:00Z"
    }
  ],
  "total": 42,
  "page": 1,
  "limit": 20
}
```

#### Error Responses (unchanged)

| Status | Code | Description |
|--------|------|-------------|
| 500 | `INTERNAL_ERROR` | Store query failed |

No new error codes are introduced. An unrecognisable query string is treated as a single-term
search (degraded gracefully), never as an error.

## No New Endpoints

This feature does not introduce any new API endpoints.
