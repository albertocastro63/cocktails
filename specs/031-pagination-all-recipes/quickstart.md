# Quickstart: Pagination for All Recipes

## What changes

- **Frontend**: The "All Recipes" page gains Previous / Next controls, a "Page X of Y"
  indicator, and a total-results count. Sorting (A→Z / Z→A) now applies across the
  whole catalog, not just the current page.
- **Backend**: `GET /v1/recipes` accepts an optional `sort` param, and the recipe store
  now reads the full table across DynamoDB Scan batches so `total` is always correct.

## Run locally

```bash
# Backend (DynamoDB Local must be running; see feature 029)
cd backend
export DYNAMODB_ENDPOINT=http://localhost:8000
go run ./cmd/server

# Frontend
cd frontend
npm run dev
```

## Verify the fix manually

1. Ensure the recipes table has more than one page of recipes (e.g., 30 with the
   default page size of 20).
2. Open the app and go to **All Recipes**.
3. Confirm the page shows the first 20 recipes, a total count (e.g., "30 recipes"),
   and a "Page 1 of 2" indicator with **Next** enabled and **Previous** disabled.
4. Click **Next** → the remaining 10 recipes appear, indicator shows "Page 2 of 2",
   **Next** disabled, **Previous** enabled.
5. Click **A→Z**, then page through: names are in order across *both* pages (the last
   name on page 1 sorts before the first name on page 2), and you return to page 1.
6. Type a search term that matches more than 20 recipes: the total and controls now
   reflect the matches; paging reaches all of them.
7. Clear/change the search: the list resets to page 1.

## Run the tests

```bash
cd backend && go test ./...
cd frontend && npm test
```

Expected new/updated coverage:

- `backend/internal/store/dynamo/recipes_test.go` — multi-batch scan returns all items;
  `total` correct; `name_asc`/`name_desc` order the full set before paging.
- `backend/internal/handler/recipes_test.go` — `page`/`limit`/`sort` behavior;
  out-of-range page returns empty page with correct total.
- `frontend/src/components/Pagination.test.js` — control rendering, disabled states,
  callbacks.
- `frontend/src/pages/RecipeList.pagination.test.js` — paging fetches the right page,
  resets to page 1 on filter/sort change, hides controls when a single page.

## Rollback

Frontend and backend changes are independent and backward compatible: reverting the
frontend leaves the endpoint working (just back to showing page 1); reverting the
`sort` param leaves pagination working with default order. The `scanAll` change is a
pure correctness fix with no contract change.
