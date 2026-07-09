# Tasks: Mobile Bottom Navigation

**Input**: Design documents from `specs/024-mobile-bottom-nav/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/nav-ui-contract.md ✓, quickstart.md ✓

**Tests**: TDD required (Constitution §II — this feature has unit-testable logic). Test tasks are written first and confirmed failing before implementation. Vitest + jsdom.

**Stack**: Vanilla JS (ES modules), Tailwind CSS, Vite, Vitest. No new dependency (inline SVG icons).

---

## Phase 1: Setup

- [X] T001 Add `viewport-fit=cover` to the viewport `<meta>` in `frontend/index.html` so `env(safe-area-inset-*)` resolves (FR-014).

---

## Phase 2: Foundational — shared destination logic + responsive shell (blocks all stories)

**⚠️ Constitution §II**: T002 (tests) MUST be written and confirmed failing before T003.

- [X] T002 [P] Write failing unit tests in `frontend/src/nav/destinations.test.js`: `navDestinations('visitor'|'user'|'admin')` returns the correct ordered id sets (U1–U3); `partitionNav(list, 5)` returns all-direct when ≤5 and 4-direct+overflow when >5 including a synthesized `more` (U4–U6); active matcher handles exact + prefix (`/recipes/new`) and never matches action items (U7).
- [X] T003 Implement `frontend/src/nav/destinations.js`: `navDestinations(state)` (source of truth per data-model.md), `partitionNav(list, maxSlots=5)`, inline SVG `icons` per destination id, and an `isActive(dest, path)` helper — making T002 pass.
- [X] T004 Refactor `frontend/src/main.js` `buildNav()` responsive shell (must keep desktop markup identical — SC-004): wrap the existing top-nav element in a `hidden md:flex` container; add a slim mobile brand header (`flex md:hidden`) whose brand text is an anchor to `#/` (keeps Home reachable on mobile, since Home is not a bottom-bar destination); add mobile-only bottom padding to the app content container (`pb-[calc(4rem+env(safe-area-inset-bottom))] md:pb-0`); add a `focusin`/`focusout` handler on the app root that hides the bottom bar while an `input`/`textarea` is focused.

**Checkpoint**: `destinations.test.js` passes; desktop nav unchanged; mobile shows the brand header (bar mounted in US1).

---

## Phase 3: User Story 1 — Logged-in user navigates from a phone (Priority: P1) 🎯 MVP

**Goal**: A signed-in regular user on a phone gets a fixed bottom bar (icon+label) with All, Mine, New, Sign out; no overflow, no horizontal scroll, content clears the bar, bar hides while typing.

**Independent Test**: Sign in as a regular user at <768px — all four destinations reachable from the bottom bar, no horizontal scroll, page end scrolls clear of the bar, active item highlighted.

- [X] T005 [US1] Write failing DOM tests in `frontend/src/components/BottomNav.test.js` (jsdom): for the `user` set, the bar renders one item per destination each with an icon **and** a text label (N7), no `More` entry (N4), the item matching the current path carries the active token (B2), and every item is ≥44×44px (N6).
- [X] T006 [US1] Implement `frontend/src/components/BottomNav.js` direct-item renderer: `<nav aria-label="Primary">` fixed at `bottom-0 inset-x-0` (`md:hidden`), one icon+label button/link per `partitionNav(...).direct` item, active token (amber) via `isActive`, ≥44px tap targets, `padding-bottom: env(safe-area-inset-bottom)` — making T005 pass.
- [X] T007 [US1] Wire `frontend/src/main.js` to mount `buildBottomNav(navDestinations(state))` into the mobile bottom slot from T004, so signing in renders the bar; confirm content bottom padding clears the fixed bar (B1).
- [X] T008 [US1] Run `cd frontend && npm test -- src/components/BottomNav.test.js src/nav/destinations.test.js` (green) and manually verify at <768px as a regular user: All/Mine/New/Sign out present, no horizontal scroll, bar fixed on scroll, bar hides when the recipe form/search input is focused.

**Checkpoint**: MVP — regular users have a working bottom bar.

---

## Phase 4: User Story 2 — Admin user with a longer menu (Priority: P2)

**Goal**: An admin's 6 destinations render as 4 direct items + an accessible "More" overflow holding the rest; nothing clipped; grows safely.

**Independent Test**: Sign in as admin at <768px — bar shows All, Mine, New, Users, More; opening More reveals Manage + Sign out; menu closes on select/Escape/outside/navigate; keyboard-operable.

