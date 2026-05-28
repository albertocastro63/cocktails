# CI Script Contracts: PR Preview Environments

**Branch**: `021-pr-preview-environments` | **Date**: 2026-05-28

---

## Script: `preview-deploy.sh`

**Location**: `.github/scripts/preview-deploy.sh`  
**Trigger**: Called by `preview-deploy.yml` workflow on every push to an open PR.

### Inputs (environment variables)

| Variable | Description | Example |
|---|---|---|
| `PR_NUMBER` | GitHub PR number | `42` |
| `AWS_REGION` | AWS region | `us-east-1` |
| `LAMBDA_ROLE_ARN` | ARN of the shared preview execution role | `arn:aws:iam::123:role/cocktails-preview-lambda-role` |
| `API_GATEWAY_ID` | ID of the existing production HTTP API Gateway | `abc123xyz` |
| `FRONTEND_BUCKET` | Production S3 frontend bucket name | `cocktails-prod-frontend` |
| `CLOUDFRONT_DISTRIBUTION_ID` | Production CloudFront distribution ID | `EX7HUB6P225MV` |
| `JWT_SECRET` | JWT signing secret | (from GitHub secret) |
| `PROD_RECIPES_TABLE` | Production recipes DynamoDB table name | `cocktails-recipes` |
| `PROD_USERS_TABLE` | Production users DynamoDB table name | `cocktails-users` |
| `PROD_FAVORITES_TABLE` | Production favorites DynamoDB table name | `cocktails-favorites` |

### Inputs (positional / artifacts)

| Input | Description |
|---|---|
| `backend/bin/bootstrap` | Pre-built Lambda binary (arm64/linux) |
| `frontend/dist/` | Pre-built frontend assets (built with `VITE_API_PATH_PREFIX=/api/pr-${PR_NUMBER}`) |

### Steps (idempotent)

1. Derive names: `PR_ID=pr-${PR_NUMBER}`, table names, function name
2. **Tables**: If `cocktails-pr-${PR_NUMBER}-recipes` does not exist → create all three tables + seed from production. If exists → skip.
3. **Lambda**: If function `cocktails-pr-${PR_NUMBER}-api` does not exist → create with env vars + role. If exists → update function code only.
4. **API Gateway**: If integration for this PR does not exist → create integration + route. If exists → update integration Lambda ARN.
5. **Frontend**: `aws s3 sync frontend/dist/ s3://${FRONTEND_BUCKET}/pr-${PR_NUMBER}/ --delete`
6. **Output**: Print preview URL: `https://cocktails.albertomcastro.com/pr-${PR_NUMBER}/`

### Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | AWS operation failed (partial state may exist; see cleanup contract) |

---

## Script: `preview-teardown.sh`

**Location**: `.github/scripts/preview-teardown.sh`  
**Trigger**: Called by `preview-teardown.yml` workflow when a PR is closed or merged.

### Inputs (environment variables)

| Variable | Description |
|---|---|
| `PR_NUMBER` | GitHub PR number |
| `AWS_REGION` | AWS region |
| `API_GATEWAY_ID` | Production HTTP API Gateway ID |
| `FRONTEND_BUCKET` | Production S3 frontend bucket name |
| `CLOUDFRONT_DISTRIBUTION_ID` | Production CloudFront distribution ID |

### Steps (idempotent — each step continues even if the resource does not exist)

1. Delete Lambda function `cocktails-pr-${PR_NUMBER}-api` (ignore NotFound)
2. Delete API Gateway route + integration for `/api/pr-${PR_NUMBER}/{proxy+}` (ignore NotFound)
3. Delete DynamoDB tables `cocktails-pr-${PR_NUMBER}-{recipes,users,favorites}` (ignore NotFound)
4. Delete S3 objects under prefix `pr-${PR_NUMBER}/`
5. Create CloudFront invalidation for paths `/pr-${PR_NUMBER}/*`

### Exit codes

| Code | Meaning |
|---|---|
| `0` | All resources removed (or were already absent) |
| `1` | One or more AWS operations failed |

---

## Workflow: `preview-deploy.yml`

**Trigger**: `pull_request` → types: `[opened, synchronize, reopened]`  
**Branches**: all branches targeting `main`

**Jobs**:
1. `test` — reuse existing test jobs from `ci.yml` (or inline)
2. `deploy-preview` (needs: `test`) — runs `preview-deploy.sh`; posts preview URL as PR comment

---

## Workflow: `preview-teardown.yml`

**Trigger**: `pull_request` → types: `[closed]`  
**Branches**: all branches targeting `main`

**Jobs**:
1. `teardown` — runs `preview-teardown.sh`

---

## Workflow: `prod-deploy.yml`

**Trigger**: `push` → branches: `[main]`

**Jobs**:
1. `deploy-prod` — builds Lambda + frontend (with `VITE_API_PATH_PREFIX=/api` default), deploys Lambda code update, syncs S3, invalidates CloudFront

---

## New Lambda Environment Variables Contract

When creating a preview Lambda, these environment variables MUST be set:

```
STORE_BACKEND=dynamodb
RECIPES_TABLE=cocktails-pr-{number}-recipes
USERS_TABLE=cocktails-pr-{number}-users
FAVORITES_TABLE=cocktails-pr-{number}-favorites
JWT_SECRET={from GitHub secret PROD_JWT_SECRET}
STRIP_PATH_PREFIX=/pr-{number}
```

---

## New Frontend Build Contract

Preview frontend builds MUST set:

```
VITE_API_PATH_PREFIX=/api/pr-{number}
```

Production builds use the default (`/api`) — no variable required.
