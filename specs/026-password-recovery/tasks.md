# Tasks: Password Recovery

**Input**: Design documents from `specs/026-password-recovery/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/{api,email}-contract.md ✓, quickstart.md ✓

**Tests**: TDD required (Constitution §II). Test tasks precede implementation. Go `testing` (handlers use a stub `email.Sender`) + Vitest.

**Stack**: Go (Lambda/DynamoDB, bcrypt, JWT), AWS SES v2 SDK (new), vanilla JS + Tailwind + Vitest, Terraform (SES/DKIM/IAM via existing Cloudflare provider).

**⚠ Operational note**: SES starts in **sandbox** (verified recipients only, ~200/day). Requesting SES production access (T028) is a prerequisite for real users receiving email; it does not block building/testing (the stub sender covers tests).

---

## Phase 1: Setup & Infrastructure (SES/IAM — parallel with code; required before real sends)

- [ ] T001 [P] Add an SES domain identity for `cocktails.albertomcastro.com` with DKIM (publish DKIM CNAMEs via the existing Cloudflare provider) in `infra/main.tf`; sender `no-reply@cocktails.albertomcastro.com`.
- [ ] T002 [P] In `infra/main.tf`, grant `ses:SendEmail`/`ses:SendRawEmail` (scoped to the SES identity) to BOTH the production Lambda role and the preview Lambda role (`cocktails-preview-lambda-role`), and add `MAIL_FROM` + `APP_BASE_URL` Lambda env vars; add SES verification/MAIL FROM to `infra/outputs.tf`.
- [ ] T003 Run `terraform fmt`, `terraform validate`, and `terraform plan` from `infra/`; confirm the plan contains only the SES identity/DKIM, the IAM SES permission, env vars, and outputs.

---

## Phase 2: Foundational (shared logic — blocks US1 and US2)

**⚠ §II**: each test task precedes its implementation.

- [X] T004 Add reset + rate-limit fields (`ResetTokenHash`, `ResetTokenExpires`, `ResetWindowStart`, `ResetRequestCount`, all `omitempty`) to `backend/internal/model/model.go` `User`, and map them in `backend/internal/store/dynamo/users.go` (persisted by the existing `Update`).
- [X] T005 [P] Write failing tests in `backend/internal/auth/password_test.go` for `ValidateComplexity`: rejects < 12 bytes, rejects > 72 bytes, rejects missing upper/lower/digit/symbol (one case each), accepts a compliant password, and treats a boundary symbol (e.g. `~` or `_`) as a valid symbol.
- [X] T006 [P] Implement `ValidateComplexity(pw) error` in `backend/internal/auth/password.go` — length 12–72 bytes and ≥ 1 upper, lower, digit, and symbol (symbol = any non-alphanumeric printable ASCII), returning the unmet rule(s) — making T005 pass.
- [X] T007 [P] Write failing tests in `backend/internal/auth/reset_test.go`: `GenerateToken` returns a high-entropy base64url string; `HashToken` is SHA-256; a constant-time verify accepts the matching token and rejects others.
- [X] T008 [P] Implement `GenerateToken()`, `HashToken(token)`, and `VerifyToken(token, hash)` (constant-time) in `backend/internal/auth/reset.go` — making T007 pass.
- [X] T009 [P] Create `backend/internal/email/email.go` (`Sender` interface + `PasswordResetData`) and `backend/internal/email/stub.go` (recording no-op sender) with a stub test — the seam that lets handlers be tested without SES.
- [ ] T010 [P] Create `frontend/src/utils/password.js` (+ `password.test.js`) — a shared JS complexity validator mirroring the backend rules, for live UX feedback.

**Checkpoint**: validator, token helpers, email seam, and model fields exist and are tested.

---

## Phase 3: User Story 1 — Request a reset link by email (Priority: P1) 🎯 MVP (part 1)

**Goal**: A user submits their email and (only if the account exists and is under the rate limit) receives a branded reset email; the on-screen response is always neutral.

**Independent Test**: Submit a registered email → neutral message + email sent (stub records it); submit an unregistered email → identical message, nothing sent; exceed 6/hour → identical message, nothing sent.

- [X] T011 [US1] Write failing handler tests in `backend/internal/handler/password_reset_test.go` for `ForgotPassword` using a stub sender: F1 (registered→token fields set + email recorded), F2 (unknown→neutral, no send), F3 (≥6 requests/hour→neutral, no send), F4 (missing/invalid email→400), F5 (identical body across F1–F3).
- [X] T012 [US1] Implement `ForgotPassword` in `backend/internal/handler/password_reset.go`: `GetByEmail`; fixed-window rate check (6/hour via `ResetWindowStart`/`ResetRequestCount`); on allow, `GenerateToken`+store `HashToken`/expiry via `Update` and call `Sender`; always return the neutral 200 — making T011 pass.
- [X] T013 [US1] Implement the reset email in `backend/internal/email/`: a pure `BuildResetEmail(data) (subject, html, text)` in `email.go` with a test (`email_test.go`) asserting the content contract (single `.../#/reset?uid=&token=` link, 15-minute note, NO credential, brand markers per `contracts/email-contract.md`); and a thin SES v2 `SendEmail` wrapper in `ses.go` that sends the built message from `MAIL_FROM` (link base = `APP_BASE_URL`).
- [X] T014 [US1] Wire the sender (SES in prod from env, stub otherwise) and register `POST /api/v1/auth/forgot-password` in `backend/cmd/lambda/main.go`.
- [ ] T015 [US1] Write failing Vitest in `frontend/src/pages/ForgotPassword.test.js` (submitting an email calls the client and shows the neutral confirmation) and add a `requestPasswordReset` case in `frontend/src/api/client.test.js`.
- [ ] T016 [US1] Implement `frontend/src/pages/ForgotPassword.js` (email form → neutral confirmation), add the "Forgot password?" link to `frontend/src/pages/Login.js`, register route `#/forgot` in `frontend/src/main.js`, and add `requestPasswordReset(email)` to `frontend/src/api/client.js` — making T015 pass.

