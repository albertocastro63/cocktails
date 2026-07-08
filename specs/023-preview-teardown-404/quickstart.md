# Quickstart: PR Preview 404 After Teardown

**Branch**: `023-preview-teardown-404` | **Date**: 2026-07-08

Manual/CI acceptance verification for the 404 behavior. Run after `terraform apply` (CloudFront change) and a prod deploy that publishes `404.html`. CloudFront config changes take a few minutes to propagate; allow for that before asserting.

## Prerequisites

- CloudFront distribution `EX7HUB6P225MV` updated with the new `custom_error_response` (403 → 404 `/404.html`, no 404→200 mapping).
- `frontend/dist/404.html` present and synced to `s3://cocktails-prod-frontend/404.html` (via prod deploy).
- CloudFront invalidation `/*` issued after the change (the prod deploy already does this).

## Acceptance checks

```bash
BASE=https://cocktails.albertomcastro.com
code() { curl -s -o /dev/null -w "%{http_code}" "$1"; }
body() { curl -s "$1"; }

# C1 production home still works
test "$(code "$BASE/")" = 200 && echo "PASS C1 home 200" || echo "FAIL C1"

# C3 error page is reachable and returns 200 when requested directly
test "$(code "$BASE/404.html")" = 200 && echo "PASS C3 /404.html 200" || echo "FAIL C3"

# C4 unknown production path -> 404 branded page
test "$(code "$BASE/random-typo-$(date +%s)")" = 404 && echo "PASS C4 unknown 404" || echo "FAIL C4"
body "$BASE/random-typo-xyz" | grep -qi "not found" && echo "PASS C4 branded body" || echo "FAIL C4 body"

# C11 API 404 stays JSON (no regression / bug fix)
test "$(code "$BASE/api/v1/recipes/does-not-exist-xyz")" = 404 && echo "PASS C11 api 404" || echo "FAIL C11"
body "$BASE/api/v1/recipes/does-not-exist-xyz" | grep -qi "<!DOCTYPE html" && echo "FAIL C11 body is HTML" || echo "PASS C11 body not HTML"

# C7/C8 removed or never-created preview -> 404 (use a PR number with no live preview)
test "$(code "$BASE/pr-9999/")" = 404 && echo "PASS C8 never-created preview 404" || echo "FAIL C8"

# C9 trailing-slash variants both 404
test "$(code "$BASE/pr-9999")" = 404 && test "$(code "$BASE/pr-9999/")" = 404 && echo "PASS C9 trailing slash" || echo "FAIL C9"
```

## Full lifecycle check (optional, with a real PR)

1. Open a PR → wait for `deploy-preview` → confirm `GET /pr-<n>/` returns **200** and the preview loads (C5).
2. Confirm `GET /pr-<n>/api/v1/recipes` returns **200** JSON (C6-adjacent).
3. Merge or close the PR → wait for `teardown` to complete (including CloudFront invalidation).
4. Confirm `GET /pr-<n>/` now returns **404** with the branded page (C7) and `GET /pr-<n>/api/v1/recipes` returns **404** (C10).
5. Reopen the PR → wait for redeploy → confirm `GET /pr-<n>/` returns **200** again (C12).

## Notes

- If C4 returns `200`, CloudFront has not yet picked up the config change (propagation) or the old `403→200 /index.html` mapping is still present — re-check the distribution's custom error responses.
- If C4 returns `404` but the body is a raw CDN/XML error instead of the branded page, `404.html` is missing at the bucket root — re-run the prod deploy.
