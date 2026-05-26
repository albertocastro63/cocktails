# Quickstart: Base Spirit Search Filter

**Feature**: 018-base-spirit-search | **Date**: 2026-05-25

---

## Prerequisites

- `npm run dev` serving the frontend (default `http://localhost:5173`)
- Backend API running (local or dev environment)
- At least 3 recipes exist with distinct base spirits (e.g. gin, rum, rye whiskey). At least one recipe should have no base spirit flagged.

---

## Acceptance Test 1 — Base-spirit filter alone

**Scenario**: Typing `base spirit is gin` returns only gin-based recipes.

### Steps

1. Navigate to `#/recipes`.
2. Type `base spirit is gin` in the search field (wait ~350 ms for debounce).
3. Confirm: only recipes whose base spirit is gin appear. Recipes with no base-spirit ingredient or a different base spirit are absent.
4. Clear the field. Confirm all recipes return.
5. Type `base spirit = Gin` (capital G, equals syntax). Confirm the same recipes as step 3 appear (case-insensitive, both syntaxes equivalent).

**Pass criteria**: Steps 3 and 5 return the same set; no non-gin-based recipe appears.

---

## Acceptance Test 2 — Combined filter (ingredient + base spirit)

**Scenario**: `absinthe base spirit is rye whiskey` returns the intersection.

### Steps

1. Navigate to `#/recipes`.
2. Type `absinthe base spirit is rye whiskey` in the search field.
3. Confirm: only recipes containing absinthe as an ingredient AND whose base spirit is rye whiskey appear.
4. If no such recipe exists, the empty state is shown (not an error).

**Pass criteria**: Every displayed recipe satisfies both constraints.

---

## Acceptance Test 3 — Whiskey/whisky spelling normalisation

**Scenario**: `rye whisky` and `rye whiskey` return identical results.

### Steps

1. Ensure at least one recipe with an ingredient named "rye whisky" (without 'e') exists, and one with "rye whiskey" (with 'e').
2. Type `rye whisky` in the search field. Note the results.
3. Clear and type `rye whiskey`. Confirm the same recipes appear.
4. Type `base spirit is rye whisky`. Note the results.
5. Clear and type `base spirit is rye whiskey`. Confirm the same recipes appear as step 4.

**Pass criteria**: Steps 2 and 3 yield identical results; steps 4 and 5 yield identical results.

---

## Acceptance Test 4 — Hint text updated

**Scenario**: The search hint advertises the new syntax.

### Steps

1. Navigate to `#/recipes`.
2. Look at the small hint text below the search field.
3. Confirm it includes a reference to the `base spirit is` syntax (e.g. "base spirit is gin").

**Pass criteria**: Hint text is visible and mentions `base spirit is`.

---

## Acceptance Test 5 — Edge cases

### 5a — Empty base-spirit value

1. Type `base spirit is ` (trailing space, no value). Confirm normal search behaviour — all recipes load as if no filter was applied.

### 5b — Base spirit with no matches

1. Type `base spirit is xyzzy123`. Confirm the empty-state message is shown (consistent with any search that yields no results).

### 5c — Only first clause used

1. Type `base spirit is gin base spirit is rum`. Confirm only gin-based recipes appear (first clause wins; second is ignored).

---

## Running Unit Tests

```bash
# Frontend
cd frontend && npm test -- --coverage

# Backend
cd backend && go test ./...
```

**Coverage requirements**: frontend ≥ 75%, backend ≥ 75%.

**New/modified test files**:
- `frontend/src/pages/RecipeList.test.js` — new base-spirit parsing tests
- `frontend/src/api/client.test.js` — new `getRecipes` parameter tests (if applicable)
- `backend/internal/handler/recipes_test.go` — new handler tests for `base_spirit` param
- `backend/internal/store/sqlite/recipes_test.go` — new `SearchByBaseSpirit` tests
- `backend/internal/store/dynamo/recipes.go` — new `SearchByBaseSpirit` tests
