# Research: Site Footer and Ingredient Popup Layout Fix

**Feature**: 017-footer-popup-fix | **Date**: 2026-05-25

## Finding 1 — Why the current popup expands page scroll height

**Decision**: Portal the popover to `document.body` instead of appending it to the card element.

**Rationale**: The current popup (`absolute top-full left-0`) is appended as a child of the card div. Even though it uses `position: absolute`, browsers include absolutely positioned descendants in the scroll height calculation of their nearest scrollable ancestor. When the card is near the bottom of the page, the popover extends beyond the existing page content, increasing `document.documentElement.scrollHeight`. Portalling to `document.body` with coordinates derived from `getBoundingClientRect() + window.scrollY` places the popover within the body's existing coordinate space, and since the body's scroll height is already determined by page content, the popup does not extend it for any card except possibly the very last row (which the spec explicitly tolerates: "the user may scroll to see it fully").

**Alternatives considered**:
- `position: fixed` — rejected; spec scenario 2 requires the popup to scroll with page content, not stay fixed to the viewport.
- `overflow: clip` / `overflow: hidden` on the grid — rejected; this hides the popup instead of overlaying it.
- CSS `contain: layout` on the grid — rejected; same effect as overflow clipping.
- Keep current approach but add `z-index` manipulation — rejected; z-index does not affect scroll height.

## Finding 2 — Dynamic copyright year in vanilla JS

**Decision**: Use `new Date().getFullYear()` evaluated at component render time.

**Rationale**: The Footer component is a function that runs when `renderPage()` is called. Calling `new Date().getFullYear()` inside the function body returns the current calendar year at render time. No build-time substitution, no environment variable, no cron job needed. On January 1st the first page load of the new year automatically shows the correct year.

**Alternatives considered**:
- Hardcoded year — rejected; spec FR-009 requires no code change when year rolls over.
- Build-time injection via Vite define — rejected; unnecessarily complex and wrong anyway (would show build year, not current year).

## Finding 3 — Footer layout to match content area width

**Decision**: Wrap footer content in a `max-w-4xl mx-auto px-4` div, matching the width constraint used by all existing pages.

**Rationale**: Existing pages (`RecipeList`, `RecipeDetail`, etc.) use `max-w-4xl mx-auto px-4` on their root element. A footer with the same constraint on its inner wrapper will produce a separator line and copyright text that visually align with page content on all viewport widths—wide viewports show whitespace on the sides, narrow viewports fill the full width.

**Alternatives considered**:
- Full-width footer — rejected; spec FR-003 requires the separator NOT extend beyond content area maximum width.
- CSS `width: inherit` — rejected; unreliable across different page contexts.

## Finding 4 — Footer placement in the SPA

**Decision**: Append the Footer component in `renderPage()` after the route-matched content block, rendering it in the normal document flow.

**Rationale**: The spec states the footer appears in normal document flow at the bottom of each page's content (not sticky/fixed). `renderPage()` already calls `root.innerHTML = ''` and then appends `buildNav()` and the page component. Appending `Footer()` as the last child follows the same pattern with zero structural change to the router.

**Alternatives considered**:
- Add footer inside each page component — rejected; requires N changes (10+ page files) instead of one; violates DRY.
- CSS sticky footer — rejected; spec explicitly says not pinned to viewport bottom.

## Finding 5 — Popup position calculation

**Decision**: Use `getBoundingClientRect()` on the card element combined with `window.scrollY` to compute absolute body coordinates. Clean up by appending a `data-popup-for` attribute keyed to a card ID so `mouseleave` removes the correct element.

**Rationale**: `getBoundingClientRect()` returns coordinates relative to the viewport. Adding `window.scrollY` converts them to page coordinates relative to `document.body`. The popup is then `position: absolute` in `document.body`'s coordinate space, which means it scrolls with the page (satisfying scenario 2) and does not change the page scroll height (satisfying SC-003) for all cards except those in the very last row.

**Implementation pattern**:
```js
const rect = cardEl.getBoundingClientRect();
popup.style.position = 'absolute';
popup.style.top = `${rect.bottom + window.scrollY + 4}px`;
popup.style.left = `${rect.left + window.scrollX}px`;
popup.style.width = `${rect.width}px`;
document.body.appendChild(popup);
```

Cleanup on `mouseleave`: `document.body.querySelector('[data-popover]')?.remove()`.
