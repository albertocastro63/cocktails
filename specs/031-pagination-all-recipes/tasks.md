# Tasks: Pagination for All Recipes

**Feature**: `031-pagination-all-recipes` | **Spec**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md)

Test-First is NON-NEGOTIABLE per the project constitution (§II), so each story's
tests come before its implementation and must be confirmed failing first.

**Story legend**: US1 = browse the entire catalog (P1); US2 = see position/total (P2);
US3 = pagination within filters & sort (P2).

---

## Phase 1: Setup

- [x] T001 Confirm local dev loop runs: DynamoDB Local up, `cd backend && go test ./...` and `cd frontend && npm test` both green on a clean `031-pagination-all-recipes` checkout (baseline before changes).

## Phase 2: Foundational (blocking prerequisites)

These fix the storage-level correctness the stories depend on (FR-009) and must land before the story phases.

- [x] T002 Add failing test in `backend/internal/store/dynamo/recipes_test.go` that seeds more recipes than a single Scan batch returns (force small pages via a large item set or injected page limit) and asserts `List` + `ListAll` return every item and the correct `total`.
- [x] T003 Create `scanAll(ctx, client, input)` in `backend/internal/store/dynamo/scan.go` that loops on `LastEvaluatedKey`/`ExclusiveStartKey` until exhausted, accumulating all items.
- [x] T004 Route every single-Scan call site in `backend/internal/store/dynamo/recipes.go` (`List`, `Search`, `Random`, `ExistsByName`, `ListByCreator`, `ListAll`, `SearchByIngredients`, `SearchByBaseSpirit`, `SearchByBaseSpiritAndIngredients`) through `scanAll`, then confirm T002 passes.

**Checkpoint**: Storage returns the full table regardless of Scan batch size; `total` is trustworthy.

---

## Phase 3: User Story 1 — Browse the entire catalog (P1) 🎯 MVP

**Goal**: Every recipe is reachable from the listing via Next/Previous.
**Independent test**: Seed 30 recipes (page size 20); open All Recipes; page 1 shows 20; Next shows the remaining 10; all 30 reachable; Next disabled on last page, Previous disabled on first.

### Tests (write first, confirm failing)

- [x] T005 [P] [US1] Create `frontend/src/components/Pagination.test.js` asserting: renders Prev/Next + "Page X of Y"; Prev disabled on page 1; Next disabled on last page; clicking enabled buttons invokes the callback with the target page.
- [x] T006 [P] [US1] Create `frontend/src/pages/RecipeList.pagination.test.js` asserting: initial load requests page 1; clicking Next fetches the next page and re-renders; clicking Previous goes back; controls hidden when `total <= limit`.

### Implementation

- [x] T007 [US1] Create `frontend/src/components/Pagination.js` — accessible Prev/Next buttons + "Page X of Y" indicator, reusing the button styling conventions from `frontend/src/components/SortButtonGroup.js`; props: `{ page, totalPages, onPage }`; disabled states via `disabled` + `aria-disabled`.
- [x] T008 [US1] In `frontend/src/pages/RecipeList.js`, track `currentPage`, `total`, and `limit`; read `total`/`page`/`limit` from the `getRecipes` response; pass `page` (and `limit`) in the request params.
- [x] T009 [US1] In `frontend/src/pages/RecipeList.js`, render `Pagination` below the grid, wire `onPage` to re-fetch that page, compute `totalPages = max(1, ceil(total/limit))`, and hide the control when `totalPages <= 1`; confirm T005/T006 pass.

**Checkpoint**: US1 delivers the MVP — the original "only 20 shown" bug is fixed.

---

## Phase 4: User Story 2 — See where I am in the catalog (P2)

**Goal**: Show total recipe count and current page position.
**Independent test**: Against a known catalog size, the displayed total matches the true count and the current-page indicator updates as pages change.

### Tests (write first, confirm failing)

- [x] T010 [P] [US2] Extend `frontend/src/pages/RecipeList.pagination.test.js` to assert a total-count label (e.g., "30 recipes") renders and matches the response `total`, and that the "Page X of Y" indicator updates on navigation.

### Implementation

- [x] T011 [US2] In `frontend/src/pages/RecipeList.js`, render a total-results label driven by `total`, updating on every fetch; ensure the `Pagination` "Page X of Y" reflects `currentPage`/`totalPages`; confirm T010 passes.

**Checkpoint**: Users can orient themselves; US1 + US2 is a complete browsing experience.

