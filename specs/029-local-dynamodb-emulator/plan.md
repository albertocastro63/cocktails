# Implementation Plan: Local DynamoDB Emulator (replace SQLite)

**Branch**: `029-local-dynamodb-emulator` | **Date**: 2026-07-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/029-local-dynamodb-emulator/spec.md`

## Summary

Collapse the backend to a **single storage model — DynamoDB** — for every environment. Local development and tests point the AWS SDK at a **DynamoDB Local container** via a `DYNAMODB_ENDPOINT` override; production keeps using the managed service through the default AWS config. The SQLite store package, its tests, and the `modernc.org/sqlite` dependency are deleted. Table/index creation is centralized in one idempotent `EnsureSchema` helper (the single source of truth for the local/test schema, matching production keys and GSIs), reused by the local server startup and the store tests. Behaviors previously covered only by SQLite-backed tests are ported onto the DynamoDB store tests so coverage does not regress. CI already runs a `dynamodb-local` service, so this mainly brings local dev in line with what CI and production already do.

## Technical Context

**Language/Version**: Go 1.25 (backend). No frontend changes.  
**Primary Dependencies**: existing only — `aws-sdk-go-v2` (dynamodb, config), `aws-lambda-go`. **Removed**: `modernc.org/sqlite` and its transitive deps (`modernc.org/libc`, `mathutil`, `memory`, `ncruces/go-strftime`, `mattn/go-isatty`).  
**Storage**: Amazon DynamoDB everywhere. Local/test = `amazon/dynamodb-local` container reached via `DYNAMODB_ENDPOINT`; production = managed DynamoDB via default AWS config. Tables: `recipes` (hash `id`), `users` (hash `id` + GSI `username-index` on `username`), `favorites` (hash `user_id` + range `recipe_id` + GSI `recipe_id-index`).  
**Testing**: `go test ./...` with `-p 1` against the emulator; store/integration tests require `DYNAMODB_ENDPOINT` and **fail fast** (no silent skip/fallback) when it is unset. Handler/auth/unit tests keep using in-memory stubs and need no emulator.  
**Target Platform**: AWS Lambda (production); local dev on macOS/Linux with a container runtime (Docker).  
**Project Type**: Web application (Go backend + vanilla-JS SPA) — this feature touches the backend and dev tooling only.  
**Performance Goals**: No change to production budgets (p95 ≤ 200 ms read / ≤ 500 ms write); production already runs on DynamoDB. Local emulator latency is not budget-relevant.  
**Constraints**: Local schema MUST match production key/index shape; provisioning MUST be idempotent; tests MUST be deterministic and isolated; no silent fallback to any non-DynamoDB store.  
**Scale/Scope**: Low-traffic personal app. Change is confined to `cmd/server`, `cmd/lambda`, `internal/store/dynamo`, deletion of `internal/store/sqlite`, `go.mod`/`go.sum`, CI config, new local-orchestration files, and docs.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Single-responsibility, below complexity limits, duplication removed? | ✅ `EnsureSchema` is one focused helper; deleting the SQLite store removes a whole duplicated implementation of every store method (net simplification). |
| II. Test-First | Failing tests before implementation; coverage ≥ 75%? | ✅ Port/write DynamoDB store tests for the behaviors currently proven only by SQLite tests **before** deleting the SQLite package; keep total coverage ≥ 75% (constitution v2.0.0). |
| III. UX Consistency | Design language + loading/empty/error states; WCAG AA? | ✅ N/A to end users — no UI surface changes. Dev-facing failures (emulator unreachable) MUST be human-readable and actionable, consistent with the "no raw stack traces / actionable errors" spirit. |
| IV. Performance | p95 ≤ 200 ms read / ≤ 500 ms write, TTI ≤ 3 s? | ✅ Production path unchanged (already DynamoDB). No new N+1s; handler benchmarks use stubs and are unaffected. |
| Quality Gates | Lint, coverage ≥ 75%, benchmarks pass? | ✅ Watch the coverage denominator: removing the SQLite package drops both its code and its tests — DynamoDB store tests must keep total ≥ 75%. CI `dynamodb-local` service already present. |

> Coverage threshold is **≥ 75%** (constitution v2.0.0), not the 80% in the stock template.

**No constitution violations.** Complexity Tracking section intentionally omitted.

## Project Structure

### Documentation (this feature)

```text
specs/029-local-dynamodb-emulator/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (table/index schema + entities)
├── quickstart.md        # Phase 1 output (local dev workflow)
├── contracts/
│   └── environment.md   # Phase 1 output (env-var + schema contract)
└── tasks.md             # Phase 2 (/speckit-tasks — not created here)
```

### Source Code (repository root)

```text
backend/
├── cmd/server/main.go                    # Remove sqlite branch; always DynamoDB; honor
│                                         #   DYNAMODB_ENDPOINT (+ static creds) for local;
│                                         #   call EnsureSchema when a local endpoint is set
├── cmd/lambda/main.go                    # Remove sqlite fallback + import; require DynamoDB config
├── internal/store/dynamo/
│   ├── client.go        (NEW)            # NewClient(ctx) honoring DYNAMODB_ENDPOINT/creds
│   ├── schema.go        (NEW)            # EnsureSchema(ctx, client, TableNames): idempotent
│   │                                     #   create of the 3 tables + GSIs (single source of truth)
│   ├── schema_test.go   (NEW)            # EnsureSchema is idempotent; creates expected shape
│   ├── dynamo_test.go                    # Use EnsureSchema instead of ad-hoc createTable;
│   │                                     #   require endpoint (fail-fast helper, no t.Skip)
│   └── *_test.go        (ADDED CASES)    # Port behaviors from the deleted sqlite tests
├── internal/store/sqlite/                # DELETE entire package (impl + tests)
├── go.mod / go.sum                       # Drop modernc.org/sqlite + transitive deps
│
docker-compose.yml            (NEW)       # dynamodb-local service for local dev
Makefile  or scripts/dev.sh   (NEW)       # One-command: start emulator + run server (SC-001/FR-004)
.github/workflows/ci.yml                  # Keep dynamodb-local service; ensure ./... passes, -p 1
README.md / CONTRIBUTING.md               # Emulator-based workflow; remove SQLite/DB_PATH refs
infra/main.tf                 (EDIT)      # FR-012: declare the users_table `username-index` GSI
```

**Structure Decision**: Keep the existing web-service layout. The store interfaces (`store.RecipeStore/UserStore/FavoriteStore`) are unchanged; only the concrete backend and its wiring change. Schema knowledge, currently duplicated between the SQLite migrations and the test-only `createTable` helper, is consolidated into `internal/store/dynamo/schema.go` so the local server, the dev provisioning step, and the tests all create tables the same way. Local orchestration lives in a root `docker-compose.yml` plus a `Makefile`/`scripts/dev.sh` entrypoint for the single documented startup command.

## Complexity Tracking

> No constitution violations. Section intentionally empty.
