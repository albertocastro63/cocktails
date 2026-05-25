# Tasks: Site Footer and Ingredient Popup Layout Fix

**Input**: Design documents from `specs/017-footer-popup-fix/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, quickstart.md ✓

**Deliverables**: `frontend/src/components/Footer.js`, `frontend/src/components/Footer.test.js` (new); `frontend/src/components/RecipeCard.js`, `frontend/src/main.js` (modified).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no shared edit conflicts)
- **[Story]**: Which user story this task maps to
- Constitution requires Test-First — failing tests are written before each implementation task

---

## Phase 1: Setup

**Purpose**: Confirm project structure is ready — no new directories needed.

- [X] T001 Verify `frontend/src/components/` directory exists and `frontend/src/main.js` is accessible (read both to confirm starting state before any changes)

---

## Phase 2: Foundational

**N/A** — No shared schema, middleware, or dependency between the two user stories. Both US1 and US2 are independently implementable after Phase 1.

**Checkpoint**: Phase 1 complete → proceed to user story phases.

---

## Phase 3: User Story 1 — Site-Wide Footer (Priority: P1) 🎯 MVP

**Goal**: Every page in the application displays a consistent footer at the bottom of its content — a horizontal separator line constrained to the content area width followed by a copyright notice with the current year.

**Independent Test**: Navigate to any page (`#/`, `#/recipes`, `#/login`) and confirm a footer with a separator line and `© [year] Cocktails` text is visible at the bottom. The separator must not exceed the content width on wide viewports.

### Tests for User Story 1

> **Write these tests FIRST and confirm they FAIL before implementing Footer.js**

- [X] T002 [US1] Write failing tests for the Footer component in `frontend/src/components/Footer.test.js`:
  - Test: `Footer()` returns a `<footer>` DOM element
  - Test: footer contains an `<hr>` (or element with role separator) and a text node containing `© [currentYear] Cocktails`
  - Test: the copyright year matches `new Date().getFullYear()`
  - Test: footer inner wrapper has class `max-w-4xl` (content-area constraint)

### Implementation for User Story 1

- [X] T003 [US1] Implement `frontend/src/components/Footer.js`: export a `Footer()` function that returns a `<footer>` element containing a wrapper `div` with classes `max-w-4xl mx-auto px-4`, inside which render an `<hr class="border-stone-300 mt-4 mb-0">` (no bottom margin to avoid extra whitespace at page bottom) followed by a `<p class="text-stone-500 text-sm text-center py-4">` with text `© ${new Date().getFullYear()} Cocktails`

- [X] T004 [US1] Wire footer into `frontend/src/main.js` in the `renderPage()` function: import `Footer` from `./components/Footer.js` and append `root.appendChild(Footer())` as the last statement before every `return` inside `renderPage()` — this includes: (1) the admin-denied early return (renders "Access denied" paragraph), (2) the write-route auth guard early return (renders Login), (3) the matched-route branch (`root.appendChild(factory(m))`), and (4) the 404 fallback. Every code path that renders page content must end with `root.appendChild(Footer())`

**Checkpoint**: Navigate to `#/`, `#/recipes`, `#/login`, and `#/recipes/new` — footer visible on all pages. Separator constrained to content width. Copyright year is current.

---

## Phase 4: User Story 2 — Ingredient Popup as Non-Expanding Overlay (Priority: P1)

**Goal**: The ingredient hover popup on recipe list cards appears as a true overlay — appended to `document.body` with JS-calculated coordinates — so the page total scroll height does not change and no surrounding content shifts when the popup appears.

**Independent Test**: On `#/recipes`, open DevTools console, record `document.documentElement.scrollHeight`, hover a recipe card, record scroll height again — both values must be identical.

### Tests for User Story 2

> **Write these tests FIRST and confirm they FAIL before modifying RecipeCard.js**

- [X] T005 [US2] Add failing popup-overlay tests to `frontend/src/components/RecipeCard.test.js`:
  - Test: on `mouseenter`, the popover is appended to `document.body` (not to the card element)
  - Test: on `mouseenter`, the popover is NOT a descendant of the card element
  - Test: on `mouseleave`, the popover is removed from `document.body`
  - Test: only one popover exists in `document.body` at a time when two cards are hovered in sequence
  - Test: popover has `position: absolute` style set
  - Test: clicking `document.body` (click-elsewhere) removes the popover (FR-007 click-outside closure); simulate by dispatching a `click` event on `document.body` while popover is open and asserting `document.body.querySelector('[data-popover]')` is null
  - Note: jsdom does not define `window.scrollX`/`window.scrollY` — mock them as `0` in `beforeEach` via `Object.defineProperty(window, 'scrollX', { value: 0, writable: true })` and similarly for `scrollY` to avoid `NaN` coordinates in tests

