# Feature Specification: GitHub Actions CI Pipeline

**Feature Branch**: `016-github-ci-pipeline`
**Created**: 2026-05-25
**Status**: Draft
**Input**: User description: "Create a GitHub Actions continuous integration pipeline for the project. The pipeline should run automatically on every pull request and push to main. It should validate the backend (Go tests, build) and the frontend (Vitest tests, build). The pipeline should block merging if any check fails."

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Automated Validation on Pull Requests (Priority: P1)

When a developer opens or updates a pull request, the pipeline automatically runs all backend and frontend checks. If any check fails, GitHub marks the PR as blocked and prevents merging. The developer can see exactly which check failed and why, fix the issue, and push again to re-trigger validation.

**Why this priority**: This is the core value of CI — catching broken code before it reaches main. Without it, regressions can silently land in production (as happened with the missing Lambda route).

**Independent Test**: Open a PR with a deliberately broken Go test; confirm the check appears as failing on the PR and the merge button is disabled.

**Acceptance Scenarios**:

1. **Given** a PR is opened or updated, **When** the pipeline runs, **Then** backend tests, backend build, frontend tests, and frontend build all execute automatically.
2. **Given** a backend test fails, **When** the pipeline completes, **Then** the PR shows a failing status check and cannot be merged.
3. **Given** a frontend test fails, **When** the pipeline completes, **Then** the PR shows a failing status check and cannot be merged.
4. **Given** all checks pass, **When** the pipeline completes, **Then** the PR shows a passing status and is eligible to be merged.
5. **Given** a developer pushes a fix, **When** the pipeline re-runs, **Then** the status updates to reflect the new result.

---

### User Story 2 — DynamoDB Integration Tests via Local Emulator (Priority: P2)

The pipeline runs the DynamoDB integration tests (currently skipped in CI) against a local DynamoDB emulator started as a sidecar service. No AWS credentials are required. Developers get confidence that the store implementations work correctly against a real DynamoDB protocol, not just stub stores.

**Why this priority**: The DynamoDB store implementations exist and have tests written, but those tests are never exercised in CI. A bug in the real DynamoDB store (e.g., a missing GSI, a wrong attribute name) would only surface in production.

**Independent Test**: Confirm that the DynamoDB integration tests (e.g., `TestDynamo_SearchByIngredients_TwoIngredients`) run and pass in the pipeline without any AWS credentials configured.

**Acceptance Scenarios**:

1. **Given** the pipeline runs, **When** the backend test job executes, **Then** a local DynamoDB emulator is available and the integration tests run against it rather than being skipped.
2. **Given** a DynamoDB integration test fails, **When** the pipeline completes, **Then** the backend check is marked failing and the PR is blocked.
3. **Given** the emulator fails to start, **When** the pipeline runs, **Then** the pipeline fails visibly rather than silently skipping the tests.

---

### User Story 3 — Secure AWS Authentication via OIDC (Priority: P3)

Any pipeline step that needs to communicate with real AWS services (e.g., a future deploy step) authenticates using short-lived credentials obtained through OpenID Connect, with no long-lived AWS access keys stored in GitHub. The IAM role granted to the pipeline is scoped to the minimum permissions required.

**Why this priority**: Establishes the secure credential pattern now so that adding a deploy step later does not require retrofitting secrets management. Long-lived access keys stored as GitHub Secrets are a common credential leak vector.

**Independent Test**: Add a step that calls `aws sts get-caller-identity` using OIDC credentials and confirm it succeeds without any `AWS_ACCESS_KEY_ID` secret being defined in the repository.

**Acceptance Scenarios**:

1. **Given** the pipeline runs on `main` and `vars.AWS_CI_ROLE_ARN` is set, **When** the OIDC step executes, **Then** it obtains temporary credentials via OIDC role assumption — no static access key is used.
2. **Given** `vars.AWS_CI_ROLE_ARN` is not set, **When** the pipeline runs, **Then** the OIDC step is skipped automatically and all other checks still pass.
3. **Given** the OIDC trust is scoped to this repository, **When** a fork or unrelated workflow attempts to assume the same role, **Then** the role assumption is denied.
4. **Given** no long-lived AWS credentials are stored in GitHub Secrets, **When** the repository settings are audited, **Then** no `AWS_ACCESS_KEY_ID` or `AWS_SECRET_ACCESS_KEY` secrets exist.

---

### User Story 4 — Validation on Pushes to Main (Priority: P2)

When code is merged or pushed directly to main, the pipeline runs the same full validation suite. Developers and maintainers can see the current health of the main branch at a glance via the commit status badges.

**Why this priority**: Protects main from broken states introduced by direct commits (such as the hotfix workflow used in this project) or failed merges.

**Independent Test**: Push a commit directly to main and confirm the pipeline triggers and all checks run to completion.

**Acceptance Scenarios**:

1. **Given** a commit lands on main, **When** the pipeline runs, **Then** backend and frontend checks all execute automatically.
2. **Given** all checks pass on main, **When** the pipeline completes, **Then** the commit is marked green in the GitHub commit history.
3. **Given** a check fails on main, **When** the pipeline completes, **Then** the commit is marked red and the team can identify which check broke.

---

### Edge Cases

