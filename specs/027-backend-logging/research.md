# Research: Backend Logging

All spec ambiguities were resolved in `/speckit-clarify` (see spec Clarifications). This document records the technical decisions that turn those requirements into an implementation approach.

## Decision 1 — Logging library: standard `log/slog`

**Decision**: Use Go's standard-library `log/slog` with `slog.NewJSONHandler` writing to `os.Stdout`.

**Rationale**:
- Go 1.25 is in use; `log/slog` is mature, provides native **structured JSON** output (FR-010) and **leveled** logging (FR-002) with zero third-party dependencies (constitution I: minimal surface).
- Lambda forwards stdout/stderr straight to CloudWatch, so a JSON-per-line handler is directly queryable in Logs Insights.
- `slog.LevelVar` allows the active level to be set once at init from the environment and, if ever needed, swapped atomically.

**Alternatives considered**:
- `zerolog` / `zap`: faster in extreme throughput scenarios, but add a dependency for no benefit at this app's traffic; `slog`'s allocation cost is irrelevant here and prod defaults to `warn` (hot path emits nothing).
- Continue with `log.Printf`: rejected — unstructured, unleveled, and the source of the current debugging pain (e.g., the password-recovery outage).

## Decision 2 — Level parsing and safe fallback

**Decision**: Read `LOG_LEVEL` at process init and map case-insensitively: `debug→slog.LevelDebug`, `info→LevelInfo`, `warn`/`warning→LevelWarn`, `error→LevelError`. Missing or unrecognized → **error-only** fallback, and emit one WARN-level line noting the fallback was applied.

**Rationale**: FR-005 requires a safe default that never crashes; error-only is the quietest safe choice. Logging the fallback (at WARN, which error-only suppresses — so actually emit it at ERROR so it is visible even under the fallback) makes misconfiguration diagnosable. *Refinement*: emit the fallback notice at ERROR so it appears even when the fallback level (error) is active.

**Alternatives considered**: Defaulting to `info` on invalid input — rejected as it could silently make production chatty/costly.

## Decision 3 — Request correlation ID source

**Decision**: In middleware, obtain the correlation id from the **API Gateway v2 request context** via `core.GetAPIGatewayV2ContextFromContext(r.Context())` (`github.com/awslabs/aws-lambda-go-api-proxy/core`), using `RequestID`. Fallbacks in order: Lambda `lambdacontext.FromContext` `AwsRequestID`; then a generated random id. Bind it as `rid` on the request-scoped logger (FR-011).

**Rationale**: The adapter (`httpadapter.NewV2(...).ProxyWithContext`) already stores the proxy request context in the Go request context, so the API Gateway request id is retrievable with no new plumbing and correlates with API Gateway access logs. Fallbacks keep it working under the local `cmd/server` path and in tests.

**Alternatives considered**: Generating our own UUID unconditionally — rejected (loses correlation with gateway logs, adds a dependency or hand-rolled id when a platform id already exists). Trusting a client-supplied header — rejected (spoofable, not present).

## Decision 4 — Request-scoped logger via context

**Decision**: A `RequestLogger` middleware builds a child logger `base.With("rid", id, "req", method+" "+path)` and stores it in the request context. Handlers fetch it with `logging.FromContext(ctx)` (which returns the process default logger if none is set, so handlers never nil-panic) and call `log.InfoContext/DebugContext/...` with action fields.

**Rationale**: Threads correlation + request context onto every line a handler emits (FR-011) without changing handler signatures. Falling back to the default logger keeps handlers usable in unit tests that don't install the middleware.

**Alternatives considered**: Passing a logger parameter through every handler/store method — rejected as invasive and churn-heavy. A global logger only — rejected (loses per-request correlation).

## Decision 5 — Severity mapping and action catalog

**Decision**: Encode FR-015 as a fixed mapping applied at each call site: successful writes (login, logout, recipe create/edit/delete, favorite add/remove, password-recovery request/reset) → **INFO**; successful reads/searches (recipe get/list/random, ingredient/base-spirit search, favorite check/list) → **DEBUG**; recoverable anomalies (rate-limit rejection, config fallback, auth rejected) → **WARN**; handled failures (validation, authz, store/dependency errors) → **ERROR**. Each call uses a stable `action` name (dotted, e.g., `recipe.create`) plus safe context fields. The full list lives in `contracts/action-catalog.md`.

**Rationale**: A single documented catalog keeps levels consistent and makes SC-001 (100% action coverage) and acceptance tests concrete.

**Alternatives considered**: Ad-hoc per-handler decisions — rejected (inconsistent levels defeat filtering).

## Decision 6 — Secret redaction strategy

**Decision**: Never pass secret-bearing values into log calls. Log helpers accept only explicit, safe key/values; call sites reference `user_id`/`username`/`recipe_id`/`outcome`/`error`, never `password`, `token`, or reset-token values. Authentication and password-recovery handlers log the event + outcome only. A unit test asserts that entries produced by these handlers contain none of the secret substrings.

**Rationale**: FR-009 forbids secrets at any level; the safest design is to make it impossible by construction rather than filtering after the fact.

**Alternatives considered**: A redaction filter that scrubs known keys — rejected as brittle (depends on remembering to name keys correctly); "don't log secrets" by construction is simpler and testable.

## Decision 7 — Best-effort logging & panic safety

**Decision**: `slog` write errors are not surfaced to handlers (the API returns nothing on log failure), satisfying best-effort (FR-012). Add a `Recover` middleware that converts a panic into one ERROR entry (action + rid + recovered value, no stack leaked to the client) and a generic `500` response.

**Rationale**: Guarantees a logging or unexpected failure never breaks request handling and that uncaught errors are still captured (spec edge case).

**Alternatives considered**: Letting panics propagate to the Lambda runtime — rejected (loses request correlation and returns an opaque platform error).

## Decision 8 — Environment defaults via infra

**Decision**: Set `LOG_LEVEL=warn` on the production Lambda (`infra/main.tf` env block) and `LOG_LEVEL=debug` on preview Lambdas (`.github/scripts/preview-deploy.sh` env block). Both are overridable in the AWS console (FR-004); a console change applies to subsequent invocations because updating a Lambda's environment replaces its execution environment.

**Rationale**: Matches the clarified per-environment defaults (FR-003) and reuses the exact mechanism already used for `MAIL_FROM`/`STRIP_PATH_PREFIX`.

**Alternatives considered**: Hard-coding per-environment levels in code keyed off another variable — rejected (not console-adjustable without redeploy).
