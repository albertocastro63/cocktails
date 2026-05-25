# Tasks: Base Spirit Search Filter

**Input**: Design documents from `specs/018-base-spirit-search/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/api.md ✓, quickstart.md ✓

**Deliverables**:
- `backend/internal/store/store.go` (modified — interface extension)
- `backend/internal/store/sqlite/recipes.go` (modified — two new methods)
- `backend/internal/store/sqlite/recipes_test.go` (modified — new tests)
- `backend/internal/store/dynamo/recipes.go` (modified — two new methods)
- `backend/internal/store/dynamo/ingredient_search_test.go` (modified — new tests)
- `backend/internal/handler/recipes.go` (modified — base_spirit dispatch)
- `backend/internal/handler/mock_test.go` (modified — stub gets new methods)
- `backend/internal/handler/recipes_test.go` (modified — new handler tests)
- `frontend/src/pages/RecipeList.js` (modified — parsing helper, hint text)
- `frontend/src/pages/RecipeList.base-spirit.test.js` (new — base-spirit + normalisation tests)

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no shared edit conflicts)
- **[Story]**: Which user story this task maps to
- Constitution requires Test-First — failing tests are written before each implementation task

---

## Phase 1: Setup

**Purpose**: Read starting state of all target files before any edits.

- [ ] T001 Read `backend/internal/store/store.go`, `backend/internal/handler/recipes.go`, `backend/internal/handler/mock_test.go`, and `frontend/src/pages/RecipeList.js` to confirm starting state before any changes

---

## Phase 2: Foundational — Store Interface Extension

**Purpose**: Extend the `RecipeStore` interface with two new methods so all downstream layers can implement and test against the contract. This blocks all other phases — Go will not compile until the stub in `mock_test.go` also satisfies the interface.

- [ ] T002 Extend `backend/internal/store/store.go`: add two method signatures to `RecipeStore` — `SearchByBaseSpirit(baseSpirit string, page, limit int) ([]*model.Recipe, int, error)` and `SearchByBaseSpiritAndIngredients(baseSpirit string, ingredients []string, page, limit int) ([]*model.Recipe, int, error)`

- [ ] T003 Extend `backend/internal/handler/mock_test.go`: add stub implementations of both new `RecipeStore` methods to `stubRecipeStore` so the compile-time interface check (`var _ store.RecipeStore = (*stubRecipeStore)(nil)`) continues to pass. `SearchByBaseSpirit` stub: filter `s.recipes` to recipes that have at least one ingredient where `ing.IsBaseSpirit == true && strings.Contains(strings.ToLower(ing.Name), strings.ToLower(baseSpirit))`; apply pagination same as `SearchByIngredients`. `SearchByBaseSpiritAndIngredients` stub: apply `SearchByBaseSpirit` filter first, then apply `stubMatchesAllIngredients` on the result; apply pagination.

**Checkpoint**: `cd backend && go build ./...` passes with zero errors before proceeding.

---

## Phase 3: User Story 1 — Base Spirit Search Filter (Priority: P1) 🎯 MVP

**Goal**: Users can type `base spirit is <name>` or `base spirit = <name>` in the search field to filter recipes by their flagged base-spirit ingredient. The filter combines with existing ingredient search terms.

**Independent Test**: Type `base spirit is gin` in `#/recipes`; only gin-based recipes appear. Type `absinthe base spirit is rum`; only recipes containing absinthe AND with rum as base spirit appear.

### Tests for User Story 1

> **Write these tests FIRST and confirm they FAIL before implementing**

- [ ] T004 [US1] Write failing tests for `SearchByBaseSpirit` and `SearchByBaseSpiritAndIngredients` in `backend/internal/store/sqlite/recipes_test.go`. Add a `TestSearchByBaseSpirit` suite:
  - Setup: create three recipes — (A) has `{Name:"gin", IsBaseSpirit:true}` + `{Name:"lime juice"}`, (B) has `{Name:"rum", IsBaseSpirit:true}`, (C) has no `IsBaseSpirit:true` ingredient
  - Test: `SearchByBaseSpirit("gin",1,20)` returns only recipe A; total=1
  - Test: `SearchByBaseSpirit("Gin",1,20)` (uppercase) returns recipe A (case-insensitive)
  - Test: `SearchByBaseSpirit("whis",1,20)` returns empty (no base-spirit ingredient contains "whis")
  - Test: `SearchByBaseSpirit("",1,20)` returns all recipes (empty filter falls through to List)
  - Add a `TestSearchByBaseSpiritAndIngredients` suite:
  - Setup: use same recipes A and B; recipe A also has ingredient `{Name:"lime juice"}`
  - Test: `SearchByBaseSpiritAndIngredients("gin", []string{"lime"}, 1, 20)` returns only recipe A
  - Test: `SearchByBaseSpiritAndIngredients("gin", []string{"rum"}, 1, 20)` returns empty (gin base but no rum ingredient)
  - Fixture clarification: recipe B = `{Name:"rum", IsBaseSpirit:true}, {Name:"ginger beer"}`. Test: `SearchByBaseSpiritAndIngredients("rum", []string{"ginger"}, 1, 20)` returns recipe B (rum base + ginger ingredient match). Test: `SearchByBaseSpiritAndIngredients("rum", []string{"lime"}, 1, 20)` returns empty (rum base but no lime ingredient in recipe B).

