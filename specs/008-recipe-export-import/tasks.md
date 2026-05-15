---

description: "Task list for Recipe Export and Import"
---

# Tasks: Recipe Export and Import

**Input**: Design documents from `specs/008-recipe-export-import/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/api.md ✅

**Organization**: Tasks grouped by user story. US1 (P1) → US2 (P2) → US3 (P3). TDD is mandatory per the Constitution: write each failing test before the corresponding implementation task.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)

---

## Phase 1: Setup

No new dependencies or project initialization required. All tooling, frameworks, and build configuration are already in place.

---

## Phase 2: Foundational (Blocking Prerequisite)

**Purpose**: Extend the `RecipeStore` interface and immediately add stub implementations to keep the build green. Both stores must satisfy the interface at every commit.

**⚠️ CRITICAL**: All three tasks below must complete together. Adding the interface methods without the stubs breaks the build in `backend/cmd/server/main.go` and `backend/cmd/lambda/main.go` because those files assign concrete store types to `store.RecipeStore` variables.

- [x] T001 Add `ListAll() ([]*model.Recipe, error)` and `ImportBatch(recipes []*model.Recipe, creatorID string) (created, skipped int, err error)` to the `RecipeStore` interface in `backend/internal/store/store.go`
- [x] T001a Add stub implementations of `ListAll` and `ImportBatch` to `backend/internal/store/sqlite/recipes.go` (return `nil, errors.New("not implemented")` / `0, 0, errors.New("not implemented")`) so the package compiles immediately after T001
- [x] T001b Add stub implementations of `ListAll` and `ImportBatch` to `backend/internal/store/dynamo/recipes.go` with the same stub bodies as T001a

**Checkpoint**: `go build ./...` passes. The stubs are replaced by real implementations in US2 (T012, T013) and US3 (T021, T022).

---

## Phase 3: User Story 1 — Download Recipe JSON Schema (Priority: P1) 🎯 MVP

**Goal**: Admin can click "Download Schema" on the admin recipes page and receive a valid JSON Schema Draft 7 file describing the recipe structure.

**Independent Test**: Log in as admin. Navigate to `#/admin/recipes`. Click "Download Schema". Open the downloaded `recipe-schema.json`. Confirm `"required": ["name"]` is present and `ingredients`, `steps`, `properties`, `notes` are all described.

### Tests for User Story 1

> **Write these tests FIRST, confirm they FAIL, then implement**

- [x] T002 [US1] Add failing test `TestExportSchema` to `backend/internal/handler/admin_recipes_test.go`: GET `/api/v1/admin/schema` with a valid admin JWT returns 200, `Content-Disposition: attachment; filename="recipe-schema.json"`, and a body containing `"required":["name"]`
- [x] T003 [P] [US1] Add failing test to `frontend/src/pages/AdminRecipes.test.js`: rendering `AdminRecipes()` produces a button with `data-download-schema` attribute

### Implementation for User Story 1

- [x] T004 [US1] Create `backend/internal/handler/admin_recipes.go` with `AdminRecipeHandler` struct (holding a `store.RecipeStore` field), a `recipeSchema` embedded constant (JSON Schema Draft 7 as per `data-model.md`), and the `ExportSchema(w, r)` method that writes `Content-Disposition: attachment; filename="recipe-schema.json"` with the schema constant as body
- [x] T005 [P] [US1] Add `downloadRecipeSchema(token)` and an internal `fetchBlob(path, token)` helper to `frontend/src/api/client.js`; `downloadRecipeSchema` calls `fetchBlob('/api/v1/admin/schema', token)` and returns the Blob
- [x] T006 [US1] In `backend/cmd/server/main.go`, instantiate a single shared handler: `adminRecipesH := handler.NewAdminRecipeHandler(recipeStore)`, then register `GET /api/v1/admin/schema` under `requireAuth(handler.RequireAdmin(http.HandlerFunc(adminRecipesH.ExportSchema)))` — T015 and T024 reuse this same `adminRecipesH` instance
- [x] T007 [US1] Create `frontend/src/pages/AdminRecipes.js` with: page header "Admin · Recipes", a Schema section with a `data-download-schema` button that calls `downloadRecipeSchema(getToken())` then triggers a DOM download via `URL.createObjectURL` + temporary `<a download="recipe-schema.json">`; show inline loading/error states on the button
- [x] T008 [US1] Add `{ pattern: /^\/admin\/recipes$/, factory: () => AdminRecipes() }` to the routes array in `frontend/src/main.js`; update the admin nav block to render two links — "Users" → `#/admin/users` and "Recipes" → `#/admin/recipes` — instead of the single "Admin" link; add a test to `frontend/src/main.test.js` asserting that `buildNav()` called with an admin session renders both an `<a href="#/admin/users">` and an `<a href="#/admin/recipes">` element

