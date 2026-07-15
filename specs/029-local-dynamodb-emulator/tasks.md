---
description: "Task list for Local DynamoDB Emulator (replace SQLite)"
---

# Tasks: Local DynamoDB Emulator (replace SQLite)

**Input**: Design documents from `/specs/029-local-dynamodb-emulator/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/environment.md, quickstart.md

**Tests**: Included. The constitution mandates Test-First (§II), and research D7 requires porting SQLite-backed test coverage onto the DynamoDB store **before** deleting SQLite.

**Organization**: Grouped by user story (US1 P1, US2 P2, US3 P3) so each is an independently testable increment, plus a cross-cutting phase for the FR-012 infra reconciliation and docs.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on incomplete tasks)
- **[Story]**: US1 / US2 / US3 (Setup, Foundational, and Polish tasks carry no story label)
- All paths are repo-relative from `/Users/albertocastro/Code/cocktails`

## Path Conventions

- Web app layout: Go backend under `backend/`, infra under `infra/`, orchestration + docs at repo root.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Local orchestration so the emulator can be started; documented one-command entrypoint (FR-004, SC-001).

- [X] T001 Add `docker-compose.yml` at repo root defining a `dynamodb-local` service (image `amazon/dynamodb-local`, port `8000:8000`, healthcheck) matching the image CI already uses.
- [X] T002 Add a `Makefile` at repo root with a `dev` target (`make dev`) that runs `docker compose up -d dynamodb-local`, then starts `backend/cmd/server` with `DYNAMODB_ENDPOINT=http://localhost:8000`, dummy AWS creds/region, and `JWT_SECRET` set, per `contracts/environment.md`. This is **the** single documented startup command (FR-004/SC-001).

**Checkpoint**: `docker compose up -d dynamodb-local` starts the emulator on `localhost:8000`.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared store-construction + schema-provisioning + test harness that BOTH US1 (server) and US2 (tests) depend on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T003 Create `backend/internal/store/dynamo/client.go` with `NewClient(ctx) (*dynamodb.Client, error)` that honors `DYNAMODB_ENDPOINT` (base-endpoint + static `test`/`test` creds + region when set) and otherwise uses the default AWS config chain, per contracts/environment.md.
- [X] T004 Create `backend/internal/store/dynamo/schema.go` with a `TableNames{Recipes,Users,Favorites}` type (validate all non-empty) and idempotent `EnsureSchema(ctx, client, names)` that creates the 3 tables + GSIs from data-model.md (recipes hash `id`; users hash `id` + `username-index` on `username`; favorites hash `user_id` + range `recipe_id` + `recipe_id-index`), waits for `ACTIVE`, and treats `ResourceInUseException` as success.
- [X] T005 [P] Write `backend/internal/store/dynamo/schema_test.go`: `EnsureSchema` creates the expected table/index shape and is idempotent on a second call (test-first; runs against the emulator).
- [X] T006 Add a shared test helper in `backend/internal/store/dynamo/dynamo_test.go` that requires `DYNAMODB_ENDPOINT` and **fails fast** with an actionable message (replace the current `t.Skip("DYNAMODB_ENDPOINT not set…")`), per FR-009/D4.
- [X] T007 Refactor `backend/internal/store/dynamo/dynamo_test.go` to provision via `EnsureSchema` (remove the ad-hoc `createTable` helper) so all existing DynamoDB tests use the single schema source; keep per-test unique table names + `t.Cleanup` and `go test -p 1`.

**Checkpoint**: `DYNAMODB_ENDPOINT=… go test ./internal/store/dynamo/...` passes using `EnsureSchema`; running it without the endpoint fails with a clear message.

---

## Phase 3: User Story 1 - Run the app locally against a production-like store (Priority: P1) 🎯 MVP

**Goal**: The local backend serves every feature from the emulator through the same store code as production.

**Independent Test**: Start the local env (Phase 1 command), then exercise recipes CRUD/search, login, favorites, and admin user management end-to-end — all data served by the emulator, no SQLite.

