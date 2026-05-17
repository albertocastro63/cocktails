# Tasks: Recipe Ownership and Per-User Recipe Listing

**Input**: Design documents from `specs/010-recipe-ownership/`  
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/api.md ✅, quickstart.md ✅

**Note**: Tests are included per the constitution's mandatory TDD requirement (Principle II). Write each test task, confirm it fails, then implement.

**Organization**: Tasks are grouped by user story. US1 and US2 share the backend handler file but are independently testable. US3 is fully independent.

---

## Phase 1: Setup

No new project initialization is needed — the project structure exists. This phase has no tasks.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Changes that MUST be complete before any user story can be implemented or tested. Adding `ListByCreator` to the interface will break the compile-time mock check until the stub is updated; both must land together. `getUserID()` is needed by both US1 and US3.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T001 Add `ListByCreator(creatorID string, page, limit int) ([]*model.Recipe, int, error)` to `RecipeStore` interface in `backend/internal/store/store.go`
- [ ] T002 Add `ListByCreator` stub implementation to `stubRecipeStore` in `backend/internal/handler/mock_test.go` (returns owned recipes; run `go build ./...` to confirm compile passes)
- [ ] T003 [P] Add `getUserID()` function to `frontend/src/api/auth.js` (decode `user_id` claim from JWT payload, same technique as existing `isAdmin()`)

**Checkpoint**: `go build ./...` passes; frontend auth module exports `getUserID`.

---

## Phase 3: User Story 1 - Recipe Edit and Delete Restricted to Owner (Priority: P1) 🎯 MVP

**Goal**: Non-owners cannot edit or delete recipes via the API; the UI hides edit/delete controls from non-owners.

**Independent Test**: Log in as user A, create a recipe, log in as user B — user B's PUT/DELETE return 403; user B sees no edit/delete buttons; user A's buttons are visible and functional.

### Tests for User Story 1

> **Write tests FIRST. Confirm they FAIL before implementing T007 and T008.**

- [ ] T004 [US1] Write tests for ownership enforcement in `backend/internal/handler/recipes_test.go`: non-owner PUT returns 403; non-owner DELETE returns 403; owner PUT succeeds; owner DELETE succeeds; recipe with empty `creator_id` (legacy) cannot be edited/deleted by non-admin (returns 403)
- [ ] T005 [P] [US1] Write failing tests for `RecipeCard` edit/delete visibility in `frontend/src/components/RecipeCard.test.js`: buttons hidden when `currentUser` is null or non-owner; buttons shown when `currentUser.id === recipe.creator_id`; buttons shown when `currentUser.isAdmin === true`

### Implementation for User Story 1

- [ ] T006 [P] [US1] Update `RecipeCard` to accept optional `currentUser` prop `{ id, isAdmin }` and append edit/delete buttons only when `currentUser?.id === recipe.creator_id || currentUser?.isAdmin === true` in `frontend/src/components/RecipeCard.js`
- [ ] T007 [US1] Update `RecipeList` to resolve `currentUser` via `getUserID()` and `isAdmin()` from auth and pass it to each `RecipeCard` call in `frontend/src/pages/RecipeList.js`
- [ ] T008 [US1] Update `RecipeDetail` to show edit/delete controls only when `currentUser.id === recipe.creator_id || currentUser.isAdmin` (currently shows to all logged-in users) in `frontend/src/pages/RecipeDetail.js`

**Checkpoint**: User Story 1 fully functional — non-owners see no edit/delete controls; API rejects unauthorized mutations.

---

## Phase 4: User Story 2 - Administrator Can Edit or Delete Any Recipe (Priority: P2)

**Goal**: An administrator can edit or delete any recipe regardless of who created it.

**Independent Test**: Log in as admin, attempt PUT/DELETE on a recipe created by a regular user — both return 200/204.

### Tests for User Story 2

> **Write tests FIRST. Confirm they FAIL before implementing T010 and T011.**

- [ ] T009 [US2] Write failing tests for admin bypass in `backend/internal/handler/recipes_test.go`: admin PUT on non-owned recipe returns 200; admin DELETE on non-owned recipe returns 204; admin can also edit/delete own recipes

### Implementation for User Story 2

- [ ] T010 [US2] Add admin bypass to `Update` handler — insert `if !claims.IsAdmin && existing.CreatorID != claims.UserID` guard (replacing bare `!=` check) in `backend/internal/handler/recipes.go`
- [ ] T011 [US2] Add admin bypass to `Delete` handler using the same guard pattern in `backend/internal/handler/recipes.go`

**Checkpoint**: Admin can edit/delete any recipe; non-owners still get 403; owners still succeed.

---

## Phase 5: User Story 3 - Per-User Recipe Listing (Priority: P3)

**Goal**: Authenticated users can retrieve and view a listing of only their own recipes via a dedicated "My Recipes" nav link and page.

**Independent Test**: Create two recipes as user A; GET `/api/v1/recipes/mine` as user A returns exactly those two; as user B returns empty; frontend "My Recipes" page shows user A's recipes with edit/delete on every card.

### Tests for User Story 3

> **Write tests FIRST. Confirm they FAIL before implementing store and handler.**

- [ ] T012 [US3] Write failing SQLite test for `ListByCreator` in `backend/internal/store/sqlite/recipes_test.go`: returns only recipes with matching creator_id; returns empty for unknown creator; respects page/limit; recipes with empty creator_id are NOT returned (legacy recipes excluded from per-user listings)
- [ ] T013 [US3] Write failing handler test for `Mine` endpoint in `backend/internal/handler/recipes_test.go`: authenticated request returns only own recipes; unauthenticated request returns 401; empty-state returns `{"data":[],"total":0,...}`

