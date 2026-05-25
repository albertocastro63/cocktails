# Tasks: GitHub Actions CI Pipeline

**Input**: Design documents from `specs/016-github-ci-pipeline/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, contracts/ci-workflow.md ✓, quickstart.md ✓

**Deliverable**: A single file — `.github/workflows/ci.yml` — built up incrementally across phases.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no shared edit conflicts)
- **[Story]**: Which user story this task maps to
- All tasks edit `.github/workflows/ci.yml` unless noted otherwise

---

## Phase 1: Setup

**Purpose**: Create the GitHub Actions directory structure.

- [ ] T001 Create `.github/workflows/` directory at the repository root

---

## Phase 2: Foundational

**N/A** — No shared database schema, middleware, or framework. The single YAML file is self-contained. Phase 1 (directory creation) is the only prerequisite.

**Checkpoint**: `.github/workflows/` directory exists → proceed to user story phases.

---

## Phase 3: User Story 1 — Automated Validation on PRs (Priority: P1) 🎯 MVP

**Goal**: Every PR targeting `main` automatically runs backend (Go tests + build) and frontend (Vitest + Vite build) checks. Failing checks block merging.

**Note**: US4 (Validation on Pushes to Main) is implemented in this same phase — it requires only the addition of a `push: branches: [main]` trigger alongside the PR trigger. Both use the same jobs.

**Independent Test**: Open a PR with a deliberately broken Go test → confirm the `backend` check appears as failing on the PR page and the merge button is disabled.

### Implementation for User Story 1

- [ ] T002 [US1] Create `.github/workflows/ci.yml` with: workflow name, `on: pull_request` (branches: [main]) AND `on: push` (branches: [main]) triggers (covering US4), and `concurrency: group: ${{ github.workflow }}-${{ github.ref }}, cancel-in-progress: true`

- [ ] T003 [US1] Add `frontend` job to `.github/workflows/ci.yml`: `runs-on: ubuntu-latest`, steps: `actions/checkout@v4`, `actions/setup-node@v4` (node-version: `lts/*`, cache: `npm`, cache-dependency-path: `frontend/package-lock.json`), `cd frontend && npm ci`, `cd frontend && npm test`, `cd frontend && npm run build`

- [ ] T004 [US1] Add `backend` job to `.github/workflows/ci.yml`: `runs-on: ubuntu-latest`, steps: `actions/checkout@v4`, `actions/setup-go@v5` (go-version-file: `backend/go.mod`, cache-dependency-path: `backend/go.sum`), a combined test+coverage step that runs `cd backend && go test -coverprofile=coverage.out ./...` then parses `go tool cover -func=coverage.out` total and exits non-zero if coverage is below 80%, and a build step running `cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/bootstrap ./cmd/lambda/`

**Checkpoint**: Both `backend` and `frontend` named checks appear on a PR. A broken test in either language blocks merging.

---

## Phase 4: User Story 2 — DynamoDB Integration Tests (Priority: P2)

**Goal**: DynamoDB integration tests (currently skipped in CI because `DYNAMODB_ENDPOINT` is unset) run against a local DynamoDB emulator started as a sidecar service. No AWS credentials required.

**Independent Test**: Examine the backend job log for a PR → confirm `TestDynamo_` tests appear and pass (not skipped).

### Implementation for User Story 2

- [ ] T005 [US2] Add `services:` block to the `backend` job in `.github/workflows/ci.yml`: service name `dynamodb-local`, image `amazon/dynamodb-local:latest`, ports `["8000:8000"]`, options with `--health-cmd "curl -s http://localhost:8000 > /dev/null" --health-interval 5s --health-timeout 3s --health-retries 10` (note: DynamoDB Local returns HTTP 400 on GET /, so `-f` must NOT be used — any response means the container is up). Also add `env:` block to the backend job: `DYNAMODB_ENDPOINT: http://localhost:8000`, `AWS_REGION: us-east-1`, `AWS_ACCESS_KEY_ID: test`, `AWS_SECRET_ACCESS_KEY: test`

**Checkpoint**: Navigate to the backend job log for any PR → `TestDynamo_SearchByIngredients_TwoIngredients` (and sibling tests) appear in output and pass, with no "skipping DynamoDB tests" message.

---

## Phase 5: User Story 3 — OIDC Authentication (Priority: P3)

**Goal**: The backend job contains a conditional OIDC credential step that activates automatically when `vars.AWS_CI_ROLE_ARN` is set, with no static AWS access keys ever stored as GitHub Secrets.

**Independent Test**: Set `vars.AWS_CI_ROLE_ARN` in repository Actions variables → trigger a pipeline run → the OIDC step succeeds without any `AWS_ACCESS_KEY_ID` secret configured.

### Implementation for User Story 3

- [ ] T006 [US3] Add `permissions: id-token: write, contents: read` to the `backend` job in `.github/workflows/ci.yml`, then add a step: name `Configure AWS credentials (OIDC)`, `if: vars.AWS_CI_ROLE_ARN != ''`, `uses: aws-actions/configure-aws-credentials@v4`, with `role-to-assume: ${{ vars.AWS_CI_ROLE_ARN }}` and `aws-region: us-east-1`. Place this step before the test step. Note: the initial IAM role (`github-ci-role`) requires **no permission policies** — the trust policy alone is sufficient until a deploy step is added. Only the trust relationship needs to be configured: `StringLike sub: repo:albertocastro63/cocktails:*`.

**Checkpoint**: When `vars.AWS_CI_ROLE_ARN` is unset, the OIDC step is skipped and all other checks pass. When the variable is set (after IAM role creation), the step runs and obtains temporary credentials.

---

## Phase 6: Polish & Verification

**Purpose**: Validate the pipeline end-to-end against the quickstart.md acceptance tests.

- [ ] T007 Verify the pipeline by pushing a test commit that deliberately breaks a Go test to a PR branch, confirming the `backend` check fails and the merge button is disabled (quickstart.md Acceptance Test 1); then revert the breakage and confirm the check turns green

- [ ] T008 [P] Verify DynamoDB integration tests run (not skipped) by inspecting the `backend` job log for any PR run and confirming `TestDynamo_` test names appear in the output (quickstart.md Acceptance Test 2)

- [ ] T009 Enable branch protection on `main`: go to Settings → Branches → Add rule for `main`, enable "Require status checks to pass before merging", add `backend` and `frontend` as required checks (FR-008). This makes failing CI checks block merging.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: N/A
- **Phase 3 (US1 + US4)**: Depends on Phase 1
- **Phase 4 (US2)**: Depends on Phase 3 (requires the backend job to exist before adding service containers)
- **Phase 5 (US3)**: Depends on Phase 3 (requires the backend job to exist before adding OIDC step)
- **Phase 6 (Polish)**: Depends on Phases 3–5

### Within Phase 3

- T002 must complete before T003 and T004 (file must exist before adding jobs)
- T003 and T004 both depend on T002 but edit different sections of `ci.yml` — complete T003 then T004 sequentially to avoid merge conflicts

### User Story Dependencies

- **US1 + US4 (Phase 3)**: Start after directory creation — no other dependencies
- **US2 (Phase 4)**: Start after US1 (backend job must exist to add service container)
- **US3 (Phase 5)**: Start after US1 (backend job must exist to add OIDC step); can run in parallel with US2

---

## Implementation Strategy

### MVP (Phase 3 only)

1. Complete Phase 1: create directory
2. Complete Phase 3: T002 → T003 → T004
3. **Validate**: Push a PR — both `backend` and `frontend` checks appear
4. **Ship**: Branch protection can now reference these check names

### Incremental Delivery

1. Phase 1 + Phase 3 → Basic CI (all PRs and pushes to main validated)
2. Phase 4 (US2) → DynamoDB integration tests no longer skipped in CI
3. Phase 5 (US3) → OIDC credential pattern in place; activates when IAM role is created
4. Phase 6 → End-to-end verification

---

## Notes

- All tasks T002–T006 edit a single file: `.github/workflows/ci.yml`. Complete them sequentially.
- T003 and T004 (frontend and backend jobs) have no shared lines and can be mentally treated as parallel concerns, but edit the same file — write one then the other.
- The coverage threshold command for Go requires parsing `go tool cover -func` output. A portable approach: `awk '/^total:/{pct=substr($3,1,length($3)-1); if (pct+0 < 80){print "Coverage " pct "% < 80%"; exit 1}}'`
- `amazon/dynamodb-local` serves a 400 response (not 200) on `GET /` — use `curl -sf http://localhost:8000/ || exit 0` OR `curl -s http://localhost:8000 > /dev/null` for the health check, since any response indicates the container is up.
- The OIDC step in T006 must come before the `go test` step so credentials are available if needed.
