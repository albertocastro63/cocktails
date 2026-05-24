# Tasks: Multi-Ingredient Search

**Input**: Design documents from `specs/015-multi-ingredient-search/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/api.md ✅, quickstart.md ✅

**TDD Mandate** (Constitution II): Every test task MUST be written and confirmed FAILING before
its corresponding implementation task begins.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[US1]**: User Story 1 — AND Search for Multiple Ingredients

---

## Phase 1: Foundational (Blocking Prerequisite)

**Purpose**: Extend the `RecipeStore` interface with `SearchByIngredients`. Both store
implementations and the handler depend on this definition — nothing else can proceed without it.

- [ ] T001 Add `SearchByIngredients(ingredients []string, page, limit int) ([]*model.Recipe, int, error)` method to `RecipeStore` interface in `backend/internal/store/store.go`

**Checkpoint**: Interface is defined. SQLite, DynamoDB, and handler work can now proceed.

---

## Phase 2: User Story 1 — AND Search for Multiple Ingredients (Priority: P1) 🎯 MVP

**Goal**: Users can type `gin and lemon` or `gin + lemon + sugar` in the search bar on the
all-recipes page and receive only recipes whose ingredient list contains all named items. A hint
below the search bar surfaces the syntax.

**Independent Test**: `GET /api/v1/recipes?q=gin+and+lemon` returns only recipes containing both
"gin" and "lemon" in their ingredient list. `GET /api/v1/recipes?q=gin` returns existing
single-term FTS results unchanged.

### Tests for User Story 1 (TDD — write and confirm FAILING before implementation)

- [ ] T002 [P] [US1] Write failing SQLite `SearchByIngredients` tests covering: 2-ingredient AND match, 3-ingredient AND match, no-match empty result, case-insensitivity, substring match (e.g. "gin" matches "Sloe Gin"), pagination (page 2), single-empty-token falls through in `backend/internal/store/sqlite/ingredient_search_test.go`
- [ ] T003 [P] [US1] Write failing DynamoDB `SearchByIngredients` tests covering: 2-ingredient AND match, no-match empty result, case-insensitivity in `backend/internal/store/dynamo/ingredient_search_test.go`
- [ ] T004 [P] [US1] Write failing handler HTTP tests covering: `?q=gin+and+lemon` routes to `SearchByIngredients` and returns filtered recipes; `?q=gin%2Blemon` (no spaces) same result; `?q=gin++%2B++lemon` (extra whitespace) same result; `?q=gin` stays on single-term path; empty `?q=` returns full list; `?q=gin+%2B+%2B+lemon` (empty middle token) treated as 2-token search; `?q=gin+and+lemon&page=2` returns correct offset (pagination unchanged per FR-007) in `backend/internal/handler/ingredient_search_test.go`
- [ ] T005 [P] [US1] Write failing Vitest test asserting a `<p>` element with text matching `/use.*and.*\+.*ingredient/i` exists in the RecipeList DOM in `frontend/src/pages/RecipeList.ingredient.test.js`

**Checkpoint**: All 4 test files exist and their tests FAIL. Proceed to implementation.

### Implementation for User Story 1

- [ ] T006 [P] [US1] Implement `SearchByIngredients` on SQLite `RecipeStore` in `backend/internal/store/sqlite/recipes.go` using a dynamically-built SQL query with `json_each(r.ingredients)` and one correlated `(SELECT COUNT(*) ... WHERE LOWER(json_extract(value,'$.name')) LIKE ?) > 0` clause per ingredient token (args: `%token%` repeated for COUNT query then SELECT query)
- [ ] T007 [P] [US1] Implement `SearchByIngredients` on DynamoDB `RecipeStore` in `backend/internal/store/dynamo/recipes.go` by adding `matchesAllIngredients(r *model.Recipe, ingredients []string) bool` helper (case-insensitive substring check per token across `r.Ingredients[*].Name`) and wiring it into a full-Scan filter, mirroring the existing `Search` method structure
- [ ] T008 [US1] Implement `parseIngredientQuery(q string) []string` in `backend/internal/handler/recipes.go` (extract to `backend/internal/handler/parse.go` if function body exceeds ~30 lines, per Constitution I ≤ 40 line limit): split on `(?i)\s+and\s+` first, then split each token on `\s*\+\s*`, trim all tokens, discard empty strings, return the resulting slice
- [ ] T009 [US1] Wire `parseIngredientQuery` into the list handler in `backend/internal/handler/recipes.go`: when `len(tokens) >= 2` call `h.recipes.SearchByIngredients(tokens, page, limit)`, when `len(tokens) == 1` call existing `h.recipes.Search(tokens[0], page, limit)`, when `len(tokens) == 0` call `h.recipes.List(page, limit)` (sequential — depends on T006, T007, T008)
- [ ] T010 [P] [US1] Add hint `<p>` element with text `Tip: use "and" or "+" to search multiple ingredients — e.g. "gin and lemon"` and class `text-stone-400 text-sm mt-1` immediately after the `searchWrap.appendChild(bar)` line in `frontend/src/pages/RecipeList.js`

**Checkpoint**: All tests pass. Multi-ingredient search is fully functional end-to-end.

---

## Phase 3: Polish & Validation

**Purpose**: Verify full test suite health and run quickstart scenarios.

- [ ] T011 Run full backend test suite with coverage: `cd backend && go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | grep -E 'store/sqlite|store/dynamo|handler'` — confirm all tests pass (including T002, T003, T004) and coverage ≥ 80% on new files
- [ ] T012 Run full frontend test suite `cd frontend && npm test` and confirm all tests pass including T005
- [ ] T013 Manual quickstart.md validation: run the 8 backend API scenarios and 6 frontend steps from `specs/015-multi-ingredient-search/quickstart.md` against a local dev server; SC-002 (p95 ≤ 200 ms) verified informally via response times observed during this step

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Foundational)**: No dependencies — start immediately
- **Phase 2 (US1)**: Requires T001 complete (interface defined) — then all T002–T005 can run in parallel, followed by T006–T010
- **Phase 3 (Polish)**: Requires all of Phase 2 complete

### Within User Story 1

```
T001 (interface)
  ├── T002 [SQLite tests]  ─┐
  ├── T003 [Dynamo tests]   ├── all parallel, then:
  ├── T004 [handler tests]  │
  └── T005 [frontend tests]─┘
        ↓
  T006 [SQLite impl] ─┐
  T007 [Dynamo impl] ─┤ parallel, then:
  T008 [parser impl]  │
        ↓             │
  T009 [handler wire]─┘  (depends on T006, T007, T008)
  T010 [frontend hint]   (independent of T006–T009)
