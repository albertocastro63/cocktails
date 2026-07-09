# Implementation Plan: Password Recovery

**Branch**: `026-password-recovery` | **Date**: 2026-07-09 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/026-password-recovery/spec.md`

## Summary

Add self-service password reset: a public "Forgot password?" flow that emails a 15-minute, single-use link (via Amazon SES from the site domain) to a set-new-password page enforcing a strong password (12+ chars, upper/lower/number/symbol). The request step is enumeration-safe (always neutral), rate-limited to 6/hour per user, and a successful reset bumps the account's `token_version` to invalidate existing sessions. Reset state lives as new attributes on the existing user record (no new table); `GetByEmail` already exists.

## Technical Context

**Language/Version**: Go 1.22 (backend, Lambda), Node 24 / Vite (frontend), Terraform ≥ 1.10 (infra)  
**Primary Dependencies**: existing — DynamoDB (`UserStore` with `GetByEmail`/`GetByID`/`Update`), `bcrypt`, JWT (`auth.Issue`, `token_version` enforced by `RequireAuthWithStore`); new — AWS **SES v2** SDK for Go; Cloudflare provider (already present) for DKIM DNS  
**Storage**: reset fields added to the existing `cocktails-users` items (schemaless DynamoDB — no new table, no new GSI)  
**Testing**: Go `testing` (validator, token, rate-limit, handlers with a stubbed email sender) + Vitest (forgot/reset pages, client). TDD.  
**Target Platform**: AWS Lambda (arm64) + DynamoDB + SES; SPA on S3/CloudFront  
**Project Type**: Full-stack web (backend + frontend + infra)  
**Performance Goals**: reset endpoints are not hot paths; no impact on p95 read/write budgets  
**Constraints**: security-first — hashed single-use token, constant-time compare, neutral response (no enumeration), rate limit 6/hour/user, session invalidation via `token_version`, no secrets in email or logs; email **from the site domain**  
**Scale/Scope**: small user base; ~2 new endpoints, 2 new frontend pages, 1 email template, one-time SES/IAM Terraform

### Key facts (from the current code)

- `UserStore` already exposes `GetByEmail` (Scan-based; emails are unique — admin enforces `EMAIL_CONFLICT`), `GetByID`, and `Update`. Reset needs no new store methods — only new fields on `model.User` persisted via `Update`.
- `RequireAuthWithStore` rejects a request when `user.TokenVersion != claims.TokenVersion` — so **incrementing `TokenVersion` invalidates existing sessions** (admin password update already does this: `TestUpdateUser_PasswordIncreasesTokenVersion`).
- Only `POST /api/v1/auth/login` exists under `/auth`; the two new endpoints are public (no `RequireAuth`).
- Routing is hash-based; the reset link `https://cocktails.albertomcastro.com/#/reset?uid=…&token=…` loads `index.html` (200) and the SPA parses the params — no server route needed.

## Constitution Check

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Single-responsibility, ≤40 lines, CC ≤ 10, no duplication? | ✅ Split: `password` validator, reset-token helper, rate limiter, `email.Sender` interface, two thin handlers |
| II. Test-First | Failing tests before implementation? | ✅ Rich unit-testable logic — complexity validator, token hash/expiry/single-use, rate-limit window, neutral response, `token_version` bump; handler tests use a stub `email.Sender`. Frontend Vitest for both pages. Tests first, coverage ≥ 75% |
| III. UX Consistency | Design tokens + error/empty/loading states + WCAG 2.1 AA? | ✅ Forgot/reset pages reuse stone/amber tokens and Login's form patterns; labelled inputs, live complexity feedback, clear error/success states; branded HTML email (FR-013) |
| IV. Performance | Budgets met? | ✅ Not hot paths; SES send only for existing, non-rate-limited emails; no read/write-path changes |
| Quality Gates | Lint, coverage ≥ 75%, benchmarks pass? | ✅ Backend + frontend suites extended |

No violations. Security MUSTs are handled explicitly (below) rather than deferred.

## Project Structure

### Documentation (this feature)

```text
specs/026-password-recovery/
├── plan.md
├── research.md          # SES/DKIM, token design, rate limit, timing, sandbox
├── data-model.md        # User reset fields + token/rate-limit lifecycle
├── quickstart.md        # Go/Vitest gates + manual + SES setup
├── contracts/
│   ├── api-contract.md   # forgot-password / reset-password endpoints
│   └── email-contract.md # reset email content + branding
└── tasks.md             # /speckit-tasks output (not created here)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── model/model.go                  # EDIT — add reset + rate-limit fields to User
│   ├── auth/password.go (+_test)       # NEW — ValidateComplexity(pw) (12+/upper/lower/number/symbol)
│   ├── auth/reset.go (+_test)          # NEW — GenerateToken(), HashToken(), constant-time verify
│   ├── email/email.go                  # NEW — Sender interface + PasswordReset template data
│   ├── email/ses.go                    # NEW — SES v2 implementation (from site domain)
│   ├── email/stub.go                   # NEW — no-op/recording sender for tests & local
│   ├── handler/password_reset.go (+_test)  # NEW — ForgotPassword + ResetPassword handlers
│   └── store/dynamo/users.go           # (uses existing GetByEmail/GetByID/Update; maps new fields)
└── cmd/lambda/main.go                  # EDIT — wire email.Sender + register the 2 routes

frontend/src/
├── pages/Login.js                      # EDIT — add "Forgot password?" link
├── pages/ForgotPassword.js (+test)     # NEW — email request → neutral confirmation
├── pages/ResetPassword.js (+test)      # NEW — new password x2 + live complexity, submit
├── api/client.js (+test)               # EDIT — requestPasswordReset(), resetPassword()
├── utils/password.js (+test)           # NEW — shared JS complexity validator (UX mirror)
└── main.js                             # EDIT — routes for #/forgot and #/reset

infra/
├── main.tf                             # EDIT — SES domain identity + DKIM (Cloudflare DNS),
│                                       #        lambda ses:SendEmail IAM, MAIL_FROM/APP_BASE_URL env
└── outputs.tf                          # EDIT — SES verification status / MAIL FROM
```

