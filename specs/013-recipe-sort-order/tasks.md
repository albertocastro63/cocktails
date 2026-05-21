# Tasks: Recipe Sort Order on All Recipes Page

**Input**: Design documents from `specs/013-recipe-sort-order/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, quickstart.md ✅

**Organization**: Phase 2 writes all failing tests first (Constitution II — TDD mandatory). US1 and US2 are delivered by the same `SortButtonGroup` + `RecipeList` implementation. US3 (accessibility) is verified as a dedicated checkpoint once the component exists.

---

## Phase 2: Foundational (TDD — Write Failing Tests First)

**Purpose**: Write all tests before any implementation begins. Both test files can be written in parallel (different files, no dependencies on each other).

**⚠️ CRITICAL**: Run `npm test` from `frontend/` after T001 and T002 — ALL new tests MUST fail before proceeding to implementation.

- [X] T001 [P] Write failing unit tests for `SortButtonGroup` in `frontend/src/components/SortButtonGroup.test.js` — import `SortButtonGroup` from `./SortButtonGroup.js`; cover: (1) renders a button labeled "A→Z" and a button labeled "Z→A"; (2) when `currentDir` is `null`, both buttons have `aria-pressed="false"`; (3) when `currentDir='asc'`, A→Z button has `aria-pressed="true"`, Z→A has `aria-pressed="false"`; (4) when `currentDir='desc'`, Z→A button has `aria-pressed="true"`, A→Z has `aria-pressed="false"`; (5) clicking A→Z calls `onSort('asc')`; (6) clicking Z→A calls `onSort('desc')`; (7) active button has class `bg-amber-100`; (8) inactive button has class `bg-white`
- [X] T002 [P] Add failing sort behavior tests to `frontend/src/pages/RecipeList.test.js` — mock `getRecipes` to return objects `[{ id: 'r1', name: 'Zombie', ingredients: [] }, { id: 'r2', name: 'Aperol Spritz', ingredients: [] }, { id: 'r3', name: 'Margarita', ingredients: [] }]`; cover: (1) page renders a button labeled "A→Z" and a button labeled "Z→A" on load; (2) neither sort button has `aria-pressed="true"` on initial render; (3) clicking "A→Z" shows recipe cards in order Aperol Spritz, Margarita, Zombie; (4) clicking "Z→A" shows them in order Zombie, Margarita, Aperol Spritz; (5) switching from A→Z to Z→A immediately updates the displayed order; (6) case-insensitivity: with a recipe named `"margarita"` (all lowercase), clicking "A→Z" still places it between Aperol Spritz and Zombie (same position as "Margarita"); (7) empty list: when `getRecipes` returns `{ data: [] }`, clicking "A→Z" produces no JavaScript error and the empty state message remains visible

**Checkpoint**: `npm test` from `frontend/` shows the new tests failing with "Cannot find module" or assertion failures. No implementation exists yet.

---

## Phase 3: User Story 1 — Sort Recipes A→Z (Priority: P1) 🎯 MVP

**Goal**: Clicking the A→Z button immediately reorders the recipe list in ascending alphabetical order. The A→Z button shows as active (highlighted).

**Independent Test**: Open the all-recipes page. Click "A→Z". Verify the first recipe card shown has a name that comes first alphabetically and the "A→Z" button appears highlighted.

### Implementation for User Story 1

- [X] T003 [US1] Create `frontend/src/components/SortButtonGroup.js` — export `SortButtonGroup({ currentDir, onSort })` that returns a `div` with `role="group"` and `aria-label="Sort recipes"`; append two `<button>` elements: label "A→Z" with `data-dir="asc"` and label "Z→A" with `data-dir="desc"`; each button gets `aria-pressed` set to `"true"` if its `data-dir` matches `currentDir`, else `"false"`; active button (matching `currentDir`) gets Tailwind classes `bg-amber-100 text-amber-800 border-amber-300 font-semibold px-3 py-1 text-sm border`; inactive button gets `bg-white text-stone-600 border-stone-200 hover:bg-stone-50 px-3 py-1 text-sm border`; both buttons get `focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-amber-500`; clicking a button calls `onSort(button.dataset.dir)`
- [X] T004 [US1] Add `sortRecipes` export and `SortButtonGroup` integration to `frontend/src/pages/RecipeList.js` — (1) add `export function sortRecipes(recipes, direction)` that returns `recipes.slice().sort((a, b) => direction === 'asc' ? a.name.localeCompare(b.name, undefined, { sensitivity: 'base' }) : b.name.localeCompare(a.name, undefined, { sensitivity: 'base' }))`; (2) add `let currentSortDir = null` alongside `currentQ`; (3) import `SortButtonGroup` from `../components/SortButtonGroup.js`; (4) insert a `sortWrap` div between `searchWrap` and `content`; (5) render `SortButtonGroup({ currentDir: currentSortDir, onSort })` into `sortWrap`; (6) implement `onSort(dir)` that sets `currentSortDir = dir`, re-renders `SortButtonGroup` with updated `currentDir`, and re-renders the recipe grid using `sortRecipes(data, dir)` on the currently loaded data; (7) after `getRecipes` resolves, store results in `let data = []`, then render using `currentSortDir ? sortRecipes(data, currentSortDir) : data`

**Checkpoint**: `npm test` from `frontend/` — all T001 and T002 tests pass. Open the dev server and verify the "A→Z" button reorders recipes alphabetically when clicked.

---

## Phase 4: User Story 2 — Sort Recipes Z→A (Priority: P2)

**Goal**: Clicking Z→A reorders the list in descending alphabetical order. Switching between A→Z and Z→A works correctly and is idempotent.

**Independent Test**: Open the all-recipes page. Click "Z→A". Verify the first recipe shown has a name that comes last alphabetically. Click "A→Z". Verify the order reverses immediately.

### Verification for User Story 2

- [X] T005 [US2] Verify Z→A sort and direction-switching in `frontend/src/pages/RecipeList.js` — run `npm test` from `frontend/` and confirm all sort tests in `RecipeList.test.js` pass (including Z→A and direction-switching scenarios); open the dev server and click Z→A then A→Z to confirm bidirectional switching works and the active button highlights correctly; confirm clicking the already-active button a second time leaves the list order and button state unchanged (idempotent); if any test fails, fix the root cause in `RecipeList.js` (check `sortRecipes` for `'desc'` direction and the `onSort` callback logic) before proceeding

**Checkpoint**: All sort-direction tests green. Both A→Z and Z→A work correctly in the browser.

---

## Phase 5: User Story 3 — Accessible Sort Controls (Priority: P3)

**Goal**: The sort buttons are keyboard-focusable, activatable with Enter/Space, announced correctly by screen readers, and meet WCAG 2.1 AA color contrast.

**Independent Test**: Tab to the sort controls. Activate each with Enter or Space. Verify visible focus ring appears. Use a screen reader to confirm it announces the button label and pressed state.

### Verification for User Story 3

- [X] T006 [US3] Audit and confirm WCAG 2.1 AA compliance of `frontend/src/components/SortButtonGroup.js` — verify the following are all present in the rendered HTML: `role="group"` on the container; `aria-label="Sort recipes"` on the container; `aria-pressed="true"|"false"` on each button (dynamic, matching `currentDir`); `data-dir` on each button; `focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-amber-500` for keyboard focus ring; amber active state contrast: `bg-amber-100 text-amber-800` passes 4.5:1 (verify visually); stone inactive contrast: `text-stone-600` on white passes 4.5:1 (verify visually); if any attribute is missing, add it to `SortButtonGroup.js` and re-run `npm test` to confirm all T001 tests still pass

**Checkpoint**: Tab to sort buttons in the browser — a visible amber focus ring appears. Screen reader (VoiceOver: Cmd+F5, then Tab to buttons) announces "A to Z, toggle button, not pressed" and "Z to A, toggle button, not pressed" (or equivalent).

---

## Phase 6: Polish & Verification

**Purpose**: Confirm coverage requirements and run all manual quickstart scenarios.

- [X] T007 Run `npm test -- --coverage` from `frontend/` and verify line coverage ≥ 80% for `frontend/src/components/SortButtonGroup.js` and `frontend/src/pages/RecipeList.js`; if below threshold, add targeted tests to cover the gaps
- [X] T008 Start dev server (`npm run dev` from `frontend/`), then manually run through quickstart.md scenarios SC-001 through SC-007 — verify A→Z sort, Z→A sort, default page load state (no direction active), keyboard navigation, screen reader announcement, empty list graceful handling, and sort+search composition all pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 2)**: No dependencies — start immediately; T001 and T002 run in parallel
- **US1 (Phase 3)**: Depends on Phase 2 complete; T003 then T004 are sequential (T004 imports T003's component)
- **US2 (Phase 4)**: Depends on Phase 3 complete; T005 is a verification step with no code changes expected
- **US3 (Phase 5)**: Depends on Phase 3 complete (T006 audits the component created in T003); can run in parallel with T005
- **Polish (Phase 6)**: Depends on all previous phases; T007 then T008 sequential

### Task Sequencing

T001 ∥ T002 (parallel, different files) → T003 (depends on T001) → T004 (depends on T002 + T003) → T005 ∥ T006 (parallel, different files) → T007 → T008

### Parallel Opportunities

```bash
# Phase 2 — both run in parallel (different files):
T001  # SortButtonGroup.test.js
T002  # RecipeList.test.js