**Checkpoint**: Backend test `TestExportSchema` passes. Frontend test for `data-download-schema` passes. All pre-existing tests still pass. Navigate to `#/admin/recipes` and trigger the download manually; inspect `recipe-schema.json`.

---

## Phase 4: User Story 2 — Export All Recipes (Priority: P2)

**Goal**: Admin can click "Export Recipes" and receive a `recipes-export.json` file containing all recipes as a JSON array, each object conforming to the recipe schema.

**Independent Test**: Create at least two recipes. Log in as admin. Navigate to `#/admin/recipes`. Click "Export Recipes". Open the downloaded file; confirm it is a JSON array, each element has at least `"name"`, and server fields (`id`, `creator_id`, `created_at`, `updated_at`) are absent.

### Tests for User Story 2

> **Write these tests FIRST, confirm they FAIL, then implement**

- [x] T009 [US2] Add failing test `TestListAll` to `backend/internal/store/sqlite/recipes_test.go`: seed two recipes, call `ListAll()`, confirm both are returned without pagination
- [x] T010 [P] [US2] Add failing test `TestExportRecipes` to `backend/internal/handler/admin_recipes_test.go`: GET `/api/v1/admin/recipes/export` with admin JWT returns 200, `Content-Disposition: attachment; filename="recipes-export.json"`, body is a JSON array where each object has `"name"` but no `"id"` field
- [x] T011 [P] [US2] Add failing test to `frontend/src/pages/AdminRecipes.test.js`: rendering `AdminRecipes()` produces a button with `data-export-recipes` attribute

### Implementation for User Story 2

- [x] T012 [US2] Implement `ListAll()` in `backend/internal/store/sqlite/recipes.go`: `SELECT id, name, ingredients, steps, properties, notes, creator_id, created_at, updated_at FROM recipes ORDER BY created_at DESC` with no LIMIT, reusing the existing `scanRecipes` helper (pass `total = len(results)`)
- [x] T013 [P] [US2] Implement `ListAll()` in `backend/internal/store/dynamo/recipes.go`: paginated `Scan` following `LastEvaluatedKey` until `nil`, collecting all items via the existing `scanToRecipes` helper
- [x] T014 [US2] Add `ExportRecipes(w, r)` method to `AdminRecipeHandler` in `backend/internal/handler/admin_recipes.go`: call `h.recipes.ListAll()`, marshal each recipe as a RecipeExportRecord (fields: `name`, `ingredients`, `steps`, `properties`, `notes` — omit server-generated fields), write `Content-Disposition: attachment; filename="recipes-export.json"`, empty array `[]` when no recipes exist
- [x] T015 [US2] Register `GET /api/v1/admin/recipes/export` under `requireAuth(handler.RequireAdmin(http.HandlerFunc(adminRecipesH.ExportRecipes)))` in `backend/cmd/server/main.go`, reusing the `adminRecipesH` instance created in T006
- [x] T016 [P] [US2] Add `exportRecipes(token)` function to `frontend/src/api/client.js`: calls `fetchBlob('/api/v1/admin/recipes/export', token)` and returns the Blob
- [x] T017 [US2] Add the Export section to `frontend/src/pages/AdminRecipes.js`: a `data-export-recipes` button that calls `exportRecipes(getToken())` and triggers a download of `recipes-export.json`; show inline loading/error states

**Checkpoint**: `TestListAll` and `TestExportRecipes` pass. Frontend `data-export-recipes` test passes. All pre-existing tests pass. Export a collection of recipes and verify the file contents manually.

---

## Phase 5: User Story 3 — Import Recipes from a JSON File (Priority: P3)

