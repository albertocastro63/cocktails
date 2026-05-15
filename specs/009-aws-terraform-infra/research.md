# Research: AWS Infrastructure Deployment

**Feature**: 009-aws-terraform-infra
**Date**: 2026-05-15

---

## Decision 1: Terraform + SAM Integration Mechanism

**Decision**: Use `null_resource "sam_metadata_*"` blocks (SAM metadata resources) alongside `aws_lambda_function` resources to enable SAM CLI local invocation of Terraform-managed functions.

**Rationale**: AWS SAM CLI's Terraform hook does not read `aws_lambda_function` attributes directly. Instead, SAM scans the Terraform plan JSON for `null_resource` blocks whose names begin with `sam_metadata_`. SAM reads the `triggers` map on those resources to locate each function's source and built artifact, then generates an internal `.aws-sam/` metadata file used by `sam local` commands.

**Required `triggers` keys** on the `null_resource`:
- `resource_name` — e.g. `"module.lambda_function.aws_lambda_function.this[0]"` when using the `terraform-aws-modules/lambda` module (use the fully-qualified module resource path, not a bare `aws_lambda_function.<name>` address)
- `resource_type` — `"ZIP_LAMBDA_FUNCTION"`
- `original_source_code` — path to Go source directory
- `built_output_path` — path to the compiled binary/zip

**`samconfig.toml` entry** (simplifies CLI usage):
```toml
[default.global.parameters]
hook_name = "terraform"
```

**Local invocation command**:
```bash
sam local start-api --hook-name terraform
```
SAM runs `terraform init` + `terraform plan -out=<json>`, parses the output, and mounts the API.

**Alternatives considered**: SAM `template.yaml` — rejected because it duplicates infrastructure definitions already in Terraform, creating a maintenance split-brain.

**Known limitations**:
- Lambda functions linked to multiple layers are not supported
- Terraform local variables used to define inter-resource links break SAM discovery
- Functions referenced in an `aws_api_gateway_rest_api.body` (OpenAPI body) are not discovered — use `aws_api_gateway_integration` resources instead (not applicable here since we use API Gateway v2)
- Docker must be installed for all `sam local *` commands

---

## Decision 2: Terraform Module Selection (serverless.tf / terraform-aws-modules)

**Decision**: Use the following registry modules, all from the `terraform-aws-modules` organization (which is the canonical serverless.tf module library):

| Component | Module | Version |
|-----------|--------|---------|
| Lambda function | `terraform-aws-modules/lambda/aws` | `~> 7.0` |
| HTTP API Gateway (v2) | `terraform-aws-modules/apigateway-v2/aws` | `~> 5.0` |
| DynamoDB tables | `terraform-aws-modules/dynamodb-table/aws` | `~> 4.0` |
| S3 (artifacts + frontend) | `terraform-aws-modules/s3-bucket/aws` | `~> 4.0` |
| CloudFront | `terraform-aws-modules/cloudfront/aws` | `~> 3.4` |

**IAM role**: Handled internally by the Lambda module (`create_role = true`, `attach_policy_statements = true`). No separate IAM module needed.

**SAM support in Lambda module**: The `terraform-aws-modules/lambda` module natively emits SAM-compatible metadata. Setting `use_serverless_terraform = true` on the module (or configuring the `sam_metadata` resources manually alongside it) enables SAM CLI discovery.

**Rationale**: All modules are battle-tested, widely adopted, and directly referenced from serverless.tf project examples. Using them reduces bespoke HCL surface area and ensures sensible defaults (encryption, versioning, least-privilege).

**Alternatives considered**: Writing inline `aws_lambda_function`, `aws_apigatewayv2_api`, etc. resources directly — rejected because it re-implements defaults already encoded in the modules and increases maintenance burden.

---

## Decision 3: Terraform Project Structure

**Decision**: Single root workspace in `infra/` with a dedicated `infra/bootstrap/` subdirectory for the one-time state backend setup.

```
infra/
├── bootstrap/          # one-time state backend provisioning
│   ├── main.tf         # S3 bucket + DynamoDB lock table
│   ├── outputs.tf      # bucket name, table name
│   └── variables.tf
├── backend.tf          # S3 remote state config (references bootstrap outputs)
├── main.tf             # Lambda, API Gateway, DynamoDB, CloudFront, S3
├── variables.tf        # region, project_name, environment
├── outputs.tf          # api_url, frontend_url
├── versions.tf         # provider + module version pins
└── samconfig.toml      # hook_name = "terraform"
```

