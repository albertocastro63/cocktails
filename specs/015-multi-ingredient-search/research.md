# Research: Multi-Ingredient Search

**Feature**: 015-multi-ingredient-search
**Date**: 2026-05-24

## Decision 1: Query Parsing Location

**Decision**: Parse the compound query string in the HTTP handler layer, not in the store layer.

**Rationale**: The handler already reads the `q` URL parameter and routes to `Search()` or `List()`. Adding a `parseIngredientQuery(q string) []string` helper at the handler level keeps the store interface clean. The store receives either a single string (existing `Search`) or a slice of strings (new `SearchByIngredients`).

**Alternatives considered**:
- Parse inside the store — rejected: mixes protocol concerns with storage concerns; harder to test the parser independently.
- Parse in the frontend and send separate `ingredient[]` params — rejected: changes the public API contract; breaks existing integrations.

---

## Decision 2: `q` Parameter Delimiters

**Decision**: Split on two patterns:
1. `\s+and\s+` (case-insensitive regex) — natural language form
2. `\s*\+\s*` (regex) — shorthand form, optional surrounding whitespace

If both patterns are present (unlikely but possible), apply `and` split first, then `+` split on the resulting tokens. Empty tokens after `strings.TrimSpace` are discarded.

**Rationale**: The spec defines exactly these two forms. The `and` delimiter requires at least one whitespace character on each side to avoid splitting ingredient names like "ginger and lime cordial" incorrectly when typed as a bare word. The `+` delimiter allows no-space form (`Gin+Lemon`) for convenience.

**Alternatives considered**:
- Single delimiter (`+` only, map `and` → `+` first) — simpler but ` and ` with space requirements differs from `+`, so treating them identically would lose the whitespace-requirement distinction for `and`.

---

## Decision 3: SQLite Multi-Ingredient Query Strategy

**Decision**: Use `json_each(ingredients)` with a dynamically-built set of correlated subqueries — one `EXISTS` clause per ingredient token.

```sql
SELECT COUNT(*) FROM recipes r
WHERE
  (SELECT COUNT(*) FROM json_each(r.ingredients)
   WHERE LOWER(json_extract(value, '$.name')) LIKE ?) > 0
  AND
  (SELECT COUNT(*) FROM json_each(r.ingredients)
   WHERE LOWER(json_extract(value, '$.name')) LIKE ?) > 0
  -- ... repeated for each ingredient token
```

Args: one `%token%` per ingredient. The WHERE clause is built programmatically in Go; the number of placeholders matches `len(ingredients)`.

**Rationale**: The existing FTS index (`recipes_fts`) stores a flattened `search_text` field covering name + ingredient names + steps + properties. A multi-token FTS query (`gin* lemon*`) would match ANY document containing both words anywhere — including recipe names or steps — which violates FR-002 (ingredient-only match). `json_each` targets ingredient names specifically. SQLite has supported `json_each` since version 3.38 (2022), available in all current Go sqlite drivers.

**Performance**: `json_each` with a LIKE scan is O(n × m) where n = total recipes and m = ingredients count. Acceptable for current scale (hundreds to low thousands of recipes). An index on ingredient names would require schema change and is deferred.

**Alternatives considered**:
- FTS5 `gin* lemon*` (implicit AND across full `search_text`) — rejected: matches name/steps/properties, not ingredient-only.
- Separate `recipe_ingredients` normalized table — rejected: requires schema migration; over-engineering for current scale.

---

## Decision 4: DynamoDB Multi-Ingredient Filter Strategy

**Decision**: Add a `matchesAllIngredients(r *model.Recipe, ingredients []string) bool` helper that iterates each required ingredient and confirms at least one recipe ingredient name contains it (case-insensitive substring). The existing full-table `Scan` already fetches all items; the new helper replaces `matchesQuery` when `len(ingredients) > 1`.

**Rationale**: The DynamoDB implementation already performs a full Scan + in-memory filter. Adding an all-ingredient filter in the same in-memory pass costs zero additional DynamoDB reads. No new GSI is needed.

**Alternatives considered**:
- DynamoDB FilterExpression on the scan — rejected: DynamoDB FilterExpression does not support `contains` on nested list attributes in a multi-condition AND across the same list. In-memory filtering is equivalent in cost for current data volume.

---

## Decision 5: Store Interface Extension

**Decision**: Add `SearchByIngredients(ingredients []string, page, limit int) ([]*model.Recipe, int, error)` as a new method on `RecipeStore`.

**Rationale**: Keeps single-term and multi-term paths cleanly separated; each is independently testable. The handler decides which method to call based on the parsed token count.

**Alternatives considered**:
- Overload the existing `Search(query string, ...)` by encoding multiple ingredients with a delimiter — rejected: leaks parsing concerns into the store; makes the store method signature misleading.

---

## Decision 6: Frontend Changes

**Decision**: Add a single-line hint text element below the `SearchBar` component on the RecipeList page. No changes to `SearchBar` itself or `client.js` — the raw query string is already forwarded to `getRecipes({ q })` unchanged.

Hint text: `Tip: use "and" or "+" to search multiple ingredients — e.g. "gin and lemon"`

**Rationale**: Users need syntax discovery without a tutorial. A one-line tip below the search bar is the minimal sufficient affordance. The hint should use `text-stone-400 text-sm` to match the existing design language (muted, non-distracting).

**Alternatives considered**:
- Placeholder text in the SearchBar input — rejected: placeholder disappears when the user starts typing; usability research shows placeholder-as-instruction is unreliable.
- Tooltip on a help icon — rejected: adds complexity; the tip is brief enough to display inline.
