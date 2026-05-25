# Quickstart: Site Footer and Ingredient Popup Layout Fix

**Feature**: 017-footer-popup-fix | **Date**: 2026-05-25

## Purpose

Manual acceptance test protocol for the two user stories. Run these after `npm run dev` is serving the frontend.

---

## Acceptance Test 1 — Footer on every page (US1)

**Independent Test**: Navigate to any page and confirm a footer with a separator line and copyright text is visible at the bottom of the content.

### Steps

1. Open `http://localhost:5173` (or the Vite dev server URL).
2. Navigate to `/` (Home). Scroll to the bottom. Confirm you see:
   - A horizontal separator line.
   - A copyright notice below it reading `© 2026 Cocktails` (current year).
3. Navigate to `#/recipes` (Recipe List). Scroll to the bottom. Confirm the same footer appears.
4. Navigate to `#/login`. Scroll to the bottom. Confirm the footer appears on the login page.
5. Navigate to `#/recipes/new`. Scroll to the bottom. Confirm the footer appears.
6. Resize the browser to a narrow viewport (e.g., 375px wide). Confirm the separator line spans the full content width but does not touch the browser edges (matches the content container's padding).
7. Resize to a very wide viewport (e.g., 2000px). Confirm the separator line does NOT extend to the browser edges—it stays within `max-w-4xl` bounds.

**Pass criteria**: Footer visible on all pages checked, separator constrained to content area, copyright legible.

---

## Acceptance Test 2 — Ingredient popup as non-expanding overlay (US2)

**Independent Test**: Trigger the ingredient popup on a recipe card and confirm that the page total height, scroll position, and positions of all other cards remain unchanged.

### Setup

Ensure at least 3 recipes exist so the grid has multiple rows. If not, create some via the admin panel or API.

### Steps

1. Navigate to `#/recipes`.
2. Open browser DevTools → Console. Note the current `document.documentElement.scrollHeight` value by running: `document.documentElement.scrollHeight`.
3. Hover over any recipe card in the **middle of the grid**. The ingredient popup should appear below the card.
4. While the popup is visible, run `document.documentElement.scrollHeight` in the console again. Confirm the value is **identical** to step 2.
5. Confirm that no other recipe cards have shifted position.
6. Move the cursor off the card. Confirm the popup closes and the layout is unchanged.
7. Hover over a recipe card in the **last row** (bottom of the list). Confirm the popup appears at its natural position below the card without causing other cards to shift.
8. If the last-row popup extends below the visible viewport: scroll down to confirm you can scroll to see it. Note: a slight scroll height increase for the very last card is acceptable per spec edge case.
9. Hover over one card, then quickly hover over another. Confirm only one popup is visible at a time.

**Pass criteria**: Step 4 shows identical scroll height; step 5 shows no card movement; step 6 shows clean close; step 9 shows single popup at a time.

---

## Running Unit Tests

```bash
cd frontend
npm test -- --coverage
```

All tests must pass. Coverage must remain ≥ 75%.

**New test files to confirm present**:
- `src/components/Footer.test.js`

**Modified test files with new cases**:
- `src/components/RecipeCard.test.js`