# Phase 3 — sequential (T004 imports from T003):
T003 → T004

# Phase 4 + 5 — parallel (different files):
T005  # RecipeList.js verification
T006  # SortButtonGroup.js audit

# Phase 6 — sequential:
T007 (coverage) → T008 (manual verification)
```

---

## Implementation Strategy

### MVP First

1. Complete Phase 2: Foundational (T001–T002 — write failing tests)
2. Complete Phase 3: US1 (T003–T004 — implement component + integration)
3. **STOP and VALIDATE**: `npm test` passes; clicking A→Z in browser reorders recipes
4. Complete Phase 4: US2 (T005 — verify Z→A already works)
5. Complete Phase 5: US3 (T006 — verify accessibility)
6. Complete Phase 6: Polish (T007–T008 — coverage + manual scenarios)

### Incremental Delivery

This feature is purely additive (no existing behavior changes). US1 alone (sort A→Z) is a shippable increment. US2 (Z→A) is delivered free by the same component; it requires no additional implementation once US1 is done. US3 (accessibility) should be baked into T003 from the start and verified in T006.

---

## Notes

- TDD is mandatory (Constitution II): always run `npm test` after T001/T002 to confirm tests fail before writing implementation.
- `sortRecipes` in `RecipeList.js` should be a named export so it can be tested directly in isolation if needed.
- The `data` variable in `RecipeList` must never be mutated — `sortRecipes` uses `.slice()` to create a copy.
- Re-render on sort is done by clearing the grid and re-appending cards (same pattern as `loadRecipes`); do not call `getRecipes` again.
- `SortButtonGroup` is stateless — `RecipeList` owns `currentSortDir` and passes it down.
