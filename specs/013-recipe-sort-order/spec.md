# Feature Specification: Recipe Sort Order on All Recipes Page

**Feature Branch**: `013-recipe-sort-order`
**Created**: 2026-05-21
**Status**: Draft
**Input**: Add alphabetical (A→Z) and reverse alphabetical (Z→A) sort to the recipes list page with a clean, accessible interface.

## Clarifications

### Session 2026-05-21

- Q: What UI pattern should the sort controls use? → A: Adjacent toggle button pair — two buttons side by side (A→Z and Z→A), one highlighted as active; both always visible.
- Q: What is the default sort order on page load? → A: Server-returned order; neither sort button is pre-selected — the user must explicitly activate a sort direction.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Sort Recipes Alphabetically A→Z (Priority: P1)

A visitor browsing the all-recipes page wants to find a specific recipe by name. They activate the A→Z sort control and the recipe list immediately reorders to display recipes in ascending alphabetical order.

**Why this priority**: Alphabetical sorting is the most natural way to scan a named list. It is the foundational sort direction and delivers immediate value as a standalone increment — a user can find any recipe by name as soon as US1 is in place.

**Independent Test**: Open the all-recipes page with multiple recipes loaded. Activate the A→Z sort control. Verify the recipe list reorders with the recipe whose name comes first alphabetically at the top, and the last alphabetically at the bottom.

**Acceptance Scenarios**:

1. **Given** the all-recipes page is loaded with multiple recipes, **When** the user activates the A→Z sort control, **Then** the recipe list immediately reorders in ascending alphabetical order (A first, Z last), case-insensitively.
2. **Given** the A→Z sort is active, **When** the user inspects the sort control, **Then** the control clearly indicates A→Z is the currently active sort direction.
3. **Given** the A→Z sort is active and the user navigates away and returns to the page, **Then** the default server-returned order is restored and no sort direction is pre-selected.

---

### User Story 2 — Sort Recipes Reverse Alphabetically Z→A (Priority: P2)

A visitor wants to browse recipes starting from the end of the alphabet. They activate the Z→A sort control and the recipe list reorders so recipes are displayed from Z to A.

**Why this priority**: Completes the bidirectional sort feature. Necessary for users scanning for recipes near the end of the alphabet, and for completeness of the sort interface.

**Independent Test**: Open the all-recipes page. Activate the Z→A sort control. Verify recipes are listed in descending alphabetical order (Z first, A last). Switch to A→Z and verify the list reverses.

**Acceptance Scenarios**:

1. **Given** the all-recipes page is loaded, **When** the user activates the Z→A sort control, **Then** the recipe list immediately reorders in descending alphabetical order (Z first, A last), case-insensitively.
2. **Given** Z→A sort is active, **When** the user activates A→Z, **Then** the list immediately switches to ascending order and the A→Z control shows as active.
3. **Given** a sort direction is active, **When** the user activates the same direction again, **Then** the list order and active state are unchanged (idempotent activation).

---

### User Story 3 — Accessible Sort Controls (Priority: P3)

A user who navigates with a keyboard or uses a screen reader can operate the sort controls without a mouse. The controls are keyboard-focusable, activatable with standard keys, and announced by assistive technology with their purpose and current state.

**Why this priority**: Accessibility ensures the feature is usable by people with motor, visual, or cognitive disabilities. The sort controls must meet WCAG 2.1 Level AA standards to be inclusive by default.

**Independent Test**: Using only the Tab key, navigate to the sort controls and activate each direction using Enter or Space. Use a screen reader to verify it announces the control label and active state when focused.

**Acceptance Scenarios**:

1. **Given** the user navigates the page using only a keyboard, **When** they Tab to the sort controls, **Then** each control receives a clearly visible focus indicator and can be activated with Enter or Space.
2. **Given** a screen reader is active, **When** the focus moves to a sort control, **Then** the screen reader announces the control's purpose and its current state (e.g., active or inactive).
3. **Given** a sort control is activated and the list reorders, **Then** the new sort state is discoverable by a screen reader user without additional navigation.
4. **Given** the sort controls are rendered on screen, **When** tested for color contrast, **Then** text and interactive elements meet WCAG 2.1 AA contrast requirements (4.5:1 for normal text, 3:1 for large text and UI components).

