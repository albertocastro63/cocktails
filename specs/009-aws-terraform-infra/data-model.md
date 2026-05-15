# Data Model: AWS Infrastructure Deployment

**Feature**: 009-aws-terraform-infra
**Date**: 2026-05-15

This document describes the Terraform resource entities, their attributes, and relationships. "Data model" here means the infrastructure object graph, not application data.

---

## Infrastructure Entities

### 1. Bootstrap State Bucket

**Terraform resource**: `aws_s3_bucket` (in `infra/bootstrap/`)
**Module**: direct resource (not a module; bootstrap is minimal by design)

| Attribute | Value |
|-----------|-------|
| `bucket` | `cocktails-tf-state-<account_id>` (unique per account) |
| `versioning` | enabled |
| `server_side_encryption` | AES256 (SSE-S3) |
| `block_public_acls` | true |
| `block_public_policy` | true |
| `force_destroy` | false |

**State locking**: Handled natively by the S3 backend via `use_lockfile = true` (Terraform ≥ 1.10). Terraform writes a `.tflock` file to the state bucket when a lock is acquired. No DynamoDB table is required.

---

### 2. Artifact Storage Bucket

**Module**: `terraform-aws-modules/s3-bucket/aws ~> 4.0`
**Purpose**: Stores the compiled Lambda deployment package (zip) before function creation.

| Attribute | Value |
|-----------|-------|
| `bucket` | `cocktails-artifacts-<environment>` |
| `versioning` | enabled (keep last 3 versions via lifecycle rule) |
| `block_public_acls` | true |
| `block_public_policy` | true |
| `server_side_encryption_configuration` | AES256 |

---

### 3. Lambda Function (API Backend)

**Module**: `terraform-aws-modules/lambda/aws ~> 7.0`
**Purpose**: Serves all backend API requests.

| Attribute | Value |
|-----------|-------|
| `function_name` | `cocktails-api-<environment>` |
| `runtime` | `provided.al2023` |
| `architectures` | `["arm64"]` |
| `handler` | `bootstrap` |
| `memory_size` | `256` MB |
| `timeout` | `30` seconds |
| `source_code_artifact` | pre-built zip uploaded to artifact bucket |
| `create_role` | true (inline IAM role) |
| `environment_variables` | `STORE_BACKEND=dynamodb`, `RECIPES_TABLE`, `USERS_TABLE`, `JWT_SECRET` (from SSM or env) |
| `sam_metadata` | enabled (`use_serverless_terraform = true` or companion `null_resource`) |

**SAM metadata resource** (companion `null_resource`):

| Trigger key | Value |
|-------------|-------|
| `resource_name` | `"module.lambda_function.aws_lambda_function.this[0]"` (fully-qualified module resource path) |
| `resource_type` | `"ZIP_LAMBDA_FUNCTION"` |
| `original_source_code` | `"../backend"` (Go source root) |
| `built_output_path` | `"../backend/bootstrap.zip"` |

---

### 4. HTTP API Gateway

**Module**: `terraform-aws-modules/apigateway-v2/aws ~> 5.0`
**Purpose**: Public-facing HTTP endpoint routing all API requests to the Lambda function.

| Attribute | Value |
|-----------|-------|
| `name` | `cocktails-api-<environment>` |
| `protocol_type` | `HTTP` |
| `cors_configuration` | `allow_origins = ["*"]`, `allow_methods = ["GET","POST","PUT","DELETE","OPTIONS"]`, `allow_headers = ["Content-Type","Authorization"]` |
| `create_domain_name` | false (use auto-generated execute-api URL) |
| `routes` | `$default` → Lambda integration (catch-all proxy) |
| `integration_type` | `AWS_PROXY` |
| `payload_format_version` | `2.0` |

---

### 5. DynamoDB — Recipes Table

**Module**: `terraform-aws-modules/dynamodb-table/aws ~> 4.0`

| Attribute | Value |
|-----------|-------|
| `name` | `cocktails-recipes` |
| `billing_mode` | `PAY_PER_REQUEST` |
| `hash_key` | `id` (type: String) |
| `point_in_time_recovery_enabled` | true |
| `server_side_encryption_enabled` | true |

