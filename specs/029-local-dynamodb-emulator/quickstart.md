# Quickstart: Local Development with the DynamoDB Emulator

Target developer experience after this feature ships. Exact command names (Make target vs. script) are finalized during implementation; the shape is fixed here.

## Prerequisites

- A container runtime (Docker Desktop or compatible) running.
- Go toolchain (per `backend/go.mod`).

## One-command startup (happy path)

```bash
# From repo root — starts the emulator and runs the API against it.
make dev            # or: scripts/dev.sh
```

Behind that command:

1. `docker compose up -d dynamodb-local` — starts `amazon/dynamodb-local` on `localhost:8000`.
2. The server starts with `DYNAMODB_ENDPOINT=http://localhost:8000` (+ dummy AWS creds/region).
3. On startup the server runs `EnsureSchema`, creating the `recipes`, `users`, and `favorites` tables (with the `username-index` and `recipe_id-index` GSIs) if they don't exist.
4. API listens on `:8080`.

Optionally seed an admin: set `ADMIN_BOOTSTRAP_PASSWORD` before starting (unchanged behavior).

## Manual equivalent (understanding the pieces)

```bash
docker compose up -d dynamodb-local

cd backend
DYNAMODB_ENDPOINT=http://localhost:8000 \
AWS_REGION=us-east-1 AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test \
JWT_SECRET=local-dev-secret \
ADMIN_BOOTSTRAP_PASSWORD=admin123 \
go run ./cmd/server
```

Data lives only in the running container — `docker compose down` discards it. That's expected for local dev.

## Running tests

```bash
cd backend
# Store/integration tests require the emulator:
docker compose up -d dynamodb-local
DYNAMODB_ENDPOINT=http://localhost:8000 AWS_REGION=us-east-1 \
AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test \
go test -p 1 ./...
```

- If you run the store tests without `DYNAMODB_ENDPOINT`, they **fail fast** with a message telling you to start the emulator (they do not skip or fall back).
- Handler/auth/model unit tests use in-memory stubs and pass without the emulator.

## Validation checklist (maps to spec Success Criteria)

- [ ] Clean checkout → running app on the emulator in < 10 min (SC-001).
- [ ] Every feature works locally on the emulator: recipes CRUD, search, login, favorites, admin users (SC-002 / User Story 1).
- [ ] `go test ./...` passes against the emulator (SC-003 / User Story 2).
- [ ] `grep -ri sqlite backend/ go.mod go.sum` returns nothing; `modernc.org/sqlite` gone (SC-004 / User Story 3).
- [ ] Behaviors from the old SQLite tests are still verified by DynamoDB store tests (SC-005).
- [ ] Local and CI both use `amazon/dynamodb-local` (SC-006).
- [ ] Emulator-down and missing-endpoint cases fail with clear, actionable messages (Edge cases / FR-009).
