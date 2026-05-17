# Research: Recipe Ownership and Per-User Recipe Listing

## Decision 1: Admin Bypass in Ownership Enforcement

**Decision**: In `Update` and `Delete` handlers, check `claims.IsAdmin` before evaluating the `CreatorID` ownership check. If the caller is an admin, skip the ownership check entirely.

**Rationale**: The existing `RequireAdmin` middleware and `claims.IsAdmin` field already provide this signal at zero cost. A single `if claims.IsAdmin { ... }` guard before the ownership check keeps the handler readable and is the standard pattern for role-based bypass in Go HTTP handlers.

**Alternatives considered**:
- Separate admin-specific update/delete endpoints — rejected; doubles the surface area with no benefit.
- Policy object/service layer — rejected; premature abstraction for two call sites.

---

## Decision 2: ListByCreator in DynamoDB

**Decision**: Implement `ListByCreator` as a DynamoDB Scan with a `FilterExpression` on `creator_id`, consistent with the existing `List()` and `Search()` implementations.

**Rationale**: The existing DynamoDB store already uses Scan for all read operations. At the current scale (hundreds of recipes) a filtered scan comfortably meets the p95 ≤ 200 ms target. Adding a GSI would require a Terraform change and a data migration outside the scope of this feature.

**Alternatives considered**:
- GSI on `creator_id` — the production-grade solution; deferred (see Complexity Tracking in plan.md).
- DynamoDB Query via GSI — same as above, deferred.

---

## Decision 3: ListByCreator in SQLite

**Decision**: Implement as a standard `SELECT … WHERE creator_id = ?` with `LIMIT` and `OFFSET`, indexed via the existing `creator_id` column.

**Rationale**: SQLite is the development/local store. A parameterised WHERE clause is the simplest correct implementation and aligns with the existing query patterns in `sqlite/recipes.go`.

**Alternatives considered**: None — no meaningful alternatives for SQLite.

---

## Decision 4: Frontend Ownership Check

**Decision**: Add a `getUserID()` helper to `src/api/auth.js` that decodes the `user_id` claim from the JWT payload (same technique as the existing `isAdmin()` function). Pass a `currentUser` object `{ id, isAdmin }` to `RecipeCard`; show edit/delete only when `currentUser.id === recipe.creator_id || currentUser.isAdmin`.

**Rationale**: The `creator_id` field is already present in every recipe JSON response (it's part of the `model.Recipe` struct). No API change is needed for the ownership check — the client already has all the information it needs. Reusing the JWT decode pattern keeps auth logic co-located in `api/auth.js`.

**Alternatives considered**:
- Server-side `can_edit` boolean in recipe responses — rejected; adds server-side per-caller computation and couples the API response to frontend rendering decisions.
- Separate `/api/v1/me` endpoint — rejected; all claims already in the JWT.

---

## Decision 5: My Recipes Page Structure

**Decision**: `MyRecipes.js` is a page component with the same grid layout as `RecipeList.js` but calls `getMyRecipes()` instead of `getRecipes()`, uses the heading "My Recipes", omits the search bar, and always passes edit/delete visibility as `true` (every listed recipe belongs to the viewer).

**Rationale**: Since every recipe on the My Recipes page is owned by the viewer, there is no need to evaluate ownership per card — edit/delete are always shown. This simplifies the component and avoids passing `currentUser` at all within this page.

**Alternatives considered**:
- Reusing `RecipeList` with a filter prop — rejected; creates conditional rendering complexity in an otherwise simple component.