- [X] T009 [US2] Write failing tests in `frontend/src/components/BottomNav.test.js`: for the `admin` set the bar renders 4 direct items + a `More` button (N5); clicking `More` opens the menu with `aria-expanded="true"` and moves focus to the first item (B4); the menu closes on item-select, `Escape` (returning focus to `More`), outside pointer-down, and route change (B5); `More` carries the active token when the current path is in the overflow set (B3); overflow items are ≥44px (N6).
- [X] T010 [US2] Implement the overflow in `frontend/src/components/BottomNav.js`: render a `More` disclosure `<button aria-expanded aria-controls="nav-more">` with an ellipsis icon+label and a popover `#nav-more` listing overflow destinations; add open/close with focus management, `Escape`, outside pointer-down, and select handlers — making T009 pass.
- [X] T011 [US2] Run the BottomNav tests (green) and manually verify at <768px as admin: All/Mine/New/Users/More; More → Manage/Sign out; open/close via tap and keyboard; no clipping or overlap.

**Checkpoint**: Overflow scaling works and is accessible.

---

## Phase 5: User Story 3 — Visitor and desktop experience unchanged (Priority: P3)

**Goal**: Signed-out visitors get the reduced set in the bottom bar on phones; tablets/desktop keep the current top nav unchanged; the switch happens without reload.

**Independent Test**: Logged-out at <768px shows All + Sign in in the bottom bar; at ≥768px the top nav is identical to today with no bottom bar; resizing across 768px switches without a reload.

- [X] T012 [US3] Write failing tests in `frontend/src/components/BottomNav.test.js` (or `main.test.js`): the `visitor` set renders All + Sign in in the bottom bar (N3); assert the top-nav container has the `hidden md:flex` wrapper and the bottom bar has `md:hidden` (N1/N2 class contract, SC-004).
- [X] T013 [US3] Confirm/adjust `frontend/src/main.js` so the visitor state wires through `navDestinations('visitor')` to the bottom bar, and the existing top-nav markup is byte-for-byte preserved inside the `hidden md:flex` wrapper (no style/token changes) — making T012 pass.
- [X] T014 [US3] Manually verify: logged-out phone shows the visitor bottom bar; desktop shows the unchanged top nav and no bottom bar; resizing/rotating across 768px flips top⇄bottom with no reload and the current page preserved (B8).

**Checkpoint**: All auth states + all viewport sizes covered; desktop provably unchanged.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T015 [P] Accessibility pass on `frontend/src/components/BottomNav.js`: `nav` labelled, overflow menu roles/labels correct, visible focus styles on all items, and WCAG 2.1 AA contrast for active vs. inactive icon+label tokens (stone/amber).
- [X] T016 [P] Run `cd frontend && npm test -- --coverage`; confirm ≥75% maintained and existing nav-dependent tests still pass (no regression from the `buildNav` refactor).
- [ ] T017 Run the `quickstart.md` manual device matrix end-to-end (desktop parity, rotation across 768px, safe-area/home-indicator, on-screen keyboard, sign-in/out without reload, growth check) and record results.

---

## Dependencies & Execution Order

### Phase order

- **Phase 1 (Setup)** → **Phase 2 (Foundational)** → **Phase 3 (US1)** → **Phase 4 (US2)** → **Phase 5 (US3)** → **Phase 6 (Polish)**.
- **Foundational blocks everything**: `destinations.js` (T003) and the responsive shell (T004) are prerequisites for all three stories.
- **US2 and US3 depend on US1**: both extend/consume `BottomNav.js` (T006). They can be worked sequentially after US1; they touch the same file so are not `[P]` against each other at the implementation step.

### Critical path

```
T001 ─┐
      ├─ T002 → T003 → T004 → T005 → T006 → T007 → T008 (US1 MVP)
      │                                        ├→ T009 → T010 → T011 (US2)
      │                                        └→ T012 → T013 → T014 (US3)
                                                              └→ T015/T016 [P] → T017
```

### Parallel opportunities

- **T001 + T002**: `index.html` edit and the destinations test file are independent — write together.
- **Test-first pairs**: each story's test task precedes its implementation task (T005→T006, T009→T010, T012→T013).
- **T015 + T016**: accessibility pass and coverage run are independent.

---

## Implementation Strategy

### MVP (User Story 1 only)

1. Phase 1 (viewport meta) → Phase 2 (destinations logic + responsive shell).
2. Phase 3: BottomNav direct-item renderer wired into `main.js`.
3. **STOP and VALIDATE**: a signed-in regular user on a phone has a working, non-overflowing bottom bar — the reported defect is fixed for the primary audience.

### Full delivery

Add US2 (admin overflow "More"), then US3 (visitor set + provable desktop parity + no-reload switch), then Phase 6 (a11y, coverage, device matrix).
