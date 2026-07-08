# Feature Specification: PR Preview Environments

**Feature Branch**: `021-pr-preview-environments`  
**Created**: 2026-05-28  
**Status**: Draft  
**Input**: User description: "I want to have test environments in AWS. For each PR the code would be deployed to a new Lambda and Frontend. The Frontend code will be deployed in the same bucket as the production frontend, but in a directory named after the PR. There would be a new Lambda for each test environment and the path to access the frontend would be the current custom path + / PR name. Each new iteration of the code while working on the PR will be deployed that way. Once the PR is successfully merged an automatic deployment to Prod would be initiated."

## Clarifications

### Session 2026-05-28

- Q: What is the data isolation strategy for preview backend environments? → A: Separate per-PR DynamoDB tables; the recipes table is seeded from a production snapshot, while the users and favorites tables are created empty (previews are public, so production user records and their favorites — PII — are never copied in). Previews have full read/write access to their own isolated tables.
- Q: How does the preview frontend reach the preview backend Lambda? → A: A shared API Gateway routes requests by path prefix (e.g., `/pr-42/*`) to the corresponding PR Lambda; no per-PR API Gateway is provisioned.
- Q: How are preview environment AWS resources (Lambda, DynamoDB tables, API Gateway routes) provisioned and torn down? → A: Via AWS CLI/SDK scripts run directly in the CI pipeline; preview resources are managed outside Terraform to keep the production infrastructure state clean.
- Q: Should preview environment URLs be access-controlled or publicly accessible? → A: Publicly accessible — anyone with the URL can view the preview, identical to production access rules.
- Q: Is preview DynamoDB seed data refreshed on subsequent pushes, or fixed at environment creation? → A: Seeded once when the preview is first created; subsequent pushes to the PR update code only, leaving preview data unchanged.

---

## Overview

Every open pull request gets its own isolated preview environment automatically. When code is pushed to a PR branch, the backend and frontend are deployed to a dedicated AWS environment scoped to that PR. The environment is accessible at a URL derived from the PR identifier. On successful merge to main, production is updated automatically. When a PR is closed or merged, its preview environment is torn down.

This enables reviewers and the author to test the actual running application before any code reaches production, without manual deployment steps.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Preview Environment Created on PR Push (Priority: P1)

When a developer pushes commits to an open pull request, a complete working environment (frontend + backend) is automatically provisioned and accessible at a predictable URL, reflecting the latest code on that branch.

**Why this priority**: This is the core capability. Without it the rest of the feature has no meaning. Reviewers can test changes in a live environment without running anything locally.

**Independent Test**: Open a PR, push a commit, wait for CI to complete — then visit `cocktails.albertomcastro.com/<pr-identifier>`, confirm the app loads and serves the PR's code.

**Acceptance Scenarios**:

1. **Given** a developer pushes commits to an open PR, **When** the CI pipeline completes successfully, **Then** the preview environment is accessible at `cocktails.albertomcastro.com/<pr-identifier>` with the latest code from the branch.
2. **Given** a preview environment already exists for a PR, **When** a subsequent commit is pushed to the same PR, **Then** the existing preview environment is updated in-place (same URL, same backend function name) with the latest code.
3. **Given** a preview environment is deployed, **When** a reviewer visits the preview URL, **Then** they see a fully working application: the frontend loads, API calls resolve, and recipes can be browsed.

---

### User Story 2 — Automatic Production Deployment on Merge (Priority: P1)

When a PR is successfully merged into main, the production environment is automatically updated with the merged code, without any manual deployment step.

**Why this priority**: Removing the manual prod deployment step is the second half of the CI/CD loop. It ensures production is always in sync with main.

**Independent Test**: Merge a PR, wait for CI to complete — then verify `cocktails.albertomcastro.com` (no PR path) serves the newly merged code.

**Acceptance Scenarios**:

1. **Given** a PR is merged into main, **When** the CI pipeline for the merge commit completes successfully, **Then** the production environment at `cocktails.albertomcastro.com` is updated with the merged code.
2. **Given** a production deployment is triggered by a merge, **When** the deployment completes, **Then** all existing production functionality continues to work (no regression).

---

### User Story 3 — Preview Environment Torn Down After PR Close or Merge (Priority: P2)

When a PR is closed (merged or abandoned), its preview environment is automatically removed, freeing AWS resources and keeping the environment inventory clean.

**Why this priority**: Without cleanup, abandoned preview environments accumulate indefinitely and incur ongoing costs.

**Independent Test**: Close or merge a PR — after CI completes the preview URL returns a 404 or is unreachable, and the Lambda function for that PR no longer exists.

**Acceptance Scenarios**:

