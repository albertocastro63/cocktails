# Feature Specification: PR Preview 404 After Teardown

**Feature Branch**: `023-preview-teardown-404`  
**Created**: 2026-07-08  
**Status**: Draft  
**Input**: User description: "The behavior of a PR preview environment has to be changed so when the PR is no longer deployed in a preview environment and the user navigates to the site https://cocktails.albertomcastro.com/pr-nn/, the site returns a 404 response."

## Clarifications

### Session 2026-07-08

- Q: Should returning 404 apply only to preview URLs, or to any unknown path on the domain? → A: Any unknown path — removed previews and any non-existent production path (e.g. `/random-typo`) return 404; only real, existing content (`/`, `/assets/*`, live previews) returns 200.
- Q: What should the 404 response show the visitor? → A: A simple branded not-found page matching the site's look, with a short message and a link back to the home page.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Removed Preview Returns 404 (Priority: P1)

A reviewer or developer opens a preview URL for a pull request whose preview environment no longer exists — because the PR was merged or closed (and its environment was torn down), or because that PR never had a preview. Instead of being shown the main production application at a `pr-nn` address (which looks like a working preview but isn't), they receive a clear "not found" response.

**Why this priority**: This is the entire feature. Today a removed preview URL silently serves the production site with a `200 OK`, which is misleading — a reviewer can't tell whether they're looking at the PR's code, stale content, or nothing at all. Returning a genuine 404 makes the state of a preview unambiguous.

**Independent Test**: Merge or close a PR so its preview is torn down, wait for teardown to complete, then request `https://cocktails.albertomcastro.com/pr-<number>/`. Confirm the response status is `404` and the main production application is not rendered.

**Acceptance Scenarios**:

1. **Given** a PR whose preview environment has been torn down, **When** a visitor navigates to `https://cocktails.albertomcastro.com/pr-<number>/`, **Then** they receive an HTTP `404` response and are not served the production application.
2. **Given** a PR number that never had a preview environment (e.g., `pr-9999`), **When** a visitor navigates to that preview URL, **Then** they receive an HTTP `404` response.
3. **Given** a torn-down preview, **When** a visitor requests any sub-path or asset under that preview (e.g., `/pr-<number>/some/path` or a preview asset), **Then** the response is `404`, not the production application.

---

### User Story 2 - Active Previews and Production Unaffected (Priority: P1)

While removed previews must return 404, environments that *do* exist must keep working exactly as before: a live preview for an open PR loads normally, and the main production site is completely unaffected.

**Why this priority**: A change that makes removed previews 404 is only acceptable if it does not break live previews or production. This non-regression guarantee is as critical as the primary behavior; shipping the 404 without it would be a net negative.

**Independent Test**: With an open PR that has a live preview, request `https://cocktails.albertomcastro.com/pr-<number>/` and confirm `200` with the PR's application. Separately request `https://cocktails.albertomcastro.com/` and confirm the production site loads normally.

**Acceptance Scenarios**:

1. **Given** an open PR with a live preview environment, **When** a visitor navigates to its preview URL, **Then** the PR's application loads normally with an HTTP `200` response.
2. **Given** the production site, **When** a visitor navigates to `https://cocktails.albertomcastro.com/` and moves through the app, **Then** the production application loads and behaves exactly as it did before this change.
3. **Given** a PR whose preview was torn down and is later re-created (e.g., the PR is reopened and redeployed), **When** a visitor navigates to its preview URL after redeployment, **Then** the preview loads normally (`200`) again, with no manual intervention.

---

### Edge Cases

- **Trailing-slash and bare forms**: `https://.../pr-42` and `https://.../pr-42/` for a removed preview both return `404` (consistent behavior regardless of trailing slash).
- **Preview API path**: Requests to a removed preview's API path (`/pr-<number>/api/...`) also return a non-success response rather than misleading production or application content.
- **Teardown timing**: Between the moment teardown begins and the moment edge caches reflect the removal, a request could briefly still succeed; the 404 must be in effect once teardown (including cache invalidation) has completed.
- **Non-preview unknown paths**: Paths on the production domain that do not correspond to existing content (e.g., `/some-random-path`) also return `404`, consistent with removed previews. Only real, existing content returns `200`.
- **Malformed preview identifiers**: Paths that resemble but are not valid preview identifiers (e.g., `/pr-`, `/pr-abc/`) are treated as ordinary unknown paths and therefore also return `404` when they do not map to existing content.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When no preview environment exists for a given PR identifier — because it was never created or has been torn down — any request to that preview's URL (`/pr-<number>/` and its sub-paths) MUST return an HTTP `404` response.
- **FR-002**: A `404` for a removed preview MUST NOT serve the production application content; a visitor MUST NOT be shown a working-looking site at a removed preview address.
- **FR-003**: Preview environments that currently exist (open PRs with a live preview) MUST continue to return their own content with `200` for their URL and assets, unchanged by this feature.
- **FR-004**: Existing production content (the home page, the application's normal in-app navigation, and all real assets) MUST remain unaffected and continue to load and behave exactly as before. (Requests to paths with no existing content are covered by FR-009.)
- **FR-005**: Requests to a removed preview's API path (`/pr-<number>/api/...`) MUST return a non-success response (`404`) rather than production or application content, so a removed preview is consistently unreachable across both its site and its API.
- **FR-006**: When a preview environment is re-created after having been removed, its URL MUST return to normal (`200`) automatically once redeployment completes, with no manual steps.
- **FR-007**: The `404` response MUST render a simple branded "not found" page that matches the site's visual style, states that the page does not exist in plain language, and provides a link back to the home page. It MUST NOT expose a stack trace, raw storage/CDN error (e.g., XML access-denied), or internal diagnostic.
- **FR-008**: The transition to `404` for a removed preview MUST take effect as part of the existing preview teardown process, without requiring a separate manual action per PR.
- **FR-009**: Any request to the production domain that does not map to existing content — whether a removed/never-created preview URL or an ordinary unknown path (e.g., `/some-random-path`) — MUST return `404`. Only real, existing resources (the production home and its assets, and live previews and their assets) return `200`.

### Key Entities

- **Preview Environment**: A per-PR deployment reachable at a `pr-<number>` URL. Relevant states for this feature are *present* (created and not yet torn down) and *absent* (never created, or torn down). The 404 behavior is driven by the *absent* state.
- **Preview URL**: The address pattern `https://cocktails.albertomcastro.com/pr-<number>/` (and its sub-paths and API path) that maps to a specific Preview Environment.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of requests to a removed or never-created preview URL return HTTP `404`.
- **SC-002**: 0 regressions for existing environments — 100% of requests to live previews and to the production site continue to return their expected content with `200`.
- **SC-003**: A reviewer can tell, within a few seconds and without inspecting page contents, whether a preview is live or removed, because a removed preview returns an unambiguous `404` instead of a loaded page.
- **SC-004**: After a PR is merged or closed, its preview URL returns `404` within the existing teardown completion window (the same window in which the environment's resources are removed).
- **SC-005**: A re-created preview (reopened PR) returns `200` again within the normal preview deployment window, confirming the 404 state is not sticky.
- **SC-006**: The `404` response shows a branded not-found page (site styling, plain-language message, and a working link to the home page) — never a raw storage/CDN error or the production application.

## Assumptions

- The preview URL pattern is `https://cocktails.albertomcastro.com/pr-<number>/`, as established by the existing PR preview environments feature.
- "No longer deployed in a preview environment" covers both a preview that was created and then torn down (PR merged or closed) and a PR identifier that never had a preview.
- This feature builds on the existing preview teardown process; teardown already removes the preview's stored assets and backend, and the desired `404` behavior keys off the absence of those resources.
- The `404` is defined by its HTTP status AND a simple branded not-found page (site styling, plain-language message, link back to the home page). No redirect to the main site is expected (the previous behavior of showing the main site is explicitly what this feature removes).
- Unknown paths on the production domain (not just preview URLs) are in scope: any path that does not map to existing content returns `404`. Because the production application uses client-side (in-app) navigation rather than server-resolved deep-link paths, this does not affect normal use of the site.
- Edge-cache propagation of the `404` follows the same timing and invalidation approach already used by preview teardown.

## Dependencies

- Depends on the existing **PR Preview Environments** feature (feature 021), which defines the preview URL scheme, deployment, and teardown that this feature adjusts.