```

### Parallel Opportunities

- T002, T003, T004, T005 — all test-writing tasks, different files, fully parallel
- T006, T007 — store implementations, different files, fully parallel
- T008, T010 — parser and frontend hint, different files, fully parallel after tests
- T009 — requires T006, T007, T008 complete before wiring

---

## Parallel Example: User Story 1 Tests

```bash
# Write all failing tests simultaneously (different files):
Task A: "Write failing SQLite SearchByIngredients tests in backend/internal/store/sqlite/ingredient_search_test.go"
Task B: "Write failing DynamoDB SearchByIngredients tests in backend/internal/store/dynamo/ingredient_search_test.go"
Task C: "Write failing handler ingredient search tests in backend/internal/handler/ingredient_search_test.go"
Task D: "Write failing Vitest hint text test in frontend/src/pages/RecipeList.ingredient.test.js"

# Then implement simultaneously:
Task E: "Implement SQLite SearchByIngredients in backend/internal/store/sqlite/recipes.go"
Task F: "Implement DynamoDB SearchByIngredients in backend/internal/store/dynamo/recipes.go"
Task G: "Implement parseIngredientQuery in backend/internal/handler/recipes.go"
Task H: "Add hint text to frontend/src/pages/RecipeList.js"
```

---

## Implementation Strategy

### MVP (this entire feature is one user story)

1. Complete Phase 1: T001 (interface) — ~5 min
2. Write all failing tests: T002–T005 in parallel — ~20 min
3. Implement backend: T006, T007, T008 in parallel → T009 — ~30 min
4. Implement frontend: T010 — ~5 min
5. Polish: T011–T013 — ~15 min
6. **Total**: ~75 min end-to-end

### Incremental Checkpoints

- After T001: `go build ./...` must compile with interface error on unimplemented stores
- After T002–T005: `go test ./...` and `npm test` fail on new tests only; existing tests still pass
- After T006–T010: all tests pass
- After T011–T013: ready to open PR

---

## Notes

- `parseIngredientQuery` starts in `backend/internal/handler/recipes.go`; extract to `backend/internal/handler/parse.go` if the body exceeds ~30 lines (Constitution I: ≤ 40 lines per function)
- The DynamoDB `ingredient_search_test.go` uses the existing stub/mock pattern from `backend/internal/store/dynamo/dynamo_test.go` — no live AWS calls
- SQLite tests use `newTestDB()` helper from `backend/internal/store/sqlite/` package — same pattern as existing `recipes_test.go`
- The hint `<p>` element is always visible (not conditional on focus), placed between the SearchBar and SortButtonGroup elements
- No Terraform or infrastructure changes are needed for this feature
