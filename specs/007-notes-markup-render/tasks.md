---

description: "Task list for Notes Rendered Markup Styling"
---

# Tasks: Notes Rendered Markup Styling

**Input**: Design documents from `specs/007-notes-markup-render/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅

**Organization**: Tasks are grouped by user story. Each story is in a different set of files, so all three user stories can be worked in parallel once Phase 1 (plugin setup) is complete.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- TDD is mandatory per the Constitution: write each failing test before the corresponding implementation task

---

## Phase 1: Setup & Configuration (Foundational)

**Purpose**: Install and configure the Tailwind Typography plugin. This phase MUST complete before any user story work begins — the `prose` class has no effect without the plugin.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T001 Add `@tailwindcss/typography` to devDependencies: run `npm install --save-dev @tailwindcss/typography` inside `frontend/` — updates `frontend/package.json`
- [x] T002 Register the typography plugin in `frontend/tailwind.config.js` by importing and adding it to the `plugins` array (depends on T001)

**Checkpoint**: Running `npm test` in `frontend/` still passes all existing tests. The `prose` class is now active.

---

## Phase 2: User Story 1 — Visually Styled Preview While Editing (Priority: P1) 🎯 MVP

**Goal**: Render the MarkdownEditor preview with full typography styling so authors can see visually distinct headings, bold, lists, blockquotes, and code when they press "Preview".

**Independent Test**: Open the recipe create or edit form. Type `## Heading\n**bold** and\n- a list\n> blockquote` in notes. Press "Preview". Confirm the heading is visually larger, bold is heavier, the list has a bullet, and the blockquote is visually distinguished.

### Tests for User Story 1

> **Write these tests FIRST, confirm they FAIL, then implement**

- [x] T003 [US1] Add a failing test to `frontend/src/components/MarkdownEditor.test.js`: after clicking Preview, assert that the `[data-preview]` element's `className` contains `prose`

### Implementation for User Story 1

- [x] T004 [US1] Update the preview div class in `frontend/src/components/MarkdownEditor.js` from `'prose w-full border border-gray-200 rounded-lg px-3 py-2 min-h-[4.5rem] text-gray-700'` to `'prose prose-stone max-w-none w-full border border-gray-200 rounded-lg px-3 py-2 min-h-[4.5rem]'`

**Checkpoint**: Run `npm test` in `frontend/`. T003 test now passes. All pre-existing MarkdownEditor tests still pass.

---

## Phase 3: User Story 2 — Consistently Styled Notes on Recipe Detail Page (Priority: P2)

**Goal**: Render the notes section on the recipe detail page with the same prose typography as the editor preview.

**Independent Test**: Create a recipe with notes `## Tips\n**Shake well.**\n- Ice\n- Lime`. Navigate to the recipe detail page. Confirm the heading is visually larger, bold is heavier, and the list has bullets — matching the editor preview.

### Tests for User Story 2

> **Write these tests FIRST, confirm they FAIL, then implement**

- [x] T005 [US2] Add a failing test to `frontend/src/pages/RecipeDetail.test.js`: when a recipe with non-empty notes is rendered, assert that the notes container div's `className` contains `prose`

### Implementation for User Story 2

- [x] T006 [US2] Update the notes div class in `frontend/src/pages/RecipeDetail.js` from `'text-stone-700'` to `'prose prose-stone max-w-none'`

**Checkpoint**: Run `npm test` in `frontend/`. T005 test now passes. All pre-existing RecipeDetail tests still pass.

---

## Phase 4: User Story 3 — Consistently Styled Notes on Homepage (Priority: P3)

**Goal**: Render the featured recipe's notes on the homepage with the same prose typography as the recipe detail page.

**Independent Test**: Ensure a recipe with markdown notes is the featured recipe on the homepage. Navigate to the homepage and confirm the notes render with the same visual formatting as observed on the detail page.

### Tests for User Story 3

> **Write these tests FIRST, confirm they FAIL, then implement**

- [x] T007 [US3] Add a failing test to `frontend/src/pages/Home.test.js`: when a recipe with non-empty notes is rendered, assert that the notes container div's `className` contains `prose`

### Implementation for User Story 3

- [x] T008 [US3] Update the notes div class in `frontend/src/pages/Home.js` from `'text-stone-700'` to `'prose prose-stone max-w-none'`

**Checkpoint**: Run `npm test` in `frontend/`. T007 test now passes. All pre-existing Home tests still pass.

---

## Phase 5: Polish & Validation

**Purpose**: Full suite confirmation and visual validation.

- [x] T009 Run the complete test suite (`npm test` in `frontend/`) and confirm all tests pass with no regressions
- [ ] T010 [P] Start the dev server (`npm run dev` in `frontend/`) and manually verify all three surfaces render styled markdown: editor preview, recipe detail page, and homepage
- [x] T011 [P] Add `overflow-x-auto` to the notes container class in all three surfaces (`frontend/src/components/MarkdownEditor.js`, `frontend/src/pages/RecipeDetail.js`, `frontend/src/pages/Home.js`) so wide markdown tables scroll horizontally rather than overflowing the layout

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately. T001 and T002 are in different files and can run in parallel.
- **Phase 2, 3, 4 (User Stories)**: All depend on Phase 1 completion (plugin must be installed and configured first).
  - The three user stories are independent of each other (different files) and can proceed in parallel after Phase 1.
  - Or sequentially in priority order: US1 → US2 → US3.
- **Phase 5 (Polish)**: Depends on all desired user stories being complete.

### User Story Dependencies

- **US1 (P1)**: Starts after Phase 1 — no dependency on US2 or US3
- **US2 (P2)**: Starts after Phase 1 — no dependency on US1 or US3
- **US3 (P3)**: Starts after Phase 1 — no dependency on US1 or US2

### Within Each User Story

1. Write failing test (T003 / T005 / T007)
2. Run tests — confirm the new test FAILS
3. Implement class change (T004 / T006 / T008)
4. Run tests — confirm the new test PASSES and existing tests still pass

---

## Parallel Example: After Phase 1 Completes

```bash
# These three tracks can run in parallel:

Track A (US1):
  Task: "T003 — Write failing prose-class test in MarkdownEditor.test.js"
  Task: "T004 — Update preview div class in MarkdownEditor.js"

Track B (US2):
  Task: "T005 — Write failing prose-class test in RecipeDetail.test.js"
  Task: "T006 — Update notes div class in RecipeDetail.js"

Track C (US3):
  Task: "T007 — Write failing prose-class test in Home.test.js"
  Task: "T008 — Update notes div class in Home.js"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Install and configure plugin
2. Complete Phase 2: US1 — styled editor preview
3. **STOP and VALIDATE**: Confirm preview renders headings, lists, bold, etc. visually
4. Ship if ready

### Incremental Delivery

1. Phase 1: Plugin setup → foundation ready
2. Phase 2: US1 → editor preview styled → demo/validate
3. Phase 3: US2 → detail page styled → demo/validate
4. Phase 4: US3 → homepage styled → demo/validate
5. Each increment adds value without breaking previous surfaces

---

## Notes

- [P] tasks = different files, no dependencies between them
- [Story] label maps each task to its user story for traceability
- TDD is mandatory: every implementation task must be preceded by a failing test task
- The only code changes in this feature are CSS class strings — no logic, no API, no data model
- Commit after each user story phase checkpoint