**Checkpoint**: request flow works end-to-end against the stub; email is enumeration-safe and rate-limited.

---

## Phase 4: User Story 2 — Set a new password via the link (Priority: P1) 🎯 MVP (part 2)

**Goal**: A valid, unexpired link lets the user set a strong new password (entered twice); on success the password changes, sessions are invalidated, and the link is consumed.

**Independent Test**: With a valid token, submit a matching strong password → password updated, old password fails, pre-reset session rejected; submit a weak password → rejected naming the rule.

- [X] T017 [US2] Write failing handler tests in `backend/internal/handler/password_reset_test.go` for `ResetPassword`: R1 (valid uid+token+strong pw → 200, `PasswordHash` changed, `TokenVersion` incremented, token cleared), R3 (weak pw → 422 naming the unmet rule), R5 (after reset the old password no longer verifies and a JWT with the old `token_version` is rejected by `RequireAuthWithStore`).
- [X] T018 [US2] Implement `ResetPassword` in `backend/internal/handler/password_reset.go`: `GetByID(uid)`, verify token (`VerifyToken` + `now < ResetTokenExpires` + non-empty hash), `ValidateComplexity`, set `bcrypt(password)`, `TokenVersion++`, clear reset fields, `Update` — making T017 pass.
- [X] T019 [US2] Register `POST /api/v1/auth/reset-password` in `backend/cmd/lambda/main.go`.
- [ ] T020 [US2] Write failing Vitest in `frontend/src/pages/ResetPassword.test.js` (two password fields; mismatch shows an error and does not submit; live complexity feedback via `utils/password.js`; a matching strong pair submits via the client) and a `resetPassword` case in `frontend/src/api/client.test.js`.
- [ ] T021 [US2] Implement `frontend/src/pages/ResetPassword.js` (parse `uid`/`token` from `location.hash`; two inputs; live complexity checklist; match check; submit → success then redirect to sign-in), register route `#/reset` in `frontend/src/main.js`, and add `resetPassword(uid, token, password)` to `frontend/src/api/client.js` — making T020 pass.

**Checkpoint**: MVP complete — a user can request and complete a reset, and sign in with the new password.

---

## Phase 5: User Story 3 — Expired, used, or invalid links (Priority: P2)

