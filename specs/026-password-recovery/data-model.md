# Data Model: Password Recovery

**Branch**: `026-password-recovery` | **Date**: 2026-07-09

No new table. Reset state is added to the existing **User** item in `cocktails-users`.

## Entity: User (added fields)

| Field | Type | Purpose | Notes |
|-------|------|---------|-------|
| `ResetTokenHash` | string | `SHA-256(token)` of the active reset link | empty when none; raw token never stored |
| `ResetTokenExpires` | epoch seconds (int) | expiry of the active link | issued + 15 min |
| `ResetWindowStart` | epoch seconds (int) | start of the current rate-limit window | rolls when > 1h old |
| `ResetRequestCount` | int | requests made in the current window | capped behavior at ≥ 6 |

Existing fields used by the flow: `ID`, `Email` (`GetByEmail`), `PasswordHash`, `TokenVersion`.

All fields are `omitempty` on the item; existing users without them behave as "no active reset, empty window".

## Reset token lifecycle

```
none ──request──▶ active (hash set, expires = now+15m)
active ──successful reset──▶ consumed (hash cleared)            [single use]
active ──15 min elapse──▶ expired (time check fails)            [expiry]
active ──new request──▶ superseded (hash overwritten)          [single active]
consumed/expired/superseded ──▶ any submitted token fails verification
```

Verification (reset step) passes only when **all** hold: user found by `uid`; `now < ResetTokenExpires`; `constant_time_equal(SHA-256(token), ResetTokenHash)`; `ResetTokenHash` non-empty.

## Rate-limit window (per user, 6/hour)

```
request arrives:
  if now - ResetWindowStart >= 3600:   ResetWindowStart = now; ResetRequestCount = 1  → allow (send)
  elif ResetRequestCount < 6:          ResetRequestCount += 1                          → allow (send)
  else:                                (unchanged)                                     → block (no send)
in every branch the HTTP response is the same neutral 200
```

## Password complexity rule (new-password validation)

Valid iff: length ≥ 12 AND contains ≥ 1 upper-case AND ≥ 1 lower-case AND ≥ 1 digit AND ≥ 1 symbol (common punctuation/special character). No comparison against the previous password (per clarification).

## Account state changed by a successful reset

| Field | Change |
|-------|--------|
| `PasswordHash` | set to `bcrypt(new password)` |
| `TokenVersion` | incremented (invalidates existing sessions) |
| `ResetTokenHash` / `ResetTokenExpires` | cleared (single-use) |
