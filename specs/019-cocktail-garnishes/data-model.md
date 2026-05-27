# Data Model: Cocktail Recipe Garnishes

**Feature**: 019-cocktail-garnishes  
**Date**: 2026-05-26

## Entities

### Recipe (extended)

New field added to the existing `Recipe` entity.

| Field      | Type       | Nullable / Default | Constraints                              |
|------------|------------|-------------------|------------------------------------------|
| `garnishes`| `[]string` | Optional, `[]`    | Ordered; each entry non-empty after trim |

No other fields are changed.

### Garnish (value type, not an entity)

A garnish is a plain string — a free-text description of a garnish instruction (e.g., `"Express orange oil over the cocktail"`). It has no independent identity; it exists only as an element of `Recipe.Garnishes`.

## Validation Rules

- Each garnish string MUST be non-empty after whitespace trimming.
- Blank/whitespace-only entries MUST NOT be persisted (filtered at the handler layer before storage).
- `nil` and `[]` are equivalent — no garnishes.
- No upper bound is enforced on the count.

## Storage

### SQLite

New column on the `recipes` table, added via idempotent `ALTER TABLE`:

```sql
ALTER TABLE recipes ADD COLUMN garnishes TEXT NOT NULL DEFAULT '[]';
```

Stored as a JSON array of strings (same pattern as `steps`). Empty garnish list → `'[]'`.  
All SELECT queries, INSERT, UPDATE, and scan functions must include this column.

### DynamoDB

New attribute on the `recipeItem` struct:

```go
Garnishes []string `dynamodbav:"garnishes"`
```

Existing items without the attribute deserialise to `nil` (treated as empty). No migration required.

## API Wire Format

### Recipe response (GET /recipes/:id, GET /recipes)

```json
{
  "id": "...",
  "name": "...",
  "ingredients": [...],
  "steps": [...],
  "garnishes": ["Express orange oil over the cocktail", "Use orange peel to garnish"],
  "properties": {},
  "notes": "",
  "creator_id": "...",
  "created_at": "...",
  "updated_at": "..."
}
```

`garnishes` is omitted from the JSON response when empty (`omitempty`).

### Recipe create/update request body

```json
{
  "name": "...",
  "ingredients": [...],
  "steps": [...],
  "garnishes": ["Express orange oil over the cocktail"],
  "properties": {},
  "notes": ""
}
```

`garnishes` is optional; absent or `null` leaves garnishes unchanged on update.

### Export format

```json
[
  {
    "name": "...",
    "ingredients": [...],
    "steps": [...],
    "garnishes": ["Express orange oil over the cocktail"],
    "properties": {},
    "notes": ""
  }
]
```

## Frontend Data Flow

```
RecipeForm (edit/create)
  → collects garnishes[] from garnish section inputs
  → filters empty strings
  → passes garnishes[] in POST/PUT payload

RecipeDetail (read)
  → receives recipe.garnishes from API
  → renders <em> list below ingredients when garnishes.length > 0

RecipeCard popover (hover preview)
  → receives recipe.ingredients and recipe.garnishes
  → shows garnish section only when ingredients.length < MAX_VISIBLE (5)
  → renders each garnish in <em>
```

## Backward Compatibility

- Existing recipes with no `garnishes` data: field absent in DynamoDB (nil slice), `'[]'` in SQLite → both treated as empty list → no garnishes section rendered.
- Existing export files without `garnishes` key: import handler ignores absent optional key → recipe imported with empty garnish list. No data loss; no error.
- Existing tests: no breakage expected; new field has zero value when absent.
