# Tasks: Backend Logging

**Feature**: 027-backend-logging · **Branch**: `027-backend-logging`
**Inputs**: [plan.md](./plan.md) · [spec.md](./spec.md) · [research.md](./research.md) · [data-model.md](./data-model.md) · [contracts/](./contracts/) · [quickstart.md](./quickstart.md)

**Tech**: Go 1.25, stdlib `log/slog` (JSON handler → stdout → CloudWatch), `aws-lambda-go-api-proxy/core` for the API Gateway v2 request id. No new dependencies.

**Conventions**: Test-First (constitution II) — write the failing test, then implement. `[P]` = parallelizable (different files, no incomplete dependency). Coverage gate ≥ 75%.

---

## Phase 1: Setup

- [X] T001 [P] Create the `backend/internal/logging/` package directory with a `doc.go` (package comment describing the structured-logging contract and linking the action catalog).
- [X] T002 [P] Add `LOG_LEVEL = "warn"` to the production Lambda `environment_variables` block in `infra/main.tf`; run `terraform fmt` + `terraform validate`.
- [X] T003 [P] Add `LOG_LEVEL=debug` to the preview Lambda env in `.github/scripts/preview-deploy.sh` (append to the `Variables={...}` set used by both the create-function and update-function-configuration paths); `bash -n` the script.

---

## Phase 2: Foundational (blocking prerequisites)

**Goal**: The logger plumbing every story relies on — JSON logger construction, request-scoped context propagation, and the request/recover middleware — independent of level parsing (which is US1).

- [X] T004 Write failing tests in `backend/internal/logging/logging_test.go` for: `New(level)` returns a slog logger whose JSON output includes `time`/`level`/`msg`; `IntoContext`/`FromContext` round-trip; and `FromContext` returns the package default logger (never nil) when none was set.
- [X] T005 Implement `backend/internal/logging/logging.go`: `New(level slog.Leveler) *slog.Logger` using `slog.NewJSONHandler(os.Stdout, ...)`; a package-level default logger + `SetDefault`; and `IntoContext(ctx, *slog.Logger)` / `FromContext(ctx) *slog.Logger` (default-logger fallback). Make T004 pass.
- [X] T006 Write failing tests in `backend/internal/handler/middleware_test.go` for `RequestLogger` (child logger carries `rid` + `req="<METHOD> <path>"`, retrievable via `logging.FromContext`) and `Recover` (a panicking handler yields one ERROR entry and a 500 response with no stack in the body). Use `httptest` and a buffer-backed logger.
- [X] T007 Implement `RequestLogger` and `Recover` middleware in `backend/internal/handler/middleware.go`: `RequestLogger` derives the correlation id from `core.GetAPIGatewayV2ContextFromContext` → Lambda `lambdacontext` → generated fallback (research Decision 3), builds `base.With("rid", id, "req", method+" "+path)`, stores it via `logging.IntoContext`; `Recover` logs the recovered panic at ERROR and writes a generic 500. Make T006 pass.

**Checkpoint**: Logger + context + middleware exist and are tested, but no level control or action logs yet.

---

## Phase 3: User Story 1 — Control log verbosity per environment (Priority: P1) 🎯 MVP

**Goal**: A single `LOG_LEVEL` selects the minimum severity; production defaults to `warn`, preview to `debug`; missing/invalid falls back to error-only; console changes take effect without redeploy.

**Independent test**: Set `LOG_LEVEL=warn`, hit an endpoint → no INFO/DEBUG lines; a failure → ERROR line. Set `LOG_LEVEL=debug` → same success now emits DEBUG. Set `LOG_LEVEL=bogus` → server still serves; ERROR notes the fallback.

