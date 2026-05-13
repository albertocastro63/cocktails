# Feature Specification: Ingredient Hover Popover

**Feature Branch**: `005-ingredient-hover-popover`  
**Created**: 2026-05-12  
**Status**: Draft  
**Input**: User description: "In the page where all cocktails are listed add the following feature: when the mouse hovers over the tile for the recipe, display a popover with the list of ingredients in that recipe. If there are more than 5 ingredients show an ellipsis in the sixth line. Hide popover when mouse is out of tile."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Ingredients on Hover (Priority: P1)

A user browsing the cocktail recipe list wants a quick way to preview a recipe's ingredients without navigating away from the list. When they hover their mouse over a recipe tile, a popover appears showing the ingredient names for that recipe. When the mouse leaves the tile, the popover disappears.

**Why this priority**: This is the core behavior of the feature. Every other scenario depends on this working correctly. It delivers immediate value by allowing users to scan ingredient lists without opening each recipe.

**Independent Test**: Can be fully tested by hovering over any recipe tile in the list and verifying the popover appears with the correct ingredients, then moving the mouse away and verifying it disappears.

**Acceptance Scenarios**:

1. **Given** the recipe list page is displayed with at least one recipe tile, **When** the user moves their mouse over a recipe tile, **Then** a popover appears showing the ingredients for that recipe.
2. **Given** a popover is visible for a recipe tile, **When** the user moves their mouse away from that tile, **Then** the popover disappears.
3. **Given** the recipe list page is displayed, **When** the user has not hovered over any tile, **Then** no popover is visible.

---

### User Story 2 - Long Ingredient Lists Truncated (Priority: P2)

A user hovers over a recipe tile that has more than 5 ingredients. The popover shows the first 5 ingredients and then an ellipsis on the sixth line to indicate there are more ingredients not shown.

**Why this priority**: Important for visual consistency and preventing popovers from becoming unwieldy, but depends on P1 being complete first.

**Independent Test**: Can be tested by hovering over a recipe tile that has 6 or more ingredients and verifying exactly 5 ingredient names appear followed by an ellipsis on the next line.

**Acceptance Scenarios**:

1. **Given** a recipe has exactly 5 or fewer ingredients, **When** the user hovers over its tile, **Then** the popover shows all ingredients with no ellipsis.
2. **Given** a recipe has more than 5 ingredients, **When** the user hovers over its tile, **Then** the popover shows exactly the first 5 ingredients followed by an ellipsis ("…") on the sixth line.
3. **Given** a recipe has exactly 6 ingredients, **When** the user hovers over its tile, **Then** the popover shows the first 5 ingredients and an ellipsis (the 6th ingredient is not shown individually).

---

### User Story 3 - Recipe with No Ingredients (Priority: P3)

A user hovers over a recipe tile that has no ingredients recorded. The popover still appears but indicates there are no ingredients listed.

**Why this priority**: Edge case handling — ensures the feature degrades gracefully for incomplete recipe data without breaking the UI.

**Independent Test**: Can be tested by hovering over a recipe tile with zero ingredients and verifying the popover appears with an appropriate empty state message.

**Acceptance Scenarios**:

1. **Given** a recipe has no ingredients, **When** the user hovers over its tile, **Then** a popover appears showing a message indicating no ingredients are listed.

---

### Edge Cases

- What happens when a recipe tile has zero ingredients? → Popover appears with an empty state message (e.g., "No ingredients listed").
- What happens if a recipe has exactly 5 ingredients? → All 5 are shown, no ellipsis.
- What happens if a recipe has exactly 6 ingredients? → First 5 shown, ellipsis on line 6 (the 6th ingredient is not individually displayed).
- What happens when the user quickly moves the mouse across multiple tiles? → Each tile shows its own popover; popover from the previous tile hides when mouse leaves it.
- What happens on touch/mobile devices where hover is not available? → Popovers are not shown; the feature is a mouse-only enhancement and does not affect touch interactions.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The recipe list page MUST display a popover when the user's mouse enters a recipe tile.
- **FR-002**: The popover MUST list the ingredient names belonging to the hovered recipe.
- **FR-003**: The popover MUST hide when the user's mouse leaves the recipe tile.
- **FR-004**: When a recipe has more than 5 ingredients, the popover MUST display only the first 5 ingredient names and show an ellipsis ("…") on the sixth line.
- **FR-005**: When a recipe has 5 or fewer ingredients, the popover MUST display all ingredient names without an ellipsis.
- **FR-006**: When a recipe has no ingredients, the popover MUST appear and display a message indicating no ingredients are available.
- **FR-007**: Only one popover MUST be visible at a time; hovering a new tile hides any previously visible popover.
- **FR-008**: The popover MUST be visually associated with (positioned near) its recipe tile.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can identify the ingredients of any recipe without navigating away from the list — verified by hovering over any tile and seeing its ingredient list within 1 second of the mouse entering the tile.
- **SC-002**: 100% of recipe tiles that have ingredients show the correct ingredient list in their popover.
- **SC-003**: 100% of recipe tiles with more than 5 ingredients show exactly 5 ingredient names followed by an ellipsis.
- **SC-004**: Popovers disappear immediately (within one animation frame) after the mouse leaves the tile, with no ghost popovers persisting on screen.
- **SC-005**: The feature does not disrupt existing navigation — clicking a recipe tile still navigates to the recipe detail page.

## Assumptions

- The ingredient data for each recipe is already available when the recipe list is rendered (no additional data fetching is required to show the popover).
- Ingredient names are plain text strings; no special formatting is required inside the popover.
- The feature is a mouse-hover enhancement only; it is not required on touch devices and does not affect the mobile experience.
- The "first 5 ingredients" refers to the order in which ingredients are stored for the recipe (no sorting is applied).
- The ellipsis displayed is the text character "…" or equivalent, not an interactive element (clicking it does not expand the list).
- The popover does not need to show ingredient quantities or units — names only.
