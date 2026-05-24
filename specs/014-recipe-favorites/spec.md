# Feature Specification: Recipe Favorites

**Feature Branch**: `014-recipe-favorites`
**Created**: 2026-05-22
**Status**: Draft
**Input**: User description: "For a logged user allow them to mark any cocktail that they have not entered as a favorite. This means that when that user navigates to a recipe, it can mark it as a favorite by clicking on a small icon that is shaped as a heart. When it is selected make the heart a tone of red, when not it is shaded gray. All the favorite cocktails will be shown in the my recipes page alongside the recipes entered by that user."

## Clarifications

### Session 2026-05-22

- Q: How should favorites be presented relative to the user's own created recipes on the "My Recipes" page? → A: Unified list with a visual badge — all recipes in one list, favorited items marked with a heart badge or "Favorited" label.
- Q: Should the heart icon appear on recipe cards in the "All Recipes" browse list, or only on the recipe detail page? → A: Detail page only — heart icon appears only when viewing a full recipe page.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Mark a Recipe as Favorite (Priority: P1)

A logged-in user navigates to any recipe detail page for a cocktail they did not create. They see a heart icon near the recipe title or header. Clicking the heart marks the recipe as a favorite — the icon turns red to confirm the action. Clicking the heart again removes it from their favorites — the icon returns to gray.

**Why this priority**: This is the core interaction of the feature. Without the ability to favorite/unfavorite, nothing else in this feature delivers value.

**Independent Test**: Log in as a user, navigate to a recipe created by a different user, and click the heart icon. Verify the icon changes to red. Reload the page — the heart should still be red, confirming the preference was saved. Click again to unfavorite and verify the icon returns to gray.

**Acceptance Scenarios**:

1. **Given** a logged-in user on a recipe detail page for a recipe they did not create, **When** the heart icon is in its default (gray/unfavorited) state and the user clicks it, **Then** the icon turns red and the recipe is added to the user's favorites.
2. **Given** a logged-in user on a recipe detail page for a recipe already in their favorites, **When** the user clicks the red heart icon, **Then** the icon returns to gray and the recipe is removed from their favorites.
3. **Given** a logged-in user who has favorited a recipe, **When** they navigate away and return to that recipe's detail page, **Then** the heart icon is shown red (persisted state).
4. **Given** a guest (not logged in) viewing a recipe detail page, **When** the page loads, **Then** no heart icon is shown (favoriting is only available to logged-in users).
5. **Given** a logged-in user viewing their own recipe, **When** the page loads, **Then** no heart icon is shown (users cannot favorite their own recipes).

---

### User Story 2 - View Favorites on My Recipes Page (Priority: P2)

A logged-in user navigates to the "My Recipes" page. All recipes they have created and all recipes they have favorited appear in a single unified list. Each favorited recipe is marked with a small heart badge so the user can distinguish it from recipes they created.

**Why this priority**: Discovering and accessing saved favorites is the payoff for the favoriting action. Without this, users have no way to benefit from what they've saved.

**Independent Test**: Log in, favorite at least one recipe created by another user, then navigate to "My Recipes." Verify the favorited recipe appears in the list with a heart badge. Unfavorite it and refresh — verify it no longer appears.

**Acceptance Scenarios**:

1. **Given** a logged-in user with at least one favorited recipe, **When** they visit the "My Recipes" page, **Then** their favorited cocktails appear in the unified list alongside their own recipes, each marked with a heart badge.
2. **Given** a logged-in user with no favorited recipes and no created recipes, **When** they visit "My Recipes," **Then** an appropriate empty state message is shown.
3. **Given** a logged-in user who unfavorites a recipe, **When** they visit or refresh "My Recipes," **Then** the unfavorited recipe no longer appears in the list.
4. **Given** a logged-in user with both created recipes and favorited recipes, **When** they view "My Recipes," **Then** the list shows all entries and favorited recipes are visually distinguishable via a heart badge.

---

### User Story 3 - Favorite Count or Social Signal (Priority: P3)

Recipe detail pages display the total number of users who have favorited that recipe, giving authors and visitors a sense of a recipe's popularity.

**Why this priority**: A nice-to-have social signal that adds value for authors and visitors. It is independent of the core save/retrieve flow and can be added without affecting P1 or P2.

**Independent Test**: Two separate users each favorite the same recipe. Navigate to that recipe's detail page and verify a count of "2 favorites" (or equivalent label) is shown.

