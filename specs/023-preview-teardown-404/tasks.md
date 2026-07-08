# Tasks: PR Preview 404 After Teardown

**Input**: Design documents from `specs/023-preview-teardown-404/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/http-response-contract.md ✓, quickstart.md ✓

**Tests**: No unit/TDD tests requested. This is a CloudFront config + static-asset change with no business-logic module; correctness is verified by `terraform plan` and the curl-based acceptance checks (the contract in `contracts/http-response-contract.md`), per plan.md Complexity Tracking.

**Note**: This feature **supersedes feature 021 task T032a** (which accepted the SPA fallback for torn-down previews).

---

## Phase 1: Setup & Baseline

**Purpose**: Record the current state and confirm the precondition the design relies on.

- [X] T001 Capture baseline and confirm precondition: record the distribution's current custom error responses (`aws cloudfront get-distribution-config --id EX7HUB6P225MV --query 'DistributionConfig.CustomErrorResponses.Items'` → expect 403→200 and 404→200 `/index.html`) and confirm the frontend bucket policy grants only `s3:GetObject` (`aws s3api get-bucket-policy --bucket cocktails-prod-frontend`), which means S3 returns `403` for missing objects. Note results in the PR description.

**Checkpoint**: Precondition confirmed (S3 misses → 403); baseline recorded for rollback reference.

---

## Phase 2: Foundational — Shared Change (blocks both user stories)

**Purpose**: The single CloudFront change plus the branded page that together deliver the feature. Both user stories are acceptance-verified against this one change.

- [X] T002 [P] Create the branded not-found page at `frontend/public/404.html`: a self-contained page (inline CSS matching the site's stone/amber palette, `<html lang="en">`, a semantic "Page not found" heading, a short plain-language message, and a clearly labeled link back to `/`). No references to hashed build assets.
- [X] T003 [P] Edit `infra/main.tf` (`module.cdn` `custom_error_response`): replace the two existing entries with a single entry `{ error_code = 403, response_code = 404, response_page_path = "/404.html", error_caching_min_ttl = 0 }`, and remove the `404 → 200 /index.html` mapping so API `404`s pass through unchanged.
- [X] T004 Run `terraform fmt`, `terraform validate`, and `terraform plan` from `infra/`; confirm the ONLY planned change is `module.cdn` custom error responses (no other resource drift).
- [X] T005 Verify the page ships in the build: `cd frontend && npm run build`, then confirm `frontend/dist/404.html` exists (so the existing `prod-deploy.yml` S3 sync publishes it to the bucket root).

**Checkpoint**: `terraform plan` shows only the error-response change; `dist/404.html` is produced by the build.

---

## Phase 3: User Story 1 — Removed Preview Returns 404 (Priority: P1) 🎯 MVP

**Goal**: A torn-down or never-created preview URL — and any unknown path — returns HTTP 404 with the branded page, instead of silently serving the production app.

**Independent Test**: Request `https://cocktails.albertomcastro.com/pr-9999/` (no such preview) and a random path; both return `404` with the branded not-found page, not the production SPA.

**⚠️ Ordering**: T006 (publish `404.html`) MUST complete before T007 (apply the CloudFront change), otherwise the distribution would reference a `/404.html` that does not yet exist.

- [X] T006 [US1] Publish `404.html` to production **first**: merge to `main` so `prod-deploy.yml` syncs `dist/404.html` to `s3://cocktails-prod-frontend/404.html` (for pre-merge testing instead: `aws s3 cp frontend/dist/404.html s3://cocktails-prod-frontend/404.html`). Confirm `GET https://cocktails.albertomcastro.com/404.html` returns `200`.
- [X] T007 [US1] Apply the CloudFront change: `terraform apply` from `infra/` (only after T006), then `aws cloudfront create-invalidation --distribution-id EX7HUB6P225MV --paths "/*"`.
- [X] T008 [US1] Acceptance C7/C8/C9: `curl -s -o /dev/null -w "%{http_code}"` for `/pr-9999/` and `/pr-9999` both return `404`; response body contains the branded "not found" marker and does NOT contain production SPA markers.
- [X] T009 [US1] Acceptance C4: an unknown non-preview path (e.g. `/random-typo-<timestamp>`) returns `404` with the branded page.
- [X] T010 [US1] Acceptance C10: a removed preview's API path (e.g. `/pr-9999/api/v1/recipes`) returns `404` (non-success), not production or application content.

