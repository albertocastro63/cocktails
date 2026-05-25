# Implementation Plan: Base Spirit Search Filter

**Branch**: `018-base-spirit-search` | **Date**: 2026-05-25 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/018-base-spirit-search/spec.md`

---

## Summary

Add a `base spirit is <name>` / `base spirit = <name>` search syntax to the recipe list page. The client parses the clause, normalises whiskey/whisky spelling, and sends two parameters to the backend (`q` for remaining ingredient terms, `base_spirit` for the extracted spirit name). The backend gains two new store methods and updated handler dispatch to filter recipes by their flagged base-spirit ingredient.

---

## Technical Context

**Language/Version**: Go 1.25 (backend), Vanilla JS ES modules (frontend)
**Primary Dependencies**: SQLite (local/dev via `modernc.org/sqlite`), DynamoDB (production); Vitest (frontend tests); standard `go test` (backend tests)
**Storage**: SQLite (dev) — JSON column with `json_each`/`json_extract`; DynamoDB (prod) — full-scan with Go-level filter (existing pattern)
**Testing**: Vitest (frontend, coverage ≥ 75%); `go test ./...` (backend, coverage ≥ 75%)
**Target Platform**: Web SPA + AWS Lambda (arm64)
**Project Type**: Web application (frontend SPA + backend REST API)
**Performance Goals**: Existing p95 ≤ 200 ms read budget; base-spirit filter uses the same in-memory/JSON-scan approach as existing ingredient search — no additional latency expected
**Constraints**: No data model changes; `IsBaseSpirit` field already persisted; no FTS index changes needed
**Scale/Scope**: Small dataset (hundreds of recipes); full-scan approach consistent with existing implementation

---

## Constitution Check

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | New functions single-responsibility, ≤ 40 lines, cyclomatic complexity ≤ 10? | ✅ Query-parsing helper extracted; store methods are focused |
| II. Test-First | Failing tests written before implementation begins (TDD cycle)? | ✅ Test tasks precede implementation tasks in tasks.md |
| II. Coverage | Unit test coverage ≥ 75% for all business logic modules? | ✅ New store methods and handler branches fully tested |
| III. UX Consistency | Hint text updated; empty-state handled; no new error surfaces? | ✅ Hint text updated (FR-007); empty results use existing EmptyState component |
| IV. Performance | p95 ≤ 200 ms for read operations? | ✅ Base-spirit filter is equivalent cost to existing ingredient search |
| Quality Gates | Lint, coverage ≥ 75%, no regressions | ✅ Existing CI pipeline; no new dependencies added |

---

## Project Structure

### Documentation (this feature)

```text
specs/018-base-spirit-search/
├── plan.md              ← this file
├── research.md          ← Phase 0 output
├── data-model.md        ← Phase 1 output
├── quickstart.md        ← Phase 1 output
├── contracts/
│   └── api.md           ← Phase 1 output
└── tasks.md             ← Phase 2 output (/speckit-tasks)
```

### Source Code Changes

```text
# Frontend — recipe list page and search API client
frontend/src/pages/RecipeList.js          ← parse base-spirit clause, normalise whisky, update hint
frontend/src/pages/RecipeList.test.js     ← new tests for base-spirit parsing + whisky normalisation

# Backend — store interface
backend/internal/store/store.go           ← add SearchByBaseSpirit, SearchByBaseSpiritAndIngredients

# Backend — SQLite store
backend/internal/store/sqlite/recipes.go ← implement two new methods
backend/internal/store/sqlite/recipes_test.go ← new tests

# Backend — DynamoDB store
backend/internal/store/dynamo/recipes.go  ← implement two new methods
backend/internal/store/dynamo/ingredient_search_test.go ← new tests (or new file)

# Backend — handler
backend/internal/handler/recipes.go      ← extend List handler; parse base_spirit query param
backend/internal/handler/recipes_test.go ← new handler tests
```

---

## Implementation Strategy

### Phase 1 — Backend store layer (foundational)

Add `SearchByBaseSpirit` and `SearchByBaseSpiritAndIngredients` to the `RecipeStore` interface and implement in both SQLite and DynamoDB stores. Write failing tests first.

**SQLite implementation** (JSON path filter):
```sql
-- base-spirit-only WHERE clause
EXISTS (
  SELECT 1 FROM json_each(r.ingredients)
  WHERE json_extract(value,'$.is_base_spirit') = 1
    AND LOWER(json_extract(value,'$.name')) LIKE ?
)
```

**DynamoDB implementation** (Go-level filter, consistent with existing pattern):
```go
for _, ing := range r.Ingredients {
    if ing.IsBaseSpirit && strings.Contains(strings.ToLower(ing.Name), q) {
        return true
    }
}
```

### Phase 2 — Backend handler

Extend `RecipeHandler.List` to read the `base_spirit` query parameter and dispatch to the correct store method using the lookup table from research.md Finding 7.

### Phase 3 — Frontend query parsing

Add a `parseBaseSpirit(rawQ string) → { baseSpirit, q }` helper inside `RecipeList.js`. Apply whiskey/whisky normalisation first. Update `onSearch` to use the helper and pass `base_spirit` to `getRecipes`. Update hint text.

### Dispatch table (handler)

| `q` tokens | `base_spirit` | Store method |
|------------|---------------|--------------|
| 0          | absent        | `List` |
| 1          | absent        | `Search` |
| ≥2         | absent        | `SearchByIngredients` |
| 0          | present       | `SearchByBaseSpirit` |
| ≥1         | present       | `SearchByBaseSpiritAndIngredients` |

---

## Complexity Tracking

No constitution violations. All new functions are single-responsibility and well within the 40-line and complexity-10 limits.