---

## Phase 5: User Story 3 — Pagination within filters & sort (P2)

**Goal**: Pagination composes with search, ingredient, base-spirit filters and global sort; totals reflect the filtered set; changing a filter/sort resets to page 1; order is consistent across pages.
**Independent test**: Apply a filter matching >1 page; total reflects matches; page through all matches; apply A→Z and confirm last name on page 1 sorts before first name on page 2; changing the query returns to page 1.

### Tests (write first, confirm failing)

- [x] T012 [P] [US3] Extend `backend/internal/handler/recipes_test.go`: `sort=name_asc`/`name_desc` orders the whole result set before paging (last item of page 1 precedes first item of page 2); unknown `sort` degrades to default order (no error); out-of-range `page` returns 200 with empty `data` and correct `total`; `sort` combined with `q`/`base_spirit` reflects filtered totals.
- [x] T013 [P] [US3] Extend `backend/internal/store/dynamo/recipes_test.go`: store applies `name_asc`/`name_desc` stable sort over the full set across pages for `List` and the search variants.
- [x] T014 [P] [US3] Extend `frontend/src/pages/RecipeList.pagination.test.js`: changing search/filter resets to page 1; changing sort resets to page 1 and sends `sort` to the API; total/controls reflect the filtered result set.

### Implementation

- [x] T015 [US3] Add an in-memory stable sort helper in `backend/internal/store/dynamo/recipes.go` (case-insensitive by name, asc/desc) applied to the full scanned set before the `page`/`limit` slice, in `List`, `Search`, `SearchByIngredients`, `SearchByBaseSpirit`, and `SearchByBaseSpiritAndIngredients`.
- [x] T016 [US3] In `backend/internal/handler/recipes.go`, parse and validate the `sort` query param (`name_asc`/`name_desc`, else default), thread it into the store calls; confirm T012/T013 pass.
- [x] T017 [US3] In `frontend/src/pages/RecipeList.js`, replace client-side `sortRecipes` usage for display with a server `sort` param; on sort change, reset to page 1 and re-fetch; on search/filter change, reset to page 1 (already partly present — verify and fix); update `frontend/src/api/client.js` only if `getRecipes` needs to forward `sort` (it already forwards arbitrary params — verify). Confirm T014 passes.
- [x] T018 [US3] Remove now-dead client-side sort logic in `frontend/src/pages/RecipeList.js` if fully superseded (keep `sortRecipes` export only if other callers/tests depend on it; otherwise delete and update `frontend/src/pages/RecipeList.test.js`).

**Checkpoint**: All three stories complete; filtered/sorted browsing is fully paginated.

---

## Phase 6: Polish & Cross-Cutting

- [x] T019 [P] Verify loading/empty/error states still render correctly with pagination (empty catalog shows empty state and no controls per FR-010); add/adjust assertions in `frontend/src/pages/RecipeList.pagination.test.js`.
- [x] T020 [P] Confirm coverage ≥ 75% and linters/formatters pass: `cd backend && go vet ./... && go test ./...`; `cd frontend && npm run lint && npm test`.
- [x] T021 Run the manual verification steps in `specs/031-pagination-all-recipes/quickstart.md` against the running app and confirm SC-001..SC-005.

---

## Dependencies & Execution Order

- **Phase 1 (Setup)** → **Phase 2 (Foundational, T002–T004)** must complete before any story phase (stories rely on correct `total`).
- **US1 (Phase 3)** is the MVP and has no dependency on US2/US3.
- **US2 (Phase 4)** depends on US1 (reuses the `Pagination` component and page state).
- **US3 (Phase 5)** depends on US1 for the frontend paging wiring; its backend tasks (T012, T013, T015, T016) depend only on Phase 2 and may proceed in parallel with US1 frontend work.
- **Phase 6** last.

### Parallel opportunities

- T005 and T006 (US1 tests) run in parallel — different files.
- T012, T013, T014 (US3 tests) run in parallel — different files.
- Backend US3 (T015–T016) can proceed in parallel with frontend US1 (T007–T009) once Phase 2 is done.
- T019 and T020 run in parallel.

## Implementation Strategy

- **MVP**: Phase 1 + Phase 2 + Phase 3 (US1). This alone fixes the reported bug (all 30 recipes reachable).
- **Increment 2**: US2 (position/total labels).
- **Increment 3**: US3 (server-side sort + filtered pagination), which also retires the per-page client-side sort.
