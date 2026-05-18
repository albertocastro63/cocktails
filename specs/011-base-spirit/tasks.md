# Tasks: Base Spirit Designation for Recipe Ingredients

**Input**: Design documents from `specs/011-base-spirit/`  
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/api.md ✅, quickstart.md ✅

**Note**: Tests are included per the constitution's mandatory TDD requirement (Principle II). Write each test task, confirm it fails, then implement.

**Organization**: Tasks are grouped by user story. Phase 2 (Foundational) must complete before any user story work begins. US2 and US3 are independently implementable after Phase 2.

---

## Phase 1: Setup

No new project initialization needed — existing structure is used. This phase has no tasks.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The `IsBaseSpirit` field must exist on `model.Ingredient` and on the DynamoDB `ingItem` struct before any frontend or handler work can reference or serialize it. T001 and T002 touch different files and run in parallel. T003 tests the backend API round-trip and can also run in parallel (different file).

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T001 [P] Add `IsBaseSpirit bool \`json:"is_base_spirit,omitempty"\`` to `Ingredient` struct in `backend/internal/model/model.go` — existing SQLite serialization tests in `go test ./...` cover the JSON round-trip; no new test file is required for this struct-field addition (C3)
- [X] T002 [P] Add `BaseSpirit bool \`dynamodbav:"is_base_spirit,omitempty"\`` to `ingItem` struct and update `toItem`/`fromItem` mapping functions in `backend/internal/store/dynamo/recipes.go` — the existing DynamoDB integration test suite covers attribute marshaling; no new test file required (C3)
- [X] T003 [P] Write a failing handler test in `backend/internal/handler/recipes_test.go` confirming `is_base_spirit: true` on one ingredient survives a full `POST /api/v1/recipes` → `GET /api/v1/recipes/{id}` round-trip: send a payload with one ingredient flagged, assert the response body contains `"is_base_spirit": true` on that ingredient and no other (C2)

**Checkpoint**: `go build ./...` passes. The field is present in both the JSON API shape and the DynamoDB attribute envelope. The handler test fails (red) and will turn green after T001 lands.

---

## Phase 3: User Story 1 — Mark Base Spirit When Authoring (Priority: P1) 🎯 MVP

**Goal**: Authors can check a base spirit toggle on one ingredient row; checking another automatically clears the previous; unchecking leaves none selected; the state persists on save and is restored on edit.

**Independent Test**: Create a recipe with three ingredients, mark the second, save — only the second shows `is_base_spirit: true` in the API response. Re-edit, mark the third, save — only the third is marked. Re-edit, uncheck the third, save — no ingredient has `is_base_spirit`.

### Tests for User Story 1

> **Write tests FIRST. Confirm they FAIL before implementing T005–T008.**

- [X] T004 [US1] Write failing tests for base spirit toggle behaviour in `frontend/src/pages/RecipeForm.test.js`:
  - A newly added ingredient row contains a base spirit checkbox
  - Checking checkbox on row B when row A is already checked: B becomes checked, A is automatically cleared
  - Unchecking the active checkbox leaves all rows unchecked
  - Deleting the row whose base spirit checkbox is checked leaves all remaining rows unchecked (FR-009) (C1)
  - Prefill (edit mode): ingredient with `is_base_spirit: true` has its checkbox restored; others do not
  - Submit payload: the marked ingredient has `is_base_spirit: true`; all others omit the field (or have `false`)

### Implementation for User Story 1

- [X] T005 [US1] Update the ingredient row builder in `frontend/src/pages/RecipeForm.js` to append a base spirit checkbox `<input type="checkbox" name="ing_base_spirit">` with an amber label to each ingredient row (after the Unit input, before the × button)
- [X] T006 [US1] Add mutual-exclusion click handler in `frontend/src/pages/RecipeForm.js`: when a base spirit checkbox is checked, find all sibling ingredient rows via `ingredientsSection._rows.children` and uncheck their base spirit inputs; when unchecked, leave all others as-is
- [X] T007 [US1] Update the edit prefill block in `frontend/src/pages/RecipeForm.js` so that when loading an existing recipe, the ingredient row with `ing.is_base_spirit === true` has its checkbox set to `checked`
- [X] T008 [US1] Update the submit handler in `frontend/src/pages/RecipeForm.js` to read `row.querySelector('[name="ing_base_spirit"]').checked` for each ingredient row and include `is_base_spirit: true` in the payload only when `checked`

**Checkpoint**: User Story 1 fully functional. Authors can set, change, and clear the base spirit. The correct `is_base_spirit` state persists through save/reload cycles.

---

## Phase 4: User Story 2 — Base Spirit Highlighted in Ingredient Hover Popover (Priority: P2)

**Goal**: When a viewer hovers over a recipe card, the base spirit ingredient is visually distinguished (bold + amber `(base spirit)` label) in the popover list. Recipes without a base spirit show uniform ingredient rows.

**Independent Test**: On the recipe list page, hover over a recipe with `is_base_spirit` set — the named ingredient shows bold text and an amber `(base spirit)` label. Hover over a recipe without one — all ingredient rows look identical.

### Tests for User Story 2

> **Write tests FIRST. Confirm they FAIL before implementing T010.**

