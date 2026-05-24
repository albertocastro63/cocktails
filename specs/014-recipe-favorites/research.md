# Research: Recipe Favorites

**Feature**: 014-recipe-favorites
**Date**: 2026-05-22

---

## Decision 1: Favorites Storage Architecture

**Decision**: Dedicated `cocktails-favorites` DynamoDB table with composite key `(user_id PK, recipe_id SK)`.

**Rationale**: Favorites are a many-to-many join between users and recipes. A separate table isolates this relationship cleanly without polluting the recipe or user items. The composite key enables:
- O(1) lookup: "has user X favorited recipe Y?" → `GetItem(user_id, recipe_id)`
- O(n) list by user: "all favorites for user X" → `Query(PK=user_id)`
- Idempotent add: `PutItem` with `attribute_not_exists(user_id)` condition (or unconditional put, same result)

**Alternatives considered**:
- Storing favorites as a list inside the User item (rejected — list size is unbounded; DynamoDB item limit is 400 KB; also complicates concurrent writes).
- Storing a `favorited_by` set on the Recipe item (rejected — same unbounded growth issue; also creates a write hotspot on popular recipes).

---

## Decision 2: SQLite Schema for Local Dev

**Decision**: Add a `favorites` table to the SQLite schema: `(user_id TEXT, recipe_id TEXT, created_at TEXT, PRIMARY KEY (user_id, recipe_id))`.

**Rationale**: The project maintains SQLite parity for local development. The table mirrors the DynamoDB access patterns. A composite PK enforces the one-favorite-per-user-per-recipe constraint at the DB level.

**Alternatives considered**: None — SQLite parity is required by the existing architecture.

---

## Decision 3: API Shape

**Decision**: Four new REST endpoints:

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| `PUT` | `/api/v1/recipes/{id}/favorite` | Required | Add to favorites (idempotent) |
| `DELETE` | `/api/v1/recipes/{id}/favorite` | Required | Remove from favorites (idempotent) |
| `GET` | `/api/v1/recipes/{id}/favorite` | Required | Check if user has favorited this recipe |
| `GET` | `/api/v1/recipes/favorites` | Required | List all recipes the user has favorited |

**Rationale**: Using `PUT` (not `POST`) for add makes the operation idempotent — repeated calls do not create duplicate favorites. Using a sub-resource path `/{id}/favorite` is RESTful (the favorite is a resource on a recipe). A separate `GET /recipes/favorites` avoids mutating the existing `GET /recipes/mine` contract.

**Alternatives considered**:
- Single `POST /recipes/{id}/favorite` toggle (rejected — non-idempotent toggles are error-prone when clients retry on network failure; separate PUT/DELETE are more predictable).
- Embedding `is_favorited` in the recipe response (rejected — `GET /recipes/{id}` is unauthenticated; adding conditional auth logic to it adds complexity with no benefit; a separate lightweight call is cleaner).
- Extending `GET /recipes/mine` to return both created and favorited recipes (rejected — breaks the single-responsibility of that endpoint and complicates the response shape; the frontend merges the two lists itself).

---

## Decision 4: Frontend Data Merge on My Recipes Page

**Decision**: `MyRecipes` fetches `GET /recipes/mine` and `GET /recipes/favorites` in parallel, merges results in the browser, deduplicates by recipe ID (created recipes take precedence — shown without heart badge), and renders a single unified list where favorited-only recipes carry a `isFavorite: true` prop that triggers a heart badge on `RecipeCard`.

**Rationale**: Keeping the two fetch calls separate preserves the simplicity of each endpoint. The merge logic is trivial (two arrays → Map → deduplicate) and runs in < 1 ms for realistic recipe counts. The heart badge is a pure display concern, so extending `RecipeCard` with an `isFavorite` prop is the minimal change.

**Alternatives considered**:
- Single backend endpoint that returns a merged list (rejected — would require a new join query, more backend complexity, and breaks the separation between "created" and "favorited" semantics).

---

## Decision 5: Heart Icon — SVG vs Emoji vs Icon Library

**Decision**: Inline SVG heart path, styled with Tailwind `text-red-500` (favorited) and `text-stone-400` (unfavorited). No additional icon library dependency.

**Rationale**: The project uses no icon library today. Adding one (Heroicons, Lucide) for a single icon is over-engineering. An inline SVG `<path>` for a heart is ~120 characters, accessible with `aria-label`, and fully styleable via Tailwind color classes. The amber/stone design system uses red as the status/danger color; `text-red-500` for a favorited heart is semantically correct and visually clear.

**Alternatives considered**:
- Unicode heart emoji `♥` (rejected — inconsistent cross-platform rendering, cannot be reliably styled with CSS color).
- Heroicons library (rejected — extra dependency for one icon).

---

## Decision 6: Favorite Count (P3)

**Decision**: Defer to a follow-up increment. The `CountByRecipe` store method will be defined in the interface but left unimplemented (returns 0) in both DynamoDB and SQLite stores until P3 is prioritized.

**Rationale**: P3 is explicitly deferrable. A count requires scanning or a GSI query — both are more expensive than the P1/P2 operations. Stubbing the interface method now avoids a breaking interface change later without blocking the MVP.

**Alternatives considered**:
- DynamoDB GSI on `recipe_id` for efficient `CountByRecipe` (viable for P3, not needed for P1/P2).
- Scan with filter expression on `recipe_id` (acceptable for small datasets, no Terraform changes needed — suitable for P3 when prioritized).
