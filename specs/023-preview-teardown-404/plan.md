# Implementation Plan: PR Preview 404 After Teardown

**Branch**: `023-preview-teardown-404` | **Date**: 2026-07-08 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/023-preview-teardown-404/spec.md`

## Summary

Today CloudFront maps both `403` and `404` origin errors to `200 /index.html`, so any unknown path — including a torn-down preview URL — silently serves the production SPA. This feature replaces that fallback: missing content returns a real `404` with a small branded not-found page, while real content (home, assets, live previews) still returns `200`, and API JSON error responses are no longer mangled.

The change is almost entirely a one-time CloudFront (Terraform) adjustment plus a static `404.html`. It supersedes feature 021's task T032a (which had accepted the SPA fallback for torn-down previews).

## Technical Context

**Language/Version**: Terraform ≥ 1.10 (HCL) for CloudFront; static HTML/CSS for the error page; existing Node 24 / Vite build (no JS logic added)  
**Primary Dependencies**: AWS CloudFront (`terraform-aws-modules/cloudfront`), S3 frontend bucket with Origin Access Control (OAC), existing Vite frontend build  
**Storage**: Static `404.html` object in the existing `cocktails-prod-frontend` S3 bucket (root)  
**Testing**: `terraform validate` + `terraform plan`; quickstart curl-based acceptance checks; existing Vitest/Go suites remain green (unaffected)  
**Target Platform**: AWS CloudFront distribution `EX7HUB6P225MV` + S3 (us-east-1)  
**Project Type**: Web application — infrastructure change + one static asset  
**Performance Goals**: 404 page served from the CloudFront edge; no change to any API hot path; production TTI unchanged  
**Constraints**: MUST NOT regress production, active previews, or API JSON error responses; change must be a one-time Terraform apply (no per-PR work)  
**Scale/Scope**: Single distribution, global; one `404.html`; one `custom_error_response` block

### Key technical facts (verified against the live environment)

- The frontend S3 bucket policy grants **only `s3:GetObject`** to CloudFront (no `s3:ListBucket`), so S3 returns **`403` AccessDenied** for a missing object — it never returns `404`.
- Therefore mapping **`403` → a 404 page** covers every "missing frontend object" case: unknown production paths and torn-down preview paths (`/pr-nn/index.html` absent after teardown).
- The only origin that returns **`404`** is the API (API Gateway / the Go Lambda's `http.NotFound`). Removing the `404 → 200 /index.html` custom response lets those pass through unchanged, which *fixes* a pre-existing bug where API 404s were rewritten to `200` HTML.
- The production app uses **hash-based in-app navigation** (`location.hash`), so it never depends on the server returning `index.html` for an unknown path — dropping the SPA fallback is safe.

## Constitution Check

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Single-responsibility, within complexity limits? | ✅ One static `404.html` (no logic) + one `custom_error_response` block edit |
| II. Test-First | Failing tests written before implementation? | ⚠ No unit-testable business logic (infra + static asset); verified via `terraform plan` and an executable quickstart acceptance script. Justified in Complexity Tracking |
| III. UX Consistency | Design language + loading/empty/**error** states handled, WCAG 2.1 AA? | ✅ Adds a proper branded **error state** where none existed; page uses semantic HTML, AA-contrast colors matching the site, and a clear "back to home" link |
| IV. Performance | p95 ≤ 200 ms read / ≤ 500 ms write, TTI ≤ 3 s? | ✅ No API path changes; tiny edge-served static page; removes an origin round-trip for unknown paths |
| Quality Gates | Lint, coverage ≥ 75%, benchmarks pass? | ✅ No application logic added; existing coverage unchanged |

## Project Structure

### Documentation (this feature)

```text
specs/023-preview-teardown-404/
├── plan.md              # this file
├── research.md          # Phase 0 — decisions & alternatives
├── data-model.md        # Phase 1 — response-behavior model (not domain data)
├── quickstart.md        # Phase 1 — curl-based acceptance verification
├── contracts/
│   └── http-response-contract.md   # path class → status code contract
└── tasks.md             # /speckit-tasks output (not created here)
```

### Source Code (repository root)

```text
frontend/
└── public/
    └── 404.html          # NEW — standalone branded not-found page (inline CSS, link home)