**Goal**: Admin can select a JSON file, submit it, and receive a success message reporting how many recipes were created and how many were skipped. Invalid files show a specific error and create no recipes.

**Independent Test**: Export all recipes. Delete all recipes. Import the export file. Confirm success message "N recipes imported, 0 skipped" and that all recipes reappear in the recipe list with all fields intact.

### Tests for User Story 3

> **Write these tests FIRST, confirm they FAIL, then implement**

- [x] T018 [US3] Add failing tests `TestImportBatch_*` to `backend/internal/store/sqlite/recipes_test.go` covering: (a) happy path — 2 recipes created, returns `created=2, skipped=0`; (b) duplicate name — 1 existing + 1 new in file, returns `created=1, skipped=1`; (c) all duplicates — returns `created=0, skipped=N`; (d) rollback on unexpected error (inject a broken recipe mid-batch, verify none persisted)
- [x] T019 [P] [US3] Add failing tests `TestImportRecipes_*` to `backend/internal/handler/admin_recipes_test.go` covering: (a) valid array → 200 `{"imported":N,"skipped":M}`; (b) body is not a JSON array → 400 BAD_REQUEST; (c) recipe missing `name` → 400 BAD_REQUEST with index in message; (d) non-admin → 403; (e) no auth → 401
- [x] T020 [P] [US3] Add failing tests to `frontend/src/pages/AdminRecipes.test.js` covering the import control: (a) an `<input type="file" data-import-file>` is present; (b) a `data-import-submit` button is present; (c) a `data-import-status` element is present

### Implementation for User Story 3

- [x] T021 [US3] Implement `ImportBatch(recipes []*model.Recipe, creatorID string) (created, skipped int, err error)` in `backend/internal/store/sqlite/recipes.go`: begin a `sql.Tx`; for each recipe check `SELECT COUNT(*) FROM recipes WHERE name = ?` within the tx — if exists increment `skipped`, else insert + call an `upsertFTSTx(tx, r)` helper and increment `created`; `tx.Commit()` on success; `defer tx.Rollback()` on any error (sets `created = 0`)
- [x] T022 [P] [US3] Implement `ImportBatch` in `backend/internal/store/dynamo/recipes.go`: for each recipe call `ExistsByName`; if exists increment `skipped`; else call `Create` and track the created ID — on any `Create` error, attempt to delete all previously created IDs for compensation and return the error with `created = 0`; note that compensation is best-effort (a secondary delete failure leaves orphan records — acceptable per `research.md` Decision 5 for the ≤500 recipe scope)
- [x] T023 [US3] Add `ImportRecipes(w, r)` method to `AdminRecipeHandler` in `backend/internal/handler/admin_recipes.go`: apply `http.MaxBytesReader(w, r.Body, 10<<20)`; decode body as `[]json.RawMessage`; validate each element (must be object with non-empty `name` string; check `ingredients[*].name`, `steps[*]`, `properties` values, `notes` types if present); on validation failure return 400 with `"recipe at index N: <reason>"`; call `h.recipes.ImportBatch(recipes, claims.UserID)`; return `{"imported": N, "skipped": M}`
- [x] T024 [US3] Register `POST /api/v1/admin/recipes/import` under `requireAuth(handler.RequireAdmin(http.HandlerFunc(adminRecipesH.ImportRecipes)))` in `backend/cmd/server/main.go`, reusing the `adminRecipesH` instance created in T006
- [x] T025 [P] [US3] Add `importRecipes(recipes, token)` to `frontend/src/api/client.js`: calls `request('POST', '/api/v1/admin/recipes/import', recipes, token)` and returns the `{imported, skipped}` result
- [x] T026 [US3] Add the Import section to `frontend/src/pages/AdminRecipes.js`: `<input type="file" accept=".json" data-import-file>` and `data-import-submit` button; on submit: read the file via `FileReader.readAsText`, `JSON.parse`, verify it is an array (show client-side error if not), call `importRecipes(parsed, getToken())`; on success show `"N recipes imported, M skipped"` in `data-import-status`; on error show the server error message; show "Importing…" while in flight

**Checkpoint**: All `TestImportBatch_*` and `TestImportRecipes_*` tests pass. Frontend import form tests pass. All pre-existing tests pass. Run the round-trip scenario from `quickstart.md` (Scenario 6) manually.

