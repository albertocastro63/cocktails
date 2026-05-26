# Research: Base Spirit Search Filter

**Feature**: 018-base-spirit-search | **Date**: 2026-05-25

---

## Finding 1 — Client-side query parsing location

**Decision**: Parse the `base spirit is / base spirit =` clause in `RecipeList.js` (the `onSearch` callback) before calling `getRecipes`. The `SearchBar` component stays generic; the recipe list page owns the domain-specific syntax.

**Rationale**: `SearchBar` is a presentable UI primitive used (or likely used) in multiple contexts. Keeping syntax parsing in the page-level `onSearch` handler avoids leaking recipe-domain logic into a shared component, mirrors how the existing `and`/`+` splitting is handled (it is in the handler layer, not in `SearchBar`).

**Alternatives considered**:
- Parse inside `SearchBar` — rejected; makes the component non-reusable and mixes concerns.
- Parse in `client.js#getRecipes` — rejected; API clients should not know about UI query syntax.

---

## Finding 2 — Whiskey/whisky normalisation location

**Decision**: Normalise in the `onSearch` handler (same location as the base-spirit parser) by replacing `/whisky\b/gi` with `whiskey` in both the remaining `q` string and the extracted `baseSpirit` value before calling `getRecipes`.

**Rationale**: The spec (clarification Q2) specifies client-side expansion. Doing it in the same function that strips the base-spirit clause keeps all query preprocessing in one place, making it easy to test and reason about.

**Alternatives considered**:
- Backend normalisation — rejected per clarification Q2 decision.
- Separate utility function — acceptable, but inline replacement is simpler for a two-form substitution; can be extracted later if more variants emerge.

---

## Finding 3 — Backend: new `SearchByBaseSpirit` store method vs. post-filter

**Decision**: Add a dedicated `SearchByBaseSpirit(baseSpirit string, page, limit int)` method to the `RecipeStore` interface and implement it in both the SQLite and DynamoDB stores. The handler calls this when `base_spirit` is present (with or without `q`).

**Rationale**: The existing search methods do not accept a base-spirit constraint. A dedicated method keeps the handler's switch statement clean and allows each store to implement the most efficient filter (JSON path expression for SQLite, in-memory check for DynamoDB — consistent with how `SearchByIngredients` works).

**Alternatives considered**:
- Extend `Search` / `SearchByIngredients` with an optional extra parameter — rejected; adds complexity to already-used signatures and changes the interface in a backward-incompatible way.
- Post-filter in the handler — rejected; means fetching a full page of un-filtered results and then discarding most, which breaks pagination counts.

---

## Finding 4 — Combined base-spirit + ingredient query

**Decision**: When both `q` (ingredient terms) and `base_spirit` are present, apply a two-stage filter: first filter by base spirit (in-memory for DynamoDB, SQL WHERE for SQLite), then apply the ingredient filter on the resulting subset.

**Rationale**: The spec requires the intersection of both constraints. The cleanest handler dispatch is: `SearchByBaseSpiritAndIngredients` (new) for the combined case, `SearchByBaseSpirit` for base-spirit-only, existing paths for ingredient-only and no-filter.

**Alternatives considered**:
- Single combined store method for all cases — simpler interface but requires all four callers to accept optional-ness.
- Two sequential queries — rejected; breaks pagination counts for the combined case.

---

## Finding 5 — SQLite base-spirit filter implementation

**Decision**: Filter using a JSON path expression on the `ingredients` column:

```sql
EXISTS (
  SELECT 1 FROM json_each(r.ingredients)
  WHERE json_extract(value,'$.is_base_spirit') = 1
    AND LOWER(json_extract(value,'$.name')) LIKE ?
)
```

With `'%' + strings.ToLower(baseSpirit) + '%'` as the bind parameter — consistent with the substring matching used in `SearchByIngredients`.

**Rationale**: The ingredients column stores JSON; `json_each` / `json_extract` is already used in `SearchByIngredients`. No schema change is needed because `is_base_spirit` is already persisted as part of the ingredients JSON blob.

**Alternatives considered**:
- FTS5 index for base-spirit — over-engineered; the FTS `search_text` column does not preserve ingredient structure, so a direct JSON query is more precise.

---

## Finding 6 — DynamoDB base-spirit filter implementation

**Decision**: Implement `SearchByBaseSpirit` in the DynamoDB store using the existing full-table scan pattern (consistent with `Search` and `SearchByIngredients`), filtering in Go:

```go
for _, ing := range r.Ingredients {
    if ing.IsBaseSpirit && strings.Contains(strings.ToLower(ing.Name), q) {
        return true
    }
}
```

**Rationale**: DynamoDB store already performs full scans for all search operations. Adding a Go-level filter is the established pattern. No DynamoDB table changes are required.

---

## Finding 7 — Handler dispatch table

The updated `List` handler will use this dispatch:

| `q` present | `base_spirit` present | Method called |
|-------------|----------------------|---------------|
| No          | No                   | `List` (existing) |
| Yes (single token) | No            | `Search` (existing) |
| Yes (multi-token)  | No            | `SearchByIngredients` (existing) |
| No          | Yes                  | `SearchByBaseSpirit` (new) |
| Yes (any)   | Yes                  | `SearchByBaseSpiritAndIngredients` (new) |

---

## Finding 8 — Placeholder hint text

**Decision**: Update the `hint` paragraph text in `RecipeList.js` to:

> `Tip: search by ingredient (use "and" or "+" for multiple) — or try "base spirit is gin"`

**Rationale**: FR-007 requires the hint text to advertise the new syntax. The existing hint already explains the `and`/`+` syntax; adding a concise example for base-spirit keeps the hint short and scannable.
