# API Contract: Base Spirit Search Filter

**Feature**: 018-base-spirit-search | **Date**: 2026-05-25

---

## Modified Endpoint: GET /api/v1/recipes

### New Query Parameter

| Parameter     | Type   | Required | Default | Constraints |
|---------------|--------|----------|---------|-------------|
| `base_spirit` | string | No       | —       | Ignored if empty or whitespace-only. Max 200 chars (same as `q`). |

### Behaviour

1. If only `base_spirit` is present (no `q`): return recipes with a base-spirit ingredient matching the substring.
2. If both `q` and `base_spirit` are present: return the intersection (recipes matching all ingredient terms AND base-spirit).
3. If neither is present: return all recipes (existing behaviour).
4. Matching is always case-insensitive substring.

### Example Requests

```
GET /api/v1/recipes?base_spirit=gin
GET /api/v1/recipes?q=absinthe&base_spirit=rye+whiskey
GET /api/v1/recipes?base_spirit=rye+whiskey&page=2&limit=10
```

### Response (unchanged shape)

```json
{
  "data": [ ...Recipe ],
  "total": 3,
  "page": 1,
  "limit": 20
}
```

A base-spirit filter that matches no recipes returns an empty `data` array with `total: 0` — identical to any search that yields no results.

---

## Frontend Query-Building Contract

The `onSearch` handler in `RecipeList.js` MUST:

1. Apply whiskey/whisky normalisation to the raw input (replace `/whisky\b/gi` → `whiskey`) before any further parsing.
2. Extract the first occurrence of `base spirit is <value>` or `base spirit = <value>` (case-insensitive) from the normalised input, capturing `<value>` (trimmed).
3. Remove the matched clause from the remaining string to produce the cleaned `q`.
4. If `<value>` is empty or whitespace-only after trimming, discard it (no `base_spirit` parameter sent).
5. Call `getRecipes({ q: cleanedQ.trim() || undefined, base_spirit: value || undefined })`.

### Regex for clause extraction

```js
/base\s+spirit\s+(?:is|=)\s+(.+?)(?:\s+base\s+spirit\s+|$)/i
```

Used with a case-insensitive, global replace to strip the matched clause from `q`.
