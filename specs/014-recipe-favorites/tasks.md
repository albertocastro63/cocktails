# Tasks: Recipe Favorites

**Input**: Design documents from `specs/014-recipe-favorites/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/api.md ✅, quickstart.md ✅

**Organization**: Phase 2 writes all failing tests first (Constitution II — TDD mandatory). US1 and US2 are independent; US3 is a P3 stub. Infrastructure (Terraform) is its own phase.

---

## Phase 2: Foundational (TDD — Write Failing Tests First)

**Purpose**: Write all tests before any implementation. Tests in different files can be written in parallel.

**⚠️ CRITICAL**: Run `go test ./...` from `backend/` and `npm test` from `frontend/` after all T001–T005 — ALL new tests MUST fail before proceeding to implementation.

- [ ] T001 [P] Write failing Go tests for SQLite `FavoriteStore` in `backend/internal/store/sqlite/favorites_test.go` — import the sqlite package; cover: (1) `Add(userID, recipeID)` inserts a row without error; (2) `IsFavorite(userID, recipeID)` returns `true` after `Add`; (3) `Remove(userID, recipeID)` deletes the row without error; (4) `IsFavorite` returns `false` after `Remove`; (5) `ListByUser(userID)` returns all favorites for that user; (6) `Add` is idempotent — calling it twice does not return an error; (7) `CountByRecipe(recipeID)` returns 0 (stub)
- [ ] T002 [P] Write failing Go tests for `FavoriteHandler` in `backend/internal/handler/favorites_test.go` — use `httptest.NewRecorder` and a mock `FavoriteStore` + mock `RecipeStore`; cover: (1) `PUT /api/v1/recipes/{id}/favorite` with valid auth returns 204; (2) same endpoint with no token returns 401; (3) same endpoint when `recipe.creator_id == claims.UserID` returns 403 with code `FORBIDDEN`; (4) same endpoint when recipe not found returns 404; (5) `DELETE /api/v1/recipes/{id}/favorite` with valid auth returns 204; (6) `GET /api/v1/recipes/{id}/favorite` returns `{"is_favorite":true}` when `IsFavorite` returns true; (7) `GET /api/v1/recipes/{id}/favorite` returns `{"is_favorite":false}` when not favorited; (8) `GET /api/v1/recipes/favorites` returns a list with `is_favorite:true` on each item
- [ ] T003 [P] Write failing frontend unit tests for `FavoriteButton` in `frontend/src/components/FavoriteButton.test.js` — import from `./FavoriteButton.js`; cover: (1) renders a `<button>` element; (2) when `isFavorited=false`, `aria-label` is `"Add to favorites"` and `aria-pressed="false"`; (3) when `isFavorited=true`, `aria-label` is `"Remove from favorites"` and `aria-pressed="true"`; (4) when `isFavorited=true`, button has class `text-red-500`; (5) when `isFavorited=false`, button has class `text-stone-400`; (6) clicking the button calls `onToggle()` exactly once; (7) button has class `focus-visible:outline-amber-500`
- [ ] T004 [P] Write failing frontend integration tests for RecipeDetail favorites in `frontend/src/pages/RecipeDetail.test.js` — mock `../api/client.js` (`getRecipe`, `getFavoriteStatus`); mock `../api/auth.js` (`getUserID`, `getToken`, `isAdmin`); cover: (1) when logged in and `recipe.creator_id !== getUserID()`, a button with `aria-label="Add to favorites"` is rendered; (2) when `getFavoriteStatus` resolves `{is_favorite:true}`, the button has `aria-pressed="true"`; (3) when `recipe.creator_id === getUserID()`, no favorite button is rendered; (4) when not logged in (`getUserID()` returns null), no favorite button is rendered; (5) clicking the favorite button when `is_favorite=false` calls `favoriteRecipe(id, token)`; (6) when `getFavoriteStatus` resolves `{is_favorite:true}` and user clicks, `unfavoriteRecipe(id, token)` is called (symmetric unfavorite path); (7) mock `favoriteRecipe` as never-resolving — after click, the button immediately has `aria-pressed="true"` (optimistic update fires before API resolves); (8) when `favoriteRecipe` rejects, `aria-pressed` reverts to `"false"` (error-revert path)
- [ ] T005 [P] Write failing frontend integration tests for MyRecipes favorites in `frontend/src/pages/MyRecipes.test.js` — mock `../api/client.js` (`getMyRecipes`, `getMyFavorites`); mock `../api/auth.js`; cover: (1) recipes returned only from `getMyFavorites` appear in the list with a heart badge (element with `data-favorite="true"`); (2) recipes returned only from `getMyRecipes` appear without a heart badge; (3) when a recipe appears in both lists (same id), it appears exactly once without a heart badge (ownership wins); (4) when both lists are empty, the empty state message is shown

**Checkpoint**: `go test ./...` from `backend/` shows T001/T002 tests failing with "undefined" or "cannot find module." `npm test` from `frontend/` shows T003–T005 failing with "Cannot find module." No implementation exists yet.

---

## Phase 3: User Story 1 — Mark a Recipe as Favorite (Priority: P1) 🎯 MVP

**Goal**: Logged-in user can click a heart icon on any recipe they did not create. The icon turns red when favorited and gray when not. State is persisted server-side.

**Independent Test**: Log in, navigate to a recipe created by another user, click the heart, reload the page — heart should still be red. Click again — turns gray. Verify the icon is absent on own recipes and when logged out.

### Implementation for User Story 1

- [ ] T006 Add `Favorite` struct to `backend/internal/model/model.go` — add `type Favorite struct { UserID string \`json:"user_id"\`; RecipeID string \`json:"recipe_id"\`; CreatedAt time.Time \`json:"created_at"\` }` — and add `FavoriteStore` interface to `backend/internal/store/store.go`: `type FavoriteStore interface { Add(userID, recipeID string) error; Remove(userID, recipeID string) error; IsFavorite(userID, recipeID string) (bool, error); ListByUser(userID string) ([]*model.Favorite, error); CountByRecipe(recipeID string) (int, error) }`
- [ ] T007 [P] Implement SQLite `FavoriteStore` in `backend/internal/store/sqlite/favorites.go` — (1) define `type FavoriteStore struct { db *sql.DB }` with constructor `func NewFavoriteStore(db *sql.DB) *FavoriteStore`; (2) implement all 5 interface methods: `Add` does `INSERT OR IGNORE INTO favorites VALUES(?,?,?)` with RFC3339Nano timestamp; `Remove` does `DELETE FROM favorites WHERE user_id=? AND recipe_id=?`; `IsFavorite` does a `SELECT COUNT(*)` returning bool; `ListByUser` does `SELECT user_id,recipe_id,created_at FROM favorites WHERE user_id=? ORDER BY created_at DESC`; `CountByRecipe` returns 0, nil (P3 stub); (3) add `CREATE TABLE IF NOT EXISTS favorites (user_id TEXT NOT NULL, recipe_id TEXT NOT NULL, created_at TEXT NOT NULL, PRIMARY KEY (user_id, recipe_id))` to the `migrate()` function in `backend/internal/store/sqlite/store.go`; (4) add `func OpenAll(path string) (*RecipeStore, *UserStore, *FavoriteStore, error)` to `backend/internal/store/sqlite/store.go` that calls `openDB` and returns all three stores sharing the same connection
- [ ] T008 [P] Implement DynamoDB `FavoriteStore` in `backend/internal/store/dynamo/favorites.go` — (1) define `type FavoriteStore struct { client *dynamodb.Client; tableName string }` with constructor `func NewFavoriteStore(client *dynamodb.Client, tableName string) *FavoriteStore`; (2) implement `Add`: `PutItem` with `{user_id: S, recipe_id: S, created_at: S}` using RFC3339Nano timestamp; (3) `Remove`: `DeleteItem(PK=user_id, SK=recipe_id)` — no error if item does not exist; (4) `IsFavorite`: `GetItem(PK=user_id, SK=recipe_id)` — returns `out.Item != nil`; (5) `ListByUser`: `Query(KeyConditionExpression="user_id = :uid", ScanIndexForward=false)` returning `[]*model.Favorite`; (6) `CountByRecipe`: returns 0, nil (P3 stub)
- [ ] T009 Implement `backend/internal/handler/favorites.go` — define `type FavoriteHandler struct { favorites store.FavoriteStore; recipes store.RecipeStore }` with `func NewFavoriteHandler(fs store.FavoriteStore, rs store.RecipeStore) *FavoriteHandler`; implement four methods: (1) `Add(w, r)`: extract claims via `ClaimsFromContext`; call `h.recipes.GetByID(id)` — 404 if not found; return 403 if `recipe.CreatorID == claims.UserID`; call `h.favorites.Add(claims.UserID, id)`; write 204; (2) `Remove(w, r)`: extract claims; call `h.favorites.Remove(claims.UserID, id)`; write 204; (3) `Check(w, r)`: extract claims; call `h.favorites.IsFavorite(claims.UserID, id)`; `writeJSON(w, 200, map[string]any{"is_favorite": ok})`; (4) `List(w, r)`: extract claims; call `h.favorites.ListByUser(claims.UserID)`; for each Favorite fetch recipe via `h.recipes.GetByID(fav.RecipeID)` (skip if not found); build response slice with `is_favorite: true` on each; `writeJSON(w, 200, map[string]any{"data": items, "total": len(items), "page": 1, "limit": len(items)})`
- [ ] T010 Wire FavoriteStore and routes in `backend/cmd/lambda/main.go` and `backend/cmd/server/main.go` — (1) in `openStore()` for the DynamoDB branch: read `os.Getenv("FAVORITES_TABLE")`; create `dynstore.NewFavoriteStore(client, favTable)`; return all three stores; (2) for the SQLite branch: replace `sqstore.Open(dbPath)` with `sqstore.OpenAll(dbPath)` — returns `rs, us, fs, err`; (3) update `buildHandler` signature to accept `fs store.FavoriteStore`; instantiate `handler.NewFavoriteHandler(fs, rs)` as `favH`; (4) register routes in this order (before the existing `GET /api/v1/recipes/{id}` wildcard): `mux.Handle("GET /api/v1/recipes/favorites", handler.RequireAuth(http.HandlerFunc(favH.List)))`, `mux.Handle("PUT /api/v1/recipes/{id}/favorite", handler.RequireAuth(http.HandlerFunc(favH.Add)))`, `mux.Handle("DELETE /api/v1/recipes/{id}/favorite", handler.RequireAuth(http.HandlerFunc(favH.Remove)))`, `mux.Handle("GET /api/v1/recipes/{id}/favorite", handler.RequireAuth(http.HandlerFunc(favH.Check)))`
- [ ] T011 [P] Add 4 API client functions to `frontend/src/api/client.js` — append: `export function getMyFavorites(token, params = {}) { const qs = new URLSearchParams(params).toString(); return request('GET', \`/api/v1/recipes/favorites\${qs ? '?' + qs : ''}\`, undefined, token); }` and `export function getFavoriteStatus(id, token) { return request('GET', \`/api/v1/recipes/\${id}/favorite\`, undefined, token); }` and `export function favoriteRecipe(id, token) { return request('PUT', \`/api/v1/recipes/\${id}/favorite\`, null, token); }` and `export function unfavoriteRecipe(id, token) { return request('DELETE', \`/api/v1/recipes/\${id}/favorite\`, null, token); }`
- [ ] T012 [P] Create `frontend/src/components/FavoriteButton.js` — export `function FavoriteButton({ isFavorited, onToggle })`; create a `<button type="button">`; set `aria-label` to `"Remove from favorites"` if `isFavorited`, else `"Add to favorites"`; set `aria-pressed` to `String(isFavorited)`; set `className` to `isFavorited ? 'text-red-500 hover:text-red-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-amber-500' : 'text-stone-400 hover:text-stone-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-amber-500'`; set `button.innerHTML` to the heart SVG: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-6 h-6"><path d="M11.645 20.91l-.007-.003-.022-.012a15.247 15.247 0 01-.383-.218 25.18 25.18 0 01-4.244-3.17C4.688 15.36 2.25 12.174 2.25 8.25 2.25 5.322 4.714 3 7.688 3A5.5 5.5 0 0112 5.052 5.5 5.5 0 0116.313 3c2.973 0 5.437 2.322 5.437 5.25 0 3.925-2.438 7.111-4.739 9.256a25.175 25.175 0 01-4.244 3.17 15.247 15.247 0 01-.383.218l-.022.012-.007.003-.003.001a.752.752 0 01-.704 0l-.003-.001z"/></svg>`; add click listener `btn.addEventListener('click', onToggle)`; return the button
- [ ] T013 Modify `frontend/src/pages/RecipeDetail.js` to add the FavoriteButton — (1) import `FavoriteButton` from `../components/FavoriteButton.js`; import `getFavoriteStatus`, `favoriteRecipe`, `unfavoriteRecipe` from `../api/client.js`; (2) after the recipe loads and `titleRow` is assembled: if `getUserID()` exists AND `recipe.creator_id !== getUserID()`: call `getFavoriteStatus(id, getToken()).then(({ is_favorite }) => { let isFavorited = is_favorite; const renderHeart = () => { const existing = titleRow.querySelector('button[data-heart]'); if (existing) existing.remove(); const btn = FavoriteButton({ isFavorited, onToggle: () => { const action = isFavorited ? unfavoriteRecipe : favoriteRecipe; isFavorited = !isFavorited; renderHeart(); action(id, getToken()).catch(() => { isFavorited = !isFavorited; renderHeart(); const err = document.createElement('p'); err.className = 'text-red-600 text-sm mt-1'; err.textContent = 'Failed to save. Try again.'; titleRow.appendChild(err); setTimeout(() => err.remove(), 4000); }); } }); btn.dataset.heart = ''; btn.className += ' ml-3'; titleRow.appendChild(btn); }; renderHeart(); }).catch(() => {})` — the error message renders inline below the heart button and auto-removes after 4 seconds via `setTimeout`; a failing status check is swallowed so it does not break the page

**Checkpoint**: `go test ./...` from `backend/` — T001 and T002 tests pass. `npm test` from `frontend/` — T003 and T004 tests pass. Start the dev server and backend server; log in, navigate to another user's recipe — heart icon appears gray; click it — turns red; reload — still red.

---

## Phase 4: User Story 2 — View Favorites on My Recipes Page (Priority: P2)

**Goal**: Logged-in user sees favorited recipes in a unified list on "My Recipes" with a heart badge on each favorited card.

**Independent Test**: Log in, favorite a recipe created by another user. Navigate to "My Recipes." The favorited recipe appears with a heart badge. Own recipes appear without a badge. Unfavorite the recipe, refresh — it disappears from the list.

### Implementation for User Story 2

- [ ] T014 [P] [US2] Modify `frontend/src/components/RecipeCard.js` to support `isFavorite` prop — add an `isFavorite` parameter to the function signature `RecipeCard({ recipe, currentUser, isFavorite = false })`; if `isFavorite` is true, render a small red heart badge element on the card: create a `<span>` with `data-favorite="true"` and `aria-label="Favorited"` placed in the top-right corner of the card (use `position: absolute` equivalent via Tailwind: `absolute top-2 right-2`); set its `innerHTML` to a small inline heart SVG with `class="w-4 h-4 text-red-500"`; wrap the card in `relative` positioning if not already
- [ ] T015 [US2] Modify `frontend/src/pages/MyRecipes.js` to fetch and merge favorites — (1) import `getMyFavorites` from `../api/client.js`; (2) replace the single `getMyRecipes(getToken())` call with `Promise.all([getMyRecipes(getToken()), getMyFavorites(getToken())])` — handle the case where `getMyFavorites` rejects gracefully (default to empty array); (3) merge results: create a `Map` keyed by recipe `id`; iterate `createdData` first (each entry has `isFavorite: false`); then iterate `favoritedData` — add to map only if `id` is NOT already present (deduplicate: created recipes take precedence), with `isFavorite: true`; (4) render the unified list from `[...map.values()]` using `RecipeCard({ recipe: item.recipe, currentUser, isFavorite: item.isFavorite })`; (5) show empty state only when the unified list is empty

**Checkpoint**: `npm test` from `frontend/` — T005 tests pass. Open "My Recipes" — favorited recipes appear with a red heart badge alongside created recipes (no badge). Unfavorite one, refresh — disappears.

---

## Phase 5: User Story 3 — Favorite Count (Priority: P3 — Deferred)

**Goal**: `CountByRecipe` stub is implemented as part of T007 (SQLite) and T008 (DynamoDB) — both return `0, nil`. Full UI for the count display is deferred to a follow-up feature. Stub correctness is verified in T018 coverage pass.

**Checkpoint**: No dedicated task. Stub implementation and Go comment (`// P3: replace with Query on recipe_id GSI when favorite count UI is implemented`) are part of T007 and T008. Verified during T018.

---

## Phase 6: Infrastructure

**Purpose**: Provision the `cocktails-favorites` DynamoDB table and extend Lambda permissions.

- [ ] T017 Add favorites DynamoDB table to `infra/main.tf` — (1) add new module block: `module "favorites_table" { source = "terraform-aws-modules/dynamodb-table/aws"; version = "~> 5.0"; name = "${local.name_prefix}-favorites"; billing_mode = "PAY_PER_REQUEST"; hash_key = "user_id"; range_key = "recipe_id"; attributes = [{ name = "user_id", type = "S" }, { name = "recipe_id", type = "S" }]; point_in_time_recovery_enabled = true; server_side_encryption_enabled = true }`; (2) extend `module.lambda_function.policy_statements.dynamodb.resources` to also include `module.favorites_table.dynamodb_table_arn` and `"${module.favorites_table.dynamodb_table_arn}/index/*"`; (3) add `FAVORITES_TABLE = module.favorites_table.dynamodb_table_id` to `module.lambda_function.environment_variables`; (4) add `favorites_table_name = module.favorites_table.dynamodb_table_id` to `infra/outputs.tf` if an outputs file exists

---

## Phase 7: Polish & Verification

- [ ] T018 Run `go test -coverprofile=coverage.out ./...` from `backend/` and `npm test -- --coverage` from `frontend/` — verify: `backend/internal/store/sqlite/favorites.go` ≥ 80% line coverage; `backend/internal/handler/favorites.go` ≥ 80% line coverage; `frontend/src/components/FavoriteButton.js` ≥ 80% coverage; `frontend/src/pages/RecipeDetail.js` ≥ 80% coverage; `frontend/src/pages/MyRecipes.js` ≥ 80% coverage; also confirm `CountByRecipe` stub in both `dynamo/favorites.go` and `sqlite/favorites.go` returns `0, nil` and carries the comment `// P3: replace with Query on recipe_id GSI when favorite count UI is implemented`; if any file is below threshold, add targeted tests before proceeding
- [ ] T019 Start dev server (`npm run dev` from `frontend/` + Go server from `backend/`) and manually run through quickstart.md scenarios SC-001 through SC-010 — verify: heart icon toggles correctly (SC-001/SC-002), absent on own recipe (SC-003), absent for guests (SC-004), unified My Recipes list (SC-005), unfavorite removes from list (SC-006), empty state (SC-007), no duplicate (SC-008), error revert (SC-009), keyboard accessibility (SC-010)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 2)**: No dependencies — start immediately; T001–T005 run in parallel
- **US1 (Phase 3)**: Depends on Phase 2 complete; T006 → T007∥T008 → T009 → T010; T011∥T012 → T013
- **US2 (Phase 4)**: Depends on Phase 3 complete (needs `FavoriteButton`, API client, and backend routes); T014∥T015 (T014 can run before T015)
- **US3 (Phase 5)**: No dedicated task — stub is part of T007/T008; verified in T018
- **Infrastructure (Phase 6)**: Independent of frontend work; can be done any time after T006
- **Polish (Phase 7)**: Depends on all previous phases; T018 then T019 sequential

