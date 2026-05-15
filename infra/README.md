# Cocktails — Infrastructure

Terraform workspace for deploying the Cocktails application to AWS.

**Stack**: Lambda (Go, arm64) + API Gateway v2 + DynamoDB + S3 + CloudFront + CloudWatch

---

## Prerequisites

Before running any command, ensure the following are installed and configured:

1. **AWS credentials** — `~/.aws/credentials`, environment variables, or an SSO profile with permissions to create Lambda, API Gateway, DynamoDB, S3, CloudFront, IAM, and CloudWatch resources.
2. **Terraform ≥ 1.10** — required for S3 native file locking (`use_lockfile = true`).
3. **Go backend binary** (for deploy):
   ```bash
   cd backend
   GOOS=linux GOARCH=arm64 go build -o bootstrap ./cmd/lambda
   zip bootstrap.zip bootstrap
   ```
4. **Frontend assets** (for deploy):
   ```bash
   cd frontend && npm run build
   ```
5. **SAM CLI + Docker** (for local testing only).

---

## Step 1: Bootstrap (one-time only)

Provisions the S3 bucket used for Terraform remote state. Run this exactly once per AWS account.

```bash
cd infra/bootstrap
terraform init
terraform apply
```

Copy the `state_bucket_name` output value into `infra/backend.tf`:

```hcl
backend "s3" {
  bucket = "<paste state_bucket_name here>"
  ...
}
```

---

## Step 2: Deploy

```bash
cd infra
cp terraform.tfvars.example terraform.tfvars   # fill in jwt_secret and lambda_binary_path
terraform init
terraform apply -var-file=terraform.tfvars
```

On success, Terraform prints the application URLs:

| Output | Description |
|--------|-------------|
| `api_url` | Backend API endpoint |
| `frontend_url` | CloudFront URL for the SPA |
| `recipes_table_name` | DynamoDB recipes table |
| `users_table_name` | DynamoDB users table |

---

## Step 3: Verify

```bash
# Check the backend API
curl "$(terraform output -raw api_url)/api/v1/recipes"   # expect HTTP 200

# Check the frontend
curl -I "$(terraform output -raw frontend_url)"           # expect HTTP 200
```

---

## Step 4: SAM Local Test (optional)

Test the Lambda function locally without incurring AWS costs.

```bash
cd infra

# Copy and optionally edit the environment file
cp env.json.example env.json
# Default uses SQLite (offline). Edit env.json to switch to DynamoDB:
# { "cocktails-api-prod": { "STORE_BACKEND": "dynamodb", "RECIPES_TABLE": "cocktails-recipes", ... } }

sam local start-api --env-vars env.json
# In a separate terminal:
curl http://127.0.0.1:3000/api/v1/recipes   # expect HTTP 200
```

SAM CLI reads the `samconfig.toml` (`hook_name = "terraform"`) and the `null_resource.sam_metadata_lambda_function` triggers to discover the function.

---

## Step 5: Destroy

Removes all application resources from AWS. **This is irreversible.**

```bash
cd infra
terraform destroy -var-file=terraform.tfvars
```

To verify: in the AWS console, filter resources by tag `project=cocktails` — the list should be empty.

To remove the state backend (run only after the main workspace is destroyed):

```bash
cd infra/bootstrap
terraform destroy
```

---

## Variables

See `terraform.tfvars.example` for all available variables and their descriptions.

| Variable | Default | Required |
|----------|---------|----------|
| `project_name` | `"cocktails"` | No |
| `environment` | `"prod"` | No |
| `aws_region` | `"us-east-1"` | No |
| `lambda_binary_path` | — | **Yes** |
| `jwt_secret` | — | **Yes** |

---

## Tagging

All resources are tagged with:

| Tag | Value |
|-----|-------|
| `project` | `var.project_name` |
| `environment` | `var.environment` |
| `managed-by` | `terraform` |

Tags are applied automatically via the provider `default_tags` block in `versions.tf`.
