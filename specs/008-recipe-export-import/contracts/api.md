# API Contracts: Recipe Export and Import

**Feature**: `008-recipe-export-import`  
**Date**: 2026-05-14

All three endpoints require admin authentication: `Authorization: Bearer <admin-jwt>`. Requests without a valid admin token receive `401 UNAUTHORIZED` or `403 FORBIDDEN`.

---

## GET /api/v1/admin/schema

Download the JSON Schema document that defines the recipe structure for import/export.

### Request

```
GET /api/v1/admin/schema
Authorization: Bearer <admin-jwt>
```

No query parameters. No request body.

### Response — 200 OK

```
Content-Type: application/json
Content-Disposition: attachment; filename="recipe-schema.json"
```

Body: the JSON Schema Draft 7 document (see `data-model.md`). Identical on every call.

### Errors

| Status | Code | Condition |
|--------|------|-----------|
| 401 | `UNAUTHORIZED` | Missing or invalid JWT |
| 403 | `FORBIDDEN` | Authenticated but not admin |

---

## GET /api/v1/admin/recipes/export

Export all recipes as a downloadable JSON file.

### Request

```
GET /api/v1/admin/recipes/export
Authorization: Bearer <admin-jwt>
```

No query parameters. No request body.

### Response — 200 OK

```
Content-Type: application/json
Content-Disposition: attachment; filename="recipes-export.json"
```

Body: a JSON array of RecipeExportRecord objects. Each object contains only user-editable fields (`name`, `ingredients`, `steps`, `properties`, `notes`). Server-generated fields (`id`, `creator_id`, `created_at`, `updated_at`) are omitted.

**Empty collection**: returns `[]` (not an error, not `null`).

**Example body**:
```json
[
  {
    "name": "Negroni",
    "ingredients": [
      { "name": "Gin", "quantity": "1", "unit": "oz" },
      { "name": "Campari", "quantity": "1", "unit": "oz" },
      { "name": "Sweet Vermouth", "quantity": "1", "unit": "oz" }
    ],
    "steps": ["Stir with ice for 30 seconds", "Strain into a rocks glass"],
    "properties": { "glass": "Rocks", "garnish": "Orange peel" },
    "notes": "A classic Italian aperitif."
  }
]
```

### Errors

| Status | Code | Condition |
|--------|------|-----------|
| 401 | `UNAUTHORIZED` | Missing or invalid JWT |
| 403 | `FORBIDDEN` | Authenticated but not admin |
| 500 | `INTERNAL_ERROR` | Store failure |

---

## POST /api/v1/admin/recipes/import

Import recipes from a JSON array payload. Validates, skips duplicates, and creates new recipes.

### Request

```
POST /api/v1/admin/recipes/import
Authorization: Bearer <admin-jwt>
Content-Type: application/json
```

Body: a JSON array of RecipeExportRecord objects (same schema as the export file).

Request body is limited to **10 MB** by the handler.

### Response — 200 OK (all validation passed)

```json
{
  "imported": 4,
  "skipped": 1
}
```

`imported`: number of recipes created.  
`skipped`: number of recipes skipped because a recipe with the same name already exists.

### Errors

| Status | Code | Condition |
|--------|------|-----------|
| 400 | `BAD_REQUEST` | Body exceeds 10 MB, is not valid JSON, is not a JSON array, or contains invalid recipe objects (missing `name`, wrong field types). Error `message` identifies the specific problem. |
| 401 | `UNAUTHORIZED` | Missing or invalid JWT |
| 403 | `FORBIDDEN` | Authenticated but not admin |
| 500 | `INTERNAL_ERROR` | Unexpected store error during creation (all created recipes are rolled back) |

**Example error response**:
```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "recipe at index 2: name is required"
  }
}
```

### Behaviour notes

- Validation is performed before any recipes are created. A validation failure (400) means zero recipes are created.
- Duplicate name check occurs per-recipe during the import pass. Duplicates are skipped without error — they are counted in `skipped`.
- Unexpected store errors during creation roll back all creates from this request. The admin receives a 500 error and may retry.
- The `creator_id` of imported recipes is set to the authenticated admin's user ID.

---

## Frontend API Client Functions

New functions to add to `frontend/src/api/client.js`:

```js
// Fetches the schema, returns a Blob for download
export async function downloadRecipeSchema(token) { ... }

// Fetches all recipes as export JSON, returns a Blob for download
export async function exportRecipes(token) { ... }

// Posts the parsed recipes array; returns { imported, skipped } on success
export function importRecipes(recipes, token) {
  return request('POST', '/api/v1/admin/recipes/import', recipes, token);
}
```

`downloadRecipeSchema` and `exportRecipes` need to handle the response as a `Blob` (not JSON) because they trigger file downloads:

```js
async function fetchBlob(path, token) {
  const headers = {};
  if (token) headers['Authorization'] = `Bearer ${token}`;
  const res = await fetch(`${BASE_URL}${path}`, { headers });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: { message: 'Request failed' } }));
    const e = new Error(err.error?.message || 'Request failed');
    e.status = res.status;
    throw e;
  }
  return res.blob();
}
```

File download helper (used in the AdminRecipes page, not in `client.js`):

```js
function triggerDownload(blob, filename) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}
```
