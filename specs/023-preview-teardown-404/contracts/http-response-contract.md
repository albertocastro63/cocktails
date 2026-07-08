# HTTP Response Contract: PR Preview 404 After Teardown

**Branch**: `023-preview-teardown-404` | **Date**: 2026-07-08

The observable contract for this feature is the HTTP status (and body kind) returned by `https://cocktails.albertomcastro.com` per request class. Each row is an acceptance assertion.

## Contract

| # | Request | Precondition | Expected status | Expected body | Maps to |
|---|---------|--------------|-----------------|---------------|---------|
| C1 | `GET /` | prod deployed | `200` | production SPA HTML | FR-004 |
| C2 | `GET /assets/<existing>` | prod deployed | `200` | asset bytes | FR-004 |
| C3 | `GET /404.html` | prod deployed | `200` | branded not-found HTML | FR-007 |
| C4 | `GET /random-typo` (no such object) | prod deployed | `404` | branded `/404.html` | FR-009 |
| C5 | `GET /pr-<n>/` where preview is **live** | preview deployed | `200` | preview SPA HTML | FR-003 |
| C6 | `GET /pr-<n>/assets/<existing>` (live) | preview deployed | `200` | asset bytes | FR-003 |
| C7 | `GET /pr-<n>/` where preview was **torn down** | after teardown | `404` | branded `/404.html` | FR-001, FR-002 |
| C8 | `GET /pr-9999/` (never existed) | — | `404` | branded `/404.html` | FR-001 |
| C9 | `GET /pr-<n>` and `GET /pr-<n>/` (removed) | after teardown | `404` (both) | branded `/404.html` | Edge: trailing slash |
| C10 | `GET /pr-<n>/api/v1/recipes` (removed) | after teardown | `404` | non-success (not prod/app content) | FR-005 |
| C11 | `GET /api/v1/recipes/<unknown-id>` | prod deployed | `404` | **JSON** error (not HTML) | FR-004 (no regression / bug fix) |
| C12 | `GET /pr-<n>/` re-created after teardown | reopened + redeployed | `200` | preview SPA HTML | FR-006 |

## Notes

- **C4 vs C11**: unknown *frontend* paths (S3 403) yield the branded HTML 404; unknown *API* paths (origin 404) yield JSON. This split is intentional and is what distinguishes this design from a naive "map everything to a 404 page."
- **C5/C6 vs C7**: the live-vs-removed distinction is entirely determined by whether `pr-<n>/index.html` (and assets) exist in S3; no request-time logic beyond the existing SPA-rewrite function is involved.
- **Body assertion for C3/C4/C7/C8**: the response must contain the branded page's marker text (e.g., a "Page not found" heading and a link to `/`), and must NOT contain production application markers that would indicate the SPA was served.