### Implementation for User Story 2

- [X] T006 [US2] Modify `frontend/src/components/RecipeCard.js`: refactor the popup logic so that:
  1. `buildIngredientPopover` remains unchanged (still returns a div with the ingredient list)
  2. The `mouseenter` handler replaces `el.appendChild(...)` with: `const rect = el.getBoundingClientRect(); const popup = buildIngredientPopover(ingredients); popup.style.position = 'absolute'; popup.style.top = \`${rect.bottom + (window.scrollY ?? 0) + 4}px\`; popup.style.left = \`${rect.left + (window.scrollX ?? 0)}px\`; popup.style.width = \`${rect.width}px\`; document.body.appendChild(popup);`
  3. The `mouseleave` handler replaces `el.querySelector('[data-popover]')?.remove()` with: `document.body.querySelector('[data-popover]')?.remove()`
  4. Add a `click` listener on `document` (added in `mouseenter`, removed in `mouseleave`) that calls `document.body.querySelector('[data-popover]')?.remove()` — this implements FR-007 click-elsewhere closure. Use a named function reference so the listener can be removed cleanly: `const onDocClick = () => document.body.querySelector('[data-popover]')?.remove(); document.addEventListener('click', onDocClick, { once: true })`

**Checkpoint**: On `#/recipes`, hover a recipe card — popup appears at card location without layout shift. Check DevTools console confirms scroll height unchanged. Move cursor away — popup disappears cleanly.

---

## Phase 5: Polish & Verification

**Purpose**: Validate the implementation end-to-end against quickstart.md acceptance tests.

- [X] T007 Run `cd frontend && npm test -- --coverage` — confirm all tests pass and coverage remains ≥ 75%

- [X] T008 [P] Execute quickstart.md Acceptance Test 1 (footer): navigate to `#/`, `#/recipes`, `#/login` and confirm footer visible with separator line and copyright notice constrained to content width on both narrow (375px) and wide (2000px) viewports

- [X] T009 [P] Execute quickstart.md Acceptance Test 2 (popup overlay): on `#/recipes`, verify `document.documentElement.scrollHeight` is identical before and after hovering a card; verify no card shifts position; verify popup closes cleanly on mouseleave

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: N/A
- **Phase 3 (US1 — Footer)**: Depends on Phase 1
- **Phase 4 (US2 — Popup)**: Depends on Phase 1; independent of Phase 3
- **Phase 5 (Polish)**: Depends on Phases 3 and 4

### User Story Dependencies

- **US1 (Phase 3)**: Start after Phase 1 — no dependency on US2
- **US2 (Phase 4)**: Start after Phase 1 — no dependency on US1

### Within Phase 3 (US1)

- T002 (failing tests) MUST complete and be confirmed failing before T003 (implementation)
- T003 MUST complete before T004 (Footer must exist before it can be imported in main.js)

### Within Phase 4 (US2)

- T005 (failing tests) MUST complete and be confirmed failing before T006 (implementation)

### Parallel Opportunities

- Phases 3 and 4 are fully independent — they edit different files and can be executed simultaneously
- T008 and T009 in Phase 5 are independent verification steps and can run in parallel

---

## Implementation Strategy

### MVP (Both User Stories — both are P1)

1. Complete Phase 1: verify structure
2. Complete Phase 3: Footer (T002 → T003 → T004)
3. Complete Phase 4: Popup fix (T005 → T006)
4. Complete Phase 5: Polish and verification (T007 → T008 + T009)

### Parallel Execution (if two agents)

- Agent A: Phase 3 (Footer) — T002 → T003 → T004
- Agent B: Phase 4 (Popup) — T005 → T006
- Both must complete before Phase 5

---

## Notes

- Both user stories are P1 — both are required for MVP; neither can be skipped
- The popup fix uses `document.body.appendChild()` — tests will need to query `document.body`, not the card element
- The Footer component year is evaluated at render time via `new Date().getFullYear()` — no mocking needed in tests unless testing a specific year
- T004 requires appending `Footer()` before every `return` in `renderPage()` — four locations: admin-denied render, write-route auth guard, matched-route branch, and 404 fallback
- Popover cleanup: `document.body.querySelector('[data-popover]')?.remove()` — the `data-popover` attribute is already set in the existing `buildIngredientPopover` function
