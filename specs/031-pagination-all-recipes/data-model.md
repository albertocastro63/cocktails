# Phase 1 Data Model: Pagination for All Recipes

This feature adds no new persisted entities and no schema changes to the DynamoDB
recipes table. It introduces presentation/transport concepts and a query parameter.

## Entities

### Recipe (existing — unchanged)

The stored cocktail entity. No attributes are added or modified. Only its
membership in and position within a result set are relevant here.

### Result Set (conceptual)

The ordered collection of recipes being browsed.

- **Unfiltered**: every recipe in the catalog.
- **Filtered**: recipes matching the active search term, ingredient tokens, and/or
  base-spirit filter.
- **Ordering**: stable scan order by default, or by name ascending/descending when a
  `sort` is selected. Ordering is applied over the *entire* result set before paging.

### Result Page (transport)

A bounded slice of the Result Set returned by the listing endpoint. Shape is the
existing response, unchanged:

| Field  | Type            | Meaning |
|--------|-----------------|---------|
| `data` | array<Recipe>   | The recipes on the requested page (may be empty if page is out of range). |
| `total`| integer         | Count of recipes in the full result set (all pages), after filtering. |
| `page` | integer         | The 1-based page number echoed back. |
| `limit`| integer         | The page size echoed back (default 20, max 100). |

**Derived values (frontend)**: `totalPages = max(1, ceil(total / limit))`;
`hasPrev = page > 1`; `hasNext = page < totalPages`.

## Query parameters (listing endpoint)

| Param         | Type    | Default | Rules |
|---------------|---------|---------|-------|
| `page`        | integer | 1       | 1-based. Values `< 1` clamp to 1. Out-of-range (beyond last page) yields empty `data` with correct `total`. |
| `limit`       | integer | 20      | Capped at 100. |
| `q`           | string  | ""      | Existing search / ingredient tokens (unchanged). |
| `base_spirit` | string  | ""      | Existing base-spirit filter (unchanged). |
| `sort`        | string  | ""      | NEW. One of `name_asc`, `name_desc`. Empty or unrecognized ⇒ default scan order. |

## Validation & invariants

- **INV-1**: `total` reflects every stored recipe in the result set, independent of
  how many were returned on the current page or how many Scan batches the store read.
- **INV-2**: For a stable catalog and a fixed `sort`, paging over all pages yields each
  recipe exactly once (no duplicates, no gaps) — the sort is applied before slicing.
- **INV-3**: `data` length ≤ `limit`.
- **INV-4**: When `total == 0`, `data` is an empty array and the frontend shows the
  empty state with no pagination controls.
- **INV-5**: An unrecognized `sort` value degrades gracefully to default order rather
  than erroring.

## State transitions (frontend listing)

```text
Initial load ─▶ fetch(page=1, sort=current) ─▶ render grid + controls
Next clicked (hasNext) ─▶ fetch(page+1) ─▶ render
Prev clicked (hasPrev) ─▶ fetch(page-1) ─▶ render
Search / filter changed ─▶ reset page=1 ─▶ fetch ─▶ render
Sort changed ─▶ reset page=1 ─▶ fetch(sort=new) ─▶ render
```