### Parallel Opportunities

```bash
# Phase 2 — all parallel (different files):
T001 ∥ T002 ∥ T003 ∥ T004 ∥ T005

# Phase 3 — backend sequential then parallel:
T006 → T007 ∥ T008 → T009 → T010
# Frontend in parallel with backend T007/T008 onwards:
T011 ∥ T012 → T013

# Phase 4 — partial parallel:
T014 ∥ T015  (T014 can start; T015 depends on T014 for RecipeCard prop)

# Infrastructure — parallel with Phase 4:
T017

# Phase 7 — sequential:
T018 → T019
```

---

## Implementation Strategy

### MVP First

1. Complete Phase 2: Foundational (T001–T005 — failing tests)
2. Complete Phase 3: US1 (T006–T013 — heart icon, backend API, client functions)
3. **STOP and VALIDATE**: `go test ./...` passes; heart icon toggles in the browser; state persists on reload
4. Complete Phase 4: US2 (T014–T015 — My Recipes unified list)
5. Phase 5 (US3): No separate task — stub already implemented in T007/T008
6. Complete Phase 6: Infrastructure (T017 — Terraform)
7. Complete Phase 7: Polish (T018–T019)

### Incremental Delivery

US1 alone is a fully shippable MVP — users can favorite recipes and the state is persisted. US2 (My Recipes unified list) adds the value of discovering saved favorites. US3 (count) is deferred and isolated behind the stub. Infrastructure (T017) should be applied before deploying to AWS but does not block local dev or testing.

---

## Notes

- TDD is mandatory (Constitution II): run `go test ./...` and `npm test` after T001–T005 to confirm tests fail before writing implementation.
- The `GET /api/v1/recipes/favorites` route MUST be registered in `main.go` BEFORE `GET /api/v1/recipes/{id}` — the stdlib mux resolves specific paths first, but registration order still matters for clarity.
- `FavoriteHandler.List` has an N+1 query pattern (fetch each recipe by ID after listing favorites). This is acceptable for the current small-scale app; document with a comment for future optimization.
- `CountByRecipe` is intentionally stubbed as `return 0, nil` — a future P3 task will replace it with a DynamoDB GSI query or scan.
- The `OpenAll()` function in `sqlite/store.go` is a non-breaking addition; existing calls to `Open()` are unchanged until T010 migrates them.