### Implementation for User Story 3

- [ ] T014 [US3] Implement `ListByCreator` in `backend/internal/store/sqlite/recipes.go` using `SELECT … WHERE creator_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
- [ ] T015 [P] [US3] Implement `ListByCreator` in `backend/internal/store/dynamo/recipes.go` using DynamoDB Scan with `FilterExpression: "creator_id = :cid"` (consistent with existing scan-based List/Search)
- [ ] T016 [US3] Add `Mine(w http.ResponseWriter, r *http.Request)` handler method to `RecipeHandler` in `backend/internal/handler/recipes.go` (reads claims, calls `ListByCreator`, returns same envelope as List)
- [ ] T017 [US3] Register `GET /api/v1/recipes/mine` route BEFORE `GET /api/v1/recipes/{id}` in `backend/cmd/lambda/main.go` (protected by `RequireAuth`)
- [ ] T018 [P] [US3] Register `GET /api/v1/recipes/mine` route BEFORE `GET /api/v1/recipes/{id}` in `backend/cmd/server/main.go` (protected by `requireAuth`)
- [ ] T019 [P] [US3] Add `getMyRecipes(token, params = {})` function to `frontend/src/api/client.js`
- [ ] T020 [US3] Create `frontend/src/pages/MyRecipes.js` page: same grid/card layout as `RecipeList`, heading "My Recipes", no search bar, calls `getMyRecipes(token)`, passes actual `currentUser` object (from `getUserID()` + `isAdmin()`) to each `RecipeCard` — ownership check always passes since all listed recipes belong to the viewer; display error message when fetch fails (consistent with `RecipeList` error pattern)
- [ ] T021 [US3] Write tests for `MyRecipes.js` in `frontend/src/pages/MyRecipes.test.js` BEFORE implementing T020 — confirm tests fail first: renders heading; shows recipe cards on success; shows empty state when no recipes; shows error state when fetch fails; renders edit/delete controls on all cards
- [ ] T022 [US3] Add `/my-recipes` route and "My Recipes" nav link (visible only when logged in, placed after "All Recipes") in `frontend/src/main.js`

**Checkpoint**: `GET /api/v1/recipes/mine` returns only the authenticated user's recipes; "My Recipes" page renders with edit/delete on every card.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T023 Run all backend tests and confirm they pass: `cd backend && go test ./...`
- [ ] T024 [P] Run all frontend tests and confirm they pass: `cd frontend && npm test`
- [ ] T025 [P] Manually validate quickstart.md scenarios 1–6 against the running dev server

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 2)**: No dependencies — start immediately
- **US1 (Phase 3)**: Depends on Phase 2 (T001, T002, T003) complete
- **US2 (Phase 4)**: Depends on Phase 2 complete; does NOT depend on US1
- **US3 (Phase 5)**: Depends on Phase 2 complete; does NOT depend on US1 or US2
- **Polish (Phase 6)**: Depends on all desired user stories complete

### User Story Dependencies

- **US1 (P1)**: Independent after Phase 2
- **US2 (P2)**: Independent after Phase 2 (shares `recipes.go` with US1 but edits different lines)
- **US3 (P3)**: Independent after Phase 2

### Within Each User Story

1. Write tests → confirm they fail
2. Implement → confirm tests pass
3. Frontend: auth helpers → component → page → route

---

## Parallel Opportunities

```bash
# Phase 2 — can all run in parallel:
T001  # store interface
T002  # mock stub
T003  # getUserID()

# Phase 3 — tests and component can run in parallel:
T004  # backend tests (US1)
T005  # frontend RecipeCard tests

# Phase 3 — after tests fail, implementation:
T006  # RecipeCard component (independent)
# T007 depends on T006; T008 independent of T006

# Phase 4 — sequential (same file):
T009 → T010 → T011

# Phase 5 — parallel opportunities:
T012  # SQLite test
T013  # handler test
# After tests fail:
T014  # SQLite impl
T015  # DynamoDB impl (different file, parallel with T014)
T017  # lambda main (after T016)
T018  # server main (parallel with T017, different file)
T019  # client.js (independent)
T020  # MyRecipes.js page
T021  # MyRecipes.test.js (parallel with T020 if writing tests-first)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Foundational (T001–T003)
2. Complete Phase 3: User Story 1 (T004–T008)
3. **STOP and VALIDATE**: Backend ownership enforcement works; non-owners see no UI controls
4. Ship MVP — ownership is enforced

### Incremental Delivery

1. Phase 2 → Phase 3 (US1): Non-owners blocked at API + UI controls hidden
2. Phase 4 (US2): Admins can edit/delete anything
3. Phase 5 (US3): "My Recipes" listing + nav link

---

## Notes

- `GET /api/v1/recipes/mine` MUST be registered before `GET /api/v1/recipes/{id}` in the mux — Go's `http.ServeMux` matches the most specific pattern first, but to avoid any ambiguity, ordering matters.
- The `RecipeCard` edit/delete button addition means the card's `currentUser` prop is optional — pages that don't need ownership control (e.g., unauthenticated views) can omit it and buttons remain hidden.
- `MyRecipes.js` always passes `showControls: true` (or equivalent) because all listed recipes belong to the viewer — no per-card ownership check needed on that page.
- Legacy recipes (empty `creator_id`) appear in the global list but are absent from any user's "My Recipes" list — `ListByCreator` filters on the exact user ID.
