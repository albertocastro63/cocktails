# Quickstart: Ingredient Hover Popover

## Running the dev server

```sh
cd frontend
npm run dev
```

Navigate to the recipe list page (usually `http://localhost:5173`). Hover over any recipe tile to see the popover.

## Running tests

```sh
cd frontend
npm test
```

The relevant test file is `src/components/RecipeCard.test.js`.

## Running tests in watch mode

```sh
cd frontend
npm run test -- --watch
```

## What to verify manually

1. Hover over a recipe tile with ≤ 5 ingredients — all ingredient names appear, no ellipsis.
2. Hover over a recipe tile with > 5 ingredients — first 5 names appear, "…" on line 6.
3. Hover over a recipe with no ingredients — popover appears with "No ingredients listed."
4. Move mouse off the tile — popover disappears.
5. Quickly move across multiple tiles — at most one popover visible at a time.
6. Click a tile — still navigates to the recipe detail page (no regression).
