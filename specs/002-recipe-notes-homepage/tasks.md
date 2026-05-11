# Tasks: Recipe Notes and Full Homepage Display

**Input**: Design documents from `specs/002-recipe-notes-homepage/`  
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/api.md ✅

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.  
**TDD**: Constitution Principle II requires tests to be written and confirmed failing before implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1, US2)

---

## Phase 1: Setup

**Purpose**: Verify the working baseline before any changes are made.

- [X] T001 Verify all existing tests pass: run `cd backend && go test ./...` and `cd frontend && npm test`; confirm zero failures before any changes

---

## Phase 2: Foundational (Blocking Prerequisite)

**Purpose**: The `Notes` field on the `Recipe` model is shared by both user stories and must be added first.

**⚠️ CRITICAL**: Both user stories depend on this change. No US1 or US2 work can proceed until T002 is complete.

- [X] T002 Add `Notes string` field with `json:"notes"` tag to the `Recipe` struct in `backend/internal/model/model.go`; note that `omitempty` must NOT be used here so that empty notes (`""`) are always present in JSON responses per the API contract

**Checkpoint**: Model updated — US1 and US2 tasks can now proceed.

---

## Phase 3: User Story 1 — Notes Field on Recipe (Priority: P1) 🎯 MVP

**Goal**: The `notes` field is stored, returned in all recipe responses, excluded from search, and editable via the recipe create/edit form and the recipe detail view.

**Independent Test**: Create a recipe with notes via the API; retrieve it and confirm notes are returned. Search for a word that exists only in notes; confirm zero results. Open the recipe form; confirm a notes textarea is present and saves correctly.

### Tests for User Story 1 — Write First, Confirm Failing

> **Write these tests FIRST. Run the test suite and confirm each new test fails before writing implementation code.**

- [X] T003 [P] [US1] Add failing backend handler tests for notes: "Create recipe with notes returns notes in response", "Update omitting notes preserves existing notes", and "Update notes as non-creator returns 403" in `backend/internal/handler/recipes_test.go`
- [X] T004 [P] [US1] Add failing SQLite store tests: "notes persisted and returned by GetByID" and "search query matching only notes returns no results" in `backend/internal/store/sqlite/coverage_test.go`
- [X] T005 [P] [US1] Add failing frontend test: "renders a textarea for notes" in `frontend/src/pages/RecipeForm.test.js`
- [X] T006 [P] [US1] Add failing frontend tests: "renders notes section when recipe has notes" and "hides notes section when notes is empty" in `frontend/src/pages/RecipeDetail.test.js`

### Implementation for User Story 1

- [X] T007 [P] [US1] Add idempotent SQLite migration for `notes TEXT NOT NULL DEFAULT ''` column: in `backend/internal/store/sqlite/store.go`, add `ALTER TABLE recipes ADD COLUMN notes TEXT NOT NULL DEFAULT ''` inside `migrate()` and silently ignore the "duplicate column name" error
- [X] T008 [P] [US1] Update SQLite recipes.go to include `notes` in all queries: update `Create` INSERT, `GetByID`/`List`/`Search`/`Random` SELECT columns, `Update` SET clause, `scanRecipe` and `scanRecipes` Scan calls; confirm `upsertFTS` does NOT append `r.Notes` to `searchText` in `backend/internal/store/sqlite/recipes.go`
- [X] T009 [P] [US1] Update DynamoDB recipes.go: add `Notes string` to `recipeItem` struct; update `toItem` to set `Notes: r.Notes`; update `unmarshalRecipe` to populate `model.Recipe.Notes`; confirm `matchesQuery` does NOT check notes in `backend/internal/store/dynamo/recipes.go`
- [X] T010 [P] [US1] Update recipe handler Create and Update body structs: add `Notes *string json:"notes"` to both body structs; in Create set `recipe.Notes` from body (empty string if nil); in Update apply partial-update logic (`if body.Notes != nil { existing.Notes = *body.Notes }`) in `backend/internal/handler/recipes.go`
- [X] T011 [P] [US1] Add notes textarea to RecipeForm: add a labeled `<textarea name="notes">` field after the properties section; include notes value in the form submission payload; on edit prefill from `recipe.notes` in `frontend/src/pages/RecipeForm.js`
- [X] T012 [P] [US1] Add notes display section to RecipeDetail: after the properties section, add a notes `<h2>` and `<p>` element; show the section only when `recipe.notes` is a non-empty string in `frontend/src/pages/RecipeDetail.js`

**Checkpoint**: All T003–T012 tests pass. Create a recipe with notes via the form; verify notes appear on the detail page. Search for a notes-only keyword; verify zero results.

---

