# Research: Related Cocktails

Spec ambiguities were resolved in `/speckit-clarify` (any editor may relate to any cocktail; case-insensitive substring name search; alphabetical display order). This document records the technical decisions.

## Decision 1 — Relation storage: a `related_ids` set on the recipe item

**Decision**: Store relations as a set of related recipe IDs (`related_ids`) directly on each recipe. No separate relations table/GSI. Symmetry is a data invariant maintained on write (Decision 2).

**Rationale**: The app is DynamoDB-first with a single-item recipe model; relations are few per recipe and always read alongside the recipe. Co-locating the set makes reads trivial (already have it when loading the recipe) and avoids a new table/GSI or query. SQLite mirrors this with a JSON `related_ids TEXT` column, exactly like the existing `garnishes`/`steps` encoding.

**Alternatives considered**:
- Dedicated `relations` table with one row per undirected pair: normalized and self-symmetric, but adds a table, a GSI for "relations of X", and a join on every recipe read — overkill at this scale and inconsistent with the current single-item style.
- Directed edges only (store A→B, derive reverse at read): breaks the "reverse is recorded" requirement and complicates deletion cleanup.

## Decision 2 — Symmetric reconciliation on save (`SetRelated`)

**Decision**: A single store method `SetRelated(recipeID string, relatedIDs []string) error` owns the invariant. It (1) normalizes the requested set — dedupe, drop self, drop IDs that don't resolve to an existing recipe; (2) diffs against the recipe's current set to get `added`/`removed`; (3) writes the recipe's new set; (4) for each `added` counterpart, adds `recipeID` to its set; (5) for each `removed` counterpart, removes `recipeID`. The recipe Create/Update handlers call it after the normal upsert.

**Rationale**: Centralizes the two-sided write in one tested place so handlers stay thin and every entry point (create, update) keeps relations symmetric. Diffing against the stored set makes it idempotent and correct for partial edits.

**Alternatives considered**:
- Reconcile in the handler: spreads cross-item writes into HTTP code and is harder to test.
- DynamoDB `TransactWriteItems` for all-or-nothing updates: stronger atomicity, but adds complexity; at this scale sequential writes (self first, then counterparts) are acceptable. If a counterpart write fails, the operation returns an error and can be retried; documented as a tradeoff.

## Decision 3 — Deletion cleanup via `Delete`

**Decision**: Extend `RecipeStore.Delete(id)` to first load the recipe, remove `id` from each of its counterparts' `related_ids` (the symmetric guarantee means exactly those recipes reference it), then delete the recipe. Bounded by the recipe's own relation count.

**Rationale**: Because relations are symmetric, a recipe's own `related_ids` is the complete list of recipes that point back at it — so cleanup is precise and cheap (no full scan). Encapsulating it in `Delete` guarantees FR-014 regardless of caller.

**Alternatives considered**: A scan for "any recipe whose related_ids contains id" — unnecessary given symmetry, and a full-table scan.

## Decision 4 — Non-transitivity is inherent

**Decision**: No propagation logic. Only explicitly requested pairs are ever written; the system never derives A–C from A–B and B–C.

**Rationale**: Non-transitivity (FR-003) is satisfied by construction — there is no code path that would infer indirect relations. Covered by an explicit test.

## Decision 5 — Client-side type-ahead over a names endpoint

**Decision**: Add `GET /api/v1/recipes/names` returning `[{id, name}]` for all recipes (public, lightweight). The picker fetches it once on mount and filters case-insensitively by substring in the browser as the user types — no per-keystroke network calls. Suggestions exclude the current recipe and already-selected IDs.

**Rationale**: At hundreds of recipes the full name list is tiny; client-side filtering is instant, needs no debounce, and keeps the keyboard interaction snappy (SC-005). A dedicated projection avoids shipping full recipe bodies.

**Alternatives considered**:
- Per-keystroke search endpoint: more requests, needs debounce, slower feel; unnecessary at this scale.
- Reuse `GET /api/v1/recipes` (full bodies, paginated): heavier payloads and pagination complicate client filtering.

## Decision 6 — Detail response enriches related IDs to `{id, name}`

**Decision**: `GET /api/v1/recipes/{id}` includes a read-only `related: [{id, name}]` array, resolved from `related_ids` and **sorted alphabetically by name** (FR-017). The stored `related_ids` remains an unordered set. The model carries a transient `Related []RelatedRef` field (`dynamodbav:"-"`, `json:"related,omitempty"`) populated only on read.

**Rationale**: Lets the detail page render links (name + id) without a second request or client-side name resolution, and centralizes the alphabetical ordering server-side. Transient field keeps storage clean.

**Alternatives considered**: Return only `related_ids` and have the detail page resolve names from the names endpoint — an extra request and duplicated sorting logic on the client.

## Decision 7 — Accessible combobox pattern

**Decision**: Build `RelatedCocktailPicker` as an ARIA combobox: a text `input` with `role="combobox"`, a `listbox` of options, `aria-activedescendant` for the highlighted option, Arrow Up/Down to move, Enter to add, Escape to close. Added relations render as removable chips with accessible remove buttons. Reuses the app's stone/amber styling.

**Rationale**: Meets WCAG 2.1 AA and constitution III (keyboard-operable, screen-reader-friendly), matching the spec's "type, arrow to select" interaction.

**Alternatives considered**: A native `<datalist>` — simpler but inconsistent styling/behavior across browsers and weaker control over keyboard selection and chips.
