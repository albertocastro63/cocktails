# Contract: Log Entry (JSON)

Every backend log line is a single JSON object on its own line (JSON Lines), written to stdout and delivered to CloudWatch. This is the contract operators rely on when filtering in Logs Insights.

## Shape

```json
{
  "time": "2026-07-10T14:03:22.481Z",
  "level": "INFO",
  "msg": "recipe created",
  "action": "recipe.create",
  "outcome": "success",
  "rid": "abcd1234-...",
  "req": "POST /api/v1/recipes",
  "user_id": "55ca5fde-cb13-453d-870e-4f66c2b36d34",
  "recipe_id": "a1b2c3d4"
}
```

## Rules

- **R1**: `time`, `level`, `msg` are always present (slog JSON handler).
- **R2**: Action events additionally carry `action` and `outcome`.
- **R3**: Requests handled through the HTTP stack carry `rid` and `req` (set by `RequestLogger` middleware).
- **R4**: `user_id` present whenever the action is authenticated; `recipe_id`/`target_id` present when an action targets a resource.
- **R5**: Read/search events carry `count` (result summary); they MUST NOT include full result payloads.
- **R6**: Failure events carry a sanitized `error` string.
- **R7 (security)**: The following MUST NEVER appear at any level: `password`, any session token/JWT, `reset_token`, or other raw credentials. Verified by test.
- **R8**: `level` ∈ {`DEBUG`,`INFO`,`WARN`,`ERROR`}; severity ordering per data-model.

## Logs Insights example queries

```text
# All lines for one request
fields @timestamp, level, action, outcome, msg
| filter rid = "abcd1234-..."
| sort @timestamp asc

# All failures in the last hour
fields @timestamp, action, error, req, user_id
| filter level = "ERROR"
| sort @timestamp desc

# Every login attempt and its outcome
fields @timestamp, outcome, user_id
| filter action = "auth.login"
```
