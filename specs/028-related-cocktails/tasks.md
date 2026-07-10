# Tasks: Related Cocktails

**Feature**: 028-related-cocktails · **Branch**: `028-related-cocktails`
**Inputs**: [plan.md](./plan.md) · [spec.md](./spec.md) · [research.md](./research.md) · [data-model.md](./data-model.md) · [contracts/](./contracts/) · [quickstart.md](./quickstart.md)

**Tech**: Go 1.25 backend (DynamoDB + SQLite), vanilla-JS SPA (Vite + Vitest). No new dependencies. Relations = a `related_ids` set per recipe; symmetry maintained by `SetRelated`; deletion cleans counterparts.

**Conventions**: Test-First (constitution II) — write the failing test, then implement. `[P]` = parallelizable (different files, no incomplete dependency). Coverage gate ≥ 75%.

---

## Phase 1: Setup

- [ ] T001 [P] Add relation fields to `backend/internal/model/model.go`: `RelatedIDs []string` (`json:"related_ids,omitempty"`, `dynamodbav:"related_ids,omitempty"`) on `Recipe`, plus a transient `Related []RelatedRef` (`json:"related,omitempty"`, `dynamodbav:"-"`) and a `RelatedRef{ID, Name string}` type for read enrichment.
- [ ] T002 [P] Add an idempotent migration in `backend/internal/store/sqlite/store.go`: `ALTER TABLE recipes ADD COLUMN related_ids TEXT NOT NULL DEFAULT '[]'` (ignore "duplicate column" errors, matching the existing `garnishes` migration).

---

## Phase 2: Foundational (blocking prerequisites)

**Goal**: The symmetric-relation write primitive every story relies on. Independent of the UI and read paths.

- [ ] T003 Write failing store tests in `backend/internal/store/sqlite/recipes_related_test.go` for `SetRelated`: relating A→B makes B list A (symmetry); duplicate ids collapse; the recipe's own id is dropped (no self); ids that don't resolve to an existing recipe are dropped; re-saving with an id removed strips it from both sides.
- [ ] T004 Add `SetRelated(recipeID string, relatedIDs []string) error` to the `RecipeStore` interface in `backend/internal/store/store.go`, and implement it in `backend/internal/store/sqlite/recipes.go` (read/write the `related_ids` JSON column in the row mapping; normalize → diff current vs requested → write recipe → add/remove `recipeID` on each counterpart). Make T003 pass.
- [ ] T005 Implement `SetRelated` in `backend/internal/store/dynamo/recipes.go` with the same semantics: add `related_ids` to `recipeItem` + `toItem`/`unmarshalRecipe`, normalize, diff, and write the recipe plus each counterpart. Writes are non-transactional (research Decision 2): write the edited recipe first, then each counterpart, and **return any counterpart write error** so the caller (handler) surfaces it for client retry rather than silently leaving a one-sided relation. Mirror the sqlite behavior.

**Checkpoint**: Relations can be set symmetrically at the store layer, fully tested — no HTTP/UI yet.

---

## Phase 3: User Story 1 — Relate cocktails while creating/editing (Priority: P1) 🎯 MVP

**Goal**: An editor adds related cocktails via a keyboard type-ahead in the recipe form; on save they persist and are symmetric.

**Independent test**: In the edit form, search + add a related cocktail, save, then confirm (API/data) that both cocktails list each other.

