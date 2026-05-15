# Quickstart & Integration Scenarios

**Feature**: 009-aws-terraform-infra
**Date**: 2026-05-15

---

## Scenario 1: First-Time Full Deploy (Greenfield)

**Goal**: Operator with zero prior AWS state gets the full application live.

**Steps**:
1. Configure AWS credentials (`aws configure` or export env vars).
2. Build Go binary: `cd backend && GOOS=linux GOARCH=arm64 go build -o bootstrap ./cmd/lambda && zip bootstrap.zip bootstrap`
3. Build frontend: `cd frontend && npm run build`
4. Bootstrap state backend (one time): `cd infra/bootstrap && terraform init && terraform apply`
5. Note the output `state_bucket_name` and `lock_table_name`; confirm they match `infra/backend.tf`.
6. Deploy: `cd infra && terraform init && terraform apply -var="lambda_binary_path=../backend/bootstrap.zip" -var="jwt_secret=<secret>"`
7. Note the output `api_url` and `frontend_url`.
8. Verify: `curl <api_url>/api/v1/recipes` → HTTP 200, JSON array.
9. Open `<frontend_url>` in browser → cocktails app loads.

**Expected result**: Full application running in `us-east-1`. Deploy completes in under 10 minutes.

---

## Scenario 2: Idempotent Re-Deploy (No Changes)

**Goal**: Re-running deploy after no code changes makes no modifications.

**Steps**:
1. `cd infra && terraform apply -var="lambda_binary_path=../backend/bootstrap.zip" -var="jwt_secret=<secret>"`
2. Observe: Terraform reports `0 to add, 0 to change, 0 to destroy`.

**Expected result**: Command exits 0 in under 30 seconds with no resource changes.

---

## Scenario 3: Re-Deploy After Backend Code Change

**Goal**: Updated Lambda binary is reflected in the live deployment.

**Steps**:
1. Make a code change in `backend/`.
2. Rebuild: `cd backend && GOOS=linux GOARCH=arm64 go build -o bootstrap ./cmd/lambda && zip bootstrap.zip bootstrap`
3. `cd infra && terraform apply -var="lambda_binary_path=../backend/bootstrap.zip" -var="jwt_secret=<secret>"`
4. Terraform detects a new artifact hash and updates the Lambda function.
5. Verify: `curl <api_url>/api/v1/recipes` returns updated behaviour.

**Expected result**: Lambda function updated; no other resources changed.

---

## Scenario 4: Re-Deploy After Frontend Change

**Goal**: Updated static assets are served from CloudFront after a frontend build.

**Steps**:
1. Make a change in `frontend/src/`.
2. Rebuild: `cd frontend && npm run build`
3. `cd infra && terraform apply` (Terraform syncs the `dist/` directory to the frontend S3 bucket).
4. Invalidate CloudFront cache (included in Terraform apply via `aws_cloudfront_invalidation` resource or manual step).
5. Open `<frontend_url>` → confirm new version loads.

**Expected result**: Frontend updated. Cache invalidation ensures immediate visibility (no stale CDN content).

---

## Scenario 5: Local SAM Test (Offline-ish)

**Goal**: Operator tests API functions locally without deploying.

**Steps**:
1. Build native binary: `cd backend && go build -o bootstrap ./cmd/lambda`
2. Ensure Docker is running.
3. `cd infra && sam local start-api` (requires `samconfig.toml` with `hook_name = "terraform"`).
4. In another terminal: `curl http://127.0.0.1:3000/api/v1/recipes`

**Note**: Local mode still hits real DynamoDB unless `STORE_BACKEND=sqlite DB_PATH=./test.db` is set in the SAM environment override file.

**Expected result**: API response within 60 seconds of SAM startup. Function exits cleanly after each invocation.

---

## Scenario 6: Local SAM Test (Fully Offline with SQLite)

**Goal**: Test functions with no AWS connectivity at all.

**Steps**:
1. Build native binary: `cd backend && go build -o bootstrap ./cmd/lambda`
2. Create SAM environment override file `infra/env.json`:
   ```json
   {
     "cocktails-api-prod": {
       "STORE_BACKEND": "sqlite",
       "DB_PATH": "/tmp/cocktails_local.db",
       "JWT_SECRET": "local-dev-secret"
     }
   }
   ```
3. `cd infra && sam local start-api --env-vars env.json`
4. `curl http://127.0.0.1:3000/api/v1/recipes` → empty array (fresh DB).
5. POST a recipe, GET it back.

**Expected result**: Full CRUD cycle works locally with zero AWS connectivity.

---

## Scenario 7: Partial Deploy Failure Recovery

**Goal**: A deploy interrupted mid-way (e.g., network loss) recovers on retry.

**Steps**:
1. Start a deploy; simulate interruption (Ctrl-C after some resources are created).
2. Observe Terraform state is partially updated.
3. Re-run: `terraform apply` (same variables).
4. Terraform reconciles actual vs desired state and completes the deploy.

**Expected result**: Final state is identical to a clean deploy. No orphaned resources.

---

## Scenario 8: Insufficient IAM Permissions

**Goal**: Operator with incomplete IAM permissions gets a clear error.

**Steps**:
1. Deploy with credentials that lack `lambda:CreateFunction`.
2. Observe: Terraform reports `AccessDeniedException` with the missing action name.
3. Add the missing permission; re-run deploy.

**Expected result**: Error message identifies exactly which permission is missing. No partial resources are left (Terraform rolls back cleanly via plan/apply separation).

---

## Scenario 9: Full Tear Down

**Goal**: All resources removed cleanly after a successful deploy.

**Steps**:
1. `cd infra && terraform destroy -var="lambda_binary_path=../backend/bootstrap.zip" -var="jwt_secret=<secret>"`
2. Confirm the prompt (`yes`).
3. Wait for destroy to complete.
4. Verify: search AWS console for tag `project=cocktails` → zero results in `us-east-1`.
5. Confirm `<api_url>` returns connection refused or DNS error.

**Expected result**: `terraform destroy` exits 0. Zero project-tagged resources remain. No ongoing charges.

---

## Scenario 10: Bootstrap Re-Run (Idempotent)

**Goal**: Running the bootstrap twice does not create duplicate resources.

**Steps**:
1. `cd infra/bootstrap && terraform apply` (second run).
2. Observe: `0 to add, 0 to change, 0 to destroy`.

**Expected result**: No changes, no errors.
