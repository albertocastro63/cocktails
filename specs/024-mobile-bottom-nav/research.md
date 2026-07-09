# Research: Mobile Bottom Navigation

**Branch**: `024-mobile-bottom-nav` | **Date**: 2026-07-08

---

## Decision 1 — CSS-driven responsive switch (render both, toggle with Tailwind `md:`)

**Decision**: Render the top nav and the bottom bar in the DOM together and use Tailwind visibility utilities (`hidden md:flex` for top, `flex md:hidden` for the bottom bar and mobile brand header). The 768px boundary is a media query.

**Rationale**: Satisfies FR-011 (switch on rotate/resize with no reload, no lost page) with zero JavaScript — the browser re-evaluates the media query instantly. Avoids a `resize` listener and its reflow/debounce complexity. Keeps the desktop path untouched (SC-004): the existing top-nav markup is preserved and merely gains a wrapper class.

**Alternatives considered**:
- *JS `matchMedia`/resize listener that swaps which nav is mounted* — rejected; more code, a state to manage, and risk of fl::desktop divergence. CSS is simpler and more robust.
- *Separate mobile route/layout* — rejected; the app is a single hash-routed SPA with one shell.

---

## Decision 2 — Single source of truth for destinations; preserve desktop rendering

**Decision**: A `navDestinations(state)` function returns the ordered destination list per auth state; both bars consume it for **which** links to show. The bottom bar renders icon+label from it; the top bar keeps its current item-specific styling (amber "New Recipe" button, text "Sign Out", `ml-auto`).

**Rationale**: The real duplication risk (Constitution §I) is the *visibility logic* — "who sees which links" drifting between two bars. Centralizing that removes the risk. Rendering differs by necessity (SC-004 demands the desktop bar look identical), so the top bar keeps its bespoke markup rather than being regenerated from a generic list.

**Alternatives considered**:
- *Fully generic renderer for both bars* — rejected; it would either change the desktop look (violating SC-004) or need so many per-item style hints that it is more complex than two renderers.

---

## Decision 3 — Bounded 5-slot bar with a "More" overflow (FR-006/007)

**Decision**: `partitionNav(list, maxSlots=5)`: ≤ 5 destinations → all direct; > 5 → first 4 (by priority) direct + a 5th "More" slot opening a menu with the remainder. Items never shrink below the min tap size; growth always lands in More.

**Rationale**: Matches the standard mobile tab-bar convention and the spec's chosen overflow strategy (horizontal scroll was rejected in the spec for hiding items without a cue). A fixed slot count guarantees no overflow/clipping regardless of role or future growth (SC-005).

**Alternatives considered**:
- *Horizontally scrolling bar* — rejected in spec (hidden items, no affordance).
- *Shrinking items to fit* — rejected; violates the ≥44px tap-target rule (FR-008/SC-003).

---

## Decision 4 — Accessible overflow disclosure (Constitution §III / WCAG 2.1 AA)

**Decision**: "More" is a `<button>` with `aria-expanded` and `aria-controls` pointing at the menu container; the menu is a list of links/buttons. Open moves focus to the first item; `Escape` closes and returns focus to the button; outside pointer-down closes; selecting an item closes (and navigation re-renders the nav, which also tears it down). The bar is a `<nav aria-label="Primary">`.

**Rationale**: The spec covers tap size (FR-008) and safe areas (FR-014) but not the disclosure semantics; §III mandates WCAG AA, and a menu is the one interactive widget here that needs keyboard + focus + ARIA to be conformant.

**Alternatives considered**:
- *CSS-only `:focus-within` popover* — rejected; no `Escape`/outside-close and weaker screen-reader semantics.

---

## Decision 5 — Safe areas via `env()` + `viewport-fit=cover`

**Decision**: Add `viewport-fit=cover` to the `index.html` viewport meta; give the bar `padding-bottom: env(safe-area-inset-bottom)` and pad page content by `calc(barHeight + env(safe-area-inset-bottom))` on mobile only.

**Rationale**: Required for FR-014 (home-indicator area) and FR-004 (content not obscured). `env()` insets only resolve when `viewport-fit=cover` is set.

**Alternatives considered**:
- *Fixed extra padding guess* — rejected; wrong on devices without a home indicator (wasted space) and insufficient on those with larger insets.

---

## Decision 6 — Icons as inline SVG (no dependency)

**Decision**: Hand-author a small inline SVG per destination (recipes, my-recipes, plus/new, users, admin-recipes, more/ellipsis, sign-out, sign-in), following the existing pattern in `FavoriteButton.js`/`RecipeCard.js`.

**Rationale**: The clarification chose icon+label; the codebase already ships hand-coded SVGs, so no icon library is added (keeps the bundle small, TTI unaffected).

**Alternatives considered**:
- *Add an icon library (lucide/heroicons)* — rejected; unnecessary weight for ~8 glyphs.

---

## Decision 7 — On-screen keyboard: hide the bar while editing

**Decision**: Toggle a "keyboard-open" class on `focusin`/`focusout` of `input`/`textarea` within the app root that hides the fixed bottom bar; restore on blur.

**Rationale**: Addresses the spec edge case (bar must not cover form inputs). Simpler and more reliable across mobile browsers than trying to reposition around the virtual keyboard.

**Alternatives considered**:
- *VisualViewport API repositioning* — rejected as over-engineered for the goal; hiding the bar during text entry is sufficient and predictable.

---

## Resolved unknowns

| Unknown | Resolution |
|---------|-----------|
| Breakpoint | < 768px (Tailwind `md`) — from clarification |
| Item presentation | Icon + label — from clarification |
| Switch mechanism | CSS media query (no reload) |
| Overflow strategy | Bounded 5 slots + "More" menu |
| Desktop parity | Preserve existing top-nav markup, hide below `md` |