- [ ] T006 [P] [US1] Write failing handler test in `backend/internal/handler/recipes_related_test.go`: `POST`/`PUT /api/v1/recipes` with `related_ids` persists them and reconciles symmetrically; absent `related_ids` on PUT leaves relations unchanged; self/duplicate/non-existent ids are normalized away; and — per FR-016 — a **non-admin editor of their own recipe A can relate it to a recipe B owned by a different user**, and B's related list is updated with **no ownership check on B** (the permission gate applies only to editing A).
- [ ] T007 [US1] In `backend/internal/handler/recipes.go`, accept `related_ids` (`*[]string`, nil = leave unchanged) in Create and Update; after the recipe upsert, call `store.SetRelated(id, *related_ids)`. Log the action per the existing logging pattern. Make T006 pass.
- [ ] T008 [P] [US1] Add `GET /api/v1/recipes/names` in `backend/internal/handler/recipes.go` (project `ListAll()` → `[{id,name}]`), register the route in `backend/cmd/lambda/main.go` and `backend/cmd/server/main.go`, and add a handler test asserting the minimal `{id,name}` shape.
- [ ] T009 [P] [US1] In `frontend/src/api/client.js` add `getRecipeNames()` (`GET /v1/recipes/names`) and include `related_ids` in the `createRecipe`/`updateRecipe` payloads; extend `frontend/src/api/client.test.js`.
- [ ] T010 [P] [US1] Write failing Vitest `frontend/src/components/RelatedCocktailPicker.test.js`: typing filters names by case-insensitive substring; Arrow Up/Down + Enter add a chip; the current recipe and already-selected are excluded; chips are removable; exposes the selected `related_ids`.
- [ ] T011 [US1] Implement `frontend/src/components/RelatedCocktailPicker.js` as an ARIA combobox (input `role=combobox`, `listbox` options, `aria-activedescendant`, Arrow/Enter/Escape) with removable chips, per `contracts/ui.md`. Make T010 pass.
- [ ] T012 [US1] Integrate the picker into `frontend/src/pages/RecipeForm.js`: load names (via `getRecipeNames`) and derive the initial chips from the recipe's `related_ids` resolved against that names list **client-side** (do NOT depend on the detail-only `related` enrichment from T014, so US1 stays independent of US2), and submit `related_ids` on save; extend `frontend/src/pages/RecipeForm.test.js`.

**Checkpoint**: Editors can curate relations end-to-end; they persist symmetrically. Shippable MVP.

---

## Phase 4: User Story 2 — Discover related cocktails on the recipe page (Priority: P2)

**Goal**: The detail page shows related cocktails (alphabetical links) at the bottom; the home random cocktail does not.

**Independent test**: Seed a recipe with relations → its detail page lists them with working links at the bottom; the home random cocktail shows no related section.

- [ ] T013 [P] [US2] Write failing handler test in `backend/internal/handler/recipes_related_test.go`: `GET /api/v1/recipes/{id}` returns `related: [{id,name}]` sorted alphabetically; `List`/`Random`/`Search` responses do NOT include `related`.
- [ ] T014 [US2] Enrich `GetByID` in `backend/internal/handler/recipes.go` to resolve `related_ids` → `related` (`{id,name}`) sorted case-insensitively by name; leave list/random/search untouched. Make T013 pass.
- [ ] T015 [P] [US2] Write failing Vitest `frontend/src/pages/RecipeDetail.test.js` cases: a recipe with relations renders a "Related cocktails" section at the bottom with `#/recipes/{id}` links in alphabetical order; a recipe with none renders no such section.
- [ ] T016 [US2] Render the "Related cocktails" section at the bottom of `frontend/src/pages/RecipeDetail.js` (links from `recipe.related`, hidden when empty); confirm `frontend/src/pages/Home.js` random cocktail does not render it, and add a `Home.test.js` assertion for FR-011. Make T015 pass.

**Checkpoint**: Relations are discoverable and navigable; never shown on the home random.

---

## Phase 5: User Story 3 — Keep relations correct and tidy (Priority: P3)

**Goal**: Removal is two-sided, relations stay non-transitive, and deleting a cocktail removes it everywhere.

**Independent test**: Remove a relation → gone from both; A–B and B–C ⇒ A not related to C; delete a cocktail → gone from all lists.

- [ ] T017 [P] [US3] Write failing store tests (`backend/internal/store/sqlite/recipes_related_test.go` and dynamo equivalent) for delete cleanup: deleting a recipe removes its id from every counterpart's `related_ids`.
- [ ] T018 [US3] Extend `Delete` in `backend/internal/store/sqlite/recipes.go` and `backend/internal/store/dynamo/recipes.go` to load the recipe, strip its id from each counterpart, then delete. Make T017 pass.
- [ ] T019 [P] [US3] Add non-transitivity + integrity assertions in `backend/internal/handler/recipes_related_test.go`: given A–B and B–C, A's `related` never contains C; and confirm no self/duplicate relation is ever stored.
- [ ] T020 [US3] Verify the remove-relation flow end-to-end in `frontend/src/pages/RecipeForm.test.js`: removing a chip and saving submits the reduced `related_ids` (which the backend reconciles to drop both sides).

