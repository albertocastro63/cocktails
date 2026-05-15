# Tasks: AWS Infrastructure Deployment

**Input**: Design documents from `specs/009-aws-terraform-infra/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓, quickstart.md ✓

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Exact file paths are included in every task description

## Phase 0: Pre-Implementation Validation Baseline (Constitution Principle II Adaptation)

**Purpose**: Establish the "failing test" baseline required by the Test-First constitution principle. For declarative IaC, the adapted test-first cycle is: run validators on an empty/incomplete workspace (they fail) → implement HCL → run validators again (they pass).

- [ ] T000 Run `terraform init -backend=false && terraform validate` inside an empty `infra/` directory (no .tf files yet — validate will error with "No configuration files"). Run `checkov --directory infra/ --framework terraform` (no findings, exit 0 on empty dir). Document both outputs as the "failing baseline" before implementation begins. This step satisfies the IaC adaptation of constitution Principle II as documented in plan.md.

**Checkpoint**: Baseline recorded. T000 must be done before T001.

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Create directory structure, version constraints, and shared configuration files before any Terraform resources are defined.

- [ ] T001 Create `infra/` and `infra/bootstrap/` directories in the repository root
- [ ] T002 [P] Create `infra/versions.tf` with `terraform { required_version = ">= 1.10" }` block, `required_providers` pinning `hashicorp/aws ~> 5.0`, and a `provider "aws"` block with `region = var.aws_region` and `default_tags { tags = { project = var.project_name, environment = var.environment, managed-by = "terraform" } }`
- [ ] T003 [P] Create `infra/variables.tf` with five variables: `project_name` (string, default `"cocktails"`), `environment` (string, default `"prod"`), `aws_region` (string, default `"us-east-1"`), `lambda_binary_path` (string, required — path to compiled `bootstrap.zip`), and `jwt_secret` (string, required, `sensitive = true`)
- [ ] T004 [P] Create `infra/.gitignore` with patterns: `.terraform/`, `*.tfstate`, `*.tfstate.backup`, `.terraform.lock.hcl`, `terraform.tfvars`, `env.json`, `*.zip`
- [ ] T005 [P] Create `infra/bootstrap/variables.tf` with `project_name` (string, default `"cocktails"`) and `aws_region` (string, default `"us-east-1"`) variables

---

## Phase 2: Foundational — Bootstrap Workspace (Blocking)

**Purpose**: Provision the one-time remote state S3 bucket. Must be complete before the main workspace can be initialized with a remote backend.

**⚠️ CRITICAL**: Run `cd infra/bootstrap && terraform init && terraform apply` exactly once before proceeding to Phase 3.

- [ ] T006 Create `infra/bootstrap/main.tf` with: `terraform { required_version = ">= 1.10" }` block, `required_providers` for `hashicorp/aws ~> 5.0`, `provider "aws"` with region and `default_tags` (project, environment, `managed-by = "terraform"`), `data "aws_caller_identity" "current" {}`, and a single `aws_s3_bucket "state"` with `bucket = "${var.project_name}-tf-state-${data.aws_caller_identity.current.account_id}"` and `force_destroy = false`. Add companion resources: `aws_s3_bucket_versioning` (enabled), `aws_s3_bucket_server_side_encryption_configuration` (AES256/SSE-S3), and `aws_s3_bucket_public_access_block` (all four `block_public_*` = true).
- [ ] T007 [P] Create `infra/bootstrap/outputs.tf` with a single output `state_bucket_name` referencing `aws_s3_bucket.state.bucket`
- [ ] T008 Create `infra/backend.tf` with `terraform { backend "s3" { bucket = "<REPLACE_WITH_BOOTSTRAP_OUTPUT>"; key = "cocktails/prod/terraform.tfstate"; region = "us-east-1"; use_lockfile = true } }`. Add a comment instructing the operator to replace the bucket placeholder with the `state_bucket_name` output from `infra/bootstrap/` before running `terraform init` in `infra/`. `use_lockfile = true` enables S3 native file locking (Terraform ≥ 1.10) — no DynamoDB table is required.

**Checkpoint**: After `terraform apply` in `infra/bootstrap/`, copy the `state_bucket_name` output into `infra/backend.tf`, then run `terraform init` in `infra/`.

---

## Phase 3: User Story 1 — Deploy Application to AWS (Priority: P1) 🎯 MVP

**Goal**: A single `terraform apply` provisions all application infrastructure (Lambda, API Gateway, DynamoDB, S3, CloudFront, CloudWatch) and outputs the API and frontend URLs.

**Independent Test**: Run `terraform apply` against an empty AWS account. Confirm exit code 0, `api_url` responds HTTP 200 to `GET /api/v1/recipes`, and `frontend_url` responds HTTP 200.

### Implementation for User Story 1

- [ ] T009 [US1] Add artifact S3 bucket to `infra/main.tf` as `module "artifact_bucket"` using source `terraform-aws-modules/s3-bucket/aws ~> 4.0`. Configure: `bucket = "${var.project_name}-artifacts-${var.environment}"`, versioning enabled, AES256 server-side encryption, all four `block_public_*` = true, and a lifecycle rule to expire noncurrent versions after keeping the 3 most recent.
- [ ] T010 [US1] Add DynamoDB recipes table to `infra/main.tf` as `module "recipes_table"` using source `terraform-aws-modules/dynamodb-table/aws ~> 4.0`. Configure: `name = "${var.project_name}-recipes"`, `billing_mode = "PAY_PER_REQUEST"`, `hash_key = "id"` (type `"S"`), `point_in_time_recovery_enabled = true`, `server_side_encryption_enabled = true`.
- [ ] T011 [US1] Add DynamoDB users table to `infra/main.tf` as `module "users_table"` using source `terraform-aws-modules/dynamodb-table/aws ~> 4.0`. Configure: `name = "${var.project_name}-users"`, `billing_mode = "PAY_PER_REQUEST"`, `hash_key = "id"` (type `"S"`), `point_in_time_recovery_enabled = true`, `server_side_encryption_enabled = true`.
- [ ] T012 [US1] Add an explicit CloudWatch log group to `infra/main.tf` as `aws_cloudwatch_log_group "lambda_logs"` with `name = "/aws/lambda/${var.project_name}-api-${var.environment}"` and `retention_in_days = 14`. Creating this resource explicitly (before Lambda) ensures the 14-day retention policy is applied — Lambda would otherwise auto-create a log group with infinite retention.
- [ ] T013 [US1] Add Lambda function to `infra/main.tf` as `module "lambda_function"` using source `terraform-aws-modules/lambda/aws ~> 7.0`. **Before writing**: verify the v7 module README for the exact parameters to use a pre-built zip and upload it to S3 — the combination is likely `create_package = false`, `local_existing_package = var.lambda_binary_path`, `store_on_s3 = true`, `s3_bucket = module.artifact_bucket.s3_bucket_id`, but parameter names may differ in v7 (check the module's CHANGELOG). Configure: `function_name = "${var.project_name}-api-${var.environment}"`, `runtime = "provided.al2023"`, `architectures = ["arm64"]`, `handler = "bootstrap"`, `memory_size = 256`, `timeout = 30`, `create_package = false`, `local_existing_package = var.lambda_binary_path`, `store_on_s3 = true`, `s3_bucket = module.artifact_bucket.s3_bucket_id`. Set `create_role = true`, `attach_policy_statements = true`, with policy statements granting `dynamodb:GetItem/PutItem/UpdateItem/DeleteItem/Query/Scan/BatchGetItem/BatchWriteItem` on both DynamoDB table ARNs (plus `/index/*` suffix) and `logs:CreateLogGroup/CreateLogStream/PutLogEvents` on `aws_cloudwatch_log_group.lambda_logs.arn`. Set environment variables: `STORE_BACKEND = "dynamodb"`, `RECIPES_TABLE = module.recipes_table.dynamodb_table_id`, `USERS_TABLE = module.users_table.dynamodb_table_id`, `JWT_SECRET = var.jwt_secret`.
- [ ] T014 [US1] Add HTTP API Gateway to `infra/main.tf` as `module "api_gateway"` using source `terraform-aws-modules/apigateway-v2/aws ~> 5.0`. Configure: `name = "${var.project_name}-api-${var.environment}"`, `protocol_type = "HTTP"`, `create_domain_name = false`. Set CORS: `allow_origins = ["*"]`, `allow_methods = ["GET","POST","PUT","DELETE","OPTIONS"]`, `allow_headers = ["Content-Type","Authorization"]`. Add a single `$default` route with an `AWS_PROXY` Lambda integration pointing to `module.lambda_function.lambda_function_arn`, `payload_format_version = "2.0"`, and a Lambda permission allowing API Gateway to invoke the function.
- [ ] T015 [US1] Add frontend origin S3 bucket to `infra/main.tf` as `module "frontend_bucket"` using source `terraform-aws-modules/s3-bucket/aws ~> 4.0`. Configure: `bucket = "${var.project_name}-frontend-${var.environment}"`, all four `block_public_*` = true. No static website hosting — CloudFront OAC handles object access.
- [ ] T016 [US1] Add CloudFront distribution to `infra/main.tf` as `module "cdn"` using source `terraform-aws-modules/cloudfront/aws ~> 3.4`. **Before writing**: check the terraform-aws-modules/cloudfront v3.4 module README to confirm the exact attribute names for OAC configuration (`create_origin_access_control`, `origin_access_control` block structure) — these changed between 3.x minor versions. Configure: `create_origin_access_control = true` with signing behavior `"always"` and protocol `"sigv4"` for the S3 origin. Set origin to `module.frontend_bucket.s3_bucket_regional_domain_name`. Configure: `default_root_object = "index.html"`, `price_class = "PriceClass_100"`, default cache behavior with `allowed_methods = ["GET","HEAD"]`, `compress = true`, `viewer_protocol_policy = "redirect-to-https"`, `http_version = "http2"`. Add custom error response: `error_code = 404`, `response_code = 200`, `response_page_path = "/index.html"` (required for SPA hash-based routing).
- [ ] T017 [US1] Add frontend S3 bucket policy to `infra/main.tf` as `aws_s3_bucket_policy "frontend_oac"`. The policy document must allow the CloudFront service principal (`cloudfront.amazonaws.com`) to `s3:GetObject` on all objects in the frontend bucket, conditioned on `aws:SourceArn` matching the CloudFront distribution ARN (`module.cdn.cloudfront_distribution_arn`). This implements Origin Access Control (OAC) — only CloudFront can read the bucket.
- [ ] T018 [US1] Add frontend asset upload to `infra/main.tf` as `resource "aws_s3_object" "frontend_assets"` with `for_each = fileset("${path.root}/../frontend/dist", "**")`. Set `bucket = module.frontend_bucket.s3_bucket_id`, `key = each.value`, `source = "${path.root}/../frontend/dist/${each.value}"`, `etag = filemd5("${path.root}/../frontend/dist/${each.value}")`. Map `content_type` by extension using a local value or `lookup`: `.html` → `text/html`, `.js` → `application/javascript`, `.css` → `text/css`, `.json` → `application/json`, `.svg` → `image/svg+xml`, `.png` → `image/png`, `.woff2` → `font/woff2`, default → `application/octet-stream`.
- [ ] T019 [US1] Add CloudFront invalidation to `infra/main.tf` as `resource "null_resource" "cdn_invalidation"`. Set `triggers = { etags = md5(join(",", [for o in aws_s3_object.frontend_assets : o.etag])) }` so it fires whenever any asset changes. Add a `local-exec` provisioner running `aws cloudfront create-invalidation --distribution-id ${module.cdn.cloudfront_distribution_id} --paths "/*"`. Add `depends_on = [aws_s3_object.frontend_assets]`.
- [ ] T020 [P] [US1] Create `infra/outputs.tf` with four outputs: `api_url` (value = `module.api_gateway.apigatewayv2_api_api_endpoint`), `frontend_url` (value = `"https://${module.cdn.cloudfront_distribution_domain_name}"`), `recipes_table_name` (value = `module.recipes_table.dynamodb_table_id`), `users_table_name` (value = `module.users_table.dynamodb_table_id`).

**Checkpoint**: User Story 1 is complete when `terraform apply` exits 0 and both output URLs return HTTP 200.

---

## Phase 4: User Story 2 — Test Infrastructure Locally with SAM CLI (Priority: P2)

**Goal**: `sam local start-api` invokes the Lambda function locally via Docker, without deploying to AWS.

**Independent Test**: After `terraform init` in `infra/`, run `sam local start-api --hook-name terraform`. Within 60 seconds, confirm `curl http://127.0.0.1:3000/api/v1/recipes` returns HTTP 200 (using SQLite backend).

### Implementation for User Story 2

- [ ] T021 [US2] Add SAM metadata resource to `infra/main.tf` as `resource "null_resource" "sam_metadata_lambda_function"` with a `triggers` map containing: `resource_name = "module.lambda_function.aws_lambda_function.this[0]"`, `resource_type = "ZIP_LAMBDA_FUNCTION"`, `original_source_code = "${path.root}/../backend"`, `built_output_path = "${path.root}/../backend/bootstrap.zip"`. Add `depends_on = [module.lambda_function]`. SAM CLI reads this resource during `terraform plan` output to discover and configure the local function emulation.
- [ ] T022 [P] [US2] Create `infra/samconfig.toml` with content:
  ```toml
  version = 0.1

  [default.global.parameters]
  hook_name = "terraform"
  ```
  This allows `sam local start-api` (without `--hook-name`) to discover the Terraform hook automatically.
- [ ] T023 [P] [US2] Create `infra/env.json.example` documenting the SAM local environment variable format. Include a top-level key matching the function name (e.g., `"cocktails-api-prod"`) with two variants documented in comments: the SQLite offline variant (`STORE_BACKEND=sqlite`, `DB_PATH=/tmp/cocktails_local.db`, `JWT_SECRET=local-dev-secret`) and the DynamoDB variant (`STORE_BACKEND=dynamodb`, `RECIPES_TABLE=cocktails-recipes`, `USERS_TABLE=cocktails-users`, `JWT_SECRET=local-dev-secret`). Operators copy this to `env.json` (gitignored) and pass `--env-vars env.json` to `sam local start-api`.

**Checkpoint**: User Story 2 is complete when SAM local environment starts in under 60 seconds and handles one successful API request.

---

## Phase 5: User Story 3 — Tear Down Infrastructure (Priority: P3)

**Goal**: `terraform destroy` removes all project-tagged resources cleanly, verified by the AWS console tag filter `project=cocktails` returning zero resources in us-east-1.

**Independent Test**: After a US1 deployment, run `cd infra && terraform destroy`. Confirm exit 0 and zero resources under the tag filter.

### Implementation for User Story 3

- [ ] T024 [US3] Audit `infra/main.tf` to confirm the `default_tags` block in `infra/versions.tf` (T002) propagates `project`, `environment`, and `managed-by` tags to all resources. For any resource that does not inherit `default_tags` (notably `aws_cloudwatch_log_group "lambda_logs"` and any `null_resource`), add an explicit `tags` block. CloudFront distributions require tags set directly on the resource — confirm the CloudFront module accepts a `tags` input and passes it through.
- [ ] T025 [P] [US3] Confirm `infra/bootstrap/main.tf` provider block (T006) includes `default_tags` with `project = var.project_name`, `environment = var.environment` (add this variable if missing), and `managed-by = "terraform"` so the state bucket itself appears in the tag filter verify step.

**Checkpoint**: User Story 3 is complete when all resources carry the three required tags and `terraform destroy` exits cleanly.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Operator convenience, static analysis, and documentation.

- [ ] T026 [P] Create `infra/terraform.tfvars.example` with example values for all five variables, each with an inline comment explaining the expected value (e.g., `lambda_binary_path = "../backend/bootstrap.zip" # pre-built arm64 binary zip`, `jwt_secret = "change-me-in-production" # minimum 32 chars recommended`)
- [ ] T027 [P] Create `infra/.terraformignore` to exclude large directories from uploads to Terraform Cloud/Enterprise (not needed for local runs, but good practice): `.git`, `specs/`, `frontend/node_modules/`, `backend/` (source code — only the compiled zip is deployed), `*.md`
- [ ] T028 Run `terraform validate` in `infra/bootstrap/` (after `terraform init`) and in `infra/` (after populating `backend.tf` and running `terraform init`). Fix any reported configuration errors. Note: `terraform init` in `infra/` requires a real S3 bucket from Phase 2; to validate without deploying, use `terraform init -backend=false` to skip remote state initialization.
- [ ] T029 [P] Run `checkov --directory infra/ --framework terraform` and review findings. Add inline `# checkov:skip=<RULE_ID>:<justification>` comments for any rules that are intentionally suppressed (e.g., `CKV_AWS_18` for S3 access logging — not required at this scale, `CKV2_AWS_62` for S3 event notifications, `CKV_AWS_86` for CloudFront logging). Document each suppression rationale.
- [ ] T030 [P] Create `infra/README.md` with step-by-step deployment instructions covering: (1) prerequisites checklist from operator-interface.md, (2) one-time bootstrap (`cd infra/bootstrap && terraform init && terraform apply`, copy `state_bucket_name` into backend.tf), (3) deploy (`cd infra && terraform init && terraform apply -var-file=terraform.tfvars`), (4) verify (curl api_url and frontend_url), (5) SAM local test (`sam local start-api`, optionally with `--env-vars env.json`), (6) destroy (`terraform destroy`). Reference `terraform.tfvars.example` for variable format. This task satisfies SC-004 ("operator can deploy by following README instructions in under 30 minutes").

---

## Dependencies & Execution Order

### Phase Dependencies

- **Pre-Implementation Baseline (Phase 0)**: No dependencies — run before any files are created
- **Setup (Phase 1)**: Depends on Phase 0 completion — all tasks can run immediately after T000
- **Foundational (Phase 2)**: Depends on Phase 1 completion — BLOCKS main workspace initialization
- **User Story 1 (Phase 3)**: Depends on Phase 2 (backend.tf must have a valid bucket name) — core application
- **User Story 2 (Phase 4)**: Depends on Phase 3 (sam_metadata references the Lambda module from US1)
- **User Story 3 (Phase 5)**: Depends on Phase 3 (tagging audit requires all resources to be defined)
- **Polish (Phase 6)**: Can start after Phase 3 is complete

### Within User Story 1 — Resource Build Order in `infra/main.tf`

Resources in main.tf are added sequentially (same file). Internal dependency order:

1. **T009** (artifact S3 bucket) — no dependencies within main.tf
2. **T010, T011** (DynamoDB tables) — no dependencies within main.tf
3. **T012** (CloudWatch log group) — no dependencies within main.tf
4. **T013** (Lambda) — depends on T009 (artifact bucket), T010+T011 (table ARNs for IAM policy), T012 (log group ARN)
5. **T014** (API Gateway) — depends on T013 (Lambda ARN for integration)
6. **T015** (frontend S3 bucket) — no dependencies within main.tf
7. **T016** (CloudFront) — depends on T015 (origin domain)
8. **T017** (bucket policy) — depends on T015 + T016 (OAC ARN)
9. **T018** (asset upload) — depends on T015 (bucket ID)
10. **T019** (CloudFront invalidation) — depends on T016 (distribution ID) + T018 (asset ETags)
11. **T020** (outputs.tf) — different file; can be written after T014 and T016 are in main.tf

### Parallel Opportunities

```bash
# Phase 1 — all tasks can run in parallel (different files):
T002: infra/versions.tf
T003: infra/variables.tf
T004: infra/.gitignore
T005: infra/bootstrap/variables.tf

# Phase 2 — T006 first, then T007 and T008 in parallel:
T006: infra/bootstrap/main.tf   ← must complete first
T007: infra/bootstrap/outputs.tf   # parallel with T008
T008: infra/backend.tf             # parallel with T007

# Phase 3 — main.tf tasks sequential; outputs.tf in parallel with T014/T016+:
T009 → T010 → T011 → T012 → T013 → T014 (main.tf chain 1)
T015 → T016 → T017 → T018 → T019 (main.tf chain 2)
T020: infra/outputs.tf            # parallel with T014 onward

# Phase 4 — T022 and T023 in parallel with T021:
T021: null_resource in main.tf
T022: infra/samconfig.toml        # parallel with T021
T023: infra/env.json.example      # parallel with T021

# Phase 6 — all polish tasks can run in parallel:
T026, T027, T029 in parallel; T028 requires T005-T020 complete
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Bootstrap (run `terraform apply` in `infra/bootstrap/`, copy bucket name)
3. Complete Phase 3: User Story 1 (all application infrastructure)
4. **STOP and VALIDATE**: Run `terraform apply` and confirm both URLs respond HTTP 200
5. Application is live — this is the minimum deliverable (SC-001, SC-002, SC-004)

### Incremental Delivery

1. Phase 1 + Phase 2 → Remote state backend ready for team use
2. Phase 3 (US1) → Full application deployed (MVP — satisfies FR-001 through FR-011, FR-013)
3. Phase 4 (US2) → Local testing without AWS costs (satisfies FR-006, SC-005)
4. Phase 5 (US3) → Tagging verified, clean destroy confirmed (satisfies FR-007, SC-003)
5. Phase 6 → checkov clean, docs complete (satisfies SC-004, SC-006)

---

## Notes

- All tasks writing to `infra/main.tf` are sequential within their chain — that file grows incrementally
- `infra/outputs.tf` (T020) is a separate file and can be written after its referenced modules exist in main.tf
- `infra/backend.tf` (T008) contains a bucket name placeholder — the operator must update it after running the bootstrap step before `terraform init` in `infra/` will succeed; for syntax-only validation, use `terraform init -backend=false`
- `infra/env.json` is gitignored — `env.json.example` (T023) is the version-controlled template operators copy and customize
- Lambda module output path for sam_metadata: the exact resource address depends on the module version — verify with `terraform state list` after a plan if `module.lambda_function.aws_lambda_function.this[0]` does not match
- checkov suppressions (T029) are expected at this scale — each suppression must include a justification comment
