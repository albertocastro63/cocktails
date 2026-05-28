# Tasks: PR Preview Environments

**Input**: Design documents from `specs/021-pr-preview-environments/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ci-scripts.md ✓, quickstart.md ✓

---

## Phase 1: Setup — One-Time Terraform Infrastructure

**Purpose**: Provision shared AWS resources required by all preview environments. These are created once and are not per-PR. Must complete before any preview CI work.

**⚠️ CRITICAL**: Phase 2 and all user story phases depend on the preview IAM role ARN produced here.

- [ ] T001 Write CloudFront Function for PR SPA routing in `infra/cloudfront-function-spa.js` (handles any path starting with `/pr-{digits}` that has no file extension → rewrites to `/pr-{digits}/index.html` for SPA deep-link support)
- [ ] T002 Add `aws_cloudfront_function` resource and `aws_iam_role` preview Lambda execution role to `infra/main.tf`; add `preview_lambda_role_arn` to `infra/outputs.tf`
- [ ] T003 Add CloudFront Function association to the CDN module's default cache behavior viewer-request handler in `infra/main.tf`
- [ ] T004 Expand CI IAM role policy in `infra/main.tf` (or via AWS console) to include Lambda create/delete/update, DynamoDB create/delete/describe/scan/batch-write, API Gateway v2 route/integration CRUD, S3 PutObject/DeleteObject on `cocktails-prod-frontend`, CloudFront CreateInvalidation, and `iam:PassRole` on the preview Lambda role
- [ ] T005 Run `terraform plan` from `infra/` and verify planned additions match plan.md spec (CloudFront Function + IAM role + policy)
- [ ] T006 Run `terraform apply` from `infra/` to provision the one-time resources

**Checkpoint**: `terraform output preview_lambda_role_arn` returns a valid ARN; CloudFront Function appears in AWS console.

---

## Phase 2: Foundational — Application Code Changes (TDD)

**Purpose**: Minimal changes to `client.js` and `main.go` that enable path-based routing for previews. These must be complete (with passing tests) before CI scripts can be validated end-to-end.

**⚠️ Constitution §II**: Tests MUST be written first and confirmed failing before implementation.

- [ ] T007 Write failing Vitest test in `frontend/src/api/client.test.js`: assert that `getRecipes()` with `VITE_API_PATH_PREFIX=/api/pr-42` calls `/api/pr-42/v1/recipes`; assert no-prefix case still calls `/api/v1/recipes`
- [ ] T008 Write failing Go test in `backend/cmd/lambda/main_test.go`: assert that a handler wrapped with `STRIP_PATH_PREFIX=/pr-42` correctly routes a `GET /pr-42/api/v1/recipes` request to the same mux handler as `GET /api/v1/recipes`
- [ ] T009 [P] Implement `VITE_API_PATH_PREFIX` support in `frontend/src/api/client.js`: replace all hardcoded `/api/v1/...` path prefixes with `${API_PREFIX}/v1/...` where `API_PREFIX = import.meta.env.VITE_API_PATH_PREFIX || '/api'`
- [ ] T010 [P] Implement `STRIP_PATH_PREFIX` middleware in `backend/cmd/lambda/main.go`: read `os.Getenv("STRIP_PATH_PREFIX")`; if non-empty, wrap handler with `http.StripPrefix(prefix, h)` before passing to `httpadapter.NewV2`
- [ ] T011 Run `cd frontend && npm test` — confirm both new client path tests pass with ≥75% coverage maintained
- [ ] T012 Run `cd backend && go test ./cmd/lambda/...` — confirm `STRIP_PATH_PREFIX` test passes

**Checkpoint**: Both test suites pass; `git diff frontend/src/api/client.js` shows `API_PREFIX` usage; `git diff backend/cmd/lambda/main.go` shows the `StripPrefix` wrapper.

---

## Phase 3: User Story 1 — Preview Environment on PR Push (Priority: P1) 🎯 MVP

**Goal**: Every push to an open PR deploys a complete isolated preview environment and posts the URL as a PR comment.

**Independent Test**: Open a real PR, push a commit, confirm `cocktails.albertomcastro.com/pr-{number}/` is reachable and serves the PR's frontend; confirm `/api/pr-{number}/v1/recipes` returns data from isolated DynamoDB tables.

### Implementation for User Story 1

- [ ] T013 [US1] Write `.github/scripts/preview-deploy.sh` — idempotent bash script per contracts/ci-scripts.md: check table existence → create + seed if missing → create or update Lambda function → create or update API Gateway route + integration → sync frontend to S3; all steps wrapped in clear log output
- [ ] T014 [US1] Implement DynamoDB table creation in `preview-deploy.sh`: `aws dynamodb create-table` for each of recipes/users/favorites using PAY_PER_REQUEST and the schema from data-model.md; wait for ACTIVE status with `aws dynamodb wait table-exists`
- [ ] T015 [US1] Implement production data seeding in `preview-deploy.sh`: scan each production table with `aws dynamodb scan`, batch-write to PR tables in 25-item batches; skip seeding if tables already existed (idempotent guard)
- [ ] T016 [US1] Implement Lambda function provisioning in `preview-deploy.sh`: `aws lambda create-function` with binary zip, `provided.al2023` runtime, `arm64` architecture, `PREVIEW_LAMBDA_ROLE_ARN`, and env vars per data-model.md; on update use `aws lambda update-function-code`
- [ ] T017 [US1] Implement API Gateway route management in `preview-deploy.sh`: check for existing route `ANY /api/pr-{number}/{proxy+}` via `aws apigatewayv2 get-routes | jq`; create integration + route if absent; add `aws lambda add-permission` for API Gateway invocation
- [ ] T018 [US1] Implement frontend upload step in `preview-deploy.sh`: `aws s3 sync frontend/dist/ s3://${FRONTEND_BUCKET}/pr-${PR_NUMBER}/ --delete` (AWS CLI infers content-type from file extensions automatically)
- [ ] T019 [US1] Write `.github/workflows/preview-deploy.yml` — triggered on `pull_request` types `[opened, synchronize, reopened]` targeting `main`; jobs: build Lambda binary (CGO_ENABLED=0 GOOS=linux GOARCH=arm64), build frontend (`VITE_API_PATH_PREFIX=/api/pr-${{ github.event.pull_request.number }}`), configure OIDC credentials, run `preview-deploy.sh` (script prints URL to stdout), post PR comment via `gh pr comment --body "Preview: $(./preview-deploy.sh | tail -1)"` using `env: GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}`
- [ ] T020 [US1] Add all required secrets and variables to GitHub Actions repository settings: `PREVIEW_LAMBDA_ROLE_ARN`, `API_GATEWAY_ID`, `FRONTEND_BUCKET`, `CLOUDFRONT_DISTRIBUTION_ID`, `PROD_RECIPES_TABLE`, `PROD_USERS_TABLE`, `PROD_FAVORITES_TABLE` (document values from `terraform output`)
- [ ] T020a [US1] Verify that the GitHub Actions secret for the JWT signing key is named `PROD_JWT_SECRET`; confirm the workflow maps it to `JWT_SECRET` in the Lambda env vars (e.g., `JWT_SECRET: ${{ secrets.PROD_JWT_SECRET }}`); update the secret name in the workflow if it differs

