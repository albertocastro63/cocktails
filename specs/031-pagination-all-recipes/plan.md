# Implementation Plan: Pagination for All Recipes

**Branch**: `031-pagination-all-recipes` | **Date**: 2026-10-24 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/031-pagination-all-recipes/spec.md`

## Summary

The "All Recipes" listing silently drops every recipe beyond the first 20 because the frontend calls the already-paginated `GET /v1/recipes` endpoint (default `limit=20`), renders only `data`, and ignores `total`/`page`. This plan closes that gap with classic page navigation (Previous / Next, current page, total count) in the frontend, while hardening the backend so the reported `total` and reachable set stay correct for catalogs larger than one DynamoDB Scan batch. Sorting moves from client-side (which could only ever sort the current page) to a server-side `sort` parameter so ordering stays consistent across page boundaries.

## Technical Context

**Language/Version**: Go 1.x (backend), ES modules / vanilla JS (frontend, Vite)
**Primary Dependencies**: Backend — `aws-sdk-go-v2` (DynamoDB), standard `net/http`; Frontend — Vite, Tailwind utility classes, native `fetch`
**Storage**: DynamoDB (single recipes table; DynamoDB Local in dev/CI via `DYNAMODB_ENDPOINT`)
**Testing**: Backend — Go `testing` (`go test`); Frontend — Vitest + jsdom
**Target Platform**: Web application (static frontend + Go HTTP API, deployed as Lambda in prod)
**Project Type**: Web application (frontend + backend)
**Performance Goals**: List responses p95 ≤ 200 ms read; listing interactive (TTI) ≤ 3 s
**Constraints**: Catalog is small today (dozens) but the total count must remain correct as it grows past a single ~1 MB Scan page; no new runtime dependencies
**Scale/Scope**: Single public listing page + one REST endpoint + the DynamoDB recipe store scan methods

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Are functions single-responsibility and below complexity limits? | ✅ Shared `scanAll` helper removes duplicated single-Scan logic across store methods; pagination math extracted to one helper. |
| II. Test-First | Are failing tests written before implementation begins? | ✅ Tasks order tests before implementation: store scan-pagination, handler `sort`/`page`, frontend Pagination component + RecipeList paging. |
| III. UX Consistency | Do all UI surfaces follow the design language and handle loading/empty/error states? | ✅ Pagination control reuses existing button styling (see `SortButtonGroup`); empty/loading/error states preserved; controls hidden when a single page. |
| IV. Performance | Do API responses meet p95 ≤ 200 ms read / ≤ 500 ms write and TTI ≤ 3 s? | ✅ Server-side paging returns one page; `scanAll` adds bounded loop over a small table. No regression expected. |
| Quality Gates | Do all CI checks (lint, coverage ≥ 75%, benchmarks) pass? | ✅ New logic is directly unit-tested; coverage threshold is 75% per constitution v2.0.0. |

> Fill in Status as ✅ (pass), ⚠ (violation with justification below), or ❌ (blocker — must fix).

## Project Structure

### Documentation (this feature)

```text
specs/031-pagination-all-recipes/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── rest-recipes-list.md
├── checklists/
│   └── requirements.md  # From /speckit-specify
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── store/
│   │   └── dynamo/
│   │       ├── recipes.go        # List/Search/*: replace single Scan with scanAll; add server-side sort
│   │       ├── scan.go           # NEW: scanAll(ctx, client, input) paginating helper
│   │       └── recipes_test.go   # Extend: multi-page scan, sort order, total correctness
│   └── handler/
│       ├── recipes.go            # Accept & validate `sort` param; pass through; keep page/limit
│       └── recipes_test.go       # Extend: page/limit/sort behavior, out-of-range page
frontend/
├── src/
│   ├── api/
│   │   └── client.js             # getRecipes already forwards params (page, limit, sort)
│   ├── components/
│   │   ├── Pagination.js         # NEW: Prev/Next + "Page X of Y" + total, accessible
│   │   └── Pagination.test.js    # NEW
│   └── pages/
│       ├── RecipeList.js         # Track page/total; render Pagination; move sort to server; reset to page 1 on filter/sort change
│       └── RecipeList.pagination.test.js  # NEW
```

**Structure Decision**: Web application (Option 2). Changes are confined to the recipe store, the recipe list HTTP handler, and the recipes listing page plus a new reusable `Pagination` frontend component. No new services or projects.

## Complexity Tracking

> No constitution violations; section intentionally empty.
