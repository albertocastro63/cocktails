# Phase 0 Research: Local DynamoDB Emulator (replace SQLite)

No open `NEEDS CLARIFICATION` items — the spec was explicit and the codebase determined the rest. This records the decisions that shape Phase 1.

## D1 — Emulator: `amazon/dynamodb-local` container

**Decision**: Use the official `amazon/dynamodb-local` image, reached over HTTP on port `8000`.

**Rationale**: CI already runs this exact image as a service (`.github/workflows/ci.yml`), and the existing DynamoDB store tests already talk to it via `DYNAMODB_ENDPOINT`. Adopting it locally makes local == CI == (closely) production with zero new moving parts. It is the closest freely-runnable emulation of DynamoDB semantics (conditional writes, Query/Scan filter ordering, GSIs) — the whole point of the feature.

**Alternatives considered**: LocalStack (heavier, more services than needed); keeping SQLite (rejected — it is the divergence we are removing); DynamoDB Local as a downloadable JAR (rejected — a container is more reproducible and matches CI).

## D2 — SDK endpoint override, not a second store implementation

**Decision**: Keep the single `internal/store/dynamo` implementation. Select local vs. production purely by how the SDK client is built: when `DYNAMODB_ENDPOINT` is set, load config with `config.WithBaseEndpoint(endpoint)` plus static dummy credentials (`AWS_ACCESS_KEY_ID`/`SECRET` = `test`) and a region; otherwise use the default AWS config chain (production).

**Rationale**: Guarantees local and production exercise identical store code (FR-002). The only difference is transport wiring. Centralizing client construction in `NewClient(ctx)` avoids duplicating the endpoint logic across `cmd/server`, `cmd/lambda`, and tests.

**Alternatives considered**: A separate "local" store type (rejected — reintroduces divergence); per-call endpoint overrides (rejected — scattered, error-prone).

## D3 — Centralized, idempotent schema provisioning (`EnsureSchema`)

**Decision**: Add `internal/store/dynamo/schema.go` with an idempotent `EnsureSchema(ctx, client, TableNames)` that creates the three tables and their GSIs if absent (treating `ResourceInUseException` as success) and waits until they are `ACTIVE`. It becomes the single source of truth for the local/test schema. The local server calls it on startup when an endpoint override is present; tests call it in their setup (replacing the ad-hoc `createTable` helper).

**Rationale**: Production tables are created by Terraform (`infra/main.tf`), but local and test environments have no provisioning today except a test-only helper. One shared helper keeps the local schema honest against the shape the code requires, and removes duplicated schema knowledge. Idempotency makes repeated `docker compose up` / repeated test runs safe.

**Schema note**: The `users` table requires a `username-index` GSI (login does a `Query` on it in `dynamo/users.go`), and the test helper already creates it. `EnsureSchema` MUST include it. This GSI is **not** present in `infra/main.tf`'s `users_table` module even though production login depends on it. Per clarification (2026-07-14 → FR-012), this feature **reconciles** the discrepancy: the `username-index` GSI is added to the production Terraform so the declared schema matches what the code requires, and `EnsureSchema` encodes the identical shape for local/test.

**Alternatives considered**: A standalone provisioning script only (rejected as the sole mechanism — auto-provision-on-startup gives the smoothest "just works" dev loop; a script can still wrap it); relying on Terraform locally (rejected — too heavy for an ephemeral emulator).

## D4 — Tests require the emulator and fail fast (no silent skip)

**Decision**: Store/integration tests in `internal/store/dynamo` require `DYNAMODB_ENDPOINT`. Replace the current `t.Skip("DYNAMODB_ENDPOINT not set…")` with a shared helper that **fails** with an actionable message (how to start the emulator) when the endpoint is unset. Handler/auth/other unit tests continue to use in-memory stubs and need no emulator.

**Rationale**: FR-009 forbids silent fallback; a skip is effectively a silent pass that lets storage regressions slip through locally. Failing fast enforces the fidelity goal. CI always sets the endpoint, so CI is unaffected. Scoping the requirement to the store package keeps quick unit-only runs (handlers, auth, model) working without a container.

**Alternatives considered**: Keep skipping when unset (rejected — violates FR-009, hides regressions); require the emulator for the entire suite including stub-based handler tests (rejected — unnecessary friction; those tests don't touch storage).

## D5 — Test isolation strategy

**Decision**: Keep per-test unique table names (uuid suffix) with `t.Cleanup` deleting them, as the existing DynamoDB tests already do, and provision via `EnsureSchema`. Keep `go test -p 1` so packages don't contend on the shared emulator.

**Rationale**: Unique tables give clean isolation and repeatability (FR-006) without cross-test coordination. `-p 1` is already in CI and avoids flakiness from concurrent table churn on a single emulator.

**Alternatives considered**: Shared tables cleared between tests (rejected — ordering/leak risk); truncate-and-reuse (rejected — DynamoDB has no fast truncate; delete/recreate per test is simpler and already proven).

## D6 — Local orchestration: `docker-compose.yml` + one-command entrypoint

**Decision**: Add a root `docker-compose.yml` defining the `dynamodb-local` service, and a `Makefile` target (or `scripts/dev.sh`) that starts the emulator and runs the server with the endpoint set — satisfying "single documented command" (FR-004, SC-001). Auto-provisioning on server startup (D3) means no separate table-creation step is required in the happy path.

**Rationale**: Compose is the conventional, low-friction way to run a local dependency container; a thin Make/script wrapper gives the one-liner and a place to document required env. Matches the project's existing habit of small shell scripts under `.github/scripts`.

**Alternatives considered**: `docker run` documented in README only (rejected — less reproducible, no single command); Testcontainers-go for tests (rejected — adds a dependency and manages container lifecycle inside tests, whereas CI already provides the service and local uses compose).

## D7 — Coverage preservation

**Decision**: Before deleting `internal/store/sqlite`, port every behavior currently proven only by its tests (recipes CRUD/list/search/random/import/exists/list-by-creator, ingredient & base-spirit search, related-recipe symmetric reconciliation & delete cleanup, users CRUD/list/count/get-by-username/email, favorites add/remove/list/check) into `internal/store/dynamo` tests. Verify total coverage stays ≥ 75% (`-coverpkg=./internal/...`).

**Rationale**: Removing the SQLite package removes both code and tests from the coverage calculation; the DynamoDB store must be the one exercised. This is the main correctness risk of the change and is handled test-first (SC-005, FR-005).

**Alternatives considered**: Delete SQLite first and backfill later (rejected — violates Test-First and risks a coverage/behavior gap window).

## D8 — Remove SQLite from both entrypoints and `go.mod`

**Decision**: `cmd/server/main.go` drops the `sqlite` default branch and `DB_PATH`; `cmd/lambda/main.go` drops the sqlite fallback and import (require DynamoDB config, error otherwise). Then `go mod tidy` removes `modernc.org/sqlite` and its now-unused transitive deps.

**Rationale**: FR-007/FR-008/SC-004/SC-007 — no SQLite option, no SQLite dependency. Lambda already always sets `STORE_BACKEND=dynamodb` in production, so the fallback is dead code.

**Alternatives considered**: Keep a hidden SQLite fallback "just in case" (rejected — that is exactly the divergence and false-confidence the feature eliminates).
