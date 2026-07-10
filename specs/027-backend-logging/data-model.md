# Data Model: Backend Logging

This feature persists nothing. The "model" is the shape of a **log entry** and the **level configuration** that governs emission. Both are contracts consumed by operators (via CloudWatch) rather than stored records.

## Entity: Log Entry

One JSON object per line, emitted to stdout → CloudWatch. Field set:

| Field | Type | Required | Source | Notes |
|-------|------|----------|--------|-------|
| `time` | string (RFC3339) | yes | slog | Emitted automatically by the JSON handler. |
| `level` | string | yes | slog | `DEBUG` \| `INFO` \| `WARN` \| `ERROR`. |
| `msg` | string | yes | call site | Short human-readable summary. |
| `action` | string | yes (action events) | call site | Stable dotted name, e.g. `recipe.create`, `auth.login`, `search.ingredients`. See action catalog. |
| `outcome` | string | yes (action events) | call site | `success` \| `failure`. |
| `rid` | string | yes | middleware | Request-correlation id (API Gateway v2 request id; fallbacks per research Decision 3). |
| `req` | string | yes | middleware | `"<METHOD> <path>"`, e.g. `POST /api/v1/recipes`. |
| `user_id` | string | when authenticated | call site | Acting user's stable id. Never the raw email where an id suffices. |
| `recipe_id` / `target_id` | string | when applicable | call site | Affected resource id. |
| `count` | number | reads/searches | call site | Result-set size summary (FR-014). |
| `error` | string | failures | call site | Sanitized failure reason (no secrets, no stack for the client). |

**Forbidden fields (MUST never appear, any level)**: `password`, `token`, session JWT, `reset_token`, or any raw credential (FR-009).

## Entity: Log Level Setting

The `LOG_LEVEL` environment variable. Governs the minimum severity emitted.

**Ordered levels** (ascending severity): `debug` < `info` < `warn` < `error`. Selecting a level emits that level and all higher ones.

**Parsing** (case-insensitive):

| Input | Active level |
|-------|--------------|
| `debug` | DEBUG (everything) |
| `info` | INFO and above |
| `warn` / `warning` | WARN and above |
| `error` | ERROR only |
| *missing / unrecognized* | **ERROR only** (safe fallback, FR-005) + one ERROR line noting the fallback |

**Per-environment defaults**:

| Environment | Default | Set by |
|-------------|---------|--------|
| Production | `warn` | `infra/main.tf` Lambda env |
| Preview (per-PR) | `debug` | `.github/scripts/preview-deploy.sh` Lambda env |
| Local (`cmd/server`) | `debug` (dev convenience) or unset→error | developer shell |

**Runtime override**: Editing `LOG_LEVEL` in the AWS console applies to subsequent invocations without a redeploy (FR-004); the change replaces the Lambda execution environment, which re-reads the variable at init.

## Severity Mapping (from FR-015)

| Category | Level | Examples |
|----------|-------|----------|
| Successful state change (write) | INFO | `auth.login`, `auth.logout`, `recipe.create`, `recipe.update`, `recipe.delete`, `favorite.add`, `favorite.remove`, `password.reset_request`, `password.reset` |
| Successful read / search | DEBUG | `recipe.get`, `recipe.list`, `recipe.random`, `ingredients.list`, `search.ingredients`, `search.base_spirit`, `favorite.check`, `favorite.list` |
| Recoverable anomaly | WARN | rate-limit rejection, config fallback applied, auth rejected (bad credentials/expired token) |
| Handled failure | ERROR | validation error, authorization failure, store/dependency error, recovered panic |

## State / Lifecycle

No persistent state. A logger instance has a per-process lifecycle (built once at init from `LOG_LEVEL`) and a per-request child (built in middleware, discarded when the request ends).
