# Quickstart: Visual Redesign

## Running the dev server

```sh
cd frontend
npm run dev
```

Navigate to `http://localhost:5173` and verify the redesign visually.

## Running tests

```sh
cd frontend
npm test
```

## Pages to verify manually

| Page | URL | What to check |
|---|---|---|
| Home | `#/` | Dark hero band, white heading, amber subtext, amber "All Recipes" CTA visible |
| Recipe List | `#/recipes` | Stone-50 background, amber-accented cards, amber left border on each card |
| Recipe Detail | `#/recipes/:id` | Uppercase amber section labels, amber-styled edit button, stone body text |
| Sign In | `#/login` | Amber focus ring on inputs, amber "Sign In" button, polished card container |
| New Recipe | `#/recipes/new` | Amber focus rings on all inputs, amber submit button |

## Responsive checks

Resize browser window to verify:
- **375px** (mobile): Nav links remain accessible, cards stack to 1 column, no horizontal overflow
- **768px** (tablet): 2-column card grid, hero text readable
- **1280px** (desktop): 3-column card grid, full layout

## Regression checks

- Navigate between all pages — routing must work unchanged
- Log in and log out — auth must work unchanged
- Create a new recipe — form submission must work unchanged
- Edit and delete a recipe — CRUD must work unchanged
