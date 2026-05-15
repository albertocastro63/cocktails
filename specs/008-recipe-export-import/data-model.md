# Data Model: Recipe Export and Import

**Feature**: `008-recipe-export-import`  
**Date**: 2026-05-14

---

## Existing Entities (unchanged)

### Recipe (`model.Recipe`)

No changes to the canonical data model. All fields already exist.

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `id` | string (UUID) | server-generated | Not included in export/import |
| `name` | string | yes | Uniqueness checked on import by name comparison |
| `ingredients` | `[]Ingredient` | no | Ordered list |
| `steps` | `[]string` | no | Ordered list of preparation steps |
| `properties` | `map[string]string` | no | Named text-value pairs |
| `notes` | string | no | Free-form text (markdown) |
| `creator_id` | string (UUID) | server-generated | Not included in export/import; set to importing admin's user ID on import |
| `created_at` | time.Time | server-generated | Not included in export/import |
| `updated_at` | time.Time | server-generated | Not included in export/import |

### Ingredient (`model.Ingredient`)

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `name` | string | yes | Ingredient name |
| `quantity` | string | no | Amount (e.g., "1.5") |
| `unit` | string | no | Unit of measure (e.g., "oz") |

---

## New Interface Types

These types exist only at the API boundary (HTTP request/response and JSON Schema). They are not persisted as separate entities.

### RecipeExportRecord (export/import wire format)

Represents one recipe in the export file or import payload. Contains only user-editable fields.

| Field | Type | Required | Schema type |
|-------|------|----------|-------------|
| `name` | string | **yes** | `string` |
| `ingredients` | array | no | `array` of `IngredientRecord` |
| `steps` | array | no | `array` of `string` |
| `properties` | object | no | `object` with `string` values |
| `notes` | string | no | `string` |

### IngredientRecord (within RecipeExportRecord)

| Field | Type | Required | Schema type |
|-------|------|----------|-------------|
| `name` | string | **yes** | `string` |
| `quantity` | string | no | `string` |
| `unit` | string | no | `string` |

### ImportResult (import response body)

Returned by `POST /api/v1/admin/recipes/import` on success.

| Field | Type | Description |
|-------|------|-------------|
| `imported` | int | Number of recipes created |
| `skipped` | int | Number of recipes skipped due to name conflicts |

---

## Store Interface Extensions

Two new methods on `RecipeStore`:

```
ListAll() ([]*model.Recipe, error)
  - Returns all recipes in the store without pagination
  - Used by the export handler
  - SQLite: SELECT without LIMIT/OFFSET
  - DynamoDB: paginated Scan (auto-follows LastEvaluatedKey)

ImportBatch(recipes []*model.Recipe, creatorID string) (created int, skipped int, err error)
  - Attempts to create each recipe whose name does not already exist
  - Skips (does not create) recipes whose name matches an existing recipe
  - On unexpected error: rolls back all creates in this batch (all-or-nothing)
  - SQLite: wraps all inserts in a single sql.Tx; rollback on any error
  - DynamoDB: sequential writes with compensation (delete created items on error)
  - Returns counts of created and skipped recipes
```

---

## JSON Schema Document

The schema document downloaded via `GET /api/v1/admin/schema` describes the RecipeExportRecord above as JSON Schema Draft 7. It is a static document embedded in the binary.

```json
{
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
    }
  }
}
```

---

## Identity and Uniqueness

- **Import uniqueness check**: Recipe names are compared case-sensitively against existing recipe names using `ExistsByName(name string)` (already in `RecipeStore`). A name that differs only in case is treated as a distinct recipe.
- **No deduplication within the import file itself**: If the import file contains two objects with the same `name`, the first is created and the second is treated as a duplicate (skipped).

---

## Validation Rules

Applied by the import handler before any DB writes begin:

| Rule | Error |
|------|-------|
| Body must be a valid JSON array | `"import file must be a JSON array"` |
| Each element must be a JSON object | `"each recipe must be a JSON object"` |
| Each element must have a non-empty `name` string | `"recipe at index N: name is required"` |
| `ingredients[*].name` must be a string if present | `"recipe at index N: ingredient name must be a string"` |
| `steps[*]` must be strings if present | `"recipe at index N: steps must be strings"` |
| `properties` values must be strings if present | `"recipe at index N: property values must be strings"` |
| `notes` must be a string if present | `"recipe at index N: notes must be a string"` |