- [X] T008 [US1] Rewrite the store wiring in `backend/cmd/server/main.go` to always build DynamoDB stores via `dynamo.NewClient`; remove the `switch STORE_BACKEND` sqlite branch, the `DB_PATH` handling, and the `sqstore` import.
- [X] T009 [US1] In `backend/cmd/server/main.go`, when `DYNAMODB_ENDPOINT` is set, call `EnsureSchema` on startup (with table names from env/local defaults) before `bootstrapAdmin`; on emulator-unreachable, exit with a clear, actionable message (FR-009).
- [X] T010 [US1] Manually validate per quickstart.md: `make dev` (or script) → create a recipe, log in as bootstrapped admin, favorite a recipe, list admin users; confirm all succeed against the emulator (SC-002).

**Checkpoint**: MVP — local app fully functional on the emulator; SQLite no longer used by `cmd/server`.

---

## Phase 4: User Story 2 - Automated tests run against the emulator (Priority: P2)

**Goal**: The test suite exercises the DynamoDB store (same as production) and preserves the behaviors previously proven only by SQLite tests (FR-005, SC-005). Test-first: write these before deleting SQLite in US3.

**Independent Test**: `go test ./...` passes against the emulator; running the store tests without `DYNAMODB_ENDPOINT` fails with a clear message.

- [X] T011 [P] [US2] Port recipe store behaviors into `backend/internal/store/dynamo/recipes_test.go` (or extend existing) covering Create/GetByID/Update/Delete, List (pagination + total), Search, Random, ExistsByName, ListByCreator, ListAll, ImportBatch — mirroring `internal/store/sqlite/recipes_test.go`.
- [X] T012 [P] [US2] Port related-recipe behaviors into `backend/internal/store/dynamo/related_test.go` covering symmetric reconciliation, non-transitivity, dedupe, self-exclusion, and delete cleanup — mirroring `internal/store/sqlite/recipes_related_test.go`.
- [X] T013 [P] [US2] Ensure `backend/internal/store/dynamo/ingredient_search_test.go` and base-spirit search cover the cases from `internal/store/sqlite/ingredient_search_test.go` (SearchByIngredients / SearchByBaseSpirit / combined).
- [X] T014 [P] [US2] Port user store behaviors into `backend/internal/store/dynamo/dynamo_test.go` (or a `users_test.go`) covering Create/GetByID/GetByUsername/GetByEmail/List (excludes admins)/Count/Update/Delete — mirroring `internal/store/sqlite/users_test.go`.
- [X] T015 [P] [US2] Port favorites store behaviors into `backend/internal/store/dynamo/favorites_test.go` covering Add/Remove/Check/List and idempotent re-add — mirroring `internal/store/sqlite/favorites_test.go`.
- [X] T016 [US2] **Coverage gate — blocks T017.** Run `DYNAMODB_ENDPOINT=… go test -p 1 -coverprofile=coverage.out -coverpkg=./internal/... ./...`. Because deleting `internal/store/sqlite` removes both its code and its tests from the denominator, confirm coverage stays ≥ 75% **with `internal/store/sqlite` excluded** from `-coverpkg` (simulating the post-removal state). If below 75%, add DynamoDB store cases (T011–T015) until it passes. Do not start T017 until this gate is green.

**Checkpoint**: Full suite green against the emulator; behavioral coverage no longer depends on SQLite.

---

## Phase 5: User Story 3 - SQLite fully removed (Priority: P3)

**Goal**: No SQLite code, tests, or dependencies remain anywhere (FR-007/FR-008, SC-004/SC-007).

**Independent Test**: `grep -ri sqlite backend/ infra/ *.md go.mod go.sum` returns nothing; app + tests build and run without the SQLite library.

