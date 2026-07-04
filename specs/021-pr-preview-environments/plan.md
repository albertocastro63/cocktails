# Implementation Plan: PR Preview Environments

**Branch**: `021-pr-preview-environments` | **Date**: 2026-05-28 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/021-pr-preview-environments/spec.md`

## Summary

Every open PR gets an isolated AWS environment (Lambda + DynamoDB tables + S3 frontend assets) automatically deployed on push and torn down on close/merge. Production deployment on merge to main becomes automatic. The frontend gains a `VITE_API_PATH_PREFIX` build variable; the Lambda handler gains a `STRIP_PATH_PREFIX` env var; a shared preview IAM role and a CloudFront Function are added via a one-time Terraform change.

---

## Technical Context

**Language/Version**: Go 1.22 (backend), Node.js 24 / Vite (frontend), Bash (CI scripts)  
**Primary Dependencies**: GitHub Actions, AWS CLI v2, `httpadapter` (aws-lambda-go-api-proxy), AWS HTTP API Gateway v2  
**Storage**: DynamoDB (PAY_PER_REQUEST) — three per-PR tables seeded from production at first push  
**Testing**: Go `testing` package + Vitest; TDD cycle mandatory per constitution  
**Target Platform**: AWS Lambda (arm64, provided.al2023), S3 + CloudFront (frontend)  
**Project Type**: CI/CD infrastructure + minimal application code changes  
**Performance Goals**: Preview environment live within 5 minutes of CI completion (NFR-001)  
**Constraints**: No per-PR Terraform state; CI role permissions must remain least-privilege  
**Scale/Scope**: Handles multiple concurrent PRs; expected ≤ 10 open at once

---

## Constitution Check

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Are functions single-responsibility and below complexity limits? | ✅ `preview-deploy.sh` broken into named functions; Go path-strip middleware is 3 lines |
| II. Test-First | Are failing tests written before implementation begins? | ✅ Unit tests for `VITE_API_PATH_PREFIX` client path construction and `STRIP_PATH_PREFIX` Go middleware written first |
| III. UX Consistency | Do all UI surfaces follow the design language and handle loading/empty/error states? | ✅ No new UI; preview serves the same SPA |
| IV. Performance | Do API responses meet p95 ≤ 200 ms read / ≤ 500 ms write and TTI ≤ 3 s? | ✅ Preview Lambda has same memory/timeout as production; no hot-path changes |
| Quality Gates | Do all CI checks (lint, coverage ≥ 75%, benchmarks) pass? | ✅ Only `client.js` and `main.go` change in application code; existing coverage maintained |

---

## Project Structure

### Documentation (this feature)

```text
specs/021-pr-preview-environments/
├── plan.md           ← this file
├── research.md       ← architectural decisions
├── data-model.md     ← resource naming conventions and schemas
├── quickstart.md     ← manual testing guide
├── contracts/
│   └── ci-scripts.md ← CI script interface contracts
└── tasks.md          ← /speckit-tasks output
```

### Source Code Changes

```text
# Application code (minimal changes)
frontend/src/api/client.js              ← Add VITE_API_PATH_PREFIX support
backend/cmd/lambda/main.go              ← Add STRIP_PATH_PREFIX middleware

# New CI workflows
.github/workflows/preview-deploy.yml   ← Deploy on PR push
.github/workflows/preview-teardown.yml ← Tear down on PR close/merge
.github/workflows/prod-deploy.yml      ← Deploy to prod on push to main

# New CI scripts
.github/scripts/preview-deploy.sh      ← Idempotent preview deploy logic
.github/scripts/preview-teardown.sh    ← Idempotent preview teardown logic