- [X] T008 [P] [US1] Write failing table-driven tests in `backend/internal/logging/logging_test.go` for `ParseLevel`: `debug/info/warn/warning/error` (case-insensitive) map to the right `slog.Level`; empty and unrecognized inputs return the error-only fallback and signal that a fallback was applied.
- [X] T009 [US1] Implement `ParseLevel(string) (slog.Level, bool)` and `LevelFromEnv() slog.Level` (reads `LOG_LEVEL`; on missing/invalid returns `LevelError` and emits one ERROR line noting the fallback) in `backend/internal/logging/logging.go`. Make T008 pass.
- [X] T010 [US1] Wire startup in `backend/cmd/lambda/main.go`: build the logger via `logging.New(logging.LevelFromEnv())`, `logging.SetDefault(...)`, and wrap the handler chain with `handler.RequestLogger` + `handler.Recover` (outermost). Mirror the same init in `backend/cmd/server/main.go`.
- [X] T011 [US1] Add an integration-style test in `backend/internal/handler/middleware_test.go` (or a new `logging_integration_test.go`) driving a request through `New(LevelWarn)` vs `New(LevelDebug)` and asserting suppression vs emission of a DEBUG line (SC-002 behavior at the handler layer).

**Checkpoint**: Level control fully works and is env-driven. Infra defaults from T002/T003 apply. This is a shippable MVP.

---

## Phase 4: User Story 2 — All main backend actions are logged (Priority: P2)

**Goal**: Every action in `contracts/action-catalog.md` emits a log entry at its mapped level (writes INFO, reads/searches DEBUG, anomalies WARN, failures ERROR).

**Independent test**: With `LOG_LEVEL=debug`, exercise each action and confirm a matching entry with `action`, `outcome`, actor, and target.

- [X] T012 [US2] **(Test-first)** Write failing action-emission tests in `backend/internal/handler/{auth,recipes,favorites,admin,password_reset}_test.go` (buffer-backed logger installed via context) asserting a representative action per handler group emits the expected `action` + `outcome` + level per `contracts/action-catalog.md` (covers SC-001). These MUST be confirmed failing before T013–T017.
- [X] T013 [P] [US2] Instrument authentication in `backend/internal/handler/auth.go`: `auth.login` at INFO on success, WARN on rejected credentials/expired token, ERROR on unexpected failure. Handle `auth.logout` per the catalog note — instrument an existing server-side session event (e.g., token-version bump) or record it N/A in this task's completion note; do **not** invent an endpoint. Makes the auth portion of T012 pass.
- [X] T014 [P] [US2] Instrument `backend/internal/handler/recipes.go`: `recipe.create`/`update`/`delete` at INFO (`user_id`, `recipe_id`); `recipe.get`/`list`/`random`, `ingredients.list`, `search.ingredients`, `search.base_spirit` at DEBUG (`count`, query params); failures at ERROR. Makes the recipes portion of T012 pass.
- [X] T015 [P] [US2] Instrument `backend/internal/handler/favorites.go`: `favorite.add`/`remove` at INFO; `favorite.check`/`list` at DEBUG; failures at ERROR. Makes the favorites portion of T012 pass.
- [X] T016 [P] [US2] Instrument `backend/internal/handler/admin.go`: `admin.user.create`/`update`/`delete` at INFO (actor `user_id`, `target_id`); `admin.user.list`/`get` at DEBUG; failures at ERROR. Makes the admin portion of T012 pass.
- [X] T017 [P] [US2] Refactor `backend/internal/handler/password_reset.go`: replace the existing ad-hoc `log.Printf` calls with `password.reset_request` / `password.reset` slog action events (INFO success, ERROR failure) using `logging.FromContext`. Makes the password-reset portion of T012 pass.

**Checkpoint**: Action-emission tests were written and confirmed failing first (T012), then satisfied by T013–T017. All catalog actions logged and verified. US1 + US2 together deliver the core value.

---

## Phase 5: User Story 3 — Logs are consistent, searchable, and safe (Priority: P3)

**Goal**: Consistent fields, per-request correlation grouping, and guaranteed absence of secrets.

**Independent test**: Trigger varied actions/failures; confirm filter-by-level and filter-by-`rid` grouping work, and no secret/token appears at any level.

