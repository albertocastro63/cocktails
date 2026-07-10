# Implementation Plan: Backend Logging

**Branch**: `027-backend-logging` | **Date**: 2026-07-10 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/027-backend-logging/spec.md`

## Summary

Add level-controlled, structured (JSON) logging to the Go backend Lambda so operators can debug from CloudWatch. A single `LOG_LEVEL` environment variable selects the minimum severity emitted (production defaults to `warn`, previews to `debug`), changeable from the AWS console without redeploy. A request-scoped logger — built in middleware from the API Gateway request ID — is threaded through handlers, which emit one structured event per main action with a consistent severity mapping (writes at INFO, reads/searches at DEBUG, recoverable anomalies at WARN, handled failures at ERROR). Secrets (passwords, session/reset tokens) are never logged. Implementation uses Go's standard `log/slog` with a JSON handler to stdout — zero new third-party dependencies.

## Technical Context

**Language/Version**: Go 1.25  
**Primary Dependencies**: stdlib `log/slog` (structured JSON logging, leveling); existing `github.com/awslabs/aws-lambda-go-api-proxy` (`core` package for the API Gateway v2 request ID); `github.com/aws/aws-lambda-go`. No new modules.  
**Storage**: N/A for this feature (DynamoDB unchanged; nothing persisted). Logs go to CloudWatch via Lambda stdout.  
**Testing**: `go test` (table-driven unit tests + `httptest` for middleware); existing Vitest suite untouched.  
**Target Platform**: AWS Lambda `provided.al2023`, arm64; logs in CloudWatch Logs (existing log groups).  
**Project Type**: Web service (Go backend behind API Gateway HTTP API + CloudFront).  
**Performance Goals**: Logging overhead negligible; a suppressed log call (below active level) is a near-free level check. Production default `warn` means routine successful requests emit nothing, preserving the p95 ≤ 200 ms read / ≤ 500 ms write budgets.  
**Constraints**: No secrets/credentials in logs at any level (FR-009); logging best-effort — a logging failure never fails a request (FR-012); level read from `LOG_LEVEL`, safe fallback to error-only on missing/invalid (FR-005).  
**Scale/Scope**: Low-traffic personal app; ~7 handler files, ~12 enumerated actions to instrument. Single Lambda (prod) + per-PR preview Lambdas share the mechanism.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Are functions single-responsibility and below complexity limits? | ✅ Small logging package + thin middleware; per-action calls are one-liners. |
| II. Test-First | Are failing tests written before implementation begins? | ✅ Tests first for level parsing/fallback, JSON shape, secret redaction, and middleware request-ID/context wiring. |
| III. UX Consistency | Do all UI surfaces follow the design language and handle loading/empty/error states? | ✅ N/A — backend-only; no UI change. Reinforces the rule that internal errors/stack traces are never exposed to end users (logs stay server-side). |
| IV. Performance | Do API responses meet p95 ≤ 200 ms read / ≤ 500 ms write and TTI ≤ 3 s? | ✅ slog level check is negligible; prod `warn` default emits nothing on the hot success path. No payload dumping (FR-014 logs summaries only). |
| Quality Gates | Do all CI checks (lint, coverage ≥ 75%, benchmarks) pass? | ✅ New `internal/logging` package is highly unit-testable; targets ≥ 75% per constitution v2.0.0. |

> Coverage threshold is **≥ 75%** (constitution v2.0.0), not the 80% shown in the stock template.

## Project Structure

### Documentation (this feature)

```text
specs/027-backend-logging/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (log entry schema + level model)
├── quickstart.md        # Phase 1 output (verify + operate)
├── contracts/           # Phase 1 output (log JSON contract, action catalog, env config)
└── tasks.md             # Phase 2 output (/speckit-tasks — not created here)
```

### Source Code (repository root)

```text
backend/
├── cmd/lambda/main.go            # Init slog logger from LOG_LEVEL; wrap handler with logging + recover middleware
├── internal/
│   ├── logging/                  # NEW package
│   │   ├── logging.go            #   ParseLevel, New (JSON handler → stdout, LevelVar), context get/set, action helpers
│   │   └── logging_test.go       #   level parse/fallback, JSON fields, redaction, context round-trip
│   └── handler/
│       ├── middleware.go         # NEW RequestLogger (rid+method+path child logger → ctx) and Recover (panic→ERROR, 500)
│       ├── auth.go               # instrument login/logout
│       ├── recipes.go            # instrument create/edit/delete (INFO), get/list/search (DEBUG)
│       ├── favorites.go          # instrument add/remove (INFO), check/list (DEBUG)
│       ├── admin.go              # instrument admin user actions
│       └── password_reset.go     # fold existing ad-hoc log.Printf into slog action events
└── (infra) infra/main.tf         # LOG_LEVEL=warn on prod Lambda env
   .github/scripts/preview-deploy.sh  # LOG_LEVEL=debug on preview Lambda env
```

**Structure Decision**: Reuse the existing `backend/` web-service layout. Add one new standalone package `internal/logging` (no dependency on `handler`, avoiding cycles) that owns logger construction, level parsing, context propagation, and redaction-safe helpers. The `handler` package gains two middlewares (request-scoped logger + panic recovery) and per-action log calls. Infra carries the environment defaults.

## Complexity Tracking

> No constitution violations. Section intentionally empty.