**Checkpoint**: Open a real PR → push a commit → `deploy-preview` job passes → preview URL comment appears on PR → `https://cocktails.albertomcastro.com/pr-{number}/` loads the SPA with isolated data.

---

## Phase 4: User Story 2 — Automatic Production Deployment on Merge (Priority: P1)

**Goal**: Merging any PR to main automatically deploys the production Lambda and frontend; no manual `aws lambda update-function-code` or `aws s3 sync` required.

**Independent Test**: Merge a PR to main, wait for `prod-deploy` workflow to pass, confirm production URL serves the merged code.

### Implementation for User Story 2

- [ ] T021 [US2] Write `.github/workflows/prod-deploy.yml` — triggered on `push` to `main`; jobs: build Lambda binary, build frontend (default `VITE_API_PATH_PREFIX=/api`), configure OIDC credentials, `aws lambda update-function-code` with new zip, `aws s3 sync frontend/dist/ s3://${FRONTEND_BUCKET}/ --delete`, `aws cloudfront create-invalidation --paths "/*"`
- [ ] T022 [US2] Add `PROD_LAMBDA_FUNCTION_NAME` (value: `cocktails-prod-api`) as a GitHub Actions variable, reuse existing `FRONTEND_BUCKET`, `CLOUDFRONT_DISTRIBUTION_ID` variables from US1
- [ ] T023 [US2] Verify `prod-deploy.yml` uses the same Lambda binary build flags as the existing manual deploy documented in README (`GOOS=linux GOARCH=arm64`); confirm the workflow uses `CGO_ENABLED=0`