**Rationale**: A single workspace keeps state management simple for a single-environment deployment. The bootstrap subdirectory addresses the chicken-and-egg problem (you need state storage before you can use remote state) without coupling it to the main workspace.

**Alternatives considered**:
- Separate workspaces per logical tier (backend/frontend) — rejected; increases operational complexity for no benefit at this scale.
- Terragrunt — rejected; adds a non-trivial dependency for a single-environment use case.

---

## Decision 4: Lambda Runtime and Architecture

**Decision**: `runtime = "provided.al2023"`, `architectures = ["arm64"]`, `handler = "bootstrap"`.

**Rationale**: Go on AWS Lambda uses the custom runtime (`provided.al2023`). The handler name `bootstrap` is the required entry point for the `provided.al2023` runtime. arm64 (Graviton2) is 20% cheaper and typically faster for Go workloads than x86_64, and is supported in `us-east-1`.

**Build prerequisite**: The binary must be compiled `GOOS=linux GOARCH=arm64` before running `terraform apply`. The infrastructure code does not build the binary (per FR-spec assumption).

**Alternatives considered**: `go1.x` runtime — deprecated; `provided.al2` — superseded by `al2023`.

---

## Decision 5: DynamoDB Table Design

**Decision**: `PAY_PER_REQUEST` billing for both tables; no provisioned throughput. `point_in_time_recovery_enabled = true`.

**Rationale**: At cocktails-app scale (< 1000 users, < 10k recipes), PAY_PER_REQUEST eliminates capacity planning overhead and is cost-optimal. PITR adds minimal cost and provides a safety net for accidental deletes during testing.

**Table schemas** (matching existing store implementation):
- `cocktails-recipes`: hash key `id` (S)
- `cocktails-users`: hash key `id` (S)

**Alternatives considered**: Provisioned throughput — rejected; over-engineering for this scale.

---

## Decision 6: Frontend Static Hosting

**Decision**: S3 bucket (private, OAC-accessible) + CloudFront distribution with Origin Access Control. SPA support via custom error response (404 → `index.html`, HTTP 200).

**Rationale**: CloudFront + S3 is the AWS-recommended pattern for static SPAs. OAC (Origin Access Control) is the current best practice, superseding OAI (Origin Access Identity). The SPA error response rule is required because the frontend uses hash-based routing (no server-side routing needed).

**Alternatives considered**: S3 static website hosting (public bucket) — rejected; public buckets are a security anti-pattern. Amplify — rejected; adds a managed service dependency not needed here.

---

## Decision 7: CloudWatch Log Retention

**Decision**: 14-day log retention for all Lambda log groups, configured explicitly in Terraform (`aws_cloudwatch_log_group` resource with `retention_in_days = 14`).

**Rationale**: Matches the clarification answer (FR-013). Explicit retention prevents unbounded CloudWatch costs. 14 days is sufficient for post-deploy debugging without long-term storage overhead.

---

## Decision 8: Remote State Backend Bootstrap Strategy

**Decision**: `infra/bootstrap/` is a standalone Terraform configuration with local state (no remote backend). It creates only one resource: an S3 bucket for Terraform state (versioning enabled, server-side encryption, public access blocked).

State locking uses Terraform's native S3 file locking (`use_lockfile = true` in `backend.tf`), available since Terraform 1.10. When a lock is acquired, Terraform writes a `.tflock` file to the state bucket alongside the state file. No DynamoDB table is required.

**Minimum Terraform version**: 1.10 (for `use_lockfile` support in the S3 backend).

**Operator workflow**:
```bash
cd infra/bootstrap && terraform init && terraform apply  # one-time: creates S3 state bucket only
cd infra && terraform init && terraform apply            # every deploy
```

**`backend.tf`** (main workspace):
```hcl
terraform {
  backend "s3" {
    bucket       = "<state_bucket_name>"
    key          = "cocktails/prod/terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true   # S3 native file lock — no DynamoDB needed
  }
}
```

**Rationale**: S3 native locking (Terraform ≥ 1.10) eliminates the need for a separate DynamoDB table, simplifying the bootstrap step and reducing the AWS resource count. The `.tflock` file is written atomically to S3 and provides the same mutual-exclusion guarantee as DynamoDB locking for this single-environment use case.

**Alternatives considered**:
- DynamoDB lock table — rejected per user requirement; adds an extra AWS resource and IAM permission scope.
- Manual `aws s3api create-bucket` — rejected (Option A in clarification); error-prone and undocumented.
- Self-provisioning main workspace — rejected (Option C); creates a circular dependency.