- [X] T017 [US3] **(Blocked by the T016 coverage gate.)** Delete the entire `backend/internal/store/sqlite/` package (implementation + tests): `store.go`, `recipes.go`, `related.go`, `users.go`, `favorites.go` and all `*_test.go`.
- [X] T018 [US3] Remove the SQLite fallback and `sqstore` import from `backend/cmd/lambda/main.go`; require DynamoDB configuration (error clearly if table env vars are unset), keeping production behavior unchanged.
- [X] T019 [US3] Run `cd backend && go mod tidy` to drop `modernc.org/sqlite` and now-unused transitive deps (`modernc.org/libc`, `mathutil`, `memory`, `ncruces/go-strftime`, `mattn/go-isatty`) from `go.mod`/`go.sum`.
- [X] T020 [US3] Verify removal: `grep -ri sqlite backend/ go.mod go.sum` is empty, `go build ./...` and `go vet ./...` pass (SC-004).

**Checkpoint**: Single storage model; leaner dependency graph.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: FR-012 production infra reconciliation, CI, and documentation.

- [X] T021 [P] Reconcile production infra (FR-012): **first** run `aws dynamodb describe-table --table-name cocktails-users --query 'Table.GlobalSecondaryIndexes[].IndexName'` to confirm whether the live table already has `username-index`. Then add the `username-index` GSI (hash `username`, projection ALL) to the `users_table` module in `infra/main.tf`; run `terraform fmt`, `terraform validate`, and `terraform plan`. In the PR, record the **actual** plan result (no change if the index already exists; an additive, non-destructive GSI creation if it doesn't) — do not assert "no change" without checking.
- [X] T022 Verify `.github/workflows/ci.yml` backend job still passes end-to-end with the `dynamodb-local` service and `go test -p 1 … ./...` after SQLite removal; adjust only if the coverage command or service config needs it.
- [X] T023 [P] Update `README.md` / `CONTRIBUTING.md` (and any dev docs) to describe the emulator-based local workflow from quickstart.md and remove all SQLite/`DB_PATH` references (FR-011).
- [X] T024 Final validation against spec Success Criteria: clean-checkout-to-running < 10 min (SC-001), all features work on the emulator (SC-002), suite green (SC-003), zero SQLite references (SC-004), coverage ≥ 75% and behaviors preserved (SC-005), local == CI emulator approach (SC-006), dependency footprint reduced (SC-007).

---

## Dependencies & Execution Order

- **Setup (Phase 1)** → **Foundational (Phase 2)** must complete before any user story.
- **US1 (Phase 3)** depends on Foundational (T003–T004). Delivers the MVP.
- **US2 (Phase 4)** depends on Foundational (T003–T007). Independent of US1's server changes; can proceed in parallel with US1 once Foundational is done.
- **US3 (Phase 5)** depends on **US2 completing first** (coverage must be ported off SQLite before the package is deleted — T011–T016 before T017) and on US1 (T008 removed the `cmd/server` sqlite import; T018 removes the `cmd/lambda` one).
- **Polish (Phase 6)**: T021 (infra GSI) is independent and may run anytime. T022–T024 run last, after US3.

Story completion order: **US1 (MVP) → US2 → US3**, with the FR-012 infra task (T021) schedulable independently.

## Parallel Execution Examples

- **Foundational**: T005 (schema test) can run alongside T006 (test helper) — different concerns.
- **US2 porting**: T011, T012, T013, T014, T015 are all `[P]` — separate test files, no interdependencies; write them concurrently, then T016 runs the aggregate coverage check.
- **Polish**: T021 (infra) and T023 (docs) are `[P]` — different files.

## Implementation Strategy

- **MVP = Phase 1 + Phase 2 + Phase 3 (US1)**: a local app running fully on the emulator. Stop here for a demoable increment.
- **Increment 2 = Phase 4 (US2)**: tests on the emulator with preserved coverage — the safety net before deletion.
- **Increment 3 = Phase 5 (US3) + Phase 6**: delete SQLite, reconcile infra, update docs, final validation.
- Test-First: within Foundational and US2, write the failing tests (T005, T011–T015) before/with the implementation they cover; do not delete SQLite (T017) until US2 coverage is green (T016).
