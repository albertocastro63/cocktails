# Research: Password Recovery

**Branch**: `026-password-recovery` | **Date**: 2026-07-09

---

## Decision 1 — Store reset state on the user item (no new table/GSI)

**Decision**: Add `ResetTokenHash`, `ResetTokenExpires`, `ResetWindowStart`, `ResetRequestCount` as attributes on the existing `cocktails-users` item. Reuse the existing `GetByEmail` / `GetByID` / `Update` store methods.

**Rationale**: DynamoDB is schemaless, so new attributes need no migration. `GetByEmail` already exists (Scan; emails are unique via the admin `EMAIL_CONFLICT` check) and the users table is tiny, so no email-index GSI is needed. Keeping reset state on the user item gives "single active link per user" for free (a new request overwrites `ResetTokenHash`) and avoids another preview-infra change (no new table to create/seed/tear down).

**Alternatives considered**:
- *Dedicated `password-resets` table keyed by token hash (with TTL)* — cleaner separation and auto-expiry, but adds a table to Terraform + the preview deploy/teardown scripts and a `user_id` GSI for supersede; overkill at this scale.
- *email-index GSI for lookup* — unnecessary; `GetByEmail` already works via Scan on a small table.

---

## Decision 2 — Token: random secret in the link, only its hash stored

**Decision**: Generate a 256-bit cryptographically random token, base64url-encode it into the link, and store only `SHA-256(token)`. Verify with a constant-time comparison. The link also carries the opaque user id (`uid`) so the reset page can be looked up by `GetByID` without a token index.

**Rationale**: Storing only the hash means a database leak does not expose usable reset links. Constant-time compare avoids timing attacks on the token. `uid` in the URL is an opaque UUID (not a secret); the token is the actual authorization.

**Alternatives considered**:
- *Signed/JWT reset token (stateless)* — harder to make single-use and to supersede without server state; also puts more in the URL. A stored hash is simpler and revocable.
- *Token that maps to the user (no uid in URL)* — would need a token→user index; the uid-in-URL approach avoids it.

---

## Decision 3 — SES from the site domain, with DKIM (and the sandbox caveat)

**Decision**: Create an SES **domain identity** for `cocktails.albertomcastro.com` with **DKIM** (CNAME records published via the existing Cloudflare provider); send from `no-reply@cocktails.albertomcastro.com`. Grant the Lambda `ses:SendEmail`/`ses:SendRawEmail` scoped to the identity; pass `MAIL_FROM` and `APP_BASE_URL` as Lambda env.

**Rationale**: "From the site's domain" (user requirement) requires domain verification + DKIM for deliverability/authenticity. Cloudflare is already the DNS provider in Terraform, so the DKIM records fit the existing pattern.

**Operational caveat (⚠)**: New SES accounts start in **sandbox** — they can only send to *verified* recipient addresses and are rate-capped (~200/day). Sending to arbitrary users needs **SES production access** (a support request). Until then, testing (incl. previews) requires verifying recipient addresses. This is the main operational prerequisite; it is not code and is flagged in quickstart.

**Alternatives considered**:
- *Email-address identity only (`no-reply@…` verified)* — simpler but does not authenticate the domain (DKIM) and still hits sandbox limits; domain identity is the right long-term choice.
- *Third-party (SendGrid/Postmark)* — rejected; the user chose SES and the stack is on AWS.

---

## Decision 4 — Rate limit: fixed-window counter, 6/hour/user, on the user item

**Decision**: Track `ResetWindowStart` (epoch) and `ResetRequestCount` on the user item. On each request: if `now − ResetWindowStart < 1h` and `count ≥ 6` → do nothing (still return neutral); else increment, or start a new window when it has elapsed.

**Rationale**: Satisfies "no more than 6 recovery attempts per hour per user" with minimal machinery and no extra store. Over-limit requests remain enumeration-safe (same neutral response, no email). Fixed-window is sufficient for abuse prevention at this scale.

**Alternatives considered**:
- *Sliding-window / token bucket* — smoother but more state; unnecessary for a 6/hour cap.
- *IP-based limiting* — the requirement is per user; unknown-email requests send nothing anyway, so per-user on the resolved account is the correct axis.

---

## Decision 5 — Neutral response and the timing residual

**Decision**: Return an identical 200 neutral body for registered, unregistered, and rate-limited emails. Accept a small residual timing signal (only real, non-limited users trigger an SES call) and mitigate by sending the email *after* writing the response where feasible, or applying a uniform minimum handler duration.

**Rationale**: FR-003/SC-005 require no enumeration. Body-level neutrality is guaranteed; perfect timing neutrality is impractical when work only happens for real users. The mitigation reduces the signal; residual risk is low for this application.

**Alternatives considered**:
- *Always perform equivalent work (e.g., a dummy send)* — wasteful and risky; rejected. The uniform-delay/after-response mitigation is proportionate.

---

## Decision 6 — Reset link + hash routing

**Decision**: The link is `https://cocktails.albertomcastro.com/#/reset?uid=<id>&token=<secret>`. The SPA's hash router adds `#/forgot` and `#/reset` routes; the reset page parses `uid`/`token` from `location.hash`.

**Rationale**: Hash routing means the emailed URL loads `index.html` (200 — unaffected by the feature-023 404 behavior) and the SPA handles the rest; no server-side route is required. Consistent with the existing router.

---

## Resolved unknowns

| Unknown | Resolution |
|---------|-----------|
| Email provider | Amazon SES (user) |
| From address | `no-reply@cocktails.albertomcastro.com` (site domain, DKIM) |
| Rate limit | 6 requests / hour / user (fixed window) |
| Enumeration UX | Neutral response (clarify) |
| New vs. old password | No comparison to previous (clarify) |
| Reset storage | Attributes on the user item; no new table/GSI |
| Session invalidation | `TokenVersion++` (enforced by `RequireAuthWithStore`) |
| SES sandbox | Production-access request is an operational prerequisite |