infra/
└── main.tf               # EDIT — module.cdn custom_error_response:
                          #   403 → response_code 404, page /404.html
                          #   remove the 404 → 200 /index.html mapping
```

**Structure Decision**: Existing web application (`frontend/` + `infra/`). No backend changes. The frontend change is a single static file placed in `frontend/public/` so Vite copies it verbatim to `dist/404.html`, which the existing `prod-deploy.yml` already syncs to the S3 bucket root. The infrastructure change is a one-time edit to the CloudFront distribution's custom error responses in `infra/main.tf`, applied via `terraform apply`.

## Architecture

### Response behavior after the change

```
Request                                   Origin result            CloudFront response
────────────────────────────────────────────────────────────────────────────────────
GET /                        S3 index.html present (200)      → 200  (unchanged)
GET /assets/app.js           S3 object present (200)          → 200  (unchanged)
GET /404.html                S3 object present (200)          → 200  (the error page itself)
GET /random-typo             S3 miss → 403                    → 404  /404.html  (CHANGED)
GET /pr-24/   (live)         spa-fn → /pr-24/index.html (200) → 200  (unchanged)
GET /pr-24/foo (live)        spa-fn → /pr-24/index.html (200) → 200  (SPA in-app not-found)
GET /pr-42/   (removed)      spa-fn → /pr-42/index.html → 403 → 404  /404.html  (CHANGED)
GET /api/v1/recipes/bad      API Gateway → Lambda 404 JSON    → 404  JSON  (was 200 HTML — FIXED)
GET /pr-42/api/... (removed) API GW $default → Lambda 404     → 404  (non-success; FR-005)
```

- The `viewer-request` SPA function (`spa_pr_routing`, feature 021) is unchanged; it only rewrites extensionless `/pr-<digits>/…` paths to `/pr-<digits>/index.html`. Whether that index.html exists is what distinguishes a live preview (200) from a removed one (403 → 404).
- Custom error responses are distribution-wide, but because S3 emits `403` (not `404`) for misses and the API emits `404` (not `403`) for not-found, mapping only `403 → /404.html` cleanly targets frontend misses while leaving API JSON `404`s intact.

### One-time Terraform change (`infra/main.tf`, `module.cdn`)

```hcl
custom_error_response = [
  {
    error_code            = 403
    response_code         = 404
    response_page_path    = "/404.html"
    error_caching_min_ttl = 0
  }
]
# (the previous 403→200 and 404→200 /index.html entries are removed)
```

`error_caching_min_ttl = 0` avoids sticky 404s so a re-created preview (reopened PR) serves `200` again as soon as its assets return (FR-006), backed by the existing teardown/deploy CloudFront invalidations.

## Phase 0: Research

See [research.md](research.md). All unknowns resolved; no `NEEDS CLARIFICATION` remain.

## Phase 1: Design

- [data-model.md](data-model.md) — models the response behavior (path classes → status), since the feature has no domain data.
- [contracts/http-response-contract.md](contracts/http-response-contract.md) — the testable status-code contract per path class.
- [quickstart.md](quickstart.md) — curl-based acceptance verification (the executable "test" for this infra change).

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| §II Test-First satisfied by acceptance verification rather than unit tests | The change is a CloudFront config edit plus a static HTML page — there is no business-logic module to unit-test. Correctness is a property of edge routing, only observable end-to-end. | A unit test would have to mock CloudFront's error-response engine, testing the mock rather than the behavior. Instead, `terraform plan` validates the config change and the quickstart curl matrix asserts the real status codes per path class (unknown→404, home→200, live preview→200, API 404→JSON). This mirrors feature 021's precedent for infra-only changes. |