- [ ] T005 [P] [US1] Write failing tests for `SearchByBaseSpirit` and `SearchByBaseSpiritAndIngredients` in `backend/internal/store/dynamo/ingredient_search_test.go`. Follow the same fixture structure as T004 using the existing DynamoDB test helpers (see other tests in that file for the `newTestStore` / helper pattern). Add equivalent test cases for case-insensitivity, empty base-spirit value, and combined-filter intersection.

- [ ] T006 [US1] Write failing handler tests in `backend/internal/handler/recipes_test.go`. Add a `TestList_BaseSpiritFilter` test group:
  - Test: `GET /api/v1/recipes?base_spirit=gin` — response body `data` contains only the gin-based recipe from the stub store; confirm `SearchByBaseSpirit` is exercised (stub can track calls or you can verify by recipe ID)
  - Test: `GET /api/v1/recipes?q=lime&base_spirit=gin` — calls `SearchByBaseSpiritAndIngredients` path; response contains intersection
  - Test: `GET /api/v1/recipes?base_spirit=` (empty value) — treated as no base-spirit filter; falls through to existing `List`/`Search` dispatch
  - Test: `GET /api/v1/recipes?base_spirit=%20` (whitespace-only) — same as empty; ignored

- [ ] T007 [US1] Create `frontend/src/pages/RecipeList.base-spirit.test.js`. Export and test a `parseBaseSpirit(rawQ)` helper (to be exported from `RecipeList.js`). Write failing tests:
  - Test: `parseBaseSpirit('base spirit is gin')` returns `{ baseSpirit: 'gin', q: '' }`
  - Test: `parseBaseSpirit('base spirit = gin')` returns `{ baseSpirit: 'gin', q: '' }` (equals syntax)
  - Test: `parseBaseSpirit('BASE SPIRIT IS GIN')` returns `{ baseSpirit: 'GIN', q: '' }` (case-insensitive keyword matching)
  - Test: `parseBaseSpirit('absinthe base spirit is rye whiskey')` returns `{ baseSpirit: 'rye whiskey', q: 'absinthe' }`
  - Test: `parseBaseSpirit('base spirit is ')` (trailing space only) returns `{ baseSpirit: '', q: '' }` (empty value → empty baseSpirit)
  - Test: `parseBaseSpirit('base spirit is gin base spirit is rum')` returns `{ baseSpirit: 'gin', q: '' }` (first clause wins)
  - Test: `parseBaseSpirit('martini')` returns `{ baseSpirit: '', q: 'martini' }` (no base-spirit clause → passthrough)
  - Test: `parseBaseSpirit('')` returns `{ baseSpirit: '', q: '' }`
  - Also write an integration test: mock `getRecipes`, render `RecipeList`, type `base spirit is gin` in search field (advance timer 350ms), assert `getRecipes` called with `{ base_spirit: 'gin' }` and without `q` key (or `q: ''` filtered out)

### Implementation for User Story 1

- [ ] T008 [US1] Implement `SearchByBaseSpirit` and `SearchByBaseSpiritAndIngredients` in `backend/internal/store/sqlite/recipes.go`. For `SearchByBaseSpirit`: build a WHERE clause using `EXISTS (SELECT 1 FROM json_each(r.ingredients) WHERE json_extract(value,'$.is_base_spirit') = 1 AND LOWER(json_extract(value,'$.name')) LIKE ?)` with bind arg `'%' + strings.ToLower(baseSpirit) + '%'`; if `baseSpirit` is empty/whitespace fall through to `s.List(page,limit)`. For `SearchByBaseSpiritAndIngredients`: combine the base-spirit EXISTS clause with the per-ingredient `json_each` clauses from `SearchByIngredients` using AND; apply pagination the same way.

