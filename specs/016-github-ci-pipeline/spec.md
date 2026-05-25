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

### User Story 2 — Validation on Pushes to Main (Priority: P2)

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
- **FR-007**: Each check (backend tests, backend build, frontend tests, frontend build) MUST be reported as an individual named status check on the PR.
- **FR-008**: A failing check MUST block pull request merging when branch protection rules are enabled on `main`.
- **FR-009**: The pipeline MUST report the reason for failure (test output, build error) in a way that is visible to the developer without leaving GitHub.
- **FR-010**: The pipeline MUST complete within a time that keeps developer feedback loops short.

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
- **SC-005**: A pipeline run that passes gives developers confidence equivalent to running `go test ./...` and `npm test` locally.

## Assumptions

- The repository is hosted on GitHub and GitHub Actions is available (no self-hosted runners needed for the initial version).
- Branch protection rules will be enabled on `main` separately by the repository owner after the pipeline is working; the pipeline itself only needs to report status checks.
- Backend tests do not require a live AWS environment — they use in-memory SQLite and stub stores, so no AWS credentials are needed in CI.
- Frontend tests run in a headless environment using jsdom (already the case with Vitest), so no browser binary is needed.
- DynamoDB integration tests (which require `DYNAMODB_ENDPOINT`) are skipped in CI, consistent with their current skip-if-no-endpoint behavior.
- The pipeline runs on the free GitHub Actions tier; paid compute or large runners are out of scope.
