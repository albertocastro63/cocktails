# Feature Specification: Favorite Heart Indicators on All Recipes Page

**Feature Branch**: `020-fix-favorites-heart`  
**Created**: 2026-05-28  
**Status**: Draft  
**Input**: User description: "There seems to have been a regression, the heart markers that show that a recipe has been favorited are no longer there. If I login and navigate to My Recipes I can see the heart markers in the recipes I have favorited, but those are absent in the All Recipes page. Please correct that."

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Favorite Indicators on the All Recipes Page (Priority: P1)

When a logged-in user browses the All Recipes page, recipes they have previously favorited display the same heart marker that is visible on the My Recipes page. Anonymous users and users with no favorites see no heart markers.

**Why this priority**: The All Recipes page is the primary discovery surface. Without favorite indicators, users lose at-a-glance awareness of which recipes they have saved, creating an inconsistent experience compared to the My Recipes page.

**Independent Test**: Log in, navigate to All Recipes — heart markers appear on all favorited recipes and are absent on non-favorited ones. Log out and revisit — no markers appear.

**Acceptance Scenarios**:

1. **Given** a logged-in user with at least one favorited recipe, **When** they view the All Recipes page, **Then** each favorited recipe card displays a heart marker identical to the one shown on the My Recipes page.
2. **Given** a logged-in user with no favorited recipes, **When** they view the All Recipes page, **Then** no heart markers appear on any recipe card.
3. **Given** an anonymous (not logged-in) visitor, **When** they view the All Recipes page, **Then** no heart markers appear.
4. **Given** a logged-in user who favorites a recipe on the detail page and then returns to All Recipes, **When** the page loads, **Then** the newly-favorited recipe shows the heart marker.
5. **Given** a logged-in user whose favorite indicators are visible on both All Recipes and My Recipes, **Then** the same set of recipes show the heart marker on both pages.

---

### Edge Cases

- If the favorites lookup fails (network error), the All Recipes page MUST still load and display all recipes without heart markers — the failure must not block the primary listing.
- Heart markers on the All Recipes page are read-only indicators; they do not need to be interactive (toggling favoriting is handled on the detail page).
- Recipes owned by the logged-in user should display heart markers if also favorited, following the same logic as the My Recipes page.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When a logged-in user views the All Recipes page, each recipe card for a recipe they have favorited MUST display the same heart marker that appears on the My Recipes page.
- **FR-002**: The heart marker MUST NOT appear on recipe cards the user has not favorited.
- **FR-003**: No heart markers MUST appear for anonymous users on the All Recipes page.
- **FR-004**: The All Recipes page MUST continue to load and display all recipes even if the favorites status cannot be retrieved.
- **FR-005**: The heart marker visual appearance (icon, colour, position) MUST be identical on both the All Recipes page and the My Recipes page.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of a logged-in user's favorited recipes display the heart marker on the All Recipes page, matching the My Recipes page.
- **SC-002**: The All Recipes page loads and shows all recipes within the same time budget as before this change; the additional favorites lookup MUST NOT cause a visible page-load regression.
- **SC-003**: No heart markers are visible to anonymous users or to logged-in users with zero favorites.

## Assumptions

- Heart markers on the All Recipes page are display-only; the interaction for adding or removing favorites remains on the recipe detail page.
- The favorites data is fetched at page load; real-time updates within the same session (e.g., favoriting from a different browser tab) are out of scope.
- The same visual component and styling used for heart markers on the My Recipes page is reused without modification on the All Recipes page.
- If the logged-in user's session token is missing or invalid at the time of the favorites fetch, the page degrades gracefully with no markers shown rather than showing an error.