1. **Given** a PR is merged into main, **When** the merge CI pipeline completes, **Then** the PR's dedicated backend function and PR-specific frontend assets are removed.
2. **Given** a PR is closed without merging (abandoned), **When** the close event triggers CI, **Then** the PR's dedicated backend function and PR-specific frontend assets are removed.
3. **Given** teardown has occurred, **When** a visitor navigates to the former preview URL, **Then** they are served the main production application (CloudFront's SPA fallback returns `index.html` for the now-missing preview assets), not a broken app experience. A hard 404 is not required; serving the main site satisfies this scenario.

---

### Edge Cases

- If the preview deployment fails (build error, misconfiguration), the CI pipeline reports the failure clearly and no partial environment is left behind.
- If two PRs are open simultaneously, they must have fully isolated backend environments (separate backend functions per PR).
- The production deployment must only trigger on pushes/merges to the `main` branch, not on PR branch pushes.
- If a PR is re-opened after being closed, its preview environment is recreated on the next push.
- Preview environments must not interfere with production traffic or production data.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Every push to an open pull request branch MUST trigger an automated deployment to a dedicated preview environment for that PR.
- **FR-002**: The preview environment MUST be accessible at `cocktails.albertomcastro.com/<pr-identifier>`, where `<pr-identifier>` is a consistent, URL-safe identifier derived from the PR number (e.g., `pr-42`).
- **FR-003**: The frontend assets for a PR preview MUST be deployed to a subdirectory of the production S3 bucket named after the PR identifier.
- **FR-004**: Each PR preview MUST have its own dedicated backend function, isolated from production and from other PR previews. API requests to preview backends MUST be routed via a shared API Gateway using path-based routing (e.g., all requests to `/pr-42/api/*` are forwarded to the `pr-42` Lambda).
- **FR-005**: Each PR preview environment MUST have its own dedicated DynamoDB tables (recipes, users, favorites). The recipes table MUST be seeded once from a production data export when the preview is first created; the users and favorites tables MUST be created empty (production user records and favorites MUST NOT be copied into the publicly accessible preview). Previews have full read/write access to their own tables. Subsequent pushes to the PR MUST update only the Lambda code; DynamoDB tables MUST NOT be re-seeded or reset. Previews MUST NOT access or modify production tables.
- **FR-006**: Subsequent pushes to the same PR MUST update the existing preview environment rather than creating a new one.
- **FR-007**: On successful merge to main, an automated production deployment MUST be triggered and completed without manual intervention.
- **FR-008**: When a PR is merged or closed, its dedicated backend function and frontend assets MUST be removed automatically.
- **FR-009**: Preview environments MUST be isolated from production — a failure or data mutation in a preview environment MUST NOT affect production.
- **FR-010**: The CI pipeline MUST report deployment success or failure as a status check visible on the PR.
- **FR-011**: Preview environment URLs MUST be publicly accessible without any additional authentication; access rules are identical to the production site.

### Non-Functional Requirements

- **NFR-001**: A preview environment MUST be accessible within 5 minutes of a push completing the CI build.
- **NFR-002**: Teardown of a preview environment MUST complete within 5 minutes of a PR being closed or merged.
- **NFR-003**: Preview deployments MUST NOT require manual infrastructure changes for each new PR; the system MUST scale to multiple concurrent PRs automatically.
- **NFR-004**: Preview AWS resources (Lambda functions, DynamoDB tables, API Gateway routes) MUST be provisioned and torn down via CI pipeline scripts, without modifying the production Terraform state.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of pushes to open PRs result in a reachable preview environment within 5 minutes of CI completion.
- **SC-002**: 100% of successful merges to main result in an automatic production deployment with no manual steps required.
- **SC-003**: 100% of closed or merged PRs have their preview environment removed within 5 minutes.
- **SC-004**: Zero production incidents caused by preview environment deployments or teardowns.
- **SC-005**: The preview URL for any open PR loads the correct version of the application, matching the latest commit on the PR branch.

---

## Assumptions

- The PR identifier used in URLs and resource names is `pr-<PR number>` (e.g., `pr-42`). This is stable for the lifetime of a PR and URL-safe.
- The existing CloudFront distribution can serve frontend assets from S3 subdirectories without requiring a per-PR CloudFront distribution change; path-based routing to S3 subdirectories is sufficient.
- Each PR preview has its own DynamoDB tables (recipes, users, favorites) provisioned at environment creation time. Only the recipes table is seeded from a production data export; users and favorites start empty. Subsequent pushes to the PR update only the Lambda code; seed data persists unchanged until the environment is torn down. These tables are removed with the rest of the preview environment when the PR is closed or merged.
- The CI system is GitHub Actions, consistent with the existing pipeline.
- Backend functions for preview environments are named deterministically from the PR number (e.g., `cocktails-pr-42-api`).
- Frontend assets for a PR are stored at a dedicated path in the production bucket and served at the corresponding sub-path of the custom domain.
- A single shared API Gateway handles routing for all preview backends using path prefixes (e.g., `/pr-42/api/*` → `cocktails-pr-42-api` Lambda). When a preview is created, a new route is added to the shared API Gateway; when torn down, the route is removed. No per-PR API Gateway is provisioned.
- Preview environment AWS resources are created and destroyed by CI pipeline scripts using the AWS CLI/SDK. These resources are not tracked in the production Terraform state to avoid state pollution.
- PR environments start with an empty users table. Authenticated flows in a preview require creating a user within that preview (the preview shares the production `JWT_SECRET` for token signing but has no seeded production users). Anonymous visitors can browse the seeded recipes without authentication (read access only); creating or modifying data requires logging in as a preview-created user. This is a UX/authorization distinction and does not contradict FR-005 — the preview environment itself has full read/write access to its own isolated tables.
