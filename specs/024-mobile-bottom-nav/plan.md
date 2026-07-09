# Implementation Plan: Mobile Bottom Navigation

**Branch**: `024-mobile-bottom-nav` | **Date**: 2026-07-08 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/024-mobile-bottom-nav/spec.md`

## Summary

On viewports below 768px, replace the wrapping top nav with a fixed bottom tab bar (icon + label) that shows a bounded set of direct items plus a "More" overflow for the rest; at 768px+ the existing top nav is unchanged. A single source of truth for navigation destinations (per auth state/role) feeds both bars, so which links appear never drifts. The bottom/top switch is driven by CSS (Tailwind `md:`), so rotation/resize needs no JS and no reload.

## Technical Context

**Language/Version**: JavaScript (ES modules), Node 24 / Vite build  
**Primary Dependencies**: Vanilla DOM (no framework), Tailwind CSS (utility classes), Vitest + jsdom (tests). No new runtime dependency — icons are inline SVGs (as `FavoriteButton`/`RecipeCard` already do).  
**Storage**: N/A (client-side UI only)  
**Testing**: Vitest (jsdom) — pure-function unit tests + DOM render/interaction tests  
**Target Platform**: Web browsers; mobile (< 768px) and desktop/tablet (≥ 768px)  
**Project Type**: Frontend web app (`frontend/`)  
**Performance Goals**: TTI ≤ 3 s unaffected; the top/bottom switch is a CSS media query (no JS resize listener, no reflow storms)  
**Constraints**: Desktop/tablet top nav MUST render identically to today (SC-004); WCAG 2.1 AA for the bar and the overflow menu; iOS safe-area support; integrate with the existing hash router in `main.js`  
**Scale/Scope**: ≤ ~7 destinations today, bounded to 5 bottom-bar slots; ~2 new modules + 2 edits

### Key facts (from the current code)

- `buildNav()` in `frontend/src/main.js` builds the nav imperatively and is called by `renderPage()` on **every route change**, so active-state and auth-state are recomputed each render for free (covers the "sign in/out without reload updates nav" edge case).
- Routing is **hash-based** (`location.hash`), so active-state matching compares against `getPath()`.
- Destinations today: visitor → All Recipes, Sign In; regular → All Recipes, My Recipes, New Recipe, Sign Out; admin → + Users, + admin Recipes.
- No responsive breakpoints exist in the nav today; Tailwind is available (`md` = 768px).

## Constitution Check

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Single-responsibility, ≤40 lines, CC ≤ 10, no duplication? | ✅ Split into `navDestinations()` (source of truth), `partitionNav()` (slot/overflow math), `buildBottomNav()` (render), and a small overflow-menu controller. The shared destination list removes "which links" duplication between top and bottom bars |
| II. Test-First | Failing tests written before implementation? | ✅ This feature **has** unit-testable logic — `navDestinations` (per-role sets), `partitionNav` (bounded direct vs overflow), active-state matcher, and DOM tests for `BottomNav`/overflow. Tests written first; coverage ≥ 75% |
| III. UX Consistency | Design tokens + loading/empty/error states + WCAG 2.1 AA? | ✅ Reuses stone/amber tokens; overflow menu is an accessible disclosure (`aria-expanded`/`aria-controls`, keyboard open/close, Escape, focus management); tap targets ≥ 44px; active indication; safe-area insets |
| IV. Performance | p95 / TTI budgets? | ✅ No API changes; CSS-driven responsive switch; no reload on rotate/resize (FR-011) |
| Quality Gates | Lint, coverage ≥ 75%, tests pass? | ✅ Frontend Vitest suite extended; no coverage regression |

No violations — §II applies fully (unlike infra-only features), which is a strength.

## Project Structure

### Documentation (this feature)

```text
specs/024-mobile-bottom-nav/
├── plan.md              # this file
├── research.md          # Phase 0 — decisions
├── data-model.md        # Phase 1 — destination model + priority/partition rules
├── quickstart.md        # Phase 1 — manual + test verification
├── contracts/
│   └── nav-ui-contract.md   # UI/behavior contract (states → rendered output)
└── tasks.md             # /speckit-tasks output (not created here)
```

### Source Code (repository root)

```text
frontend/
├── index.html                          # EDIT — add viewport-fit=cover for safe-area insets
├── src/
│   ├── nav/
│   │   ├── destinations.js             # NEW — navDestinations(state) source of truth,
│   │   │                               #        partitionNav(list, maxSlots), inline SVG icons
│   │   └── destinations.test.js        # NEW — unit tests (roles, ordering, partition/overflow)
│   ├── components/
│   │   ├── BottomNav.js                # NEW — renders bottom bar (icon+label, direct + More,
│   │   │                               #        active state, accessible overflow menu)
│   │   └── BottomNav.test.js           # NEW — DOM tests (render, overflow, a11y, close behavior)
│   └── main.js                         # EDIT — buildNav() consumes navDestinations(); wraps the
│                                       #        existing top nav in `hidden md:flex`, adds a slim
│                                       #        mobile brand header + bottom nav (`md:hidden`)
```

**Structure Decision**: Frontend-only change. To guarantee the desktop nav stays pixel-identical (SC-004), the existing top-bar rendering is preserved and simply hidden below `md`; the bottom bar is a separate component. The **destination set/visibility logic** is centralized in `navDestinations()` so both bars agree on which links a user gets — eliminating the real duplication risk without reworking the desktop markup.

## Architecture

### Responsive switch (no JS, no reload — FR-011)

Render both bars; let CSS decide which shows:

- Top nav wrapper: `hidden md:flex` (hidden < 768px, shown ≥ 768px).
- Mobile brand header (slim): `flex md:hidden`; the brand text is an anchor to `#/` so Home stays reachable on mobile (Home is not a bottom-bar item).
- Bottom bar: `fixed bottom-0 inset-x-0 flex md:hidden`.

