# Quickstart: PR Preview Environments

**Branch**: `021-pr-preview-environments` | **Date**: 2026-05-28

This guide covers how to manually trigger a preview deployment for testing during implementation, and how to verify the full environment works end-to-end.

---

## Prerequisites

- AWS CLI configured with credentials that have the required permissions (see below)
- `PR_NUMBER` set to a test PR number (e.g., `99`)
- All environment variables from the CI script contract set in your shell
- Go 1.22+ and Node.js 24+ installed locally

## Manual Preview Deployment (for testing the deploy script)

```bash
# 1. Build Lambda binary
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/bootstrap ./cmd/lambda/
cd ..

# 2. Build frontend with PR-specific API prefix
cd frontend
VITE_API_PATH_PREFIX=/api/pr-99 npm run build
cd ..

# 3. Run deploy script
export PR_NUMBER=99
export AWS_REGION=us-east-1
export LAMBDA_ROLE_ARN=arn:aws:iam::<account-id>:role/cocktails-preview-lambda-role
export API_GATEWAY_ID=<existing-api-gw-id>
export FRONTEND_BUCKET=cocktails-prod-frontend
export CLOUDFRONT_DISTRIBUTION_ID=EX7HUB6P225MV
export JWT_SECRET=<from-env>
export PROD_RECIPES_TABLE=cocktails-recipes
export PROD_USERS_TABLE=cocktails-users
export PROD_FAVORITES_TABLE=cocktails-favorites

.github/scripts/preview-deploy.sh
```

Expected output:
```
[preview-deploy] Creating DynamoDB tables for pr-99...
[preview-deploy] Seeding recipes table from production...
[preview-deploy] Seeding users table from production...
[preview-deploy] Creating Lambda function cocktails-pr-99-api...
[preview-deploy] Adding API Gateway route ANY /api/pr-99/{proxy+}...
[preview-deploy] Uploading frontend assets to s3://cocktails-prod-frontend/pr-99/...
[preview-deploy] ✓ Preview environment ready: https://cocktails.albertomcastro.com/pr-99/
```

## Verifying the Preview Environment

```bash
# Check Lambda exists
aws lambda get-function --function-name cocktails-pr-99-api --region us-east-1

# Check DynamoDB tables
aws dynamodb describe-table --table-name cocktails-pr-99-recipes --region us-east-1

# Check API Gateway route (should show a route for /api/pr-99/{proxy+})
aws apigatewayv2 get-routes --api-id <API_GATEWAY_ID> --region us-east-1 | jq '.Items[] | select(.RouteKey | contains("pr-99"))'

# Check S3 objects
aws s3 ls s3://cocktails-prod-frontend/pr-99/ --human-readable

# Test API endpoint via CloudFront
curl -s https://cocktails.albertomcastro.com/api/pr-99/v1/recipes | jq '.total'

# Open preview in browser
open https://cocktails.albertomcastro.com/pr-99/
```

## Manual Teardown (cleanup test resources)

```bash
export PR_NUMBER=99
export AWS_REGION=us-east-1
export API_GATEWAY_ID=<existing-api-gw-id>
export FRONTEND_BUCKET=cocktails-prod-frontend
export CLOUDFRONT_DISTRIBUTION_ID=EX7HUB6P225MV

.github/scripts/preview-teardown.sh
```

Expected output:
```
[preview-teardown] Deleting Lambda cocktails-pr-99-api...
[preview-teardown] Removing API Gateway route for pr-99...
[preview-teardown] Deleting DynamoDB tables for pr-99...
[preview-teardown] Deleting S3 objects at s3://cocktails-prod-frontend/pr-99/...
[preview-teardown] Invalidating CloudFront cache for /pr-99/*...
[preview-teardown] ✓ Preview environment pr-99 torn down.
```

## Acceptance Scenario Walkthrough

1. Open a real PR against `main` in GitHub.
2. Wait for the `deploy-preview` CI job to complete.
3. Visit the preview URL posted as a PR comment.
4. Log in with admin credentials (seeded from production).
5. Browse recipes — confirm production data is visible.
6. Create a new recipe — confirm it saves without error.
7. Verify the new recipe does NOT appear on production (`cocktails.albertomcastro.com`).
8. Merge the PR — wait for `prod-deploy` to complete.
9. Verify the merged code is live on production.
10. Verify the preview URL `cocktails.albertomcastro.com/pr-{number}/` is no longer accessible (returns 404 or redirects).
