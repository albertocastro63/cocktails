# Contract: CI Workflow

**File**: `.github/workflows/ci.yml`  
**Trigger events**: `pull_request` (targeting `main`), `push` (to `main`)

## Named Status Checks

These are the check names that appear on GitHub PRs and that branch protection rules reference:

| Check Name | Job | What It Validates |
|------------|-----|------------------|
| `backend`  | backend | Go tests (unit + DynamoDB integration) + Lambda build |
| `frontend` | frontend | Vitest tests + Vite production build |

Both checks must pass for a PR to be merge-eligible (when branch protection is enabled).

## Job: `backend`

**Runner**: `ubuntu-latest`

**Service containers**:
- `dynamodb-local`: `amazon/dynamodb-local:latest`, port 8000

**Environment variables**:
- `DYNAMODB_ENDPOINT: http://localhost:8000`
- `AWS_REGION: us-east-1` (required by AWS SDK even with local endpoint)
- `AWS_ACCESS_KEY_ID: test` (local emulator accepts any non-empty value)
- `AWS_SECRET_ACCESS_KEY: test` (local emulator accepts any non-empty value)

**Steps**:
1. Checkout source
2. Set up Go (version from `backend/go.mod`)
3. Restore Go module/build cache
4. `cd backend && go test ./...` — must exit 0
5. `cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/bootstrap ./cmd/lambda/` — must exit 0
6. (P3 placeholder) OIDC role assumption via `aws-actions/configure-aws-credentials@v4`

**Failure behavior**: Any non-zero exit code marks the `backend` check as failed. GitHub displays the failing step output inline on the PR (satisfying FR-009).

## Job: `frontend`

**Runner**: `ubuntu-latest`

**Steps**:
1. Checkout source
2. Set up Node.js (version from `frontend/package.json` engines field, or LTS)
3. `cd frontend && npm ci` — install exact locked dependencies
4. `cd frontend && npm test` — Vitest run (exits non-zero if any test fails)
5. `cd frontend && npm run build` — Vite production build (exits non-zero on error)

**Failure behavior**: Any non-zero exit code marks the `frontend` check as failed.

## Concurrency Contract

```
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

Only the latest workflow run for a given branch determines the merge status. Older in-progress runs are cancelled automatically.

## OIDC Role Contract (US3 — implemented when IAM role is ready)

**Role ARN**: stored as GitHub Actions variable `vars.AWS_CI_ROLE_ARN` (not a secret)  
**Trust scope**: `repo:albertocastro63/cocktails:*`  
**Required job permission**: `id-token: write`  
**Action**: `aws-actions/configure-aws-credentials@v4`

No `AWS_ACCESS_KEY_ID` or `AWS_SECRET_ACCESS_KEY` repository secrets are used for real AWS calls.