Crossing 768px (rotate/resize) is a pure media-query change — no reload, current page/hash preserved.

### Destination model & partitioning (FR-005/006/007)

`navDestinations(state)` → ordered `[{ id, label, icon, href|action, priority, match }]` for `state ∈ {visitor, user, admin}`. `partitionNav(list, maxSlots=5)`:
- if `list.length ≤ maxSlots` → all direct, no More.
- else → first `maxSlots-1` by priority as direct + a **More** slot whose menu holds the remainder.

Priority table (direct-slot preference) lives in `data-model.md`. Result today: visitor (2) and regular (4) are all-direct; admin (6) → 4 direct + More(2). Sign Out is always reachable (direct when it fits, else in More — FR-012).

### Content not obscured & safe areas (FR-004/014)

- App content container gets bottom padding on mobile: `pb-[calc(4rem+env(safe-area-inset-bottom))] md:pb-0` so the end of the page scrolls clear of the bar.
- Bar has `padding-bottom: env(safe-area-inset-bottom)`; `index.html` viewport meta gains `viewport-fit=cover` so `env()` resolves.

### Accessible overflow menu (FR-013, §III)

`More` is a `<button aria-expanded aria-controls="nav-more">`; the menu is a popover above the bar. Closes on: select, outside pointer-down, `Escape`, or route change (the nav re-renders on navigation, which tears the menu down). Focus moves to the first menu item on open and returns to the More button on Escape.

### On-screen keyboard edge case

When a text input/textarea gains focus on mobile, hide the fixed bar (`focusin`/`focusout` on the app root toggling a class) so it never covers the field being edited; restore on blur.

### Active state (FR-009)

At render time compare `getPath()` to each destination's `match` (exact or prefix, e.g. `/recipes/new`) and apply the active token (amber icon+label) to the matching direct item, or to the More button when the active route lives in the overflow set.

## Phase 0: Research

See [research.md](research.md) — responsive strategy, safe-area handling, overflow-menu a11y, keyboard mitigation, icon approach. No `NEEDS CLARIFICATION` remain (spec clarifications resolved item style + breakpoint).

## Phase 1: Design

- [data-model.md](data-model.md) — destination entity, per-role sets, priority table, partition rules, active-match rules.
- [contracts/nav-ui-contract.md](contracts/nav-ui-contract.md) — state → rendered-output/behavior contract (the acceptance assertions).
- [quickstart.md](quickstart.md) — Vitest + manual device verification.

## Complexity Tracking

No constitution violations; no entries required.
