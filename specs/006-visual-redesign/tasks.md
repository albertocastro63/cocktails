# Tasks: Visual Redesign

**Input**: Design documents from `specs/006-visual-redesign/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, quickstart.md ✅

**Tests**: Included — constitution mandates Test-First Development. For visual changes, tests assert the presence of key new CSS class tokens on rendered DOM elements (they fail against the current classes, pass after the update).

**Design system summary** (from research.md):
- Base bg: `bg-stone-50` | Card bg: `bg-white`
- Accent: `bg-amber-500` with `text-stone-900` (WCAG AA compliant — contrast ≈ 9:1)
- Accent text/labels: `text-amber-700` (sufficient contrast on light backgrounds)
- Nav: `bg-stone-900 text-stone-100 hover:text-amber-400`
- Headings: `text-stone-900` | Body: `text-stone-700` | Muted: `text-stone-500`
- Buttons: `bg-amber-500 text-stone-900 font-semibold rounded-xl`
- Inputs: `border-stone-300 rounded-xl focus:ring-amber-400`
- Cards: `rounded-2xl border border-stone-200 shadow-sm border-l-4 border-l-amber-400`
- Section labels: `text-sm font-semibold uppercase tracking-widest text-amber-700`

**Organization**: Tasks grouped by user story. All changes are in `frontend/src/`.

---

## Phase 1: Setup

**Purpose**: Apply the one global base style that all pages inherit.

- [x] T001 Add `body { @apply bg-stone-50; }` base style to `frontend/src/index.css`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: No package or config changes needed — `stone` and `amber` are built-in Tailwind v3 palettes. No foundational infrastructure work required beyond T001.

**Checkpoint**: Phase 1 complete → user story implementation can begin.

---

## Phase 3: User Story 1 — Home Page & Navigation (Priority: P1) 🎯 MVP

**Goal**: Visitor lands on a branded, visually rich home page with a dark stone navigation bar and an amber-accented hero section.

**Independent Test**: Load home page — dark nav bar visible at top, hero band with stone-900 background present, amber accent visible on CTA or heading.

### Tests for User Story 1 ⚠️ Write first — confirm failure before T004

- [x] T002 [US1] Add failing test in `frontend/src/pages/Home.test.js`: assert home page wrapper contains an element with a dark gradient/stone class (e.g. `from-stone-900`)
- [x] T003 [US1] Create `frontend/src/main.test.js` with failing test: assert `buildNav()` returns an element with class `bg-stone-900`

### Implementation for User Story 1

- [x] T004 [US1] Redesign nav in `frontend/src/main.js`: replace `bg-white border-b border-gray-200` with `bg-stone-900`; update link classes to `text-stone-100 hover:text-amber-400`; update "New Recipe" button to `bg-amber-500 text-stone-900 font-semibold rounded-xl`; update "Sign In" link to amber style (depends on T001)
- [x] T005 [US1] Update `frontend/src/pages/Home.js`: add full-width hero band (`bg-gradient-to-br from-stone-900 to-stone-800 text-white`) with large heading and amber-tinted subtext; update section heading styles to use `text-amber-700 text-sm font-semibold uppercase tracking-widest`; update page body to `max-w-2xl mx-auto px-4 py-8` (depends on T004)

**Checkpoint**: US1 complete — dark nav and hero band visible. T002 and T003 tests pass.

---

## Phase 4: User Story 2 — Recipe List & Cards (Priority: P2)

**Goal**: Recipe list displays visually engaging amber-accented cards on a stone-50 background.

**Independent Test**: Open recipe list — cards have visible amber left border, rounded-2xl corners, and shadow; background is warm stone not white; search input has amber focus ring.

### Tests for User Story 2 ⚠️ Write first — confirm failure before T007

- [x] T006 [P] [US2] Add failing tests in `frontend/src/components/RecipeCard.test.js`: assert card element has class `rounded-2xl` and contains element with class `border-l-amber-400`
- [x] T007 [P] [US2] Add failing test in `frontend/src/components/EmptyState.test.js`: assert empty state uses `text-stone-500` class (not legacy `text-gray-400`)
- [x] T008 [P] [US2] Add failing test in `frontend/src/components/SearchBar.test.js`: assert input has class `focus:ring-amber-400`

### Implementation for User Story 2

- [x] T009 [P] [US2] Update `frontend/src/components/RecipeCard.js`: replace `rounded-lg shadow` with `rounded-2xl border border-stone-200 shadow-sm border-l-4 border-l-amber-400`; update hover to `hover:shadow-lg hover:border-l-amber-500`; update heading to `text-stone-900`; update sub-text to `text-stone-500` (depends on T001)
- [x] T010 [P] [US2] Update `frontend/src/pages/RecipeList.js`: update page heading to `text-3xl font-bold text-stone-900`; ensure wrapper uses stone-50 background (inherited from body, no class change needed)
- [x] T011 [P] [US2] Update `frontend/src/components/EmptyState.js`: replace `text-gray-400` with `text-stone-500`; update icon/container styling to match design language
- [x] T012 [P] [US2] Update `frontend/src/components/SearchBar.js`: replace `focus:ring-indigo-500` with `focus:ring-amber-400 focus:border-transparent`; replace `border-gray-300` with `border-stone-300`; update clear button hover to `hover:text-amber-600`; apply `rounded-xl` to input

**Checkpoint**: US2 complete — amber-accented cards and updated search bar. T006–T008 tests pass.

---

## Phase 5: User Story 3 — Recipe Detail Page (Priority: P3)

**Goal**: Recipe detail page uses amber uppercase section labels, updated button styles, and stone text colors throughout.

**Independent Test**: Open any recipe detail page — "Ingredients", "Steps", "Notes" labels appear in small uppercase amber text; Edit button uses updated secondary outline style.

### Tests for User Story 3 ⚠️ Write first — confirm failure before T014

- [x] T013 [P] [US3] Add failing test in `frontend/src/components/IngredientList.test.js`: assert ingredient name `<span>` has class `text-stone-800` (not legacy `text-gray-800`)
- [x] T014 [P] [US3] Add failing test in `frontend/src/pages/RecipeDetail.test.js`: assert recipe title `<h1>` has class `text-stone-900`
- [x] T014b [P] [US3] Add failing test in `frontend/src/components/PropertyTable.test.js`: assert key `<span>` has class `text-stone-700`

### Implementation for User Story 3

- [x] T015 [P] [US3] Update `frontend/src/components/IngredientList.js`: replace `text-gray-800` with `text-stone-800`; replace `text-gray-500` with `text-stone-500`; replace `divide-gray-100` with `divide-stone-100`
- [x] T015b [P] [US3] Update `frontend/src/components/PropertyTable.js`: replace `text-gray-600` with `text-stone-600`; replace `text-gray-800` with `text-stone-800`; replace `divide-gray-100` with `divide-stone-100`
- [x] T016 [P] [US3] Update `frontend/src/pages/RecipeDetail.js`: replace `text-gray-900` with `text-stone-900` on title; replace section heading classes `text-lg font-semibold text-gray-700` with `text-sm font-semibold uppercase tracking-widest text-amber-700 mt-6 mb-2`; update Edit button to `border border-stone-300 text-stone-700 hover:border-amber-500 hover:text-amber-700 rounded-xl`; update Delete button to `border border-red-300 text-red-600 hover:bg-red-50 rounded-xl`

**Checkpoint**: US3 complete — amber section labels and updated buttons on detail page. T013–T014 tests pass.

---

## Phase 6: User Story 4 — Sign In Page (Priority: P4)

**Goal**: Sign In page presents a polished card form with amber-accented button and focus rings.

**Independent Test**: Open Sign In page — form is contained in a white card on stone-50 background; submit button is amber with dark text; input focus ring is amber.

### Tests for User Story 4 ⚠️ Write first — confirm failure before T018

- [x] T017 [US4] Add failing test in `frontend/src/pages/Login.test.js`: assert submit button has class `bg-amber-500`

### Implementation for User Story 4

- [x] T018 [US4] Update `frontend/src/pages/Login.js`: wrap form in `bg-white rounded-2xl shadow-sm border border-stone-200 p-8`; replace `text-gray-900` heading with `text-stone-900`; replace `border-gray-300` inputs with `border-stone-300 rounded-xl focus:ring-amber-400`; replace `bg-indigo-600 hover:bg-indigo-700` submit button with `bg-amber-500 hover:bg-amber-600 text-stone-900 font-semibold rounded-xl`; replace `text-gray-700` labels with `text-stone-700`

**Checkpoint**: US4 complete — polished login form with amber styling. T017 test passes.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Apply design system to RecipeForm, run full suite, verify manually.

- [x] T019a Add failing test in `frontend/src/pages/RecipeForm.test.js`: assert submit button has class `bg-amber-500`; confirm test fails before T019
- [x] T019 Update `frontend/src/pages/RecipeForm.js`: replace all `border-gray-300` inputs with `border-stone-300 rounded-xl focus:ring-amber-400`; replace `bg-indigo-600` submit button with `bg-amber-500 text-stone-900 font-semibold rounded-xl`; replace all `text-gray-*` labels/headings with stone equivalents (depends on T019a passing)
- [x] T019b Update loading text in `frontend/src/pages/Home.js`, `frontend/src/pages/RecipeList.js`, and `frontend/src/pages/RecipeDetail.js`: replace plain `textContent = 'Loading…'` assignments with a styled `<p>` element using `text-stone-500 animate-pulse py-4 text-center`
- [x] T020 [P] Run full Vitest suite in `frontend/` and confirm all tests pass (`npm test` from `frontend/`)
- [ ] T021 [P] Manually verify all pages and responsive breakpoints per `specs/006-visual-redesign/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: N/A — skipped
- **US1 (Phase 3)**: Depends on T001 (body base style)
- **US2 (Phase 4)**: Depends on T001; independent of US1 (different files)
- **US3 (Phase 5)**: Depends on T001; independent of US1 and US2 (different files)
- **US4 (Phase 6)**: Depends on T001; independent of all other stories
- **Polish (Phase 7)**: Depends on all desired stories being complete

