# Research: GitHub Actions CI Pipeline

**Feature**: 016-github-ci-pipeline  
**Phase**: 0 — Research  
**Date**: 2026-05-25

## Decision 1: Workflow File Structure

**Decision**: Single `ci.yml` file with two parallel jobs (`backend`, `frontend`).

**Rationale**: One workflow → one entry point to read and debug. Two separate jobs run in parallel and each becomes an independently named status check on GitHub PRs. Branch protection can require both checks, satisfying FR-007 and FR-008. Splitting into multiple workflow files adds no value here and complicates status-check naming.

**Alternatives considered**:
- Separate `backend.yml` + `frontend.yml`: Rejected — no benefit; harder to read the full CI picture at a glance.
- Single job with sequential steps: Rejected — loses the parallel execution speedup and produces only one merged status check, losing granularity.

---

## Decision 2: Trigger Configuration

**Decision**: `on: pull_request` (types: opened, synchronize, reopened) targeting `main`, plus `on: push` (branches: [main]).

**Rationale**: This covers FR-001 (PR triggers) and FR-002 (push-to-main triggers). `synchronize` fires on force-push or commit addition — the intended re-trigger scenario in US1.

**Concurrency**: Set `concurrency.group: ${{ github.workflow }}-${{ github.ref }}` with `cancel-in-progress: true`. This discards stale runs when a developer pushes a second commit to the same PR — only the latest run determines merge eligibility, matching the edge case in the spec.

---

## Decision 3: DynamoDB Local Service Container

**Decision**: Use `amazon/dynamodb-local:latest` as a GitHub Actions service container on the `backend` job. Expose port 8000 and set `DYNAMODB_ENDPOINT: http://localhost:8000`.

**Rationale**: The existing test helper in `backend/internal/store/dynamo/dynamo_test.go` already implements the skip-if-no-endpoint pattern: `t.Skip("DYNAMODB_ENDPOINT not set; skipping DynamoDB tests")`. Setting the env var in CI is the only change needed to activate these tests — no test code modification required. The `amazon/dynamodb-local` image is public (Docker Hub), requires no authentication, and no AWS credentials. Satisfies FR-011, FR-012.

**Health check**: GitHub Actions service containers support `options: --health-cmd "curl -f http://localhost:8000/ || exit 1"`. This ensures the emulator is accepting connections before the test step begins, satisfying FR-012.

**Alternatives considered**:
- DynamoDB via real AWS: Rejected — requires credentials, adds cost and external dependency, contradicts FR-011.
- `localstack`: Rejected — heavier image, more startup time, overkill for DynamoDB-only use.
- Mock/stub stores: Already used for unit tests. Integration tests need a real DynamoDB protocol, which is the spec's motivation.

---

## Decision 4: Dependency Caching

**Decision**:
- **Go**: Cache `~/go/pkg/mod` and `~/.cache/go-build` using `actions/cache@v4`, keyed on `hashFiles('backend/go.sum')`.
- **Node**: Use `actions/setup-node@v4` with `cache: 'npm'` and `cache-dependency-path: frontend/package-lock.json`. This is the built-in npm cache — no separate `actions/cache` step needed.

**Rationale**: Cache hits eliminate `go mod download` (~20–40 s) and `npm ci` download time (~15–30 s). Both caches are invalidated only when their respective lock files change, ensuring correctness.

---

## Decision 5: OIDC Authentication (US3, P3)

**Decision**: Use `aws-actions/configure-aws-credentials@v4` with `role-to-assume`. Requires `permissions: id-token: write` at job level.

**Trust policy scope**: `"StringLike": { "token.actions.githubusercontent.com:sub": "repo:albertocastro63/cocktails:*" }`. This restricts role assumption to this repository only, satisfying FR-014.

**Implementation phasing**: US3 is P3. The initial CI implementation (US1, US2, US4) will include a commented placeholder OIDC step so the pattern is established. It becomes an active step only after the IAM role is created in AWS. This avoids blocking US1/US4 delivery on the manual AWS setup.

**Role ARN**: Will be captured as a GitHub Actions variable (not a secret) once created: `vars.AWS_CI_ROLE_ARN`. No `AWS_ACCESS_KEY_ID` or `AWS_SECRET_ACCESS_KEY` secrets are stored — satisfying FR-013 and SC-006.

**Alternatives considered**:
- Long-lived access keys in GitHub Secrets: Rejected — credential leak risk, contradicts FR-013.
- AWS IAM Identity Center: Rejected — requires additional tooling, OIDC is the GitHub-recommended pattern.

---

## Decision 6: Go Build Command

**Decision**: `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/bootstrap ./cmd/lambda/`

**Rationale**: The Lambda binary is deployed to `arm64` (Graviton2) — this must match the production build. CGO disabled for a static binary. This is the same command used in the existing `infra/` deployment scripts.

---

## Decision 7: No data-model.md Required

**Decision**: Skip `data-model.md` for this feature.

**Rationale**: CI pipelines have no application data entities — only workflow configuration. The "entities" are jobs, steps, and events, which are better documented in the workflow contract file.
