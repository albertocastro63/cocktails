# Quickstart: Verifying the CI Pipeline

## Acceptance Test 1 — Pipeline triggers on PR (US1)

1. Create a branch: `git checkout -b test/ci-validation`
2. Introduce a deliberate failure — e.g., in `backend/internal/handler/recipes.go`, add `panic("test failure")` inside any handler.
3. Push the branch and open a PR against `main`.
4. Observe the GitHub PR page: the `backend` check should appear as failing within 30 seconds of the push.
5. The merge button should be disabled (after branch protection is enabled on `main`).
6. Revert the panic, push again — the `backend` check should turn green.

## Acceptance Test 2 — DynamoDB integration tests run (US2)

1. On any PR run, navigate to the `backend` job log in GitHub Actions.
2. Expand the `go test ./...` step output.
3. Confirm `TestDynamo_` tests appear in the output (not skipped).
4. Confirm they pass (or fail with a meaningful DynamoDB error if there's a real bug).

**Expected log evidence** (no skip messages):
```
=== RUN   TestDynamo_SearchByIngredients_TwoIngredients
--- PASS: TestDynamo_SearchByIngredients_TwoIngredients (0.12s)
```

## Acceptance Test 3 — Frontend failure blocks merge (US1)

1. In `frontend/src/pages/RecipeList.js`, introduce a syntax error.
2. Push to a PR branch.
3. The `frontend` check should fail; merge is blocked.

## Acceptance Test 4 — Push to main triggers pipeline (US4)

1. Merge a PR (or push directly to main in the initial setup phase before branch protection is active).
2. Navigate to the commit in GitHub's commit history.
3. Confirm a status check badge appears on the commit.

## Acceptance Test 5 — OIDC authentication works (US3)

Prerequisite: IAM role `github-ci-role` created with OIDC trust policy scoped to this repository.

1. Ensure `vars.AWS_CI_ROLE_ARN` is set in the repository's Actions variables.
2. Uncomment the OIDC step in `.github/workflows/ci.yml`.
3. Push to a PR — the OIDC step should complete successfully.
4. Confirm `aws sts get-caller-identity` (if added as a verification step) shows the role ARN, not an IAM user.
5. Confirm no `AWS_ACCESS_KEY_ID` or `AWS_SECRET_ACCESS_KEY` secrets exist under Settings → Secrets and Variables → Actions.

## Local Validation (before pushing)

Run backend tests with DynamoDB Local locally:

```bash
# Terminal 1: start emulator
docker run --rm -p 8000:8000 amazon/dynamodb-local

# Terminal 2: run tests
cd backend
DYNAMODB_ENDPOINT=http://localhost:8000 AWS_REGION=us-east-1 \
  AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test \
  go test ./...
```

Run frontend checks:
```bash
cd frontend
npm test
npm run build
```
