# Research: PR Preview Environments

**Branch**: `021-pr-preview-environments` | **Date**: 2026-05-28

---

## Decision 1 — Frontend Path Prefix for API Calls

**Decision**: Introduce a `VITE_API_PATH_PREFIX` build-time env var (default: `/api`) in `frontend/src/api/client.js`. All API calls become `${VITE_API_PATH_PREFIX}/v1/...`. Production builds use the default. Preview builds use `/pr-42/api` (the PR segment leads so the Go handler can strip it as a single leading prefix).

**Rationale**: The current `VITE_API_BASE_URL` is for a full URL override (e.g., pointing to a different host). A separate path-prefix env var allows preview frontends to route to their own API paths without changing the domain. This is the minimal change to `client.js` needed to support path-based API routing.

**Alternatives considered**:
- `VITE_API_BASE_URL` pointing to the API Gateway Lambda Function URL directly — rejected because: (1) the clarification chose path-based routing via shared API Gateway; (2) CORS configuration on Lambda Function URLs adds complexity.
- Rewriting all call-site paths to include the prefix at build time — rejected; too invasive and error-prone.

---

## Decision 2 — Path Stripping in the Go Lambda Handler

**Decision**: Add a `STRIP_PATH_PREFIX` env var to `backend/cmd/lambda/main.go`. When set (e.g., `/pr-42`), the handler wraps the mux with `http.StripPrefix(prefix, mux)` before passing to `httpadapter.NewV2`. Production Lambdas leave this unset.

**Rationale**: The API Gateway route for PR previews is `ANY /pr-42/api/{proxy+}`. API Gateway forwards the full original path (e.g., `/pr-42/api/v1/recipes`) to the Lambda. The existing Go mux patterns are registered as `/api/v1/...`. Because `/pr-42` is a *leading* prefix, `http.StripPrefix("/pr-42", mux)` transforms the path back to `/api/v1/...` — exactly what the mux expects — with no changes to handler registrations. (A middle-segment layout like `/api/pr-42/v1/...` would not work, since `http.StripPrefix` only removes a leading prefix.)

**Alternatives considered**:
- API Gateway path parameter mapping to strip the prefix before Lambda invocation — rejected; HTTP API v2 parameter mapping for path rewriting is more complex to configure via AWS CLI and harder to test.
- Re-registering all mux routes with the PR prefix — rejected; requires code generation or duplication.

---

## Decision 3 — CloudFront Function for PR SPA Routing

**Decision**: Add a CloudFront viewer-request Function (one-time Terraform change) that rewrites requests to `/pr-{number}/` or `/pr-{number}` to `/pr-{number}/index.html`. This makes `cocktails.albertomcastro.com/pr-42/` load the PR SPA correctly.

**Rationale**: S3 does not auto-serve `index.html` for subdirectory requests (unlike website hosting mode). Without this function, visiting `/pr-42/` returns a 404 which CloudFront rewrite to the production `index.html`. The CloudFront Function is 10 lines of JS, runs at edge with no cold start, and handles all PRs generically without per-PR changes. The function is attached to the **default** (S3) cache behavior only. Preview API traffic (`/pr-42/api/*`) is matched by a separate `/pr-*/api/*` cache behavior that forwards to the API Gateway origin and does **not** run the SPA function — otherwise the function would rewrite API calls to `index.html`.

**Alternatives considered**:
- Using `/pr-42/index.html` as the direct preview URL — viable fallback, avoids Terraform changes, but produces an ugly URL in PR status checks.
- Lambda@Edge for path rewriting — rejected; CloudFront Functions are lighter, cheaper, and sufficient for simple URL rewriting.

---

## Decision 4 — Shared Preview Lambda IAM Execution Role

**Decision**: Create a single `cocktails-preview-lambda-role` IAM role via Terraform (one-time change) with DynamoDB permissions scoped to `arn:aws:dynamodb:*:*:table/cocktails-pr-*`. All preview Lambda functions use this shared role.

