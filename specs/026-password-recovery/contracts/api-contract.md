# API Contract: Password Recovery

**Branch**: `026-password-recovery` | **Date**: 2026-07-09

Two new **public** (no auth) endpoints under `/api/v1/auth`. Each row is an acceptance assertion.

## POST /api/v1/auth/forgot-password

Request: `{ "email": "user@example.com" }`

| # | Precondition | Response | Body | Maps to |
|---|--------------|----------|------|---------|
| F1 | email belongs to an account, under rate limit | `200` | neutral message; a reset email is sent | FR-002/003/004 |
| F2 | email is not registered | `200` | **same** neutral message; no email sent | FR-003, SC-005 |
| F3 | email registered but ≥ 6 requests this hour | `200` | same neutral message; no email sent | FR-012 |
| F4 | malformed body / missing email | `400` | validation error | — |
| F5 | any success case | `200` | body is byte-identical across F1–F3 | SC-005 |

Neutral message (example): `"If an account exists for that email, a password reset link has been sent."`

## POST /api/v1/auth/reset-password

Request: `{ "uid": "<user id>", "token": "<link token>", "password": "<new password>" }`

| # | Precondition | Response | Body | Maps to |
|---|--------------|----------|------|---------|
| R1 | valid uid+token, unexpired, password meets complexity | `200` | success; password updated; `token_version` bumped; token cleared | FR-007/008/009/010 |
| R2 | token/uid invalid, expired, used, or superseded | `400` | **generic** "invalid or expired reset link" (no detail on why) | FR-005/006/011 |
| R3 | password fails a complexity rule | `422` | error naming the unmet rule(s) (min length / upper / lower / number / symbol) | FR-008 |
| R4 | (password confirmation handled on the client) | — | the two-entry match (FR-007) is enforced on the reset page; the API receives one password | FR-007 |
| R5 | after R1 | — | the old password no longer authenticates; pre-reset sessions are rejected | FR-009/010 |

## Cross-cutting

- The raw token is never returned, stored (only its hash), or logged.
- Error bodies for R2 never reveal whether the uid, token, or expiry was the problem.
- Both endpoints are unauthenticated; `reset-password` is authorized solely by the token.