---

## Phase 6: Polish & Validation

**Purpose**: Full suite confirmation and integration validation.

- [x] T027 Run `go test ./...` in `backend/` and confirm all tests pass with zero regressions
- [x] T028 [P] Run `npm test` in `frontend/` and confirm all tests pass with zero regressions
- [x] T029 [P] Add `BenchmarkExportRecipes` and `BenchmarkImportBatch` to `backend/internal/handler/admin_recipes_test.go`; run `go test -bench=. -benchtime=5s ./internal/handler/` and confirm export of 500 recipes completes within 5 s and import of 500 recipes completes within 30 s (satisfying SC-002 and SC-003 in CI)
- [ ] T030 [P] Start the dev server and manually validate quickstart.md scenarios 1–9: schema download (time < 2 s, SC-001), export (confirm file, SC-002), export empty, import happy path, import duplicates, round-trip, invalid JSON, missing name, non-admin access denied

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 2 (Foundational)**: No dependencies — start immediately. Must complete before US2/US3 store implementations.
- **Phase 3 (US1)**: Depends on Phase 2. T002 and T003 can start immediately after Phase 2. Backend (T004→T006) and frontend (T005→T007→T008) tracks proceed in parallel.
- **Phase 4 (US2)**: Depends on Phase 2. Can start after Phase 2 (does not depend on US1 completion for the backend). Frontend US2 tasks (T016, T017) depend on US1 frontend being complete (AdminRecipes.js and client.js exist).
- **Phase 5 (US3)**: Depends on Phase 2. Backend US3 is independent of US1/US2. Frontend US3 extends the same files touched by US1 and US2 — sequential.
- **Phase 6 (Polish)**: Depends on all desired user stories being complete. T029 (benchmarks) depends on T021 (SQLite ImportBatch real implementation replacing stub).

### User Story Dependencies

- **US1 (P1)**: Only depends on Phase 2
- **US2 (P2)**: Backend depends on Phase 2 only; Frontend depends on US1 frontend being done (AdminRecipes.js must exist)
- **US3 (P3)**: Backend depends on Phase 2 only; Frontend depends on US1 + US2 frontend being done

### Within Each User Story

1. Write failing tests first (test tasks before implementation tasks)
2. Run tests — confirm new tests FAIL
3. Implement — run tests again to confirm they PASS
4. Verify all pre-existing tests still pass

### Parallel Opportunities per Phase

**Phase 2 (Foundational)**:
```
T001 (interface) → T001a (SQLite stub) + T001b (DynamoDB stub) [parallel]
```

**Phase 3 (US1)**:
```
Track A (backend): T002 → T004 → T006
Track B (frontend): T003 → T005 → T007 → T008
```

**Phase 4 (US2)**:
```
Track A (SQLite):   T009 → T012
Track B (DynamoDB): T013
Track C (handler):  T010 → T014 → T015
Track D (frontend): T011 → T016 → T017
```

**Phase 5 (US3)**:
```
Track A (SQLite):   T018 → T021
Track B (DynamoDB): T022
Track C (handler):  T019 → T023 → T024
Track D (frontend): T020 → T025 → T026
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Extend store interface
2. Complete Phase 3: US1 — schema download
3. **STOP and VALIDATE**: Download the schema; confirm it is a valid JSON Schema Draft 7 document with `name` required
4. Ship if ready

### Incremental Delivery

1. Phase 2 → store interface extended
2. Phase 3 (US1) → schema download works → demo/validate
3. Phase 4 (US2) → export works → demo/validate round-trip
4. Phase 5 (US3) → import completes the data portability story → demo/validate round-trip (Scenario 6)
5. Each phase adds value without breaking previous functionality

---

## Notes

- [P] tasks = different files, no write conflicts between them
- [Story] label maps each task to its user story for traceability
- TDD is mandatory: every implementation task must be preceded by a failing test task (Constitution II)
- `admin_recipes_test.go` grows across US1/US2/US3 phases — tasks are sequential within that file across phases
- `AdminRecipes.js` and `AdminRecipes.test.js` grow across US1/US2/US3 phases — sequential across phases
- `client.js` receives additions in US1, US2, and US3 — sequential additions, one per phase
- Commit after each user story phase checkpoint
