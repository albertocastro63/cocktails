# Quickstart: Base Spirit Designation — Integration Test Scenarios

## Prerequisites

- Dev server running (`sam local start-api` or `go run ./cmd/server`)
- Frontend dev server running (`cd frontend && npm run dev`)
- At least one user account (owner) and one admin account available
- Browser open to `http://localhost:5173`

---

## Scenario 1: Mark a Base Spirit When Creating a Recipe

1. Sign in as a regular user.
2. Navigate to **New Recipe** (`#/recipes/new`).
3. Add three ingredients: "Rye Whiskey", "Sweet Vermouth", "Angostura Bitters".
4. Click the base spirit toggle next to **Rye Whiskey**.
5. **Verify**: only Rye Whiskey's toggle is active; the others are clear.
6. Save the recipe.
7. **Verify**: the recipe detail page shows Rye Whiskey highlighted (bold + amber label) in the ingredient list.
8. **Verify**: hovering over the recipe card on the list page shows Rye Whiskey highlighted in the popover.

---

## Scenario 2: Change the Base Spirit

1. Open the recipe created in Scenario 1 and click **Edit**.
2. Click the base spirit toggle next to **Sweet Vermouth**.
3. **Verify**: Sweet Vermouth's toggle activates and Rye Whiskey's toggle clears automatically.
4. Save the recipe.
5. **Verify**: Sweet Vermouth is now highlighted on the detail page and popover; Rye Whiskey is not.

---

## Scenario 3: Clear the Base Spirit

1. Open the recipe and click **Edit**.
2. Click the active base spirit toggle (currently Sweet Vermouth) to deselect it.
3. **Verify**: no ingredient's toggle is active.
4. Save the recipe.
5. **Verify**: the detail page and popover show all ingredients with equal visual weight (no highlight).

---

## Scenario 4: Recipe with No Base Spirit

1. Create a new recipe without clicking any base spirit toggle.
2. Save the recipe.
3. **Verify**: the API response's `ingredients` array contains no `is_base_spirit` field on any ingredient.
4. **Verify**: the detail page and popover display all ingredients uniformly.

---

## Scenario 5: Legacy Recipe Display

1. Open any recipe that was created before this feature (no `is_base_spirit` in its data).
2. **Verify**: the detail page shows all ingredients with equal visual weight.
3. **Verify**: the card popover shows all ingredients with equal visual weight.
4. **Verify**: editing the legacy recipe works correctly — all ingredient rows load without errors; base spirit toggle is available but unset.

---

## Scenario 6: Delete the Base Spirit Ingredient During Editing

1. Open the recipe from Scenario 1 (Rye Whiskey marked as base spirit).
2. Click **Edit**.
3. Delete the Rye Whiskey ingredient row using the × button.
4. **Verify**: no other ingredient's base spirit toggle activates automatically.
5. Save the recipe.
6. **Verify**: the detail page and popover show all remaining ingredients with equal visual weight.

---

## Scenario 7: Non-Editor Sees No Base Spirit Control

1. Sign out (or view the recipe as a different user who is not the owner).
2. Navigate to the recipe detail page.
3. **Verify**: the base spirit toggle/control is not visible anywhere on the page.
4. **Verify**: the base spirit ingredient highlight IS visible (read-only display).
5. Hover over the recipe card on the list page.
6. **Verify**: the popover highlights the base spirit ingredient without showing any control.
