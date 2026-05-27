# Tasks: Cocktail Recipe Garnishes

**Input**: Design documents from `specs/019-cocktail-garnishes/`  
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/api.md ✅

**Tests**: Included — TDD cycle is NON-NEGOTIABLE per the project constitution (§II).

**Organization**: Tasks are grouped by phase. Backend model + storage + API changes are Foundational (all three user stories depend on them). Frontend changes are per user story.

> **Note**: Phase 1 (Setup) is skipped — this feature extends an existing project with no new infrastructure required.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on in-progress tasks)
- **[Story]**: User story label (US1, US2, US3)

---

## Phase 2: Foundational — Backend Model, Storage & API

**Purpose**: All three user stories need the backend to accept, persist, and return `garnishes`. This phase must complete before any frontend story work begins.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Tests (write first — must FAIL before implementation)

- [X] T001 [P] Write failing handler tests for create/update/get with garnishes field in `backend/internal/handler/recipes_test.go`
- [X] T002 [P] Write failing SQLite store tests for garnishes persistence (create, update, read round-trip) in `backend/internal/store/sqlite/recipes_test.go`
- [X] T003 [P] Write failing admin handler tests for export/import garnishes round-trip in `backend/internal/handler/admin_recipes_test.go`

### Implementation

- [X] T004 Add `Garnishes []string` with `json:"garnishes,omitempty"` to `Recipe` struct in `backend/internal/model/model.go`
- [X] T005 Add idempotent `ALTER TABLE recipes ADD COLUMN garnishes TEXT NOT NULL DEFAULT '[]'` migration (follow existing pattern for `notes` column) in `backend/internal/store/sqlite/store.go`
- [X] T006 Update `Create` (INSERT), `Update` (SET), all SELECT queries, `scanRecipe`, and `scanRecipes` to marshal/unmarshal the `garnishes` column in `backend/internal/store/sqlite/recipes.go`
- [X] T007 [P] Add `Garnishes []string` field with `dynamodbav:"garnishes"` to `recipeItem`; update `toItem` and `unmarshalRecipe` to map the field in `backend/internal/store/dynamo/recipes.go`
- [X] T008 Add `Garnishes []string` to `Create` and `Update` request body structs; filter blank entries before assigning to recipe; initialise to `[]string{}` when nil in `backend/internal/handler/recipes.go`
- [X] T009 Add `Garnishes []string` to `recipeExport`; include in `ExportRecipes`; handle `garnishes` key in `ImportRecipes`; add `garnishes` array property to `recipeSchema` in `backend/internal/handler/admin_recipes.go`

**Checkpoint**: `go test ./...` passes. Create/update/get recipes with garnishes; export and re-import a recipe with garnishes — garnishes survive the round-trip.

---

## Phase 3: User Story 1 — Add Garnishes in Recipe Form (Priority: P1) 🎯 MVP

**Goal**: Authors can add, edit, and remove garnish entries in the recipe create/edit form. Garnishes are saved and restored on edit.

**Independent Test**: Create a recipe with two garnish entries, save, reopen in edit mode — both entries appear pre-populated. Remove one, save, reopen — only one remains.

### Tests (write first — must FAIL before implementation)

- [X] T010 [US1] Write failing Vitest tests covering: garnish section renders in form, add-row creates input, remove-row deletes input, blank entries are excluded from payload, saved garnishes are pre-populated on edit; note that auth-gating of the form itself is inherited from the existing recipe edit mechanism and does not require new test coverage in `frontend/src/pages/RecipeForm.test.js`

### Implementation

- [X] T011 [US1] Add `garnishesSection` using `buildDynamicSection('Garnishes', 'Add Garnish', ...)` after `stepsSection`; each row is a single text input + remove button in `frontend/src/pages/RecipeForm.js`
- [X] T012 [US1] Collect garnishes in submit handler (filter blank strings, same pattern as `steps`); prefill garnish rows from `recipe.garnishes` in the edit pre-fill block in `frontend/src/pages/RecipeForm.js` (depends on T011)

**Checkpoint**: Create a new recipe with garnishes via the form, save, open edit — garnishes persist and prefill correctly.

---

## Phase 4: User Story 2 — Garnishes on Recipe Detail Page (Priority: P2)

**Goal**: The recipe detail page shows a "Garnishes" section below ingredients when garnishes are present. Each entry is in italics. No section appears when the recipe has no garnishes.

**Independent Test**: Open the detail page for a recipe with garnishes — "Garnishes" heading and italic entries appear below ingredients. Open a recipe with no garnishes — no garnish section visible.

### Tests (write first — must FAIL before implementation)

- [X] T013 [US2] Write failing Vitest tests covering: garnish section appears when recipe has garnishes, each entry renders in `<em>`, section absent when garnishes empty/absent in `frontend/src/pages/RecipeDetail.test.js`

### Implementation

- [X] T014 [US2] After the ingredients block, add a `Garnishes` section: heading styled as `text-sm font-semibold uppercase tracking-widest text-amber-700 mt-6 mb-2`, each garnish entry wrapped in `<em>` in a list; section only rendered when `recipe.garnishes && recipe.garnishes.length > 0` in `frontend/src/pages/RecipeDetail.js`

**Checkpoint**: Detail page shows garnishes in italics below ingredients; no section for garnish-free recipes.