**Checkpoint**: Removed/unknown paths return branded `404`; core feature value delivered (MVP).

---

## Phase 4: User Story 2 — Production & Active Previews Unaffected (Priority: P1)

**Goal**: The 404 change causes zero regressions — real production content, API JSON errors, and live previews keep working.

**Independent Test**: Request the production home, an existing asset, and `/404.html` (all `200`); confirm an API 404 returns JSON (not HTML); confirm a live preview still returns `200`.

**Dependency**: Verifies the same change applied in T007.

- [X] T011 [US2] Acceptance C1/C2/C3: `GET /`, an existing `/assets/<hashed>.js`, and `GET /404.html` each return `200` and load normally.
- [X] T012 [US2] Acceptance C11 (regression + bug-fix): `GET /api/v1/recipes/<unknown-id>` returns `404` AND the body is JSON (assert it does NOT contain `<!DOCTYPE html`), confirming API errors are no longer rewritten to HTML.
- [ ] T013 [US2] Acceptance C5/C6 with a live preview: on an open PR with a deployed preview, `GET /pr-<n>/` and an existing `/pr-<n>/assets/*` return `200`, and `GET /pr-<n>/api/v1/recipes` returns `200` JSON. If no live preview is open, verify on the next preview PR and record the result.

**Checkpoint**: No regressions — production, API errors, and live previews behave correctly.

---

## Phase 5: Polish & Cross-Cutting Concerns

- [X] T014 [P] Accessibility & branding review of `frontend/public/404.html`: WCAG 2.1 AA — sufficient color contrast, `lang` attribute, a document `<title>`, semantic landmark/heading structure, and a keyboard-focusable "back to home" link.
- [X] T015 [P] Update `README.md` (Preview Environments section) and add a note in `specs/021-pr-preview-environments/` that torn-down and unknown URLs now return `404` (branded page), explicitly superseding feature 021 task T032a and its US3 acceptance scenario 3 wording.
- [X] T016 Run the full `quickstart.md` acceptance matrix (C1–C12) end-to-end and record pass/fail results.

---

## Dependencies & Execution Order

### Phase order

- **Phase 1 (Setup)** → **Phase 2 (Foundational)** → **Phase 3 (US1)** → **Phase 4 (US2)** → **Phase 5 (Polish)**.
- **Critical intra-US1 ordering**: T006 (publish `404.html`) **before** T007 (apply CloudFront change) to avoid a window where the error page is missing.
- **US2 (Phase 4)** depends on T007 (same applied change); it adds no new build, only acceptance assertions.

### Critical path

```
T001 → T002/T003 (parallel) → T004 → T005 → T006 → T007 → T008/T009/T010 (US1)
                                                        └→ T011/T012/T013 (US2)
                                                        └→ T014/T015 (parallel) → T016
```

### Parallel opportunities

- **T002 + T003**: the static page and the Terraform edit are different files — write in parallel.
- **T008 + T009 + T010** and **T011 + T012 + T013**: independent curl assertions — can run together once T007 is applied.
- **T014 + T015**: docs and accessibility review are independent.

---

## Implementation Strategy

### MVP (User Story 1 only)

1. Phase 1 (baseline) → Phase 2 (create `404.html` + edit `custom_error_response`).
2. Phase 3: publish `404.html`, apply the CloudFront change, verify removed/unknown paths return branded `404`.
3. **STOP and VALIDATE**: `/pr-9999/` and `/random-typo` return `404`; the core value is delivered.

### Full delivery

Continue to Phase 4 (confirm no regressions to production, API errors, and live previews) and Phase 5 (accessibility, docs, full acceptance matrix).
