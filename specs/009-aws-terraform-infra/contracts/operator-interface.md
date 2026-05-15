# Operator Interface Contract

**Feature**: 009-aws-terraform-infra
**Date**: 2026-05-15

This document defines the interface between the infrastructure code and the operator who deploys, tests, and tears down the application. "Interface" here means the commands, inputs, and outputs the operator interacts with.

---

## Prerequisites (Operator Responsibilities)

Before running any command the operator MUST ensure:

1. **AWS credentials** configured (`~/.aws/credentials`, env vars, or SSO profile) with permissions to create Lambda, API Gateway, DynamoDB, S3, CloudFront, IAM, and CloudWatch resources.
2. **Terraform ≥ 1.10** installed. (`use_lockfile = true` in the S3 backend requires Terraform 1.10 or later.)
3. **SAM CLI** installed (for local testing only).
4. **Docker** installed and running (for `sam local` commands only).
5. **Go backend binary** compiled: `cd backend && GOOS=linux GOARCH=arm64 go build -o bootstrap ./cmd/lambda && zip bootstrap.zip bootstrap`
6. **Frontend assets** built: `cd frontend && npm run build`

---

## Command Contract 1: Bootstrap (One-Time)

**Purpose**: Provision the remote Terraform state backend before the first deploy. Creates only an S3 bucket; state locking uses Terraform's native S3 file lock (`use_lockfile = true`) — no DynamoDB table is required.

**Command**:
```bash
cd infra/bootstrap
terraform init
terraform apply
```

**Inputs**: None (uses AWS credentials from environment).

**Outputs**:
| Output | Description |
|--------|-------------|
| `state_bucket_name` | Name of the S3 bucket created for Terraform state |

**Post-condition**: Copy `state_bucket_name` into `infra/backend.tf` if not already templated. Run this command EXACTLY ONCE. Re-running is idempotent. The `.tflock` file used for locking is written automatically by Terraform during `apply`/`destroy` — no manual setup needed.

**Failure modes**:
- Insufficient IAM permissions → error lists missing action; operator must update IAM policy.
- S3 bucket name already exists globally → change the `project_name` variable or add a unique suffix.

---

## Command Contract 2: Deploy

**Purpose**: Provision or update all application infrastructure in `us-east-1`.

**Command**:
```bash
cd infra
terraform init          # first time or after provider changes
terraform apply
```

**Required inputs** (via `terraform.tfvars` or `-var` flags):
| Variable | Type | Required | Description |
|----------|------|----------|-------------|
| `lambda_binary_path` | string | Yes | Path to compiled `bootstrap.zip` |
| `jwt_secret` | string | Yes | JWT signing secret (mark sensitive) |
| `project_name` | string | No | Default: `"cocktails"` |
| `environment` | string | No | Default: `"prod"` |
| `aws_region` | string | No | Default: `"us-east-1"` |

**Outputs** (printed on success):
| Output | Example |
|--------|---------|
| `api_url` | `https://abc123.execute-api.us-east-1.amazonaws.com` |
| `frontend_url` | `https://d1234abcd.cloudfront.net` |
| `recipes_table_name` | `cocktails-recipes` |
| `users_table_name` | `cocktails-users` |

**Success condition**: `terraform apply` exits 0; `api_url` responds HTTP 200 to `GET /api/v1/recipes`; `frontend_url` responds HTTP 200.

**Idempotency**: Re-running with the same inputs makes no changes. Running after a partial failure recovers cleanly.

---

## Command Contract 3: SAM Local Test

**Purpose**: Invoke API functions locally without deploying to AWS.

**Command**:
```bash
cd infra
sam local start-api --hook-name terraform
# or, with samconfig.toml present:
sam local start-api
```

**Prerequisites**: Terraform project must be initialised (`terraform init` run at least once). Lambda binary must be built locally (native OS binary acceptable for local invocation; arm64 binary required for deployment).

**Behaviour**: SAM runs `terraform plan`, discovers Lambda functions via `sam_metadata` resources, mounts the API at `http://127.0.0.1:3000`, and routes requests to locally-executed function binaries via Docker.

**Inputs**: None beyond environment (AWS credentials are read for DynamoDB calls if `STORE_BACKEND=dynamodb`; set `DB_PATH=cocktails_local.db` + `STORE_BACKEND=sqlite` to run fully offline).

**Success condition**: `curl http://127.0.0.1:3000/api/v1/recipes` returns HTTP 200 within 60 seconds of startup.

---

## Command Contract 4: Destroy

**Purpose**: Remove all application infrastructure from AWS, leaving no billable resources.

**Command**:
```bash
cd infra
terraform destroy
```

**Inputs**: Same variables as Deploy (Terraform reads from `terraform.tfvars`).

**Outputs**: None.

**Success condition**: `terraform destroy` exits 0. No project-tagged resources remain in `us-east-1` (verifiable via AWS console tag filter `project=cocktails`).

**Warning**: This command is irreversible. DynamoDB PITR allows point-in-time recovery of table data for up to 35 days after deletion if recovery is initiated before the table is fully purged — but the table itself is gone.

**Partial destroy**: Bootstrap resources (`infra/bootstrap/`) are NOT destroyed by `terraform destroy` in the main workspace. To remove the state backend, run `terraform destroy` inside `infra/bootstrap/` separately, AFTER the main workspace has been destroyed.

---

## Tagging Contract

All AWS resources MUST be tagged with:

| Tag key | Value |
|---------|-------|
| `project` | `cocktails` |
| `environment` | value of `var.environment` |
| `managed-by` | `terraform` |

This enables cost attribution and the "zero resources" verify step in Command Contract 4.