- [X] T018 [P] [US3] Add a redaction test in `backend/internal/handler/password_reset_test.go` and `auth_test.go`: capture entries from login and password-reset flows and assert the serialized JSON contains none of `password`, the JWT, or `reset_token` values at DEBUG level (SC-004).
- [X] T019 [P] [US3] Add a correlation test (`backend/internal/handler/middleware_test.go`): drive one request that produces multiple log lines and assert they all share the same `rid` and carry `req` (SC-005).
- [X] T020 [US3] Consistency sweep: verify every instrumented call site uses the canonical `action` names and safe field keys from `contracts/action-catalog.md` (no `password`/token keys anywhere); fix any drift. Cross-check the field set against `contracts/log-entry.md`.

**Checkpoint**: All three user stories independently testable and complete.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T021 [P] Coverage: `cd backend && go test -p 1 -coverprofile=coverage.out -coverpkg=./internal/... ./...`; confirm `internal/logging` and the new middleware are ≥ 75% and no regressions.
- [X] T022 [P] Lint/format/vet: `gofmt -l` (zero output), `go vet ./...` clean (constitution I: zero warnings).
- [X] T023 [P] Document `LOG_LEVEL` (values, per-environment defaults, how to change it in the AWS console) in `README.md` under backend/operations.
- [X] T024 Run the `quickstart.md` verification checklist end-to-end locally (`cmd/server`, sqlite) for SC-001, SC-002, SC-006; note results.
- [ ] T025 Deploy/verify (after merge): confirm `LOG_LEVEL` on prod (`warn`) and a preview (`debug`) via `aws lambda get-function-configuration`; exercise one failing request and locate its lines by `rid` in Logs Insights (SC-005); toggle the level in the console and confirm effect < 1 min (SC-003), then restore.

---

## Dependencies & Execution Order

```text
Setup (T001–T003)  ─┐
                    ├─→ Foundational (T004→T005, T006→T007)
                    │        │
                    │        └─→ US1 (T008→T009→T010→T011)  🎯 MVP
                    │                 │
                    │                 └─→ US2 (T012 tests → T013–T017 [P])
                    │                          │
                    │                          └─→ US3 (T018,T019 [P] → T020)
                    │                                   │
                    └───────────────────────────────────→ Polish (T021–T025)
```

- **Foundational blocks everything**: T005 and T007 must exist before any handler can log or any level test runs.
- **US1 before US2/US3**: action logs and correlation assertions need the logger installed at startup (T010).
- **T012 (US2 tests) first**, then **T013–T017 run in parallel** (different handler files, each making its portion of T012 pass). **T018/T019** parallel (different test files).
- **Infra (T002/T003)** is independent and can be done anytime before deploy (T025).

## Parallel Execution Examples

- **Setup**: T001, T002, T003 together (package stub, terraform, preview script — all different files).
- **US2 instrumentation** (after the failing tests in T012): T013 (auth), T014 (recipes), T015 (favorites), T016 (admin), T017 (password_reset) in parallel — one handler file each.
- **US3 tests**: T018 (redaction) and T019 (correlation) in parallel.
- **Polish**: T021, T022, T023 in parallel.

## Implementation Strategy

- **MVP = Phase 1 + 2 + US1 (T001–T011)**: level-controlled structured logging with correlation, env-driven defaults, and safe fallback. Shippable on its own.
- **Increment 2 = US2 (T012–T017)**: full action coverage — the day-to-day debugging value.
- **Increment 3 = US3 (T018–T020)**: consistency + safety hardening.
- **Then Polish (T021–T025)**: coverage, lint, docs, and post-deploy verification.

## Notes

- `auth.logout`: backend is stateless-JWT; confirm during T013 whether a server-side session event exists to log, else mark N/A (do not add an endpoint just to log). Reconciled in spec FR-006.
- Keep action names stable once shipped — they are the filtering contract (`contracts/action-catalog.md`).
- Best-effort: never let a logging call change a handler's control flow or error path (FR-012).
