# Research: PR Preview 404 After Teardown

**Branch**: `023-preview-teardown-404` | **Date**: 2026-07-08

---

## Decision 1 — Map S3 `403` to a `404` page instead of the SPA fallback

**Decision**: Change the CloudFront distribution's custom error responses so that an origin `403` returns `response_code = 404` with `response_page_path = /404.html`, and remove the existing `404 → 200 /index.html` mapping.

**Rationale**: The frontend S3 bucket uses Origin Access Control and its bucket policy grants CloudFront **only `s3:GetObject`** (verified live — no `s3:ListBucket`). S3 therefore returns **`403` AccessDenied** for any missing object, not `404`. Mapping `403 → /404.html` (404 status) converts every missing-frontend-object case into a real 404:

- Unknown production paths (`/random-typo`) — S3 miss → 403 → 404.
- Torn-down preview paths (`/pr-42/…`) — the `spa_pr_routing` function rewrites them to `/pr-42/index.html`, which no longer exists → S3 403 → 404.

Live previews and real production content still return `200` because their objects exist.

**Alternatives considered**:
- *Keep the SPA fallback and special-case preview paths in the CloudFront Function* — rejected; a CloudFront Function cannot check whether an S3 object exists, so it cannot tell a live preview from a torn-down one.
- *Use S3 static-website hosting with an error document* — rejected; that requires dropping OAC and exposing the bucket via the website endpoint (no OAC, weaker security), a larger and riskier change.

---

## Decision 2 — Remove the `404 → 200 /index.html` mapping (do not remap `404`)

**Decision**: Delete the custom error response for `404` entirely; let origin `404`s pass through unchanged.

**Rationale**: The only origin that returns `404` is the API (API Gateway for an unmatched route, or the Go Lambda's `http.NotFound`). Today's `404 → 200 /index.html` mapping means a legitimate API 404 (e.g., `GET /api/v1/recipes/{unknown}`) is rewritten to `200` HTML — a pre-existing bug that breaks JSON error handling. Removing the mapping lets API 404s return their real `404` JSON. It also satisfies FR-005: a torn-down preview's API path (`/pr-42/api/…`) falls through to the `$default` route → production Lambda → `404`, which now passes through as a genuine non-success response.

**Alternatives considered**:
- *Remap `404 → /404.html` too* — rejected; it would replace API JSON error bodies with the HTML page, breaking programmatic API consumers and the frontend's error parsing.

---

## Decision 3 — Ship the branded page as a standalone `frontend/public/404.html`

**Decision**: Add `frontend/public/404.html` as a self-contained HTML page with **inline CSS** (no dependency on hashed build assets), matching the site's look, stating the page does not exist, and linking to `/`.

**Rationale**: Vite copies `public/` to `dist/` verbatim, so the file lands at `dist/404.html` and the existing `prod-deploy.yml` S3 sync places it at the bucket root — exactly where CloudFront's `response_page_path = /404.html` looks. Inline CSS avoids coupling the error page to per-build hashed asset filenames (which would break whenever the main bundle is rebuilt) and keeps it renderable even if other assets are missing. The custom-error-page fetch is an origin sub-request and does **not** pass through the viewer-request SPA function, so no rewrite interferes.

**Alternatives considered**:
- *Reuse the app's in-app "Page not found" view* — rejected; that view is rendered client-side by the hash router at `200`, and requires the full JS bundle to load. It does not help for an HTTP 404 where we want to avoid serving/booting the app at all.
- *Reference the site's Tailwind stylesheet* — rejected; the stylesheet filename is content-hashed per build, so a static page cannot reliably link it. Inline styles are self-sufficient.

---

## Decision 4 — Rely on `error_caching_min_ttl = 0` + existing invalidations for re-created previews

**Decision**: Keep `error_caching_min_ttl = 0` on the custom error response and rely on the invalidations already issued by preview deploy/teardown (feature 021).

**Rationale**: A reopened PR redeploys its preview, restoring `/pr-nn/index.html`; with zero error caching and the existing `/pr-nn/*` invalidation on deploy, the edge stops serving the cached 404 and returns `200` again (FR-006, SC-005) without new machinery.

**Alternatives considered**:
- *Add a dedicated invalidation step for the 404* — rejected as unnecessary; the existing preview-deploy invalidation already covers the preview path prefix.

---

## Resolved unknowns

| Unknown | Resolution |
|---------|-----------|
| Does S3 (OAC) return 403 or 404 for a missing object? | **403** — bucket policy grants only `s3:GetObject`; verified against the live bucket policy. |
| Will changing error responses break the production SPA? | No — the app uses hash-based navigation; it never relies on server-side `index.html` fallback for unknown paths. |
| Will API error responses be affected? | Improved — removing the `404→200` remap lets API JSON 404s pass through correctly (fixes a latent bug). |
| Where must the 404 page live? | `/404.html` at the S3 bucket root, produced by `frontend/public/404.html` and synced by the existing prod deploy. |