- [X] T009 [US2] Write failing tests for base spirit highlight in popover in `frontend/src/components/RecipeCard.test.js`:
  - When one ingredient has `is_base_spirit: true`, the rendered popover list item contains a `(base spirit)` label or `font-semibold` class (I1)
  - When no ingredient has `is_base_spirit`, all list items are rendered identically (no label, no extra class)
  - The highlight appears only on the ingredient with the flag, not on neighbouring rows

### Implementation for User Story 2

- [X] T010 [US2] Update `buildIngredientPopover()` in `frontend/src/components/RecipeCard.js`: when rendering each ingredient `li`, if `ingredient.is_base_spirit` is truthy apply `font-semibold text-stone-900` to the name span and append `<span class="text-amber-600 text-xs ml-1">(base spirit)</span>` inline

**Checkpoint**: Hovering over a recipe card with a base spirit shows the highlighted ingredient; recipes without one remain uniform.

---

## Phase 5: User Story 3 — Base Spirit Highlighted on Recipe Detail Page (Priority: P3)

**Goal**: On the recipe detail page, the base spirit ingredient is visually distinguished in the full ingredient list using the same visual treatment as the popover (FR-008 consistency requirement).

**Independent Test**: Open the recipe detail page for a recipe with a base spirit — the designated ingredient shows bold + amber `(base spirit)` label in the ingredient list. The visual treatment is identical to the popover. Open a recipe without a base spirit — all ingredients are uniform.

### Tests for User Story 3

> **Write tests FIRST. Confirm they FAIL before implementing T012.**

- [X] T011 [US3] Write failing tests for base spirit highlight in `frontend/src/components/IngredientList.test.js`:
  - When one ingredient has `is_base_spirit: true`, the rendered list item contains a `(base spirit)` label or `font-semibold` class
  - When no ingredient has `is_base_spirit`, all list items are rendered identically
  - Visual treatment matches the popover pattern (same class/label strings as T009/T010)

### Implementation for User Story 3

- [X] T012 [US3] Update `IngredientList.js` in `frontend/src/components/IngredientList.js`: when rendering each ingredient `li`, if `ingredient.is_base_spirit` is truthy apply `font-semibold text-stone-900` to the name span and append `<span class="text-amber-600 text-xs ml-1">(base spirit)</span>` — identical token usage to T010

**Checkpoint**: The detail page ingredient list highlights the base spirit with the same visual treatment as the card popover.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T013 Run all backend tests and confirm they pass: `cd backend && go test ./...`
- [X] T014 [P] Run all frontend tests and confirm they pass: `cd frontend && npm test`
- [ ] T015 [P] Manually validate quickstart.md scenarios 1–7 against the running dev server

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 2)**: No dependencies — start immediately; T001, T002, T003 are all parallel
- **US1 (Phase 3)**: Depends on Phase 2 complete (T001 must land for model to be usable in frontend payload)
- **US2 (Phase 4)**: Depends on Phase 2 complete; does NOT depend on US1
- **US3 (Phase 5)**: Depends on Phase 2 complete; does NOT depend on US1 or US2
- **Polish (Phase 6)**: Depends on all desired user stories complete

### User Story Dependencies

- **US1 (P1)**: Independent after Phase 2
- **US2 (P2)**: Independent after Phase 2 — `RecipeCard.js` is a different file from `RecipeForm.js`
- **US3 (P3)**: Independent after Phase 2 — `IngredientList.js` is a different file from both

### Within Each User Story

1. Write tests → confirm they fail
2. Implement → confirm tests pass
3. Mark tasks [X] as each completes

---

## Parallel Opportunities

```bash
# Phase 2 — all three run in parallel (different files):
T001  # model.go
T002  # dynamo/recipes.go
T003  # handler/recipes_test.go

# Phase 3 — tests must precede implementation:
T004 → T005 → T006 → T007 → T008  # sequential (same file: RecipeForm.js)

# After Phase 2, all user story phases can start in parallel:
Phase 3 (US1 — RecipeForm.js)     # independent file
Phase 4 (US2 — RecipeCard.js)     # independent file
Phase 5 (US3 — IngredientList.js) # independent file

# Phase 6 — T014 and T015 parallel with each other (after T013):
T013 → T014 [P]
      T015 [P]
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Foundational (T001–T003)
2. Complete Phase 3: User Story 1 (T004–T008)
3. **STOP and VALIDATE**: Recipes can be created and edited with a base spirit; the field persists in the API
4. Deliver — authoring is fully functional even before the display highlights land

### Incremental Delivery

1. Phase 2 → Phase 3 (US1): Base spirit authoring works end-to-end
2. Phase 4 (US2): Popover highlight ships — viewers see base spirit on hover
3. Phase 5 (US3): Detail page highlight ships — consistent experience across all surfaces

---

## Notes

- `is_base_spirit` uses `omitempty` in both JSON and DynamoDB tags — the field is absent from the wire format when `false`, so legacy recipe data requires no migration.
- The mutual-exclusion logic in T006 operates on the live DOM (iterating `ingredientsSection._rows.children`) so it works correctly even when ingredient rows are added or removed during editing.
- The visual treatment (T010, T012) uses identical class strings and label text to satisfy FR-008 (consistency between popover and detail page); keep them in sync if either is changed later.
- The `(base spirit)` label is rendered as an inline `<span>` so it does not break the flex layout of quantity/unit on the right side of the ingredient row.
- FR-009 (deletion of base spirit ingredient) is covered by a test bullet in T004 and implicitly by the DOM removal pattern — no additional handler is needed since the checkbox is removed with its row.
