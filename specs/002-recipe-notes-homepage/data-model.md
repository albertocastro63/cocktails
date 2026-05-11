# Data Model: Recipe Notes and Full Homepage Display

**Branch**: `002-recipe-notes-homepage` | **Date**: 2026-05-10

## Changed Entity: Recipe

The `Recipe` entity gains one new optional field. All other fields are unchanged.

### New Field

| Field | Type | Required | Default | Searchable | Description |
|-------|------|----------|---------|------------|-------------|
| `notes` | string | No | `""` (empty) | **No** | Free-form text notes authored by the recipe creator. Excluded from full-text search indexing. Read-publicly; write-creator-only (enforced by existing access control). |

### Updated Recipe Object Shape

```json
{
  "id": "uuid",
  "name": "string",
  "ingredients": [
    { "name": "string", "quantity": "string", "unit": "string" }
  ],
  "steps": ["string"],
  "properties": { "key": "value" },
  "notes": "string",
  "creator_id": "uuid",
  "created_at": "ISO 8601",
  "updated_at": "ISO 8601"
}
```

The `notes` field is always present in responses (empty string when not set). It is omitted from the full-text search index.

### SQLite Schema Delta

```sql
-- Added to existing migrate() function, idempotent:
ALTER TABLE recipes ADD COLUMN notes TEXT NOT NULL DEFAULT '';
```

The `recipes_fts` virtual table is unchanged. The `upsertFTS` function excludes `notes` from `search_text`.

### DynamoDB Schema Delta

No table-level changes required. DynamoDB is schema-flexible; the `notes` attribute is added to the item when present and read back transparently. The `recipeItem` internal struct gains a `Notes` field. The `matchesQuery` function does not check `notes`.

## Unchanged Entities

- `User` — no changes.
- `Ingredient` — no changes.
