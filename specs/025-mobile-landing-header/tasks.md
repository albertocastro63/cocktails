# Tasks: Compact Landing Header on Mobile

**Input**: Design documents from `specs/025-mobile-landing-header/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/header-ui-contract.md ✓, quickstart.md ✓

**Tests**: TDD (Constitution §II). The test is a class-contract assertion on the hero element (the change is presentational; the testable contract is the responsive class set + unchanged text), written first. Vitest + jsdom.

**Stack**: Vanilla JS + Tailwind CSS, Vitest. No new dependency. Single-file change.

---

## Phase 1: Setup

- [X] T001 Confirm the current Home hero classes in `frontend/src/pages/Home.js` (hero `py-16 px-4`, title `text-4xl mb-3`, subtitle `text-lg mb-6`) and that `frontend/src/pages/Home.test.js` exists — this is the baseline the `md:` overrides must preserve (SC-003).

---

## Phase 2: Foundational

_None — no shared prerequisites; the single user story owns the whole change._

---

## Phase 3: User Story 1 — Compact landing header on a phone (Priority: P1) 🎯 MVP

**Goal**: On phones (< 768px) the landing hero is noticeably shorter (smaller padding + title/subtitle), while at ≥ 768px it is pixel-for-pixel unchanged.

**Independent Test**: On a phone-sized viewport the hero banner is clearly shorter and the "All Recipes" CTA is reachable with little/no scrolling; on desktop the header is identical to today.

- [X] T002 [US1] Write failing tests in `frontend/src/pages/Home.test.js` asserting the header class contract (H1–H7): the hero wrapper includes `py-8` and `md:py-16`; the title `h1` includes `text-2xl` and `md:text-4xl`; the subtitle `p` includes `text-base` and `md:text-lg`; the title text is "Cocktail Recipes"; the subtitle text is "Discover your next favorite drink"; the CTA keeps `href="#/recipes"` / "All Recipes"; the stone/amber text tokens are retained.
- [X] T003 [US1] Update the hero classes in `frontend/src/pages/Home.js`: hero wrapper `py-16` → `py-8 md:py-16`; title `text-4xl mb-3` → `text-2xl md:text-4xl mb-1 md:mb-3`; subtitle `text-lg mb-6` → `text-base md:text-lg mb-4 md:mb-6` — making T002 pass. Do not change any text content or the CTA.
- [X] T004 [US1] Run `cd frontend && npm test -- src/pages/Home.test.js` (green) and manually verify at < 768px (hero clearly shorter, CTA reachable with little/no scroll, no horizontal overflow) and at ≥ 768px (header identical to today, ideally side-by-side with production).

**Checkpoint**: Landing header is compact on phones and unchanged on desktop — feature delivered.

---

## Phase 4: Polish & Cross-Cutting Concerns

- [X] T005 [P] Verify legibility/accessibility of the compact header on phones (title/subtitle still readable, contrast unchanged since color tokens are unchanged) including a narrow/landscape phone (no clip/awkward wrap) — WCAG 2.1 AA (§III).
- [X] T006 [P] Run `cd frontend && npm test -- --coverage`; confirm ≥ 75% maintained and no regression in the existing `Home.test.js` / other suites.
- [ ] T007 Run the `quickstart.md` manual matrix (V1–V4) and record results.

---

## Dependencies & Execution Order

- **Phase 1 (Setup)** → **Phase 3 (US1)** → **Phase 4 (Polish)**. (No Foundational phase.)
- **T002 (test) before T003 (impl)** per §II.
- T004 depends on T003; Polish depends on US1.

### Parallel opportunities

- **T005 + T006**: accessibility check and coverage run are independent.

---

## Implementation Strategy

Single user story = the whole feature. Write the class-contract test (T002), apply the responsive classes (T003), verify on both phone and desktop viewports (T004), then polish (a11y, coverage, manual matrix). Desktop parity is guaranteed by the `md:` overrides equaling today's values.