**Goal**: Expired/used/superseded/tampered links show a clear, generic message and a path to request a new one, and never change the password.

**Independent Test**: Try an expired, a used, a superseded, and a tampered token → each returns the generic error and no password change; the reset page shows a friendly "invalid or expired" message with a link to request again.

- [X] T022 [US3] Write failing handler tests in `backend/internal/handler/password_reset_test.go`: expired token, already-used (cleared) token, superseded (overwritten hash) token, and tampered/unknown token all return the **same generic 400** with no password change and no disclosure of the cause (R2).
- [X] T023 [US3] Ensure `ResetPassword` in `backend/internal/handler/password_reset.go` returns the single generic error for every invalid case (add any missing branches) — making T022 pass.
- [ ] T024 [US3] Update `frontend/src/pages/ResetPassword.js` to render a generic "This reset link is invalid or has expired" message with a link to `#/forgot` when the API returns the invalid-link error; assert it in `frontend/src/pages/ResetPassword.test.js`.

**Checkpoint**: the security model (time-limited, single-use, generic errors) is enforced and surfaced clearly.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T025 [P] Accessibility pass on `ForgotPassword.js` and `ResetPassword.js` (labelled inputs, error/success announced, visible focus, keyboard operable) — WCAG 2.1 AA (§III) — and a visual review of the branded email against `contracts/email-contract.md`.
- [ ] T026 [P] Add the timing-neutrality mitigation for `ForgotPassword` (send the email after writing the response, or apply a uniform minimum handler duration) per research Decision 5, in `backend/internal/handler/password_reset.go`.
- [ ] T027 [P] Coverage: `cd backend && go test -p 1 -coverprofile=coverage.out -coverpkg=./internal/... ./...` and `cd frontend && npm test -- --coverage`; confirm ≥ 75% and no regressions.
- [ ] T028 [P] Add a coarse edge/gateway throttle on `POST /api/v1/auth/forgot-password` (CloudFront/API Gateway rate limit in `infra/main.tf`) to mitigate unauthenticated blanket abuse of the email-resolving Scan (M2), independent of the per-user 6/hour limit.
- [ ] T029 Deploy/verify: `terraform apply` the SES/DKIM/IAM/throttle changes, confirm domain + DKIM verified, **request SES production access** (and verify recipient addresses while in sandbox), then run the `quickstart.md` end-to-end flow with a real email.

---

## Dependencies & Execution Order

### Phase order

- **Phase 1 (Infra)** runs in parallel with code but must be applied (T028) before real emails send.
- **Phase 2 (Foundational)** blocks US1 and US2 (validator, token helpers, email seam, model fields).
- **US1 (Phase 3)** and **US2 (Phase 4)** both depend on Phase 2; US2's handler tests (R5) rely on the model `TokenVersion` behavior. They can be built in parallel after Phase 2 (different pages/handlers, but both touch `password_reset.go` and `main.go` — coordinate those two files).
- **US3 (Phase 5)** extends US2's handler and reset page.
- **Polish (Phase 6)** last; T028 needs the infra applied.

### Critical path

```
T001/T002/T003 (infra) ─ parallel ─┐
T004 → T005..T010 (foundational) ──┼─→ US1 (T011→T016) ─┐
                                    └─→ US2 (T017→T021) ─┼→ US3 (T022→T024) → Polish (T025..T028)
```

### Parallel opportunities

- **T001 + T002** (infra) and **T005/T006, T007/T008, T009, T010** (independent foundational files) run in parallel.
- **US1 and US2** are largely parallel after Phase 2 (coordinate the shared `password_reset.go` and `main.go`).
- **T025 + T026 + T027** are independent.

---

## Implementation Strategy

### MVP (US1 + US2)

1. Phase 2 foundational logic (validator, token, email seam, model).
2. US1 (request → email via stub) and US2 (verify → set password, invalidate sessions).
3. **STOP and VALIDATE** with the stub sender + Go/Vitest suites: full request→reset→sign-in works.
4. Apply infra (T028) and swap in the real SES sender for a live end-to-end test.

### Full delivery

Add US3 (generic invalid/expired handling + friendly UI), then Phase 6 (a11y, timing mitigation, coverage, SES production access + live verification).
