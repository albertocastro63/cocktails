# Data Model: Recipe Sort Order

**Feature**: 013-recipe-sort-order
**Date**: 2026-05-21

This feature introduces no new backend data entities and no API changes. All changes are confined to the frontend UI layer.

---

## New UI Components

### SortButtonGroup

A stateless UI component that renders an adjacent pair of toggle buttons for selecting sort direction.

| Property | Type | Description |
|----------|------|-------------|
| `currentDir` | `'asc' \| 'desc' \| null` | Currently active sort direction; `null` means default server order (no direction selected) |
| `onSort` | `(dir: 'asc' \| 'desc') => void` | Callback invoked when the user clicks a sort button; receives the direction of the clicked button |

**State transitions**:
```
null  ──click A→Z──▶  'asc'  ──click Z→A──▶  'desc'
 ▲                      │                        │
 └──────────────────────┘                        │
 ▲                                               │
 └───────────────────────────────────────────────┘
(page navigation resets to null)
```

**Rendered structure**:
```html
<div role="group" aria-label="Sort recipes">
  <button aria-pressed="true|false"  data-dir="asc">A→Z</button>
  <button aria-pressed="true|false"  data-dir="desc">Z→A</button>
</div>
```

---

## Modified UI State (RecipeList page)

The `RecipeList` page gains one additional state variable:

| Variable | Type | Initial Value | Description |
|----------|------|---------------|-------------|
| `currentSortDir` | `'asc' \| 'desc' \| null` | `null` | The active sort direction; `null` = server order |

The page renders recipe cards from a derived `displayRecipes` array:
- `currentSortDir === null` → `displayRecipes = data` (server order, unchanged)
- `currentSortDir === 'asc'`  → `displayRecipes = sortRecipes(data, 'asc')`
- `currentSortDir === 'desc'` → `displayRecipes = sortRecipes(data, 'desc')`

The `data` variable always holds the original server response; sorting creates a shallow copy and never mutates `data`.

---

## New Utility Function

### sortRecipes(recipes, direction)

A pure function that returns a sorted shallow copy of the recipes array.

| Parameter | Type | Description |
|-----------|------|-------------|
| `recipes` | `Recipe[]` | Array of recipe objects from the API |
| `direction` | `'asc' \| 'desc'` | Sort direction |

**Returns**: `Recipe[]` — a new array sorted by `recipe.name` using `localeCompare` with `sensitivity: 'base'` (case-insensitive Unicode collation).

**Location**: Implemented as a named export in `frontend/src/pages/RecipeList.js` (or extracted to `frontend/src/utils/sort.js` if reuse is anticipated — deferred to implementation).

---

## Existing Entities (unchanged)

- **Recipe** (API response): `{ id, name, ingredients, creator_id, ... }` — `name` is used as the sort key; no new fields required.
- **API** (`getRecipes`): No changes — sorting happens client-side on the returned `data` array.

---

## Component Relationship

```
RecipeList (page)
  ├── SearchBar          (existing — unchanged)
  ├── SortButtonGroup    (NEW — receives currentSortDir + onSort callback)
  │     ├── Button "A→Z" (aria-pressed, amber active state)
  │     └── Button "Z→A" (aria-pressed, stone inactive state)
  ├── loading state      (existing — unchanged)
  ├── EmptyState         (existing — unchanged)
  └── recipe grid
        └── RecipeCard[] (existing — receives sorted displayRecipes)
```
