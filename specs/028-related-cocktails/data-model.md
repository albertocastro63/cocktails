# Data Model: Related Cocktails

## Entity changes

### Recipe (existing) — new fields

| Field | Type | Stored | JSON | Notes |
|-------|------|--------|------|-------|
| `related_ids` | set of string (recipe IDs) | yes | `related_ids` | The recipe's related cocktails. Unordered set; unique; excludes own id. Empty by default. |
| `related` | list of `{id, name}` | **no** (transient) | `related` (omitempty) | Read-only, populated by `GET /recipes/{id}`, sorted alphabetically by name (FR-017). Never persisted (`dynamodbav:"-"`). |

- **DynamoDB**: `related_ids` is an attribute on the existing recipes table item (list/string-set of ids). No new table or GSI.
- **SQLite**: `related_ids TEXT NOT NULL DEFAULT '[]'` column added via idempotent `ALTER TABLE` migration; JSON-encoded, like `garnishes`/`steps`.

### Cocktail Relation (conceptual)

An **undirected, symmetric** association between two **distinct** recipes, represented implicitly by each recipe carrying the other's id in `related_ids`.

- **Identity**: the unordered pair `{idA, idB}`. At most one relation per pair.
- **Invariants**:
  - **Symmetry**: `idB ∈ A.related_ids` ⟺ `idA ∈ B.related_ids`.
  - **Irreflexive**: `idA ∉ A.related_ids` (no self-relation).
  - **No duplicates**: set semantics.
  - **Referential**: every id in a `related_ids` set resolves to an existing recipe; when a recipe is deleted, its id is removed from all sets.
  - **Non-transitive**: `{A,B}` and `{B,C}` imply nothing about `{A,C}`.

## Operations (store layer)

| Operation | Behavior |
|-----------|----------|
| `SetRelated(recipeID, requested []string)` | Normalize requested (dedupe, drop self, drop non-existent). Diff vs. stored set → `added`/`removed`. Write recipe's new set. For each `added`: add `recipeID` to counterpart. For each `removed`: remove `recipeID` from counterpart. Maintains symmetry + dedupe + irreflexivity. |
| `Delete(id)` (extended) | Load recipe; remove `id` from each counterpart in its `related_ids`; then delete the recipe. Guarantees no dangling references (FR-014). |
| `ListAll` / names projection | Provides `[{id, name}]` for the type-ahead (Decision 5). |
| `GetByID(id)` (read enrichment done in handler) | Recipe returned with `related` resolved to `{id,name}` sorted by name. |

## Validation rules (from requirements)

- A requested relation to a non-existent recipe id is dropped silently (only existing cocktails can be related — Assumptions).
- A requested relation to the recipe's own id is dropped (FR-007/FR-008).
- Duplicate ids in the request collapse to one (FR-008).
- Removing an id from a recipe's set removes the reverse (FR-013).

## State / lifecycle

Relations have no independent lifecycle; they exist as long as both recipes exist and both sets contain each other. Creating/removing is always a paired write; deleting a recipe tears down all of its relations.
