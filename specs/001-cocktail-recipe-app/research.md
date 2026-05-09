# Research: Cocktail Recipe App

**Phase**: 0 | **Date**: 2026-05-08 | **Branch**: `001-cocktail-recipe-app`

## Decision Log

---

### 1. Go Backend — Local Server vs. AWS Lambda Entry Points

**Decision**: Two entry points sharing one handler package.
- `backend/cmd/server/main.go` — starts a standard `net/http` server for local use.
- `backend/cmd/lambda/main.go` — wraps the same HTTP handler using `github.com/awslabs/aws-lambda-go-api-proxy/httpadapter`, which translates API Gateway proxy events into `http.Request` objects.

**Rationale**: All handler, middleware, and business logic lives in `backend/internal/`. Neither entry point contains application logic. Switching between local and Lambda requires only a different binary — no code changes. The `httpadapter` pattern is the standard recommended approach in the aws-lambda-go-api-proxy project.

**Alternatives considered**:
- Single binary detecting `AWS_LAMBDA_FUNCTION_NAME` env var at startup — rejected because it adds conditional startup logic to the main package and complicates local testing.
- AWS SAM or CDK wrappers — deferred; not needed at planning stage.

---

### 2. Storage Backend — Flexible Schema

**Decision**: Abstract behind a `store.RecipeStore` interface with two implementations:
- **Local**: `modernc.org/sqlite` (pure Go, no CGo, Lambda-compatible). Recipes stored with a `properties` column as a JSON blob. Full-text search via SQLite FTS5.
- **AWS**: DynamoDB. Flexible properties stored as a DynamoDB attribute map. Search via `Scan` with `FilterExpression` (viable for small dataset; can be upgraded to DynamoDB + OpenSearch if scale demands it).

**Rationale**: The `Store` interface means the handler layer has zero awareness of the underlying storage. Switching from SQLite to DynamoDB in production is a configuration change, not a code change. `modernc.org/sqlite` compiles to a self-contained Go binary — no system libraries, safe for Lambda deployment if a local/SQLite Lambda option is ever needed.

**Alternatives considered**:
- PostgreSQL — rejected for over-engineering a personal-use app; adds infrastructure dependency locally.
- S3 as primary data store (JSON files) — rejected because search across S3 objects requires S3 Select or Athena, which adds complexity and latency for interactive queries.
- Single SQLite implementation only — rejected because DynamoDB is a better fit for Lambda cold starts and multi-user concurrent writes on AWS.

---

### 3. Authentication — JWT

**Decision**: JWT Bearer tokens (`golang-jwt/jwt/v5`). On login, the server issues a signed JWT (HS256, 24-hour expiry). All write endpoints require a valid JWT in the `Authorization: Bearer <token>` header. A separate admin claim in the JWT gates the user-creation endpoint.

**Rationale**: Stateless auth is a natural fit for Lambda (no shared session store needed). HS256 with a secret stored in an environment variable is sufficient for a small trusted-contributor app. `golang-jwt/jwt/v5` is the maintained fork of the widely-used `dgrijalva/jwt-go`.

**Alternatives considered**:
- Session cookies — rejected because stateful sessions require a shared store (Redis/DynamoDB) between Lambda invocations; adds infrastructure.
- AWS Cognito — deferred; appropriate if the user base grows, but over-engineered for admin-created accounts.

---

### 4. Full-Text Search Strategy

**Decision**:
- **SQLite**: FTS5 virtual table mirroring the recipes table. Indexed fields: name, ingredients (flattened), steps (joined), and all property values. FTS5 `MATCH` queries handle substring and prefix matching.
- **DynamoDB**: `Scan` with a `FilterExpression` testing `contains()` across all relevant attributes. Acceptable for a small personal-use dataset (< a few thousand recipes).

**Rationale**: FTS5 is built into SQLite and requires no external service. DynamoDB Scan is simple and sufficient at the stated scale. Both implementations satisfy SC-002 (search results within 2 seconds) at small scale.

**Alternatives considered**:
- DynamoDB + OpenSearch — rejected as over-engineering for personal use; can be added later if needed.
- Client-side search on fetched data — rejected because it breaks the external API contract (FR-010 requires server-side filtering).

---

### 5. Frontend Build & Deployment

**Decision**: Vite builds the frontend to `frontend/dist/`. For AWS deployment, `dist/` contents are synced to an S3 bucket configured for static website hosting, fronted by a CloudFront distribution. API calls from the frontend target the API Gateway URL (configured via a Vite environment variable).

**Rationale**: Standard Vite + S3 + CloudFront is the simplest static hosting pattern on AWS. The frontend is fully decoupled from the backend — no server-side rendering needed.

**chart.js** is included in frontend dependencies for recipe statistics visualization (e.g., breakdown of recipes by base spirit or drink style). It is tree-shaken by Vite so unused chart types add no bundle weight.

**Alternatives considered**:
- Server-side rendered frontend from Go — rejected; complicates Lambda deployment and breaks the clean frontend/backend separation.
- Amplify Hosting — deferred; adds managed CI/CD but is not required for initial deployment.

---

### 6. Recipe Property Search — Indexing Strategy

**Decision**: All property values are included in the search index regardless of key name. When a new property is added to a recipe, the FTS5 index (SQLite) is updated automatically via triggers; the DynamoDB filter expression already scans the full attribute map.

**Rationale**: Satisfies FR-003 (search must work on fields added after launch) without requiring schema migrations or index rebuilds.

---

### 7. CORS & API Gateway Configuration

**Decision**: The Go handler sets permissive CORS headers during local development. In AWS, API Gateway handles CORS configuration at the gateway level; the Lambda handler does not need to set headers.

**Rationale**: Keeps CORS logic out of application code for the AWS path while still allowing the Vite dev server to call the local Go server without a proxy.