## Phase 4: User Story 2 — Full Recipe Display on Homepage (Priority: P2)

**Goal**: The homepage shows the full detail of the featured recipe — name, ingredients, steps, properties, and notes — without requiring navigation to the detail page.

**Independent Test**: Load the homepage with at least one recipe in the database. Confirm that ingredients, steps, properties, and notes are all visible on the page. Confirm that the empty-state message still displays when no recipes exist.

### Tests for User Story 2 — Write First, Confirm Failing

> **Write these tests FIRST. Run the test suite and confirm the new test fails before writing implementation code.**

- [X] T013 [US2] Add failing frontend tests to `frontend/src/pages/Home.test.js`: "renders ingredients when recipe has ingredients", "renders steps when recipe has steps", "renders notes when recipe has notes", and "renders empty state when recipe is null"

### Implementation for User Story 2

- [X] T014 [US2] Update `frontend/src/pages/Home.js`: replace the `RecipeCard` import and usage with an inline full-detail display; render the recipe name as `<h2>`, use the existing `IngredientList` component for ingredients, render steps as an `<ol>`, use the existing `PropertyTable` component for properties, and add a notes `<p>` (hidden when empty); preserve the existing empty-state handling for null recipe

**Checkpoint**: Homepage shows full recipe detail. All T013 tests pass. Empty state still displays when no recipes exist.

---

## Phase 5: Polish & Cross-Cutting Concerns

- [X] T015 [P] Run `cd backend && go test ./... -coverprofile=coverage.out` and confirm ≥80% coverage in all changed packages (`model`, `sqlite`, `dynamo`, `handler`)
- [X] T016 [P] Run `cd frontend && npm test` and confirm all tests pass with no regressions
- [X] T017 Update the base API contract to reflect the `notes` field in recipe object examples in `specs/001-cocktail-recipe-app/contracts/api.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — can start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — BLOCKS both user stories
- **Phase 3 (US1)**: Depends on Phase 2 (T002) completion
- **Phase 4 (US2)**: Depends on Phase 2 (T002) completion; frontend tasks (T013–T014) may run in parallel with Phase 3 frontend tasks (T005, T006, T011, T012) since they are in different files
- **Phase 5 (Polish)**: Depends on all desired user stories being complete

### User Story Dependencies

- **US1 (P1)**: Starts after T002. T003–T006 (tests) can run in parallel. T007–T012 (impl) can run in parallel — all different files.
- **US2 (P2)**: Starts after T002. T013 and T014 can run in parallel with US1 frontend tasks.

### Within Each User Story

- All test tasks (T003–T006, T013) MUST be written and confirmed failing before implementation tasks in the same story begin
- T007 (migration) and T008 (SQLite recipes.go) are independent files within the sqlite package — run in parallel
- T008 (SQLite) and T009 (DynamoDB) are different store backends — run in parallel
- T010 (handler) and T011 (RecipeForm) are different layers — run in parallel after tests pass

---

## Parallel Execution Examples

### User Story 1 — Tests (run together)

```
Task T003: "failing handler tests for notes in recipes_test.go"
Task T004: "failing SQLite store tests for notes in coverage_test.go"
Task T005: "failing RecipeForm test for notes textarea in RecipeForm.test.js"
Task T006: "failing RecipeDetail test for notes section in RecipeDetail.test.js"
```

### User Story 1 — Implementation (run together)

```
Task T007: "SQLite migration for notes column in store.go"
Task T008: "SQLite recipes.go notes in all queries and scan functions"
Task T009: "DynamoDB recipes.go Notes in recipeItem, toItem, unmarshalRecipe"
Task T010: "handler notes partial-update in recipes.go"
Task T011: "notes textarea in RecipeForm.js"
Task T012: "notes section in RecipeDetail.js"
```

---

## Implementation Strategy

### MVP (User Story 1 Only)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 2: Foundational (T002)
3. Write and confirm failing tests: T003–T006
4. Implement: T007–T012
5. **STOP and VALIDATE**: notes create/update/search/display all work
6. Ship MVP — notes feature is fully functional

### Full Delivery (Both Stories)

1. Complete Setup + Foundational (T001–T002)
2. Complete US1 (T003–T012) → validate independently
3. Complete US2 (T013–T014) → validate independently
4. Polish (T015–T017)

---

## Notes

- [P] tasks = different files, no incomplete dependencies
- The SQLite `notes` column is NOT indexed in FTS — confirmed by omission from `upsertFTS`
- The DynamoDB `matchesQuery` must NOT be updated with notes — exclusion by omission
- The `Notes *string` pointer pattern in the handler is consistent with the existing `Name *string` partial-update approach
- `Home.js` uses `IngredientList` and `PropertyTable` from existing components — no new components required