**Acceptance Scenarios**:

1. **Given** a recipe that has been favorited by two different users, **When** any user (logged in or guest) views the recipe detail page, **Then** a count showing 2 favorites is visible.
2. **Given** a recipe with zero favorites, **When** any user views the detail page, **Then** a count of 0 (or no count) is shown without errors.

---

### Edge Cases

- What happens when a user tries to favorite a recipe that has since been deleted? The heart icon should disappear or the item should be removed from favorites gracefully.
- What happens when favoriting fails due to a connectivity issue? The icon should revert to its previous state and show a brief error message.
- Can a user favorite their own recipe? No — the heart icon is not shown on recipes the logged-in user created.
- What if the same user opens two tabs and favorites/unfavorites from both? The most recent action wins; no data corruption should occur.
- What if a recipe appears in both "my created recipes" and "my favorites" (edge case: user favorites a recipe, then it gets transferred)? The unified list must not show duplicates; the recipe appears once without a heart badge (ownership takes precedence).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST display a heart icon on recipe detail pages for recipes not created by the currently logged-in user. The heart icon does NOT appear on recipe cards in any browse or list view.
- **FR-002**: The heart icon MUST appear in a red tone when the recipe is in the user's favorites, and in a gray tone when it is not.
- **FR-003**: Clicking the heart icon MUST toggle the favorite state — adding the recipe to favorites if not favorited, or removing it if already favorited.
- **FR-004**: The favorite state MUST be persisted so that it survives page reloads and browser sessions.
- **FR-005**: The heart icon MUST NOT be shown to guests (unauthenticated users).
- **FR-006**: The heart icon MUST NOT be shown on recipes created by the currently logged-in user.
- **FR-007**: The "My Recipes" page MUST display favorited cocktails in a unified list alongside recipes created by the user.
- **FR-008**: Each favorited recipe in the "My Recipes" unified list MUST be marked with a visible heart badge to distinguish it from the user's own created recipes.
- **FR-009**: Removing a recipe from favorites MUST cause it to disappear from the "My Recipes" list on next page load or refresh.
- **FR-010**: If a favoriting action fails (e.g., network error), the heart icon MUST revert to its previous state and display a brief error indication as an inline message below the heart button (e.g., "Failed to save. Try again."), disappearing automatically after 4 seconds.
- **FR-011**: The unified "My Recipes" list MUST NOT show a recipe twice if it is both created by and favorited by the same user; the recipe appears once without a heart badge.
- **FR-012**: The system MUST display the total number of users who have favorited each recipe on the recipe detail page (P3 — may be deferred).

### Key Entities

- **Favorite**: A relationship between a user and a recipe they did not create, indicating the user has saved that recipe. Attributes: user identifier, recipe identifier, timestamp when favorited.
- **Recipe**: An existing entity; this feature adds a "favorited by count" derived attribute and the ability to be linked via Favorites.
- **User**: An existing entity; this feature adds a "favorites list" derived from the Favorite relationships they own.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A logged-in user can favorite or unfavorite a recipe in under 2 seconds with a single click, with visible feedback immediately within the same page view (optimistic UI).
- **SC-002**: Favorited recipes appear on the "My Recipes" page within one page load after being saved — no additional user actions required.
- **SC-003**: The heart icon correctly reflects the persisted favorite state on 100% of page loads for recipes a user has favorited.
- **SC-004**: The feature is accessible to keyboard-only users — the heart icon is reachable via Tab and activatable via Enter or Space.
- **SC-005**: The heart icon meets WCAG 2.1 AA color contrast requirements in both its red (favorited) and gray (unfavorited) states.

## Assumptions

- The existing authentication system correctly identifies the logged-in user and their authored recipes; no changes to authentication are in scope.
- A recipe's author/owner is already stored and accessible — the system can determine whether the logged-in user created a given recipe.
- The "My Recipes" page already exists and lists the user's created recipes; this feature extends it to also show favorited recipes in the same unified list with a heart badge.
- Favoriting is a per-user, per-recipe relationship (one user can only favorite a given recipe once).
- The feature applies to all recipes in the system, not just a subset.
- Mobile responsiveness of the heart icon follows the existing design system; no separate mobile-specific design is required.
- Real-time sync across multiple open tabs is out of scope — stale tab state is acceptable.
