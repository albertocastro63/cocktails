# Research: Ingredient Hover Popover

## Language & Framework

**Decision**: Vanilla JavaScript (ES Modules) with Tailwind CSS utility classes.  
**Rationale**: The entire frontend codebase is plain JS with no framework. New components must follow the same pattern — a function that returns a DOM element.  
**Alternatives considered**: None — framework choice is already locked in.

## Popover Approach

**Decision**: Render the popover as an absolutely-positioned child element of the `RecipeCard` div, shown/hidden by toggling a CSS class on `mouseenter`/`mouseleave` events on the card element.  
**Rationale**: Keeping the popover as a child of its card gives natural hover scoping with no external event listeners. The card already uses `relative`-compatible layout. No third-party library is needed for a tooltip/popover at this scale.  
**Alternatives considered**:
- Appending popover to `document.body` — avoids stacking/overflow issues but requires manual positioning calculation and teardown.
- CSS-only `:hover` + sibling selector — not possible here since the popover needs data binding from JS.

## Ingredient Truncation

**Decision**: Show the first 5 ingredient names (index 0–4). If `ingredients.length > 5`, append a sixth item rendered as the ellipsis character "…".  
**Rationale**: Directly maps to the spec requirement. The ellipsis is a static text node, not interactive.  
**Alternatives considered**: Showing ingredient count instead of ellipsis — rejected because the spec explicitly calls for an ellipsis as the sixth line item.

## Existing `IngredientList` Component

**Decision**: Do **not** reuse `IngredientList.js` for the popover content.  
**Rationale**: `IngredientList` renders quantities and units with a full divider-based layout suited for the detail page. The popover is a lightweight name-only list. Reusing it would introduce unwanted visual complexity inside the popover and violate single-responsibility by forcing `IngredientList` to conditionally render two different layouts.  
**Alternatives considered**: Extending `IngredientList` with a `compact` prop — rejected as unnecessary complexity.

## Component Location

**Decision**: Popover rendering logic lives inside `RecipeCard.js`. No separate file.  
**Rationale**: The popover is not independently reused anywhere; it is tightly coupled to the card. A private helper function within the same file avoids over-engineering a one-use component.  
**Alternatives considered**: New `IngredientPopover.js` component — rejected; overkill for a private implementation detail.

## Testing

**Decision**: Vitest with jsdom (already used for `RecipeCard.test.js`). Tests simulate `mouseenter`/`mouseleave` events with `dispatchEvent`.  
**Rationale**: Consistent with existing test approach. jsdom supports `dispatchEvent` and DOM traversal needed to assert popover visibility.  
**Alternatives considered**: Playwright/end-to-end tests — out of scope for a unit-level component test.

## Accessibility

**Decision**: No ARIA role added to the popover (informational tooltip only, not interactive).  
**Rationale**: The popover surfaces data that is also accessible by navigating to the recipe detail page; it is a convenience enhancement. Adding `role="tooltip"` with a full `aria-describedby` binding is a follow-up if accessibility audits flag it.  
**Alternatives considered**: `role="tooltip"` with `aria-describedby` on the card — deferred, out of scope for this feature.
