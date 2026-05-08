# Implementation Plan: Cocktail Recipe App

**Branch**: `001-cocktail-recipe-app` | **Date**: 2026-05-08 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/001-cocktail-recipe-app/spec.md`

## Summary

A web application for storing, browsing, and searching cocktail recipes with a flexible schema (any key-value properties per recipe). The backend is a Go `net/http` service with two entry points: a local HTTP server and an AWS Lambda handler sharing identical application logic. The frontend is a Vite + TailwindCSS SPA. Storage is SQLite locally and DynamoDB on AWS. Authentication is JWT-based; only admin-created accounts can write. The external API is publicly readable.

## Technical Context

**Language/Version**: Go 1.22+ (backend); Node.js 20+ / Vite 5 (frontend)  
**Primary Dependencies**:
- Backend: `golang-jwt/jwt/v5`, `modernc.org/sqlite`, `aws/aws-sdk-go-v2`, `awslabs/aws-lambda-go-api-proxy`
- Frontend: Vite 5, TailwindCSS 3, chart.js (tree-shaken via Vite)

**Storage**: SQLite (local, via `modernc.org/sqlite` — pure Go, no CGo); DynamoDB (AWS)  
**Testing**: `go test ./...` (backend); Vitest (frontend unit tests)  
**Target Platform**: Local (any OS); AWS Lambda (`provided.al2023`, ARM64) + API Gateway + S3 + CloudFront  
**Performance Goals**: Search results ≤ 2s (SC-002); homepage load ≤ 3s (SC-005); API reads p95 ≤ 200ms; writes p95 ≤ 500ms  
**Constraints**: No CGo (Lambda compatibility); minimal frontend dependencies; no self-registration; public read API  
**Scale/Scope**: Personal/small-group use; < a few thousand recipes; < 10 concurrent users

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Are functions single-responsibility and below complexity limits? | ✅ Handler, store, and auth layers are strictly separated; each handler function maps to one endpoint |
| II. Test-First | Are failing tests written before implementation begins? | ✅ Tasks require failing tests first for every handler, store method, and auth helper |
| III. UX Consistency | Do all UI surfaces follow the design language and handle loading/empty/error states? | ✅ TailwindCSS design tokens; loading/empty/error states required for all data-driven components |
| IV. Performance | Do API responses meet p95 ≤ 200ms read / ≤ 500ms write and TTI ≤ 3s? | ✅ SQLite FTS5 and DynamoDB Scan are both sub-100ms at stated scale; CloudFront TTI ≤ 3s |
| Quality Gates | Do all CI checks (lint, coverage ≥ 80%, benchmarks) pass? | ✅ `golangci-lint` + `go test -cover`; Vitest coverage; benchmarks on search handler |

## Project Structure

### Documentation (this feature)

```text
specs/001-cocktail-recipe-app/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0: decisions and rationale
├── data-model.md        # Phase 1: entity definitions and storage schema
├── quickstart.md        # Phase 1: local dev and deployment guide
├── contracts/
│   └── api.md           # Phase 1: REST API contract
└── tasks.md             # Phase 2 output (created by /speckit-tasks)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   ├── server/
│   │   └── main.go              # Local HTTP server entry point
│   └── lambda/
│       └── main.go              # AWS Lambda entry point (wraps same handlers)
├── internal/
│   ├── handler/
│   │   ├── recipes.go           # Recipe CRUD + search + random handlers
│   │   ├── auth.go              # Login handler
│   │   ├── admin.go             # User creation handler
│   │   └── middleware.go        # JWT auth middleware, CORS
│   ├── store/
│   │   ├── store.go             # RecipeStore and UserStore interfaces
│   │   ├── sqlite/
│   │   │   ├── recipes.go       # SQLite RecipeStore implementation
│   │   │   └── users.go         # SQLite UserStore implementation
│   │   └── dynamo/
│   │       ├── recipes.go       # DynamoDB RecipeStore implementation
│   │       └── users.go         # DynamoDB UserStore implementation
│   ├── model/
│   │   └── model.go             # Recipe, Ingredient, User types
│   └── auth/
│       └── jwt.go               # Token issue, parse, claims
├── go.mod
└── go.sum

frontend/
├── src/
│   ├── pages/
│   │   ├── Home.js              # Random recipe + navigation
│   │   ├── RecipeList.js        # Browse + search
│   │   ├── RecipeDetail.js      # Full recipe view
│   │   ├── RecipeForm.js        # Create / edit form (auth-gated)
│   │   └── Login.js             # Login page
│   ├── components/
│   │   ├── RecipeCard.js        # Recipe summary display
│   │   ├── SearchBar.js         # Search input
│   │   ├── IngredientList.js    # Ingredients display
│   │   ├── PropertyTable.js     # Flexible properties display
│   │   └── EmptyState.js        # Empty / error / loading states
│   └── api/
│       └── client.js            # Fetch wrapper for all API endpoints
├── index.html
├── package.json
└── vite.config.js
```

**Structure Decision**: Standard web application layout. Backend and frontend are independent projects under the repository root. The two backend entry points (`cmd/server`, `cmd/lambda`) import from `internal/` only — no application logic leaks into `cmd/`. The store interface allows SQLite and DynamoDB to be swapped via configuration without touching handler code.

## Complexity Tracking

No Constitution violations. No additional complexity justification required.

## Phase 0 Artifacts

- [research.md](research.md) — all NEEDS CLARIFICATION resolved; 7 decisions documented with rationale and alternatives.

## Phase 1 Artifacts

- [data-model.md](data-model.md) — Recipe, Ingredient, User entities; SQLite and DynamoDB schemas; FTS search index design.
- [contracts/api.md](contracts/api.md) — Full REST API contract for all 8 endpoints across public, authenticated, and admin tiers.
- [quickstart.md](quickstart.md) — Local development setup, build commands, and manual AWS deployment steps.