---

## Phase 5: User Story 3 — Garnishes in Hover Preview Card (Priority: P3)

**Goal**: The recipe card hover popover shows garnishes in italics below the ingredient list when `ingredients.length < MAX_VISIBLE` (5). When the ingredient list fills the cap, garnishes are omitted and the ingredient list is not truncated.

**Independent Test**: Hover a recipe with ≤4 ingredients and garnishes — garnish entries appear in italics below ingredients in the popover. Hover a recipe with ≥5 ingredients — no garnishes shown; ingredient list is unchanged.

### Tests (write first — must FAIL before implementation)

- [X] T015 [US3] Write failing Vitest tests covering: garnishes shown in popover when ingredients.length < MAX_VISIBLE, garnishes absent when ingredients.length >= MAX_VISIBLE, each garnish renders in `<em>`, no garnish section when recipe has no garnishes in `frontend/src/components/RecipeCard.test.js`

### Implementation

- [X] T016 [US3] In `buildIngredientPopover`, after rendering the ingredient `<ul>`, check `if (ingredients.length < MAX_VISIBLE && garnishes && garnishes.length > 0)`: if true, append a garnish section (a `<ul>` of `<li><em>` entries) below the ingredient list; update function signature to accept `garnishes` array; update the `mouseenter` handler to pass `recipe.garnishes || []` in `frontend/src/components/RecipeCard.js`

**Checkpoint**: Hovering a short-ingredient recipe shows garnishes in italics; hovering a 5-ingredient recipe omits garnishes.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T017 Run full test suite (`go test ./...` + `npm test` in `frontend/`) and confirm coverage ≥ 75% with no failures
- [ ] T018 [P] Verify italic rendering is visually consistent between detail page and hover preview, and confirm each garnish add/remove action completes in under 10 seconds (SC-001), by manually testing at least two recipes with garnishes
- [ ] T019 [P] Verify legacy recipes (no garnishes) display correctly on detail page, hover preview, and round-trip through export/import with no data change

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 2)**: No dependencies — start immediately. BLOCKS all user story phases.
- **US1 (Phase 3)**: Depends on Foundational complete.
- **US2 (Phase 4)**: Depends on Foundational complete. Independent of US1.
- **US3 (Phase 5)**: Depends on Foundational complete. Independent of US1 and US2.
- **Polish (Phase 6)**: Depends on all desired user stories complete.

### User Story Dependencies

- **US1 (P1)**: No dependency on US2 or US3.
- **US2 (P2)**: No dependency on US1 or US3.
- **US3 (P3)**: No dependency on US1 or US2.

### Within the Foundational Phase

```
T001, T002, T003  [parallel — different test files]
     ↓
T004  (model.go — needed by all store/handler tasks)
     ↓
T005, T007  [parallel — SQLite migration & DynamoDB struct, different files]
     ↓
T006  (SQLite CRUD — depends on T005)
     ↓
T008, T009  [parallel — handler and admin handler, different files]
```

### Parallel Opportunities

| Parallel group | Tasks |
|----------------|-------|
| Foundational tests | T001, T002, T003 |
| Store implementations (after T004) | T005, T007 |
| Handler implementations (after T006) | T008, T009 |
| US2 and US3 frontend tests (after Foundational) | T013, T015 |

---

## Parallel Example: Foundational Phase

```bash
# Step 1 — write all failing tests together:
Task T001: Write failing handler tests in backend/internal/handler/recipes_test.go
Task T002: Write failing SQLite store tests in backend/internal/store/sqlite/recipes_test.go
Task T003: Write failing admin handler tests in backend/internal/handler/admin_recipes_test.go

# Step 2 — model first (unblocks everything):
Task T004: Add Garnishes to model.go

# Step 3 — parallel store work:
Task T005: SQLite migration in store.go
Task T007: DynamoDB struct update in dynamo/recipes.go

# Step 4 — SQLite CRUD (after migration):
Task T006: Update SQLite queries and scan functions in sqlite/recipes.go

# Step 5 — parallel handler work:
Task T008: Update recipes handler in handler/recipes.go
Task T009: Update admin handler in handler/admin_recipes.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Foundational (backend model + storage + API)
2. Complete Phase 3: US1 (garnish form inputs)
3. **STOP and VALIDATE**: Create and edit recipes with garnishes
4. Deploy/demo if ready

### Incremental Delivery

1. Complete Foundational → backend supports garnishes end-to-end
2. Add US1 → authors can create/edit garnishes (MVP!)
3. Add US2 → viewers see garnishes on detail page
4. Add US3 → garnishes appear in hover previews
5. Each story adds visible value without breaking previous stories

---

## Notes

- [P] tasks = different files, no dependencies on in-progress tasks
- TDD is mandatory (constitution §II): write and confirm tests fail BEFORE implementing
- `MAX_VISIBLE = 5` is defined in `frontend/src/components/RecipeCard.js` — do not hardcode this value elsewhere; reference the constant
- SQLite scan functions (`scanRecipe`, `scanRecipes`) must have columns and `Scan()` arguments updated together to avoid index mismatches
- DynamoDB: no migration needed; nil `Garnishes` on old items is safe (treated as empty)
- Commit after each phase checkpoint, not after every individual task
