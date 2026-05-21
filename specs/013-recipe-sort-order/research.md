# Research: Recipe Sort Order

**Feature**: 013-recipe-sort-order
**Date**: 2026-05-21

---

## ARIA Pattern for Toggle Button Group

**Decision**: Use `aria-pressed` on each individual button within a `role="group"` container.

**Rationale**: The sort control has exactly two always-visible, mutually-exclusive options (A→Z and Z→A). `aria-pressed` on a toggle button communicates "selected / not selected" cleanly to screen readers without requiring arrow-key navigation (which `role="radiogroup"` would demand). The pattern is:
```html
<div role="group" aria-label="Sort recipes">
  <button aria-pressed="false">A→Z</button>
  <button aria-pressed="true">Z→A</button>
</div>
```
When `aria-pressed="true"`, screen readers announce the button as "pressed" or "selected" — fulfilling FR-006 without additional live regions.

**Alternatives considered**:
- `role="radiogroup"` + `role="radio"`: More semantically correct for mutually-exclusive choices, but mandates arrow-key navigation between options (Tab moves away from the group). Worse ergonomics for a two-option sort control.
- `<select>` dropdown: Hides options behind a click; adds unnecessary interaction cost for two visible choices. Rejected per clarification Q1.
- Single cycling button: Hides the inactive option entirely; user cannot see the other direction at a glance. Rejected per clarification Q1.

---

## Sort Algorithm

**Decision**: Client-side sort using `Array.prototype.sort` with `String.prototype.localeCompare` (case-insensitive, `sensitivity: 'base'`).

**Rationale**: All recipes are already loaded into memory when the page renders (the API returns the full list). No additional network request is needed. `localeCompare` with `sensitivity: 'base'` handles Unicode, accented characters, and case-insensitivity correctly and is universally supported.

```js
recipes.slice().sort((a, b) =>
  a.name.localeCompare(b.name, undefined, { sensitivity: 'base' })
);
```

Using `.slice()` creates a shallow copy so the original server-returned array is preserved for the default (unsorted) state.

**Alternatives considered**:
- Server-side sort (query param `?sort=name_asc`): Requires an extra API round-trip per sort action; adds backend complexity for a UI-only preference. Rejected.
- `toLowerCase()` comparison: Less robust than `localeCompare` for non-ASCII characters. Rejected.

---

## Component Architecture

**Decision**: Extract a dedicated `SortButtonGroup` component in `frontend/src/components/SortButtonGroup.js`.

**Rationale**: The existing codebase follows a clear pattern — `SearchBar`, `RecipeCard`, `EmptyState` are all standalone component files with co-located test files. A `SortButtonGroup` component follows this pattern, keeps `RecipeList.js` focused on page orchestration, and allows the component to be unit-tested in isolation. The component accepts `{ currentDir, onSort }` props (same pattern as `SearchBar`'s `{ value, onSearch }`).

**Alternatives considered**:
- Inline in `RecipeList.js`: Simpler but mixes UI rendering with page logic; harder to test the button group in isolation. Rejected.

---

## Visual Design

**Decision**: Use the amber/stone Tailwind design system already established in feature 006 for button states.

Active button (currently selected sort direction):
- `bg-amber-100 text-amber-800 border-amber-300 font-semibold`

Inactive button (other direction):
- `bg-white text-stone-600 border-stone-200 hover:bg-stone-50`

Group wrapper:
- `inline-flex rounded-lg border border-stone-200 overflow-hidden`
- No gap between buttons (joined appearance with shared border)

Focus ring:
- `focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-amber-500`
  (matches WCAG 2.1 AA 3:1 contrast for UI components)

**Alternatives considered**:
- Pill/chip style with full rounded corners per button: Less clearly grouped; may look disconnected from the search bar above. Rejected in favor of joined button group.

---

## Test Strategy

**Decision**: Two test files — `SortButtonGroup.test.js` (unit) and new sort tests appended to `RecipeList.test.js`.

**SortButtonGroup.test.js** covers:
- Renders both buttons with correct labels
- Marks the active button with `aria-pressed="true"`, inactive with `aria-pressed="false"`
- Neither button is pressed when `currentDir` is `null`
- Calls `onSort('asc')` when A→Z is clicked
- Calls `onSort('desc')` when Z→A is clicked
- Clicking the already-active button calls `onSort` with the same value (idempotent — list re-render is a no-op)
- Active button has the amber highlight class; inactive has the stone class

**RecipeList.test.js additions** cover:
- Renders the SortButtonGroup on page load
- Clicking A→Z reorders the displayed recipe cards alphabetically
- Clicking Z→A reorders in reverse
- Switching directions updates the display
- Default state (no sort active) reflects server order

Constitution II (Test-First) is satisfied: test files are written as Task T001/T002 before any implementation tasks.