**Rationale**: Creating and deleting per-PR IAM roles from CI requires broad IAM permissions (`iam:CreateRole`, `iam:DeleteRole`, `iam:AttachRolePolicy`). A pre-created shared role reduces the blast radius of CI credentials and eliminates IAM management from the per-PR deploy scripts. The wildcard `cocktails-pr-*` ARN pattern correctly scopes access to preview tables only.

**Alternatives considered**:
- Per-PR IAM role created from CI — rejected; requires broad IAM permissions in the CI role, higher risk.
- Reusing the production Lambda execution role — rejected; violates isolation (preview Lambdas must not access production tables).

---

## Decision 5 — DynamoDB Table Provisioning and Seeding in CI

**Decision**: CI scripts use `aws dynamodb create-table` to provision three per-PR tables (`cocktails-pr-42-recipes`, `cocktails-pr-42-users`, `cocktails-pr-42-favorites`). Only the **recipes** table is seeded from production using `aws dynamodb scan` + `aws dynamodb batch-write-item`; the users and favorites tables are created empty. Seeding is skipped on subsequent pushes to the same PR (idempotent check: if tables exist, skip creation and seeding).

**Rationale**: DynamoDB's on-demand billing means tables cost nothing until accessed. Scan + batch-write is simple and requires no external tooling. Previews are publicly accessible (FR-011), so copying production user records or their favorites (PII) into them is unacceptable — only the non-personal recipes data is seeded. The idempotent check (table exists → skip) enables subsequent PR pushes to update code only, consistent with the spec's seed-once requirement.

**Alternatives considered**:
- DynamoDB export/import (S3 export format) — rejected; adds S3 intermediate step and is slower for small datasets.
- Seeding users/favorites from production — rejected; previews are public and must not contain production user PII.

---

## Decision 6 — API Gateway Route Management

**Decision**: The existing Terraform-managed HTTP API Gateway (`cocktails-prod-api`) gains per-PR routes added and removed by CI scripts via `aws apigatewayv2 create-route` / `delete-route` + `create-integration` / `delete-integration`. The existing `$default` route for production is untouched.

**Rationale**: HTTP API v2 supports multiple explicit routes alongside the `$default` catch-all. An explicit route `ANY /pr-42/api/{proxy+}` is more specific than `$default` and takes priority. CI manages these routes without Terraform to avoid state file pollution.

**Alternatives considered**:
- A separate API Gateway per PR — rejected; requires DNS or CloudFront changes per PR.
- Lambda Function URLs — rejected; the clarification selected shared API Gateway routing.

---

## Decision 7 — Production Deployment Workflow

**Decision**: Add a new `prod-deploy.yml` GitHub Actions workflow triggered on `push` to `main`. It builds the Lambda binary and frontend, deploys both (Lambda code update + S3 sync), and triggers a CloudFront invalidation. The existing `ci.yml` is extended with a `deploy-preview` job on `pull_request` events.

**Rationale**: The existing `ci.yml` only runs tests and builds — it does not deploy. Separating test (ci.yml) from deploy (prod-deploy.yml, preview-deploy.yml) keeps the CI workflow readable and allows deploy steps to be added/modified without touching the test pipeline.

**Alternatives considered**:
- Merging all logic into ci.yml with conditionals — viable but makes the workflow harder to read.

---

## Decision 8 — JWT Secret for Preview Lambdas

**Decision**: Preview Lambdas use the same `JWT_SECRET` as production, stored as a GitHub Actions secret and injected at deploy time.

**Rationale**: Preview environments share the same user pool (seeded from prod). Using the same JWT secret means a token obtained from production login also works in preview — useful for testing authenticated flows without re-logging in. For a personal cocktail app, this is an acceptable trade-off.

**Alternatives considered**:
- Per-PR JWT secret — more isolated but adds operational complexity; no meaningful security benefit for this use case.