# Infrastructure (one-time Terraform additions)
infra/cloudfront-function-spa.js       ← CloudFront Function for PR SPA routing
infra/main.tf                          ← Add preview IAM role + CloudFront Function
```

---

## Architecture

### Request Flow — Preview Environment (PR 42)

```
Browser: https://cocktails.albertomcastro.com/pr-42/
  └─ CloudFront: viewer-request Function rewrites /pr-42/ → /pr-42/index.html
  └─ S3 origin: serves s3://cocktails-prod-frontend/pr-42/index.html
     └─ SPA loads; frontend built with VITE_API_PATH_PREFIX=/pr-42/api
        └─ API call: GET https://cocktails.albertomcastro.com/pr-42/api/v1/recipes
           └─ CloudFront: /pr-*/api/* behavior → HTTP API Gateway (SPA function NOT attached)
              └─ API Gateway: route ANY /pr-42/api/{proxy+} → cocktails-pr-42-api Lambda
                 └─ Lambda: STRIP_PATH_PREFIX=/pr-42 → http.StripPrefix strips leading /pr-42 → Go mux /api/v1/recipes
                    └─ DynamoDB: cocktails-pr-42-recipes (seeded, isolated)
```

**Routing note**: The PR segment is a *leading* prefix (`/pr-42/api/...`), not embedded after `/api`. This is required so `http.StripPrefix("/pr-42", h)` can strip it in one operation and hand the Go mux its native `/api/v1/...` paths. A dedicated CloudFront behavior `/pr-*/api/*` (added in Phase 1) forwards these to the API Gateway origin and is matched before the default S3 behavior, so the SPA viewer-request function never rewrites preview API calls to `index.html`.

### One-Time Terraform Additions

**1. Preview Lambda Execution Role** (`cocktails-preview-lambda-role`)
- Trust policy: Lambda service
- Permission: DynamoDB full access on `arn:aws:dynamodb:*:*:table/cocktails-pr-*`
- Permission: CloudWatch Logs on `/aws/lambda/cocktails-pr-*`

**2. CloudFront Function** (`spa-pr-routing`)
- Event: `viewer-request`
- Logic: if path matches `/pr-{digits}` or `/pr-{digits}/`, append `/index.html`
- Attached to: CloudFront **default** cache behavior only (S3 origin)

**3. CloudFront cache behavior** (`/pr-*/api/*`)
- Path pattern: `/pr-*/api/*` → API Gateway origin (same policies as the production `/api/*` behavior)
- SPA function **not** attached, so preview API calls are forwarded, not rewritten to `index.html`
- Ordered before the default behavior so it matches preview API traffic first

### Per-PR CI Operations (no Terraform)

**Deploy** (idempotent — safe to re-run):
1. Check if DynamoDB tables exist → if not, create all three; seed **only** the recipes table from production (users and favorites stay empty — no production PII in a public preview)
2. Check if Lambda exists → if not, create; if yes, update code
3. Check if API GW route exists (`ANY /pr-{number}/api/{proxy+}`) → if not, create; if yes, update integration
4. Build frontend with `VITE_API_PATH_PREFIX=/pr-{number}/api`
5. Sync frontend to `s3://cocktails-prod-frontend/pr-{number}/`
6. Post preview URL as PR comment

**Teardown** (idempotent — safe to re-run, ignores NotFound):
1. Delete Lambda function
2. Delete API GW route + integration
3. Delete DynamoDB tables
4. Delete S3 prefix `pr-{number}/`
5. Invalidate CloudFront `/pr-{number}/*`

---

## Phase 0: Application Code Changes

### 0.1 Frontend — `VITE_API_PATH_PREFIX`

**File**: `frontend/src/api/client.js`

Current:
```js
const BASE_URL = import.meta.env.VITE_API_BASE_URL || '';
// calls: `${BASE_URL}/api/v1/recipes`
```

Change:
```js
const BASE_URL = import.meta.env.VITE_API_BASE_URL || '';
const API_PREFIX = import.meta.env.VITE_API_PATH_PREFIX || '/api';
// calls: `${BASE_URL}${API_PREFIX}/v1/recipes`
```

All hardcoded `/api/v1/...` paths in `client.js` are replaced with `${API_PREFIX}/v1/...`.

**Tests** (written first):
- `client.test.js`: `getRecipes()` with `VITE_API_PATH_PREFIX=/pr-42/api` calls `/pr-42/api/v1/recipes`
- `client.test.js`: `getRecipes()` with no prefix env var calls `/api/v1/recipes` (unchanged behaviour)

### 0.2 Backend — `STRIP_PATH_PREFIX`

**File**: `backend/cmd/lambda/main.go`

Current:
```go
lambda.Start(httpadapter.NewV2(h).ProxyWithContext)
```

Change:
```go
h := buildHandler(...)
if prefix := os.Getenv("STRIP_PATH_PREFIX"); prefix != "" {
    h = http.StripPrefix(prefix, h)
}
lambda.Start(httpadapter.NewV2(h).ProxyWithContext)
```

**Tests** (written first):
- Integration test: request to `/pr-42/api/v1/recipes` with `STRIP_PATH_PREFIX=/pr-42` is handled identically to `/api/v1/recipes` without the env var

---

## Phase 1: One-Time Infrastructure (Terraform)

### 1.1 CloudFront Function

**New file**: `infra/cloudfront-function-spa.js`
```js
function handler(event) {
  var request = event.request;
  var uri = request.uri;
  var prMatch = uri.match(/^(\/pr-\d+)(\/.*)?$/);
  if (prMatch && !/\.[a-zA-Z0-9]+$/.test(uri)) {
    request.uri = prMatch[1] + '/index.html';
  }
  return request;
}
```

**`infra/main.tf` additions**:
- `aws_cloudfront_function` resource referencing `cloudfront-function-spa.js`
- Attach function to `module.cdn` default cache behavior as `viewer-request`

### 1.2 Preview Lambda Execution Role

**`infra/main.tf` addition**:
```hcl
resource "aws_iam_role" "preview_lambda" {
  name = "${var.project_name}-preview-lambda-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
}

resource "aws_iam_role_policy" "preview_lambda_dynamo" {
  role = aws_iam_role.preview_lambda.id
  policy = data.aws_iam_policy_document.preview_dynamo.json
}

data "aws_iam_policy_document" "preview_dynamo" {
  statement {
    actions   = ["dynamodb:*"]
    resources = ["arn:aws:dynamodb:*:*:table/${var.project_name}-pr-*"]
  }
  statement {
    actions   = ["logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents"]
    resources = ["arn:aws:logs:*:*:log-group:/aws/lambda/${var.project_name}-pr-*:*"]
  }
}
```

**Output**: `preview_lambda_role_arn` — consumed by CI scripts.

---

## Phase 2: CI Workflows and Scripts

### 2.1 `preview-deploy.sh`

Located at `.github/scripts/preview-deploy.sh`. Idempotent bash script. See [contracts/ci-scripts.md](contracts/ci-scripts.md) for full interface.

Key implementation details:
- Use `aws dynamodb describe-table` to check if tables exist before creating
- Use `aws dynamodb scan` + `jq` to extract items; `aws dynamodb batch-write-item` to seed (handles 25-item batch limit with a loop)
- Use `aws lambda get-function` to check if Lambda exists; branch on `create-function` vs `update-function-code`
- Use `aws apigatewayv2 get-routes` + `jq` to find existing route; branch on `create-route` vs skip
- Print preview URL to stdout: `echo "https://cocktails.albertomcastro.com/pr-${PR_NUMBER}/"`. The calling workflow step captures this output and posts the PR comment via `gh pr comment` (requires `GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}` in the workflow env block)

### 2.2 `preview-teardown.sh`

Located at `.github/scripts/preview-teardown.sh`. All steps wrapped in `|| true` / NotFound-safe. See [contracts/ci-scripts.md](contracts/ci-scripts.md).

### 2.3 `preview-deploy.yml`

```yaml
on:
  pull_request:
    branches: [main]
    types: [opened, synchronize, reopened]

jobs:
  deploy-preview:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      contents: read
      pull-requests: write  # for PR comments
    env:
      GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    steps:
      - checkout
      - build Lambda binary (CGO_ENABLED=0 GOOS=linux GOARCH=arm64)
      - build frontend (VITE_API_PATH_PREFIX=/pr-${{ github.event.pull_request.number }}/api)
      - configure AWS credentials (OIDC)
      - run preview-deploy.sh (captures stdout for PR comment URL)
      - post PR comment: gh pr comment ${{ github.event.pull_request.number }} --body "Preview: https://cocktails.albertomcastro.com/pr-${{ github.event.pull_request.number }}/"
```

### 2.4 `preview-teardown.yml`

```yaml
on:
  pull_request:
    branches: [main]
    types: [closed]

jobs:
  teardown:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      contents: read
    steps:
      - checkout
      - configure AWS credentials (OIDC)
      - run preview-teardown.sh
```

### 2.5 `prod-deploy.yml`

```yaml
on:
  push:
    branches: [main]

jobs:
  deploy-prod:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      contents: read
    steps:
      - checkout
      - build Lambda binary
      - build frontend (VITE_API_PATH_PREFIX default /api)
      - configure AWS credentials (OIDC)
      - aws lambda update-function-code
      - aws s3 sync frontend/dist/ s3://$FRONTEND_BUCKET/ --delete
      - aws cloudfront create-invalidation --paths "/*"
```

### 2.6 Required GitHub Actions Secrets/Variables

| Name | Type | Description |
|---|---|---|
| `AWS_CI_ROLE_ARN` | Variable | Already exists; must be expanded with new permissions |
| `PROD_JWT_SECRET` | Secret | JWT signing secret; passed to Lambda as env var `JWT_SECRET` (i.e., `JWT_SECRET: ${{ secrets.PROD_JWT_SECRET }}` in workflow env block) |
| `PREVIEW_LAMBDA_ROLE_ARN` | Variable | ARN of the Terraform-created preview execution role |
| `API_GATEWAY_ID` | Variable | Production HTTP API Gateway ID |
| `FRONTEND_BUCKET` | Variable | `cocktails-prod-frontend` |
| `CLOUDFRONT_DISTRIBUTION_ID` | Variable | `EX7HUB6P225MV` |
| `PROD_RECIPES_TABLE` | Variable | `cocktails-recipes` (only table seeded into previews) |
| `GH_TOKEN` | Workflow env | Set to `${{ secrets.GITHUB_TOKEN }}` (auto-provided by GitHub Actions); required for `gh pr comment` in `preview-deploy.yml` |

### 2.7 CI IAM Role Permission Additions

The `AWS_CI_ROLE_ARN` role needs these additional permissions:
- `lambda:CreateFunction`, `lambda:DeleteFunction`, `lambda:UpdateFunctionCode`, `lambda:GetFunction`, `lambda:AddPermission`, `lambda:RemovePermission`
- `dynamodb:CreateTable`, `dynamodb:DeleteTable`, `dynamodb:DescribeTable`, `dynamodb:Scan`, `dynamodb:BatchWriteItem`
- `apigatewayv2:CreateRoute`, `apigatewayv2:DeleteRoute`, `apigatewayv2:GetRoutes`, `apigatewayv2:CreateIntegration`, `apigatewayv2:DeleteIntegration`, `apigatewayv2:GetIntegrations`, `apigatewayv2:UpdateIntegration`
- `s3:PutObject`, `s3:DeleteObject`, `s3:ListBucket` on `cocktails-prod-frontend`
- `cloudfront:CreateInvalidation`
- `iam:PassRole` on the preview Lambda execution role ARN
- `logs:CreateLogGroup`, `logs:DeleteLogGroup` on `/aws/lambda/cocktails-pr-*`

---

## Complexity Tracking

No constitution violations. All complexity is justified:
- The CI scripts are shell scripts, not application code — complexity limits (40-line functions, CC ≤ 10) apply to the Go and JS application code only.
- The one-time Terraform changes are additive and do not touch production routing logic.
- NFR-001/NFR-002 (5-minute timing SLAs): Enforced by an automated CI gate (T033a) that timestamps the workflow job and fails if end-to-end deploy/teardown exceeds 300 seconds. Also spot-checked during the T033 acceptance walkthrough.