---

### Edge Cases

- What happens when the recipes page has zero recipes? Sort controls remain visible; the empty list state displays normally with no errors or layout breakage.
- What happens when two recipes share the same name? A stable sort preserves their original relative order from the server response.
- What is the default sort order when the page first loads? The server-returned order is shown and neither sort button is active — this is the explicit default, not a fallback.
- What if a recipe name begins with a number or special character? Standard Unicode collation applies: digits sort before letters, special characters sort by their Unicode code point.
- What if the recipe list is very long? Sorting is performed immediately on the already-loaded data; no additional loading state is needed for collections of typical cocktail recipe scale.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The all-recipes page MUST display a pair of adjacent toggle buttons — one labeled "A→Z" and one labeled "Z→A" — that allow sorting the recipe list in ascending and descending alphabetical order respectively. Both buttons are always visible; at most one is active at a time.
- **FR-002**: When a sort control is activated, the recipe list MUST reorder immediately without a full page reload.
- **FR-003**: Sort comparisons MUST be case-insensitive (e.g., "Margarita" and "margarita" sort identically).
- **FR-004**: The currently active sort direction MUST be visually distinguished from the inactive direction so the user can tell at a glance which order is in effect.
- **FR-005**: The sort controls MUST be keyboard-operable: reachable via Tab and activatable with Enter or Space.
- **FR-006**: Each sort control MUST have an accessible label that communicates its purpose and current active/inactive state to assistive technologies (screen readers).
- **FR-007**: Sort control text and interactive elements MUST meet WCAG 2.1 Level AA color contrast requirements (4.5:1 for normal text, 3:1 for large text and UI components).
- **FR-008**: On page load, the recipe list MUST display in the default server-returned order with neither sort button in an active state. The sort preference MUST NOT persist across page navigations or browser refreshes.
- **FR-009**: Sorting MUST handle an empty recipe list gracefully, with no errors and no layout breakage.
- **FR-010**: The sort controls MUST be visually consistent with the existing design system used across the application.

### Key Entities

- **Recipe**: Has a name (string) used as the sort key. Other recipe attributes are unaffected by this feature.
- **Sort Button Group**: A pair of adjacent toggle buttons ("A→Z" and "Z→A") that trigger reordering of the displayed recipe list. Each button has a direction property and an active/inactive state; at most one is active at a time. When neither is active, the list is in its default server-returned order.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After activating a sort control, the recipe list reorders in under 200 milliseconds for collections of up to 500 recipes.
- **SC-002**: All sort controls are reachable and operable via keyboard alone, with a visible focus indicator present at all times.
- **SC-003**: All sort control text and interactive elements pass WCAG 2.1 AA color contrast checks (4.5:1 for normal text, 3:1 for UI components).
- **SC-004**: A screen reader correctly announces the sort control label and active/inactive state when focus moves to each control.
- **SC-005**: The sort controls are discoverable — a first-time visitor can locate and use them within 10 seconds of landing on the all-recipes page without any instructions.

## Assumptions

- The all-recipes page already exists and displays a list of recipe cards or tiles; this feature adds sort controls to the existing page without redesigning the overall layout.
- Sorting is performed client-side on the already-loaded recipe data — no additional server requests are needed when changing sort direction.
- Recipe name is the sole sort key for this feature; sorting by other fields (spirit, author, date created) is out of scope.
- The sort preference is session-ephemeral: the user's sort choice is not saved between visits (covered by FR-008).
- The visual design of the sort controls follows the existing amber/stone design language established for this application.
- Accessibility target is WCAG 2.1 Level AA for the sort controls specifically; AAA conformance is not required.
- The application already has an "all recipes" page and route; no new page or navigation entry is needed.