- [ ] T009 [P] [US1] Implement `SearchByBaseSpirit` and `SearchByBaseSpiritAndIngredients` in `backend/internal/store/dynamo/recipes.go`. Add helper `matchesBaseSpirit(r *model.Recipe, q string) bool` that checks `ing.IsBaseSpirit && strings.Contains(strings.ToLower(ing.Name), q)` across all ingredients. `SearchByBaseSpirit`: scan all items, filter with `matchesBaseSpirit`, paginate (same pattern as `Search`). `SearchByBaseSpiritAndIngredients`: scan all items, filter with both `matchesBaseSpirit` AND `matchesAllIngredients`, paginate.

- [ ] T010 [US1] Extend `backend/internal/handler/recipes.go`: in the `List` method, after reading `q`, read `baseSpirit := strings.TrimSpace(r.URL.Query().Get("base_spirit"))`. Update the dispatch switch to the five-case table: (no q, no bs) → `List`; (single token q, no bs) → `Search`; (multi-token q, no bs) → `SearchByIngredients`; (any q length ≥ 0, bs present) where no q tokens → `SearchByBaseSpirit`; where q tokens ≥ 1 → `SearchByBaseSpiritAndIngredients`. Note: `baseSpirit` is "present" only if non-empty after trim.

- [ ] T011 [US1] Export `parseBaseSpirit` from `frontend/src/pages/RecipeList.js` and wire it into `onSearch`. Steps: (1) add exported function `export function parseBaseSpirit(rawQ)` that extracts the first `baseSpirit` value and strips ALL base-spirit clauses from the remaining string using two steps: (a) match the first clause with `/base\s+spirit\s+(?:is|=)\s*(.*?)(?:\s+(?=base\s+spirit\s+)|$)/i` to capture `baseSpirit` (trim the capture); (b) strip every base-spirit clause from `rawQ` to produce the cleaned `q`: `rawQ.replace(/base\s+spirit\s+(?:is|=)\s*[^\n]*/gi, '').trim()`; (2) in the `onSearch` callback, call `parseBaseSpirit(q)`, build params object: include `q` only if non-empty, include `base_spirit` only if non-empty; (3) update the hint text paragraph to: `'Tip: search by ingredient (use "and" or "+" for multiple) — or try "base spirit is gin"'`

**Checkpoint**: `cd backend && go test ./...` passes. `cd frontend && npm test` passes for base-spirit test file.

---

## Phase 4: User Story 2 — Whiskey/Whisky Spelling Normalisation (Priority: P1)

**Goal**: The application treats `whiskey` and `whisky` as equivalent in all ingredient searches. A user searching for "rye whisky" finds the same recipes as "rye whiskey", regardless of how the recipe creator spelled the ingredient.

**Independent Test**: Type `rye whisky` in the search field; same recipes appear as when typing `rye whiskey`. Type `base spirit is rye whisky`; same results as `base spirit is rye whiskey`.

### Tests for User Story 2

> **Write these tests FIRST and confirm they FAIL before implementing**

- [ ] T012 [US2] Add whisky-normalisation tests to `frontend/src/pages/RecipeList.base-spirit.test.js`. Import and test a `normaliseWhisky(q)` helper (to be exported from `RecipeList.js`):
  - Test: `normaliseWhisky('rye whisky')` returns `'rye whiskey'`
  - Test: `normaliseWhisky('scotch whisky and lime')` returns `'scotch whiskey and lime'`
  - Test: `normaliseWhisky('Rye Whisky')` returns `'Rye Whiskey'` (preserves case of surrounding words, replaces only the variant spelling)
  - Test: `normaliseWhisky('rye whiskey')` returns `'rye whiskey'` (unchanged — already normalised)
  - Test: `normaliseWhisky('gin')` returns `'gin'` (unaffected — no whisky/whiskey term)
  - Test: `normaliseWhisky('whiskysoaked')` returns `'whiskysoaked'` (word-boundary: does NOT replace when not a standalone word — use `\bwhisky\b` regex)
  - Integration test: mock `getRecipes`, render `RecipeList`, type `rye whisky` (advance timer), assert `getRecipes` called with `{ q: 'rye whiskey' }` (normalised form sent)
  - Integration test: type `base spirit is rye whisky`, assert `getRecipes` called with `{ base_spirit: 'rye whiskey' }` (normalisation applied to base_spirit value too)

### Implementation for User Story 2

