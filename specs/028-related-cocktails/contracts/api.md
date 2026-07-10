# Contract: Related Cocktails API

All under the existing `/api/v1` prefix. Auth/permission for writes is unchanged (a user who may edit the recipe), with the clarified rule that relating to **any** cocktail is allowed (no ownership check on the counterpart — FR-016).

## GET /api/v1/recipes/{id} — enriched with related (read)

Existing endpoint; response now includes a read-only `related` array resolved from `related_ids`, sorted alphabetically by name.

```json
{
  "id": "neg-1",
  "name": "Negroni",
  "ingredients": [ ... ],
  "steps": [ ... ],
  "related_ids": ["lh-1", "rh-1"],
  "related": [
    { "id": "lh-1", "name": "Left Hand" },
    { "id": "rh-1", "name": "Right Hand" }
  ]
}
```

- `related` is present only for the single-recipe detail read; it is omitted (or empty) when there are no relations.
- `related` is sorted by `name` (case-insensitive). It is **not** returned by list/search/random endpoints (FR-011).

## POST /api/v1/recipes and PUT /api/v1/recipes/{id} — accept related_ids (write)

The create/update body accepts an optional `related_ids` array of recipe ids.

```json
{ "name": "Negroni", "ingredients": [], "steps": [], "related_ids": ["lh-1", "rh-1"] }
```

Semantics:
- **Absent** `related_ids` on PUT → relations left unchanged (partial update).
- **Present** (including `[]`) → relations set to exactly this set for the recipe, and the symmetric counterparts are reconciled (added/removed).
- Server normalizes: dedupe, drop the recipe's own id, drop ids that don't resolve to an existing recipe.
- Response is the saved recipe (as today). Symmetric effects on counterparts are applied server-side.

**Rules**
- **R1 (symmetry)**: After the write, every id in the recipe's `related_ids` has this recipe's id in its own `related_ids`.
- **R2 (no self)**: The recipe's own id never appears in its `related_ids`.
- **R3 (dedupe)**: No id appears twice.
- **R4 (non-transitive)**: Only the requested pairs change; no indirect relations are created.

## DELETE /api/v1/recipes/{id} — cascades relation cleanup

Existing endpoint; additionally removes `{id}` from the `related_ids` of every recipe it was related to (FR-014). No response body change.

## GET /api/v1/recipes/names — name list for the picker (read, public)

New lightweight endpoint returning every recipe's id and name for the client-side type-ahead.

```json
[
  { "id": "neg-1", "name": "Negroni" },
  { "id": "lh-1",  "name": "Left Hand" },
  { "id": "rh-1",  "name": "Right Hand" }
]
```

- Public (no auth), read-only, minimal payload (id + name only).
- The picker filters this list case-insensitively by substring and excludes the current recipe and already-selected ids.
