# Quickstart: Password Recovery

**Branch**: `026-password-recovery` | **Date**: 2026-07-09

## Automated (primary gate)

```bash
# Backend — validator, token, rate-limit, handlers (stub email sender)
cd backend && go test ./internal/auth/... ./internal/handler/... ./internal/email/...
go test -p 1 -coverprofile=coverage.out -coverpkg=./internal/... ./... && go tool cover -func=coverage.out

# Frontend — forgot/reset pages, shared validator, client
cd ../frontend && npm test -- src/pages/ForgotPassword.test.js src/pages/ResetPassword.test.js src/utils/password.test.js src/api/client.test.js
npm test -- --coverage
```

Covers: F1–F5 and R1–R5 (via handler tests with a recording stub sender), complexity rule, token hash/expiry/single-use/supersede, rate-limit window, `token_version` bump, and the two page flows (neutral confirmation; live complexity + submit).

## One-time infrastructure (SES + DKIM + IAM)

⚠ **SES sandbox**: a new SES account can only send to *verified* addresses (~200/day cap). Request **SES production access** before real users can receive email; until then verify test recipients.

1. `terraform plan` / `apply` in `infra/` to create the SES domain identity + DKIM (Cloudflare DNS records), the Lambda `ses:SendEmail` permission, and the `MAIL_FROM` / `APP_BASE_URL` env.
2. Confirm the SES domain identity shows **verified** and DKIM **success** (allow for DNS propagation).
3. (Sandbox only) Verify the recipient email(s) you'll test with.

## Manual walkthrough

1. On the sign-in page, click **Forgot password?** → enter a registered email → confirm the neutral message; enter an unregistered email → confirm the **identical** message.
2. Open the reset email → confirm it's branded, states the 15-minute expiry, and contains only the link (no password).
3. Open the link → set a new password twice:
   - mismatch → clear error, not accepted;
   - weak (e.g., `abcdefghijkl`) → rejected naming the unmet rule(s);
   - strong matching pair → success.
4. Sign in with the **new** password (works) and the **old** one (fails).
5. Confirm a session that was signed in before the reset is now signed out (protected action rejected).
6. Open the used link again → generic "invalid or expired" message.
7. Wait > 15 min on a fresh link (or force expiry) → generic expired message.
8. Request 7 resets within an hour for one account → the 7th sends no email (still neutral response).