**Checkpoint**: Merge a test PR to main → `prod-deploy` workflow passes → `https://cocktails.albertomcastro.com` serves the newly merged code → no manual deployment step required.

---

## Phase 5: User Story 3 — Preview Environment Teardown on PR Close/Merge (Priority: P2)

**Goal**: When a PR is merged or closed, all its AWS resources (Lambda, DynamoDB tables, API Gateway route, S3 assets) are removed automatically.

**Independent Test**: Close a PR (without merging) → `teardown` job passes → Lambda function no longer exists (`aws lambda get-function` returns NotFound) → preview URL returns 404.

### Implementation for User Story 3

- [ ] T024 [US3] Write `.github/scripts/preview-teardown.sh` — idempotent teardown script per contracts/ci-scripts.md: delete Lambda function (ignore NotFound), delete API GW route + integration (iterate `get-routes` to find PR-specific routes), delete DynamoDB tables (ignore NotFound, wait for deletion), delete S3 objects under `pr-{number}/`, create CloudFront invalidation for `/pr-{number}/*`
- [ ] T025 [US3] Implement Lambda deletion in `preview-teardown.sh`: `aws lambda delete-function --function-name cocktails-pr-${PR_NUMBER}-api`; use `|| true` to continue on NotFound
- [ ] T026 [US3] Implement API Gateway cleanup in `preview-teardown.sh`: query routes with `aws apigatewayv2 get-routes | jq`; extract route ID for `/api/pr-{number}/{proxy+}`; `aws apigatewayv2 delete-route` then `aws apigatewayv2 delete-integration`; continue on NotFound
- [ ] T027 [US3] Implement DynamoDB teardown in `preview-teardown.sh`: `aws dynamodb delete-table` for all three tables; wait for deletion with `aws dynamodb wait table-not-exists`; use `|| true` on NotFound. Also delete the Lambda log group: `aws logs delete-log-group --log-group-name /aws/lambda/cocktails-pr-${PR_NUMBER}-api || true`
- [ ] T028 [US3] Implement S3 cleanup in `preview-teardown.sh`: `aws s3 rm s3://${FRONTEND_BUCKET}/pr-${PR_NUMBER}/ --recursive`
- [ ] T029 [US3] Implement CloudFront invalidation in `preview-teardown.sh`: `aws cloudfront create-invalidation --distribution-id ${CLOUDFRONT_DISTRIBUTION_ID} --paths "/pr-${PR_NUMBER}/*"`
- [ ] T030 [US3] Write `.github/workflows/preview-teardown.yml` — triggered on `pull_request` type `[closed]` targeting `main`; job: configure OIDC credentials, run `preview-teardown.sh`; ensure `PR_NUMBER` is derived from `${{ github.event.pull_request.number }}`

