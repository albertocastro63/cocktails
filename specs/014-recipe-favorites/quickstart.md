# Quickstart: Recipe Favorites Verification Scenarios

**Feature**: 014-recipe-favorites
**Date**: 2026-05-22

Use these scenarios to manually verify the feature after implementation, and as the basis for integration test design.

---

## SC-001: Favorite a recipe (P1 — core happy path)

**Setup**: Two users exist — User A (recipe author) and User B (another logged-in user). User A has created at least one recipe.

1. Log in as User B.
2. Navigate to a recipe detail page for a recipe created by User A.
3. **Expect**: A gray heart icon is visible near the recipe title.
4. Click the heart icon.
5. **Expect**: The heart icon turns red immediately (optimistic update or post-save render).
6. Reload the page.
7. **Expect**: The heart icon is still red — state was persisted.

---

## SC-002: Unfavorite a recipe (P1)

1. With User B logged in, navigate to a recipe that is already favorited (heart is red).
2. Click the red heart icon.
3. **Expect**: The heart turns gray.
4. Reload the page.
5. **Expect**: The heart is still gray.

---

## SC-003: Heart icon not shown on own recipe (P1)

1. Log in as User A (the recipe creator).
2. Navigate to a recipe created by User A.
3. **Expect**: No heart icon is visible anywhere on the page.

---

## SC-004: Heart icon not shown to guests (P1)

1. Log out (or open in an incognito window).
2. Navigate to any recipe detail page.
3. **Expect**: No heart icon is visible.

---

## SC-005: Favorited recipe appears on My Recipes (P2 — unified list)

1. Log in as User B.
2. Favorite a recipe created by User A (SC-001 above).
3. Navigate to "My Recipes."
4. **Expect**: The favorited recipe appears in the list with a small heart badge visible on the card.
5. **Expect**: If User B also has their own created recipes, those appear in the same list without a heart badge.

---

## SC-006: Unfavorited recipe disappears from My Recipes (P2)

1. With User B on the "My Recipes" page, verify the favorited recipe is shown.
2. Click on that recipe to open the detail page.
3. Click the red heart to unfavorite it.
4. Navigate back to "My Recipes."
5. **Expect**: The previously favorited recipe is no longer in the list.

---

## SC-007: Empty My Recipes page (P2 — empty state)

1. Log in as a brand-new user with no created recipes and no favorites.
2. Navigate to "My Recipes."
3. **Expect**: An empty state message is shown (e.g., "You haven't added any recipes yet.").

---

## SC-008: No duplicate in My Recipes (P2 — deduplication edge case)

*This is a theoretical edge case — in practice a user cannot favorite their own recipe. Verify the guard works:*

1. Log in as User A.
2. Navigate to a recipe created by User A.
3. **Expect**: No heart icon is shown. The recipe does NOT appear twice on "My Recipes" — it appears once without a badge.

---

## SC-009: Favoriting failure reverts UI (P1 — error handling)

1. Log in as User B, navigate to an unfavorited recipe (gray heart).
2. Disable network (browser dev tools → Offline).
3. Click the heart icon.
4. **Expect**: The heart briefly shows red (optimistic), then reverts to gray.
5. **Expect**: A brief error message or indication is shown (e.g., "Failed to save favorite").
6. Re-enable network and click again.
7. **Expect**: The heart turns red and the state persists on reload.

---

## SC-010: Keyboard accessibility (P1 — WCAG)

1. Log in as User B, navigate to a recipe not created by User B.
2. Tab through the page until the heart icon button is focused.
3. **Expect**: A visible focus ring appears on the heart button.
4. Press Enter or Space.
5. **Expect**: The heart toggles (same behavior as a mouse click).
6. Verify screen reader (VoiceOver: Cmd+F5) announces: "Add to favorites, button" (unfavorited) or "Remove from favorites, button" (favorited).