**Checkpoint**: All three stories independently testable and complete.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T021 [P] Backend coverage: `cd backend && go test -p 1 -coverprofile=coverage.out -coverpkg=./internal/... ./...`; confirm the new store/handler logic is ≥ 75% and no regressions.
- [ ] T022 [P] Frontend coverage: `cd frontend && npm test -- --coverage`; confirm the picker + detail + form changes are covered and the suite is green.
- [ ] T023 [P] Lint/format: `gofmt -l` (zero output) + `go vet ./...` clean; `npm run lint`/prettier clean on the changed frontend files.
- [ ] T024 [P] Accessibility pass on `RelatedCocktailPicker` per `contracts/ui.md` (combobox/listbox roles, `aria-activedescendant`, full keyboard operation, visible focus, removable chips) — WCAG 2.1 AA.
- [ ] T025 Run the `quickstart.md` verification end-to-end locally (SQLite): SC-001 (symmetry), SC-002 (non-transitive), SC-003 (detail links), SC-004 (not on home random), SC-006 (removal + delete cleanup), SC-007 (no self/dup). Note results.

---

## Dependencies & Execution Order

```text
Setup (T001, T002)
   └─→ Foundational: SetRelated (T003 → T004 → T005)
          ├─→ US1 (T006→T007, T008, T009, T010→T011→T012)   🎯 MVP
          ├─→ US2 (T013→T014, T015→T016)        [needs model field only; seed related_ids]
          └─→ US3 (T017→T018, T019, T020)
                 └─→ Polish (T021–T025)
```

- **Foundational blocks the stories**: `SetRelated` (T004/T005) must exist before handler wiring (US1) and the integrity tests.
- **US2 depends on the model field + read enrichment**, not on US1 authoring — it can be tested by seeding `related_ids` directly.
- **US3 delete cleanup (T018)** depends on the model field; independent of US1/US2.
- **Frontend picker chain** T010 → T011 → T012 is sequential (same component/page); T008 (names endpoint) must land before T012 loads names. T012 resolves initial chips from `related_ids` + names client-side, so it does **not** depend on US2's T014 enrichment (US1 stays fully independent of US2).

## Parallel Execution Examples

- **Setup**: T001 (model) and T002 (sqlite migration) together.
- **US1**: after T007, the backend names endpoint (T008), client changes (T009), and the picker test (T010) can proceed in parallel (different files).
- **US2**: handler test (T013) and detail Vitest (T015) authored in parallel.
- **Polish**: T021, T022, T023, T024 in parallel.

## Implementation Strategy

- **MVP = Setup + Foundational + US1 (T001–T012)**: editors can assign symmetric relations via the type-ahead and they persist. Shippable alone.
- **Increment 2 = US2 (T013–T016)**: the visible discovery payoff on the detail page.
- **Increment 3 = US3 (T017–T020)**: removal, non-transitivity, and delete cleanup hardening.
- **Then Polish (T021–T025)**: coverage, a11y, lint, quickstart.

## Notes

- Keep `SetRelated` a single focused method (normalize → diff → write self → write counterparts) to stay within the complexity limit.
- Non-transitivity needs no code — it is guaranteed by only ever writing requested pairs; T019 asserts it.
- `related` (enriched `{id,name}`) is returned **only** by the single-recipe detail read, never by list/search/random (FR-011).
- **Permissions (FR-016)**: the edit-permission gate applies only to the recipe being edited (A); `SetRelated` writes the symmetric relation onto any counterpart B with no ownership check on B. T006 asserts this cross-ownership case.
- **Non-atomic writes (R1)**: `SetRelated`/`Delete` update multiple items sequentially (not in a transaction). On a mid-operation failure the store returns the error and the handler surfaces it (client can retry); this accepted tradeoff is documented in research Decision 2 — do not silently swallow a counterpart write error.
- **Terminology**: "cocktail" (user-facing / UI copy) and "recipe" (the stored entity and code identifiers) are the same thing; use "Related cocktails" in all UI text and `related_ids`/`recipe` in code, consistent with `contracts/ui.md` and `data-model.md`.
- Deleting a recipe uses its own `related_ids` as the exact set of counterparts to clean (symmetry) — no table scan.