**Checkpoint**: Close a PR → `teardown` workflow passes → `aws lambda get-function --function-name cocktails-pr-{number}-api` returns NotFound → `aws s3 ls s3://cocktails-prod-frontend/pr-{number}/` returns empty → preview URL returns 404.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T031 [P] Make `preview-deploy.sh` and `preview-teardown.sh` executable (`chmod +x`) and add `#!/usr/bin/env bash` + `set -euo pipefail` headers; verify no shellcheck warnings (`shellcheck .github/scripts/*.sh`)
- [ ] T032 [P] Update `README.md`: add a "Preview Environments" subsection under Infrastructure documenting the preview URL pattern, how to find the URL on a PR, and that production deploys automatically on merge
- [ ] T032a Verify that CloudFront returns a 404 (not 403) when a torn-down preview URL is visited: configure a custom error response in the CDN module (403 → 404 with a redirect to `/`) or confirm the existing error page is acceptable; update spec User Story 3 acceptance scenario 3 if 403 is sufficient
- [ ] T033 Run the full acceptance scenario from `quickstart.md` end-to-end: open a test PR, verify deploy, verify data isolation (write a recipe in preview, confirm it is absent in production), merge PR, verify prod deploy, verify teardown
- [ ] T034 Verify frontend test coverage remains ≥75% after client.js change (`cd frontend && npm test -- --coverage`)
- [ ] T035 Verify backend test coverage remains ≥75% after main.go change (`cd backend && go test -p 1 -coverprofile=coverage.out -coverpkg=./internal/... ./... && go tool cover -func=coverage.out`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately. Produces `preview_lambda_role_arn` needed by T016/T019.
- **Phase 2 (Foundational)**: Independent of Phase 1 — application code changes can proceed in parallel with Terraform work.
- **Phase 3 (US1)**: Depends on Phase 1 (IAM role) + Phase 2 (client.js + main.go) being complete.
- **Phase 4 (US2)**: Independent of Phase 3 scripts, but T022 reuses `FRONTEND_BUCKET` and `CLOUDFRONT_DISTRIBUTION_ID` variables first provisioned in T020. T020 must complete before T022.
- **Phase 5 (US3)**: Independent of Phase 3 — teardown scripts do not depend on deploy scripts.
- **Phase 6 (Polish)**: Depends on all prior phases complete.

### Critical Path

```
Phase 1 (T001-T006) ─┐
                      ├─→ Phase 3 (T013-T020) → Phase 6
Phase 2 (T007-T012) ─┘

Phase 4 (T021-T023) ─ independent, any time after Phase 2
Phase 5 (T024-T030) ─ independent, any time after Phase 1
```

### Within-Phase Parallel Opportunities

**Phase 1**: T001 + T002 can be written in parallel (different file areas); T003 depends on T002; T004 independent; T005 depends on T002/T003 (plan); T006 depends on all.

**Phase 2**: T007 + T008 written in parallel (different languages); T009 + T010 implemented in parallel (different codebases); T011 + T012 run in parallel.

**Phase 3**: T013-T018 are sequential (each step adds to the same script); T019 (workflow file) can be drafted while T013-T018 are being written.

**Phase 4 + Phase 5**: Can be worked on in parallel with each other after Phase 2 is done.

---

## Parallel Execution Examples

### Phase 2 Parallel (Foundational)

```
Session A: T007 → T009 → T011  (frontend test → implement → verify)
Session B: T008 → T010 → T012  (backend test → implement → verify)
```

### Phase 3 + Phase 4 Parallel (after Phase 2)

```
Session A: T013 → T014 → T015 → T016 → T017 → T018 → T019 → T020  (preview deploy)
Session B: T021 → T022 → T023  (prod deploy — independent)
```

---

## Implementation Strategy

### MVP (US1 + US2 only — teardown deferred)

1. Complete Phase 1 (Terraform one-time setup)
2. Complete Phase 2 (application code changes with tests)
3. Complete Phase 3 (preview deploy workflow)
4. **STOP and VALIDATE**: Open a real PR, confirm preview is live
5. Complete Phase 4 (prod deploy workflow)
6. **STOP and VALIDATE**: Merge a real PR, confirm prod auto-deploys
7. Deploy MVP — teardown can be added in a follow-up

### Full Delivery (all stories)

Follow phases 1→2→3→4→5→6 sequentially, validating each checkpoint before proceeding.
