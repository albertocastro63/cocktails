# Quickstart: Backend Logging

How to exercise and verify the logging feature locally and in AWS. Maps to the spec's Success Criteria (SC-00x).

## Local (fastest inner loop)

The local dev binary (`backend/cmd/server`) uses the same handler stack.

```bash
cd backend
# Verbose: see every action, including reads at DEBUG
LOG_LEVEL=debug STORE_BACKEND=sqlite go run ./cmd/server &
# In another shell, exercise actions:
curl -s localhost:8080/api/v1/recipes >/dev/null           # -> DEBUG recipe.list {count}
curl -s -X POST localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"wrong"}'              # -> WARN auth.login outcome=failure
```

Expected: JSON lines on stdout, one per event, each with `time`, `level`, `msg`, `action`, `rid`, `req`.

Then flip the level and confirm suppression (SC-002):

```bash
# Errors + warnings only: reads/writes at INFO/DEBUG disappear
LOG_LEVEL=warn STORE_BACKEND=sqlite go run ./cmd/server &
curl -s localhost:8080/api/v1/recipes >/dev/null           # -> (no line: DEBUG suppressed)
```

Fallback (SC-006):

```bash
LOG_LEVEL=bogus STORE_BACKEND=sqlite go run ./cmd/server &  # starts fine; ERROR line notes fallback
curl -s localhost:8080/api/v1/recipes >/dev/null            # server still serves normally
```

## Tests

```bash
cd backend
go test ./internal/logging/ ./internal/handler/ -run 'Log|Level|Redact|RequestID|Recover' -v
go test ./... -p 1                                          # full suite, no regressions
```

Key assertions:
- Level parse table incl. missing/invalid → error fallback.
- JSON handler emits expected fields; `action`/`outcome` present on action events.
- **Redaction**: entries from login and password-reset contain no `password`/token/reset-token substrings (SC-004).
- `RequestLogger` sets `rid` + `req`; `FromContext` returns default logger when middleware absent.
- `Recover` turns a panic into one ERROR line + 500 without leaking a stack to the client.

## AWS (preview or prod)

Verify the environment default and console override:

```bash
aws lambda get-function-configuration --function-name cocktails-prod-api \
  --region us-east-1 --query 'Environment.Variables.LOG_LEVEL'          # "warn"

# Exercise a failing request, then find its lines by correlation id in Logs Insights:
#   filter rid = "<request id>" | sort @timestamp asc     (SC-005)
```

Console override (SC-003): set `LOG_LEVEL=debug` on the prod function in the Lambda console, issue a request, and confirm DEBUG lines appear within a minute — then set it back to `warn`.

## Success-criteria checklist

- [ ] SC-001 — each catalog action produces a log entry identifying action/actor/target/outcome.
- [ ] SC-002 — prod (`warn`): success emits nothing, failure emits ERROR.
- [ ] SC-003 — console level change takes effect < 1 min, no redeploy.
- [ ] SC-004 — no secret/token appears at any level (incl. debug).
- [ ] SC-005 — all lines of one failing request found in < 2 min via `rid`.
- [ ] SC-006 — missing/invalid `LOG_LEVEL` still serves requests (error-only fallback).