- What happens when the pipeline itself has a configuration error? The workflow must fail visibly rather than silently pass.
- What happens when a dependency installation step fails (e.g., network issue)? The pipeline should report failure, not a misleading success.
- What happens when tests pass but the build fails? Both must be treated as independent blocking checks.
- What happens if the pipeline is triggered concurrently by multiple pushes to the same PR? Only the latest run should determine the merge status (handled automatically by GitHub).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The pipeline MUST trigger automatically on every pull request targeting `main` (opened, updated, or synchronized).
- **FR-002**: The pipeline MUST trigger automatically on every direct push to `main`.
- **FR-003**: The pipeline MUST run backend tests and confirm they all pass.
- **FR-004**: The pipeline MUST run a backend production build and confirm it succeeds without errors.
- **FR-005**: The pipeline MUST run frontend tests and confirm they all pass.
- **FR-006**: The pipeline MUST run a frontend production build and confirm it succeeds without errors.
- **FR-007**: The pipeline MUST report two named status checks on the PR — `backend` and `frontend` — one per job. Each check reflects the combined result of all steps in that job (tests, coverage, and build).
- **FR-008**: A failing check MUST block pull request merging when branch protection rules are enabled on `main`.
- **FR-009**: The pipeline MUST report the reason for failure (test output, build error) in a way that is visible to the developer without leaving GitHub.
- **FR-010**: The pipeline MUST complete within a time that keeps developer feedback loops short (see SC-004: under 5 minutes for a standard code change).
- **FR-011**: The pipeline MUST run DynamoDB integration tests using a local DynamoDB emulator, without requiring any AWS credentials.
- **FR-012**: The local DynamoDB emulator MUST be started automatically as part of the pipeline and be ready before the integration tests run.
- **FR-013**: Any pipeline step that accesses real AWS services MUST obtain credentials via OIDC role assumption — no static AWS access keys may be stored as repository secrets.
- **FR-014**: The OIDC trust policy MUST be scoped to this specific repository to prevent credential use from forks or unrelated workflows.
- **FR-015**: The IAM role assumed via OIDC MUST follow least-privilege — granting only the permissions required by the pipeline steps that use it. For the initial pipeline (no deploy step), the role requires no permission policies; only the OIDC trust relationship is needed.
- **FR-016**: The pipeline MUST measure backend Go test coverage and fail the backend check if coverage falls below 80% of business logic modules.
- **FR-017**: The pipeline MUST measure frontend test coverage and fail the frontend check if coverage falls below 80% (consistent with the existing Vitest coverage configuration).

### Key Entities

- **Pipeline run**: A single execution of the CI workflow, triggered by a git event. Has a status (pending, passing, failing) and a log of each step's output.
- **Status check**: A named result reported to GitHub for a specific commit. Determines whether the PR is mergeable.
- **Job**: A logical group of steps within a pipeline run (e.g., "backend", "frontend"). Jobs may run in parallel.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of pull requests targeting `main` automatically receive a pipeline run within 30 seconds of being opened or updated.
- **SC-002**: A PR with any failing test or build check cannot be merged — the GitHub merge button is disabled.
- **SC-003**: A developer can identify which specific check failed and read its output without leaving the GitHub PR page.
- **SC-004**: The full pipeline (backend + frontend) completes in under 5 minutes for a standard code change.
- **SC-005**: A pipeline run that passes gives developers confidence equivalent to running `go test ./...` and `npm test` locally — including DynamoDB integration tests, not just stub-based tests.
- **SC-006**: No static AWS access keys exist as repository secrets; all AWS authentication uses short-lived OIDC credentials.
- **SC-007**: The OIDC IAM role cannot be assumed by any workflow outside this repository.
- **SC-008**: The backend CI check fails if Go test coverage drops below 80%; the frontend CI check fails if JavaScript coverage drops below 80%.

## Clarifications

### Session 2026-05-25

- Q: Should the backend CI job fail if Go test coverage drops below 80%? → A: Yes — fail the backend job if Go coverage < 80%, enforced via `go test -coverprofile` and a threshold check. Consistent with the frontend's existing coverage enforcement and the project constitution (Principle II).
- Q: How should the OIDC step appear in the initial workflow file before the IAM role exists? → A: Conditional active step — `if: vars.AWS_CI_ROLE_ARN != ''`. The step is always present in the YAML but skipped automatically until the GitHub Actions variable is set; activates without any code change once the IAM role is ready.

## Assumptions

- The repository is hosted on GitHub and GitHub Actions is available (no self-hosted runners needed for the initial version).
- Branch protection rules will be enabled on `main` separately by the repository owner after the pipeline is working; the pipeline itself only needs to report status checks.
- Unit and handler tests do not require a live AWS environment — they use in-memory SQLite and stub stores, so no AWS credentials are needed for those checks.
- Frontend tests run in a headless environment using jsdom (already the case with Vitest), so no browser binary is needed.
- DynamoDB integration tests use the local emulator; the existing skip-if-no-endpoint pattern means they will run when `DYNAMODB_ENDPOINT` is set and skip gracefully when it is not.
- The local DynamoDB emulator is available as a public container image and requires no licensing or authentication to pull.
- An AWS account and IAM permissions to create an OIDC identity provider and IAM role are available; setting up the OIDC trust is a one-time manual step performed by the repository owner.
- The pipeline runs on the free GitHub Actions tier; paid compute or large runners are out of scope.
