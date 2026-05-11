# API Contract Delta: Recipe Notes

**Version**: v1 delta | **Date**: 2026-05-10 | **Branch**: `002-recipe-notes-homepage`  
**Base contract**: `specs/001-cocktail-recipe-app/contracts/api.md`

This document describes only the changes to the existing API contract. All endpoints, error shapes, pagination rules, and authentication requirements defined in the base contract remain unchanged.

---

## Changed: Recipe Object Schema

The recipe object returned by all recipe endpoints gains a `notes` field.

### Before

```json
{
  "id": "uuid",
  "name": "string",
  "ingredients": [...],
  "steps": ["string"],
  "properties": { "key": "value" },
  "creator_id": "uuid",
  "created_at": "ISO 8601",
  "updated_at": "ISO 8601"
}
```

### After

```json
{
  "id": "uuid",
  "name": "string",
  "ingredients": [...],
  "steps": ["string"],
  "properties": { "key": "value" },
  "notes": "string",
  "creator_id": "uuid",
  "created_at": "ISO 8601",
  "updated_at": "ISO 8601"
}
```

`notes` is always present (empty string `""` when not set). This is a **non-breaking additive change** — consumers that ignore unknown fields are unaffected.

---

## Changed: POST `/api/v1/recipes` — Request Body

The request body accepts an optional `notes` field.

```json
{
  "name": "string (required, non-empty)",
  "ingredients": [...],
  "steps": ["string"],
  "properties": { "key": "value" },
  "notes": "string (optional)"
}
```

If `notes` is omitted, it defaults to `""`.

---

## Changed: PUT `/api/v1/recipes/{id}` — Request Body

The request body accepts an optional `notes` field for partial update.

```json
{
  "name": "string (optional)",
  "ingredients": [...],
  "steps": ["string"],
  "properties": { "key": "value" },
  "notes": "string (optional)"
}
```

If `notes` is **omitted**, the existing notes value is preserved (consistent with all other optional fields).  
If `notes` is set to `""`, the existing notes value is cleared.

---

## Search Behaviour (unchanged contract, clarified)

The `q` parameter on `GET /api/v1/recipes` searches across: name, ingredient names, steps, and property values.  
**Notes are excluded from search** — a query matching only a recipe's notes content returns no results for that recipe.

---

## curl Examples (additions)

```bash
BASE=http://localhost:8080

# Create a recipe with notes
curl -X POST "$BASE/api/v1/recipes" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Negroni",
    "ingredients": [{"name":"gin","quantity":"30","unit":"ml"}],
    "steps": ["Stir over ice", "Strain into glass"],
    "properties": {"base_spirit":"gin"},
    "notes": "Try with a barrel-aged gin for a smoother finish."
  }'

# Update only the notes (all other fields preserved)
curl -X PUT "$BASE/api/v1/recipes/<id>" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"notes": "Updated tasting notes."}'

# Clear notes
curl -X PUT "$BASE/api/v1/recipes/<id>" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"notes": ""}'
```