- [ ] T013 [US2] Export `normaliseWhisky(q)` from `frontend/src/pages/RecipeList.js` and apply it in `onSearch` before calling `parseBaseSpirit`. Implementation: `export function normaliseWhisky(q) { return q.replace(/\bwhisky\b/gi, m => m.slice(0,-1) + 'ey'); }` — this replaces `whisky` → `whiskey` and `Whisky` → `Whiskey` (preserving capitalisation). In `onSearch`, change the first line to `const normQ = normaliseWhisky(q); const { baseSpirit, q: cleanQ } = parseBaseSpirit(normQ);`.

**Checkpoint**: `cd frontend && npm test -- --coverage` passes with coverage ≥ 75%.

---

## Phase 5: Polish & Verification

**Purpose**: Validate the full implementation end-to-end.

- [ ] T014 Run `cd backend && go test ./... -cover` — all tests pass; coverage ≥ 75% on modified packages

- [ ] T015 [P] Run `cd frontend && npm test -- --coverage` — all tests pass; coverage ≥ 75%

- [ ] T016 [P] Execute quickstart.md Acceptance Test 1 (base-spirit filter alone): navigate to `#/recipes`, type `base spirit is <spirit>` where at least one matching recipe exists; confirm only matching recipes appear. Clear and re-type with `=` syntax; confirm same results.

- [ ] T017 [P] Execute quickstart.md Acceptance Test 3 (whisky normalisation): search `rye whisky` vs `rye whiskey`; confirm identical result sets. Repeat with `base spirit is rye whisky` vs `base spirit is rye whiskey`.

- [ ] T018 [P] Execute quickstart.md Acceptance Test 2 (combined filter): navigate to `#/recipes`, type `absinthe base spirit is rye whiskey`; confirm only recipes containing absinthe AND with rye whiskey as base spirit appear. If no such recipe exists, confirm the empty-state message is shown (not an error).

- [ ] T019 [P] Execute quickstart.md Acceptance Test 4 (hint text): navigate to `#/recipes`, confirm the hint text below the search field contains a reference to `base spirit is` syntax (e.g. includes the phrase "base spirit is gin").

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — BLOCKS all phases (compile error otherwise)
- **Phase 3 (US1)**: Depends on Phase 2
- **Phase 4 (US2)**: Depends on Phase 3 (normalisation applied before `parseBaseSpirit` call; T013 must follow T011)
- **Phase 5 (Polish)**: Depends on Phases 3 and 4

### Within Phase 3 (US1)

- T004, T005, T006, T007 (failing tests) MUST be written and confirmed failing before T008-T011
- T004 and T005 are [P] — touch different store packages
- T008 and T009 are [P] — touch different store packages; depend on T002+T003
- T010 (handler) depends on T008 and T009 completing (so the new methods exist)
- T011 (frontend) is independent of T008-T010; can run in parallel with the backend tasks after T007

### Within Phase 4 (US2)

- T012 (failing tests) MUST be written and confirmed failing before T013
- T013 depends on T011 (the `onSearch` function must exist to be updated)

### Parallel Opportunities

- T004 and T005: parallel (different directories)
- T008 and T009: parallel (different directories)
- T011 (frontend US1 impl) and T008+T009 (backend store impl): parallel
- T014, T015, T016, T017, T018, T019 in Phase 5: all parallel

---

## Implementation Strategy

### MVP (Both User Stories — both are P1)

1. Phase 1: read target files
2. Phase 2: extend interface + stub (T002 → T003 → verify `go build ./...`)
3. Phase 3: write tests (T004, T005, T006, T007) → confirm failing → implement (T008, T009 parallel; T010; T011)
4. Phase 4: write tests (T012) → confirm failing → implement (T013)
5. Phase 5: full test suite + manual acceptance tests

### Parallel Execution (if two agents)

- Agent A: backend store + handler (T004 → T005 → T008 → T009 → T010)
- Agent B: frontend (T007 → T011 → T012 → T013)
- Both must complete before Phase 5

---

## Notes

- The compile-time check in `mock_test.go` (`var _ store.RecipeStore = (*stubRecipeStore)(nil)`) will cause a build error until T003 is complete — prioritise T002+T003 before any other work.
- `parseBaseSpirit` is exported so it can be unit-tested directly without mounting the full `RecipeList` component.
- `normaliseWhisky` uses `\bwhisky\b` (word-boundary regex) to avoid replacing substring occurrences like "whiskysoaked".
- The `onSearch` processing order is: `normaliseWhisky(raw)` → `parseBaseSpirit(normalised)` → `getRecipes({ q, base_spirit })`.
- Recipes with no ingredient flagged `IsBaseSpirit=true` are excluded from base-spirit-filtered results (no special handling needed — the WHERE clause naturally excludes them).
- Empty `base_spirit` parameter is ignored by the handler (whitespace trimmed; dispatch falls through to existing paths).
