# Data Model: PR Preview 404 After Teardown

**Branch**: `023-preview-teardown-404` | **Date**: 2026-07-08

This feature introduces no domain entities or persisted data. The relevant "model" is the mapping from **request path class** and **resource existence** to the **HTTP response**. This section captures that model so acceptance tests can assert it directly.

---

## Path classes

| Class | Example | Backing origin | Exists when |
|-------|---------|----------------|-------------|
| Production root | `/` | S3 `index.html` | Always (after any prod deploy) |
| Production asset | `/assets/index-*.js` | S3 object | Referenced object present |
| Error page | `/404.html` | S3 object | Always (shipped in every prod build) |
| Unknown production path | `/random-typo` | S3 (miss) | Never |
| Live preview root/sub-path | `/pr-<n>/`, `/pr-<n>/foo` | S3 `pr-<n>/index.html` (via SPA rewrite) | While PR preview is deployed |
| Removed preview path | `/pr-<n>/…` after teardown | S3 (miss) | Never (assets deleted) |
| API path | `/api/v1/…`, `/pr-<n>/api/…` | API Gateway → Lambda | Route/resource dependent |

## Resource state

The only state that matters is **object existence in S3** (frontend) and **route/resource existence** (API):

- `present` → origin returns `200`.
- `absent` (frontend) → S3 returns `403` (OAC, GetObject-only) → CloudFront serves `/404.html` with status `404`.
- `absent` (API) → API returns `404` (JSON or Lambda not-found) → passes through unchanged.

## Response mapping (the contract this feature establishes)

| Request | Resource state | Response status | Response body |
|---------|----------------|-----------------|---------------|
| `/`, `/assets/*`, `/404.html` | present | `200` | the object |
| `/random-typo` and any other unknown non-preview path | absent | `404` | branded `/404.html` |
| `/pr-<n>/` (+ sub-paths) — live preview | present | `200` | preview SPA |
| `/pr-<n>/` (+ sub-paths) — removed/never-created | absent | `404` | branded `/404.html` |
| `/api/…` or `/pr-<n>/api/…` — found | present | `200` | JSON |
| `/api/…` — not found | absent | `404` | JSON (unchanged, no longer mangled to HTML) |
| `/pr-<n>/api/…` — removed preview | absent | `404` | non-success (Lambda not-found) |

## State transitions (preview lifecycle → response)

```
never created ──deploy──▶ live (200) ──teardown──▶ removed (404) ──reopen+deploy──▶ live (200)
     │                                                   │
     └────────────────── 404 ───────────────────────────┘
```

The 404 state is not sticky: `error_caching_min_ttl = 0` plus the existing preview-deploy CloudFront invalidation return the path to `200` once the preview is redeployed.
