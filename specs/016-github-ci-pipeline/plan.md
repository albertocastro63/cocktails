# Implementation Plan: GitHub Actions CI Pipeline

**Branch**: `016-github-ci-pipeline` | **Date**: 2026-05-25 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/016-github-ci-pipeline/spec.md`

## Summary

Create `.github/workflows/ci.yml` — a single GitHub Actions workflow that runs two parallel jobs (`backend`, `frontend`) on every PR and push to `main`. The backend job starts a DynamoDB Local service container and runs the full Go test suite (including previously-skipped DynamoDB integration tests) plus a Lambda binary build. The frontend job runs Vitest and a Vite production build. A commented OIDC step in the backend job establishes the secure credential pattern for future deploy steps.

## Technical Context

**Language/Version**: YAML (GitHub Actions workflow syntax); Go 1.25 (backend); Node.js LTS / npm (frontend)  
**Primary Dependencies**: `actions/checkout@v4`, `actions/setup-go@v5`, `actions/setup-node@v4`, `actions/cache@v4`, `aws-actions/configure-aws-credentials@v4`  
**Storage**: N/A (CI only; DynamoDB Local ephemeral service container)  
**Testing**: `go test ./...` (backend), `npm test` via Vitest (frontend)  
**Target Platform**: GitHub Actions hosted runners (`ubuntu-latest`)  
**Project Type**: CI/CD pipeline configuration  
**Performance Goals**: Full pipeline completes in < 5 minutes (SC-004)  
**Constraints**: Free GitHub Actions tier (2,000 min/month); no self-hosted runners; no long-lived AWS credentials  
**Scale/Scope**: Single repository; two jobs; one workflow file

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Are workflow steps single-responsibility and below complexity limits? | ✅ Each step does one thing; YAML is declarative — cyclomatic complexity N/A |
| II. Test-First | Are failing tests written before implementation begins? | ⚠ See justification below |
| III. UX Consistency | Do all UI surfaces follow the design language and handle loading/empty/error states? | ✅ N/A — infrastructure, no user-facing surfaces |
| IV. Performance | Do API responses meet p95 ≤ 200 ms read / ≤ 500 ms write and TTI ≤ 3 s? | ✅ N/A for CI config; SC-004 (<5 min) is the applicable metric |
| Quality Gates | Do all CI checks (lint, coverage ≥ 80%, benchmarks) pass? | ✅ This feature IS the quality gate; validated by push + observation |

**⚠ Principle II Justification**: GitHub Actions workflow YAML cannot be unit-tested before writing it — there is no local runner that faithfully reproduces the full service container + OIDC environment. Verification is empirical: push a deliberately broken change to a PR branch and observe the pipeline fail; push the fix and observe it pass. The acceptance tests in `quickstart.md` codify this manual validation protocol.

## Project Structure

### Documentation (this feature)

```text
specs/016-github-ci-pipeline/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── contracts/
│   └── ci-workflow.md   # Workflow job/step contract
├── quickstart.md        # Manual acceptance test protocol
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```text
.github/
└── workflows/
    └── ci.yml           # The ONLY file this feature creates

backend/                 # Existing — no changes required
├── internal/store/dynamo/dynamo_test.go  # Already has skip-if-no-endpoint pattern
└── cmd/lambda/main.go   # Existing Lambda entry point

frontend/                # Existing — no changes required
├── package.json         # `npm test` runs vitest; `npm run build` runs vite build
└── vitest.config.js     # jsdom environment, coverage thresholds
```

**Structure Decision**: Single `.github/workflows/ci.yml` — the canonical GitHub Actions location. No new application code is added; only CI configuration.

## Phases

### Phase 1 — Core CI (US1, US4): Backend + Frontend validation on PR and main

**Deliverable**: `.github/workflows/ci.yml` with `backend` and `frontend` jobs.

**Workflow skeleton**:

```yaml
name: CI

on:
  pull_request:
    branches: [main]
  push:
    branches: [main]

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  backend:
    runs-on: ubuntu-latest
    services:
      dynamodb-local:
        image: amazon/dynamodb-local:latest
        ports: ["8000:8000"]
        options: >-
          --health-cmd "curl -sf http://localhost:8000/ || exit 1"
          --health-interval 5s
          --health-timeout 3s
          --health-retries 10
    env:
      DYNAMODB_ENDPOINT: http://localhost:8000
      AWS_REGION: us-east-1
      AWS_ACCESS_KEY_ID: test
      AWS_SECRET_ACCESS_KEY: test
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: backend/go.mod
          cache-dependency-path: backend/go.sum
      - name: Run backend tests
        run: cd backend && go test ./...
      - name: Build Lambda binary
        run: |
          cd backend
          CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/bootstrap ./cmd/lambda/

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: lts/*
          cache: npm
          cache-dependency-path: frontend/package-lock.json
      - name: Install dependencies
        run: cd frontend && npm ci
      - name: Run frontend tests
        run: cd frontend && npm test
      - name: Build frontend
        run: cd frontend && npm run build
```

Key points:
- `actions/setup-go@v5` handles Go module cache automatically when `cache-dependency-path` is set — no separate `actions/cache` step needed.
- `actions/setup-node@v4` with `cache: npm` handles npm cache similarly.
- DynamoDB Local health check ensures the emulator is ready before `go test ./...` runs (FR-012).
- Dummy AWS credentials (`test`/`test`) satisfy the AWS SDK credential requirement for local endpoint calls — no real AWS access.

### Phase 2 — OIDC Authentication (US3, P3)

**Deliverable**: Add OIDC step to `backend` job after IAM role is created in AWS.

**Prerequisite** (manual, performed by repository owner):
1. Create OIDC identity provider in IAM: `token.actions.githubusercontent.com`
2. Create IAM role `github-ci-role` with trust policy scoped to `repo:albertocastro63/cocktails:*`
3. Set GitHub Actions variable `AWS_CI_ROLE_ARN` to the role ARN

**OIDC step to add to backend job**:
```yaml
      - name: Configure AWS credentials (OIDC)
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ vars.AWS_CI_ROLE_ARN }}
          aws-region: us-east-1
```

**Job permission required**:
```yaml
  backend:
    permissions:
      id-token: write
      contents: read
```

This step is a placeholder in Phase 1 (commented out) and activated in Phase 2. No `AWS_ACCESS_KEY_ID` or `AWS_SECRET_ACCESS_KEY` secrets are ever created.

## Complexity Tracking

No constitution violations requiring tracking. Principle II justification documented above.

## Implementation Notes

- The DynamoDB Local service container health check uses `curl`. The `ubuntu-latest` runner image includes `curl` by default — no additional installation needed.
- `amazon/dynamodb-local` serves HTTP on port 8000 and returns a 200/400 on `GET /` — the health check curl will work.
- The `go-version-file: backend/go.mod` option reads the `go` directive from go.mod, keeping the CI Go version in sync with the project automatically.
- `npm test` maps to `vitest run` (not watch mode) per `package.json` scripts — correct for CI.
- The `concurrency.cancel-in-progress: true` setting means if two commits arrive on the same PR branch in quick succession, only the second run completes. This is acceptable and matches the spec edge case note.
