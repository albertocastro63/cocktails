# Tasks: Ingredient Hover Popover

**Input**: Design documents from `specs/005-ingredient-hover-popover/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, quickstart.md ✅

**Tests**: Included — constitution mandates Test-First Development (write failing test before implementation).

**Organization**: Tasks grouped by user story. All work is confined to two files:
- `frontend/src/components/RecipeCard.js` (implementation)
- `frontend/src/components/RecipeCard.test.js` (tests)

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no shared dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Exact file paths included in every task description

---

## Phase 1: Setup

**Purpose**: Apply the one structural prerequisite that all popover tasks depend on.

- [ ] T001 Add `relative` positioning class to the RecipeCard outer container `div` in `frontend/src/components/RecipeCard.js` (required so the absolute-positioned popover is anchored to the card)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: No foundational infrastructure work required — this feature adds to a fully initialized project with an existing build system, test runner, and component structure.

**Checkpoint**: Phase 1 complete → user story implementation can begin.

---

## Phase 3: User Story 1 — View Ingredients on Hover (Priority: P1) 🎯 MVP

**Goal**: Hovering a recipe tile shows a popover with ingredient names; moving the mouse away hides it.

**Independent Test**: Hover over any recipe tile in the dev server — popover appears with ingredient names and disappears when the mouse leaves.

### Tests for User Story 1 ⚠️ Write first — confirm failure before T003

- [ ] T002 [US1] Add failing tests: "shows popover with ingredient names on mouseenter" and "hides popover on mouseleave" in `frontend/src/components/RecipeCard.test.js`

### Implementation for User Story 1

- [ ] T003 [US1] Implement `buildIngredientPopover(ingredients)` private helper that returns an absolutely-positioned DOM element listing ingredient names in `frontend/src/components/RecipeCard.js` (depends on T001)
- [ ] T004 [US1] Attach `mouseenter` listener to append the popover and `mouseleave` listener to remove it from the card in `frontend/src/components/RecipeCard.js` (depends on T003)

**Checkpoint**: US1 complete — hover shows ingredient names, mouseleave hides popover. All T002 tests pass.

---

## Phase 4: User Story 2 — Ingredient List Truncation (Priority: P2)

**Goal**: Recipes with more than 5 ingredients show only the first 5 in the popover, followed by "…" on the sixth line.

**Independent Test**: Hover over a recipe with 6+ ingredients — exactly 5 names appear followed by "…"; hover over one with ≤5 — all names shown, no ellipsis.

### Tests for User Story 2 ⚠️ Write first — confirm failure before T006

- [ ] T005 [US2] Add failing tests for truncation: recipe with ≤5 ingredients shows all (no ellipsis), recipe with exactly 6 shows first 5 + "…", recipe with >6 shows first 5 + "…" in `frontend/src/components/RecipeCard.test.js`

### Implementation for User Story 2

- [ ] T006 [US2] Update `buildIngredientPopover` to slice the ingredients array to the first 5 and append a "…" list item when `ingredients.length > 5` in `frontend/src/components/RecipeCard.js` (depends on T003)

**Checkpoint**: US2 complete — truncation works for all boundary values. All T005 tests pass.

---

## Phase 5: User Story 3 — Empty Ingredient State (Priority: P3)

**Goal**: Hovering a recipe tile with no ingredients still shows a popover with a "No ingredients listed." message.

**Independent Test**: Hover over a recipe with zero ingredients — popover appears with the message "No ingredients listed."

### Tests for User Story 3 ⚠️ Write first — confirm failure before T008

- [ ] T007 [US3] Add failing test: "shows 'No ingredients listed.' when recipe has no ingredients" in `frontend/src/components/RecipeCard.test.js`

### Implementation for User Story 3

- [ ] T008 [US3] Update `buildIngredientPopover` to render a "No ingredients listed." message when the `ingredients` array is empty (or undefined) in `frontend/src/components/RecipeCard.js` (depends on T003)

**Checkpoint**: US3 complete — empty state renders correctly. All T007 tests pass.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all stories.

- [ ] T009 [P] Run full Vitest suite in `frontend/` and confirm all tests pass with coverage ≥ 80% (`npm test` from `frontend/`)
- [ ] T010 [P] Manually verify all 6 scenarios listed in `specs/005-ingredient-hover-popover/quickstart.md` using the dev server (`npm run dev` from `frontend/`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: N/A — skipped
- **User Stories (Phases 3–5)**: All depend on T001 (Phase 1 completion)
  - US1 must complete before US2 and US3 (they extend the helper introduced in US1)
  - US2 and US3 can be implemented in either order after US1
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **US1 (P1)**: Depends on T001 only — no story dependencies
- **US2 (P2)**: Depends on US1 completion (extends `buildIngredientPopover`)
- **US3 (P3)**: Depends on US1 completion (extends `buildIngredientPopover`); independent of US2

### Within Each User Story

1. Write tests → confirm they **fail**
2. Implement feature → confirm tests **pass**
3. Run full suite → confirm no regressions

### Parallel Opportunities

- T009 and T010 (Phase 6) can run in parallel — different activities (automated vs. manual)
- After US1 is complete, US2 and US3 tests (T005, T007) can be written in parallel (both in the same file, but can be authored without conflicts if staged carefully)
- US2 and US3 implementations (T006, T008) touch the same function sequentially — no parallelism

---

## Parallel Example: Polish Phase

```bash
# Both can run simultaneously:
Task T009: npm test  (automated Vitest suite)
Task T010: Manual browser verification per quickstart.md
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: T001
2. Complete Phase 3: T002 → T003 → T004
3. **STOP and VALIDATE**: All US1 tests pass; hover works in dev server
4. Demo / confirm before proceeding to truncation and empty state

### Incremental Delivery

1. T001 (Setup) → T002–T004 (US1: basic hover) → validate
2. T005–T006 (US2: truncation) → validate
3. T007–T008 (US3: empty state) → validate
4. T009–T010 (Polish) → ship

---

## Notes

- [P] tasks = different activities/files, no shared dependencies
- [Story] label maps every task to its user story for traceability
- Constitution mandates TDD: every test task MUST run and FAIL before its paired implementation task begins
- Commit after each checkpoint (end of each user story phase)
- `buildIngredientPopover` starts simple in T003 and is extended incrementally in T006 and T008 — keep each change minimal and targeted
- Total tasks: 10 (1 setup + 3×US1 + 2×US2 + 2×US3 + 2 polish)