---

### 6. DynamoDB — Users Table

**Module**: `terraform-aws-modules/dynamodb-table/aws ~> 4.0`

| Attribute | Value |
|-----------|-------|
| `name` | `cocktails-users` |
| `billing_mode` | `PAY_PER_REQUEST` |
| `hash_key` | `id` (type: String) |
| `point_in_time_recovery_enabled` | true |
| `server_side_encryption_enabled` | true |

---

### 7. Lambda IAM Role

**Managed by**: Lambda module internally (`create_role = true`)
**Policy statements** (attached via `attach_policy_statements = true`):

| Action | Resource |
|--------|----------|
| `dynamodb:GetItem`, `PutItem`, `UpdateItem`, `DeleteItem`, `Query`, `Scan`, `BatchGetItem`, `BatchWriteItem` | `cocktails-recipes` ARN + `/index/*` |
| same | `cocktails-users` ARN + `/index/*` |
| `logs:CreateLogGroup`, `logs:CreateLogStream`, `logs:PutLogEvents` | CloudWatch log group ARN |

---

### 8. CloudWatch Log Group

**Terraform resource**: `aws_cloudwatch_log_group` (explicit, not auto-created)

| Attribute | Value |
|-----------|-------|
| `name` | `/aws/lambda/cocktails-api-<environment>` |
| `retention_in_days` | `14` |

---

### 9. Frontend Origin Bucket

**Module**: `terraform-aws-modules/s3-bucket/aws ~> 4.0`
**Purpose**: Stores compiled frontend static assets served via CloudFront.

| Attribute | Value |
|-----------|-------|
| `bucket` | `cocktails-frontend-<environment>` |
| `block_public_acls` | true |
| `block_public_policy` | true |
| OAC bucket policy | Allows `s3:GetObject` from CloudFront OAC principal only |

---

### 10. CloudFront Distribution

**Module**: `terraform-aws-modules/cloudfront/aws ~> 3.4`
**Purpose**: Globally distributes frontend assets; enforces HTTPS.

| Attribute | Value |
|-----------|-------|
| `create_origin_access_control` | true |
| `origin` | Frontend origin bucket (S3, OAC) |
| `default_root_object` | `index.html` |
| `price_class` | `PriceClass_100` (US + EU only — cost optimised) |
| `viewer_certificate` | CloudFront default certificate (HTTPS on cloudfront.net domain) |
| `default_cache_behavior` | `GET,HEAD` cached; `compress = true` |
| `custom_error_response` | 404 → `index.html`, status 200 (SPA hash routing) |
| `http_version` | `http2` |

---

## Resource Relationships

```
infra/bootstrap/
  └── S3 State Bucket (S3 native file lock via .tflock) ──► infra/ backend.tf (references)

infra/
  Artifact S3 Bucket ──────────► Lambda Function (deployment package source)
  Lambda Function ──────────────► API Gateway (integration target)
  Lambda Function ──────────────► DynamoDB Recipes Table (IAM policy)
  Lambda Function ──────────────► DynamoDB Users Table (IAM policy)
  Lambda Function ──────────────► CloudWatch Log Group (log destination)
  Lambda IAM Role ──────────────► Lambda Function (execution role)
  Frontend S3 Bucket ───────────► CloudFront Distribution (origin)
  CloudFront OAC ───────────────► Frontend S3 Bucket (access policy)
```

## Terraform Input Variables

| Variable | Type | Description |
|----------|------|-------------|
| `project_name` | string | Short identifier prefix (default: `"cocktails"`) |
| `environment` | string | Deployment environment label (default: `"prod"`) |
| `aws_region` | string | Target region (default: `"us-east-1"`) |
| `lambda_binary_path` | string | Path to compiled Go binary zip (required) |
| `jwt_secret` | string | JWT signing secret (sensitive; not logged) |

## Terraform Outputs

| Output | Description |
|--------|-------------|
| `api_url` | HTTPS URL of the API Gateway endpoint |
| `frontend_url` | HTTPS URL of the CloudFront distribution |
| `recipes_table_name` | DynamoDB recipes table name |
| `users_table_name` | DynamoDB users table name |