### User Story Dependencies

- **US1 (P1)**: Depends on T001 only — `main.js` and `Home.js`, no story dependencies
- **US2 (P2)**: Depends on T001 only — `RecipeCard.js`, `RecipeList.js`, `EmptyState.js`, `SearchBar.js`
- **US3 (P3)**: Depends on T001 only — `RecipeDetail.js`, `IngredientList.js`
- **US4 (P4)**: Depends on T001 only — `Login.js`

All user stories can be implemented in parallel after T001, since they touch different files.

### Within Each User Story

1. Write tests → confirm they **fail** (assert new class tokens not yet present)
2. Implement changes → confirm tests **pass**
3. Run full suite → confirm no regressions

### Parallel Opportunities

Within US2, US3: component files are independent, all [P]-marked tasks in those phases can be done simultaneously.

T020 and T021 (Phase 7) can run in parallel — automated vs. manual.

---

## Parallel Example: User Story 2

```bash
# All can run simultaneously after T006–T008 tests are written:
Task T009: Update RecipeCard.js
Task T010: Update RecipeList.js
Task T011: Update EmptyState.js
Task T012: Update SearchBar.js
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. T001 (body bg) → T002–T003 (failing tests) → T004–T005 (nav + home)
2. **STOP and VALIDATE**: Dark nav and home hero visible in browser; T002–T003 pass
3. Demo before proceeding to recipe list

### Incremental Delivery

1. T001 → US1 (nav + home) → validate
2. US2 (recipe list + cards) → validate
3. US3 (recipe detail) → validate
4. US4 (login) → validate
5. T019–T021 (polish + verify) → ship

---

## Notes

- [P] tasks = different files, no shared dependencies
- All gray-* class replacements: `gray-900→stone-900`, `gray-800→stone-800`, `gray-700→stone-700`, `gray-600→stone-600`, `gray-500→stone-500`, `gray-400→stone-400`, `gray-300→stone-300`, `gray-200→stone-200`, `gray-100→stone-100`
- All indigo-* class replacements: `indigo-600→amber-500`, `indigo-700→amber-600`, `indigo-500→amber-400`
- WCAG note: `bg-amber-500 text-stone-900` (not `text-white`) on primary buttons — contrast ≈ 9:1, WCAG AA compliant
- Section labels use `text-amber-700` (not `text-amber-500`) for sufficient contrast on white/stone-50 backgrounds
- Total tasks: 21 (1 setup + 2 tests US1 + 2 impl US1 + 3 tests US2 + 4 impl US2 + 2 tests US3 + 2 impl US3 + 1 test US4 + 1 impl US4 + 3 polish)
