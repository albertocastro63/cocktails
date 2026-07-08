# Data Model: PR Preview Environments

**Branch**: `021-pr-preview-environments` | **Date**: 2026-05-28

This feature does not introduce new application domain entities. The domain model (Recipe, Ingredient, User, Favorite) is unchanged. This document describes the **resource naming conventions** and **structural schemas** for preview environment AWS resources.

---

## Resource Naming Convention

All preview resources follow the pattern `cocktails-pr-{number}-{resource}` where `{number}` is the GitHub PR number.

| Resource Type | Production Name | Preview Name (example for PR 42) |
|---|---|---|
| Lambda function | `cocktails-prod-api` | `cocktails-pr-42-api` |
| DynamoDB recipes table | `cocktails-recipes` | `cocktails-pr-42-recipes` |
| DynamoDB users table | `cocktails-users` | `cocktails-pr-42-users` |
| DynamoDB favorites table | `cocktails-favorites` | `cocktails-pr-42-favorites` |
| S3 frontend prefix | `(root)` | `pr-42/` |
| API Gateway route | `$default` | `ANY /pr-42/api/{proxy+}` |
| API Gateway integration | `production-integration` | `cocktails-pr-42-integration` |

---

## DynamoDB Table Schemas (unchanged from production)

### Recipes Table — `cocktails-pr-{number}-recipes`

| Attribute | Type | Role |
|---|---|---|
| `id` | String (S) | Hash key (partition key) |
| `name` | String | Recipe display name |
| `ingredients` | List | Ingredient objects |
| `steps` | List | Ordered step strings |
| `properties` | Map | Arbitrary key-value pairs |
| `notes` | String | Markdown notes |
| `garnishes` | List | Garnish instruction strings |
| `creator_id` | String | References a user `id` |
| `created_at` | String | ISO 8601 timestamp |
| `updated_at` | String | ISO 8601 timestamp |

Billing: `PAY_PER_REQUEST`. No GSI required (full-scan search acceptable for preview datasets).

### Users Table — `cocktails-pr-{number}-users`

Seeded with a **fixed, small set of test accounts** (`PREVIEW_SEED_USERNAMES`, default `admin alberto`) copied verbatim from production so previews are testable via login with the accounts' real passwords. This intentionally copies real user records (incl. bcrypt hashes) into a public preview, scoped to those test accounts only — it overrides the original "created empty, no PII" approach. All other production users are **not** copied.

| Attribute | Type | Role |
|---|---|---|
| `id` | String (S) | Hash key |
| `username` | String | Unique login name |
| `password_hash` | String | bcrypt hash |
| `is_admin` | Boolean | Admin flag |
| `created_at` | String | ISO 8601 timestamp |

The table is created with the `username-index` GSI (`username` HASH, projection ALL), matching production — the backend looks up users by username through it for login/auth. Without the GSI, any user later added to the preview cannot authenticate.

### Favorites Table — `cocktails-pr-{number}-favorites`

Created **empty** — not seeded from production (favorites reference user records and are treated as user PII).

| Attribute | Type | Role |
|---|---|---|
| `user_id` | String (S) | Hash key |
| `recipe_id` | String (S) | Range key |

GSI on `recipe_id` omitted for preview tables (not required for preview functionality; saves provisioning complexity).

---

## Environment Variable Schema

### Lambda Function Environment Variables

| Variable | Production Value | Preview Value (PR 42) |
|---|---|---|
| `STORE_BACKEND` | `dynamodb` | `dynamodb` |
| `RECIPES_TABLE` | `cocktails-recipes` | `cocktails-pr-42-recipes` |
| `USERS_TABLE` | `cocktails-users` | `cocktails-pr-42-users` |
| `FAVORITES_TABLE` | `cocktails-favorites` | `cocktails-pr-42-favorites` |
| `JWT_SECRET` | (GitHub secret) | (same GitHub secret) |
| `STRIP_PATH_PREFIX` | (unset) | `/pr-42` |

### Frontend Build Variables

| Variable | Production Value | Preview Value (PR 42) |
|---|---|---|
| `VITE_API_PATH_PREFIX` | `/api` (default) | `/pr-42/api` |

---

## Preview Environment Lifecycle

```
PR opened / push to PR branch
  → Tables exist? No → Create all 3 tables + seed recipes only from prod (users/favorites empty) → Deploy Lambda → Add API GW route → Build & upload frontend
  → Tables exist? Yes → Deploy Lambda (code update only) → Update API GW integration → Build & upload frontend

PR merged or closed
  → Delete Lambda function
  → Delete DynamoDB tables (recipes, users, favorites)
  → Remove API Gateway route + integration
  → Delete S3 objects at pr-{number}/
  → Invalidate CloudFront /pr-{number}/*
```

---

## One-Time Infrastructure Additions (Terraform)

These are added once to the production Terraform state and are not created per PR:

| Resource | Description |
|---|---|
| `cocktails-preview-lambda-role` | IAM execution role for all preview Lambdas; DynamoDB access scoped to `cocktails-pr-*` table ARNs |
| CloudFront Function `spa-pr-routing` | Viewer-request function rewriting `/pr-{number}/` → `/pr-{number}/index.html` |
