# Quickstart: Cocktail Recipe App

**Date**: 2026-05-08 | **Branch**: `001-cocktail-recipe-app`

## Prerequisites

- Go 1.22+
- Node.js 20+
- (AWS deployment only) AWS CLI v2, configured with appropriate credentials

---

## Local Development

### 1. Backend

```bash
cd backend
go mod download

# Run the local HTTP server (default port: 8080)
go run ./cmd/server

# With a custom port and database path
PORT=9090 DB_PATH=./cocktails.db go run ./cmd/server
```

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `DB_PATH` | `./cocktails.db` | SQLite database file path |
| `JWT_SECRET` | — | **Required.** Secret key for JWT signing |
| `ADMIN_BOOTSTRAP_PASSWORD` | — | Optional. Sets the initial admin password on first run |

On first run with an empty database, the server creates the schema and an initial `admin` user if `ADMIN_BOOTSTRAP_PASSWORD` is set.

### 2. Frontend

```bash
cd frontend
npm install

# Development server with HMR (proxies /api to the Go backend)
npm run dev
```

The Vite dev server runs on `http://localhost:5173` by default. API calls to `/api/*` are proxied to `http://localhost:8080` (configure in `vite.config.js`).

Environment variables (`.env.local`):

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API_BASE_URL` | `/api/v1` | API base URL; override for AWS target |

### 3. Run Tests

```bash
# Backend
cd backend && go test ./...

# Frontend (unit tests)
cd frontend && npm run test
```

---

## Building for Production

### Frontend

```bash
cd frontend
npm run build
# Output: frontend/dist/
```

### Backend (local binary)

```bash
cd backend
go build -o bin/server ./cmd/server
```

### Backend (Lambda binary)

```bash
cd backend
GOOS=linux GOARCH=arm64 go build -o bin/bootstrap ./cmd/lambda
zip bin/lambda.zip bin/bootstrap
```

The Lambda binary must be named `bootstrap` for the `provided.al2023` runtime on ARM64.

---

## AWS Deployment (Manual)

### Architecture

```
Browser → CloudFront → S3 (frontend static files)
                     → API Gateway → Lambda (Go backend) → DynamoDB
```

### Steps

**1. Deploy frontend to S3 + CloudFront**

```bash
# Create S3 bucket (one-time)
aws s3 mb s3://cocktails-frontend-<your-suffix>

# Sync build output
aws s3 sync frontend/dist/ s3://cocktails-frontend-<your-suffix> --delete

# Invalidate CloudFront cache after each deploy
aws cloudfront create-invalidation --distribution-id <DIST_ID> --paths "/*"
```

**2. Deploy Lambda function**

```bash
# Build Lambda binary
cd backend
GOOS=linux GOARCH=arm64 go build -o bin/bootstrap ./cmd/lambda
zip bin/lambda.zip bin/bootstrap

# Update function code (assumes function already created)
aws lambda update-function-code \
  --function-name cocktails-api \
  --zip-file fileb://bin/lambda.zip
```

Lambda environment variables to configure:
- `JWT_SECRET` — secret for JWT signing (use AWS Secrets Manager or Parameter Store)
- `DYNAMODB_REGION` — AWS region
- `DYNAMODB_RECIPES_TABLE` — DynamoDB recipes table name
- `DYNAMODB_USERS_TABLE` — DynamoDB users table name

**3. API Gateway**

Configure a REST API (or HTTP API) with a catch-all `{proxy+}` route pointing to the Lambda function. Enable CORS at the gateway level.

---

## Project Structure

```
cocktails/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go        # Local HTTP server entry point
│   │   └── lambda/main.go        # AWS Lambda entry point
│   ├── internal/
│   │   ├── handler/              # HTTP handlers (shared)
│   │   ├── store/
│   │   │   ├── store.go          # RecipeStore / UserStore interfaces
│   │   │   ├── sqlite/           # SQLite implementation
│   │   │   └── dynamo/           # DynamoDB implementation
│   │   ├── model/                # Domain types (Recipe, Ingredient, User)
│   │   └── auth/                 # JWT middleware and helpers
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── pages/                # Page-level components
│   │   ├── components/           # Reusable UI components
│   │   └── api/                  # API client functions
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
└── specs/
    └── 001-cocktail-recipe-app/
        ├── plan.md
        ├── spec.md
        ├── research.md
        ├── data-model.md
        ├── quickstart.md
        └── contracts/
            └── api.md
```