**Structure Decision**: Full-stack. Reset state is added to the existing user item (no new DynamoDB table or GSI). SES + DKIM + IAM are one-time Terraform additions. The two endpoints are public and live under the existing `/api/v1/auth` group.

## Architecture

### Flow

```
Forgot:  POST /api/v1/auth/forgot-password {email}
  → GetByEmail(email); if missing OR over rate limit → return neutral 200 (no email)
  → else: token = 32 random bytes (base64url); user.ResetTokenHash = sha256(token);
          user.ResetTokenExpires = now+15m; record request in rate-limit window; Update(user)
  → SES send branded email with link .../#/reset?uid={user.ID}&token={token}
  → return the SAME neutral 200 in all cases

Reset:   POST /api/v1/auth/reset-password {uid, token, password}
  → GetByID(uid); constant-time compare sha256(token)==ResetTokenHash AND now<ResetTokenExpires
        → invalid/expired → generic 400 (no detail)
  → ValidateComplexity(password) → 422 with the unmet rule(s) on failure
  → user.PasswordHash = bcrypt(password); user.TokenVersion++ (invalidate sessions);
    clear ResetTokenHash/Expires (single-use); Update(user) → 200 success
```

### Security design (spec §Requirements)

- **Token**: 256-bit random, delivered in the link; only its **SHA-256 hash** is stored (raw token never persisted or logged). Verify with constant-time compare.
- **Single-use / expiry / supersede**: `ResetTokenExpires` (15 min); cleared on success (single-use); a new request overwrites `ResetTokenHash` → prior link stops matching (supersede) — all via the single user item.
- **Neutral response (FR-003, SC-005)**: identical 200 body for found/absent/rate-limited. Residual timing signal (SES latency for real users) noted in research with a mitigation.
- **Rate limit (6/hour/user)**: fixed-window counter on the user item (`ResetWindowStart`, `ResetRequestCount`); over-limit requests still return neutral and send nothing.
- **Session invalidation (FR-010)**: `TokenVersion++` on reset; `RequireAuthWithStore` then rejects old JWTs.
- **Complexity (FR-008)**: backend is authoritative (`auth.ValidateComplexity`); frontend mirrors it for live UX only.

### Email (SES, from the site domain)

- SES **domain identity** for `cocktails.albertomcastro.com` with **DKIM**; DNS (DKIM CNAMEs, optional SPF/DMARC) via the existing Cloudflare provider. From address `no-reply@cocktails.albertomcastro.com`.
- Branded HTML (stone/amber, "Cocktail Recipes" header, a single amber button linking to the reset page) + plain-text alternative. No credentials in the email.
- `email.Sender` interface → SES impl in prod; recording stub in tests/local. Lambda gains `ses:SendEmail`/`SendRawEmail` (scoped to the identity) and `MAIL_FROM` / `APP_BASE_URL` env.
- **⚠ SES sandbox**: a fresh SES account can only send to *verified* addresses and is capped (~200/day). Production access is a support request — an operational prerequisite; in sandbox, testing requires verifying recipient addresses. Flagged in research + quickstart.

## Phase 0: Research

See [research.md](research.md) — SES from-domain + DKIM + sandbox, token design, rate-limit window, reset-data-on-user-record vs. a table, and the timing-enumeration residual. No `NEEDS CLARIFICATION` remain (clarify + this command's input resolved enumeration UX, new-vs-old password, SES, from-domain, and the 6/hour limit).

## Phase 1: Design

- [data-model.md](data-model.md) — new `User` fields and the token/rate-limit lifecycle & states.
- [contracts/api-contract.md](contracts/api-contract.md) — request/response contract for both endpoints (incl. neutral-response and error cases).
- [contracts/email-contract.md](contracts/email-contract.md) — reset email content, link format, and branding.
- [quickstart.md](quickstart.md) — test gates + manual walkthrough + SES/DKIM setup steps.

## Complexity Tracking

| Item | Why | Notes |
|------|-----|-------|
| New external dependency (SES) + DNS/DKIM | Password reset requires outbound email, which the app lacks today | One-time Terraform; the SES **sandbox** production-access request is the main operational gate and is called out in research/quickstart |
| Timing-based enumeration residual | Perfect constant-time neutrality is impractical when only real users trigger an SES send | Documented in research with a mitigation (send after responding / uniform delay); accepted as low risk for this app |
