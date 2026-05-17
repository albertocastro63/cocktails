# Implementation Plan: Recipe Ownership and Per-User Recipe Listing

**Branch**: `010-recipe-ownership` | **Date**: 2026-05-17 | **Spec**: [spec.md](./spec.md)

## Summary

Enforce per-user recipe ownership on edit/delete operations (with admin bypass), and add a "My Recipes" listing page accessible from the main nav. The backend `Recipe` model already stores `CreatorID` and the existing Update/Delete handlers already check it — this feature adds the admin bypass, a new `GET /api/v1/recipes/mine` endpoint, and the corresponding frontend page and nav link.

## Technical Context

**Language/Version**: Go 1.22 (backend) / Vanilla JS ES modules with Vite (frontend)  
**Primary Dependencies**: `golang-jwt/jwt/v5`, `aws-sdk-go-v2` (backend); Tailwind CSS, Vitest (frontend)  
**Storage**: DynamoDB (production) + SQLite (development/local)  
**Testing**: Go `testing` package + `testify` (backend); Vitest + `@testing-library/dom` (frontend)  
**Target Platform**: AWS Lambda + API Gateway (backend), CloudFront SPA (frontend)  
**Project Type**: Full-stack web application  
**Performance Goals**: p95 ≤ 200 ms reads, ≤ 500 ms writes (constitution baseline)  
**Constraints**: No new infrastructure required; DynamoDB Scan with FilterExpression is acceptable at current scale (see Complexity Tracking)  
**Scale/Scope**: Small-scale personal app; recipe counts in the hundreds

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Are functions single-responsibility and below complexity limits? | ✅ |
| II. Test-First | Are failing tests written before implementation begins? | ✅ |
| III. UX Consistency | Do all UI surfaces follow the design language and handle loading/empty/error states? | ✅ |
| IV. Performance | Do API responses meet p95 ≤ 200 ms read / ≤ 500 ms write and TTI ≤ 3 s? | ⚠ |
| Quality Gates | Do all CI checks (lint, coverage ≥ 80%, benchmarks) pass? | ✅ |

**IV. Performance — justified violation**: `ListByCreator` in DynamoDB uses a Scan with FilterExpression, which is O(n) on table size. This is consistent with the existing `List()` and `Search()` implementations (also Scan-based). At current scale (hundreds of recipes) the p95 target is met. A GSI on `creator_id` is the production-grade alternative; deferred because it requires a Terraform change and data migration outside this feature's scope.

## Project Structure

### Documentation (this feature)

```text
specs/010-recipe-ownership/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (affected files)

```text
backend/
├── internal/
│   ├── store/
│   │   ├── store.go                        # Add ListByCreator to RecipeStore interface
│   │   ├── sqlite/
│   │   │   └── recipes.go                  # Implement ListByCreator (WHERE creator_id = ?)
│   │   └── dynamo/
│   │       └── recipes.go                  # Implement ListByCreator (Scan + FilterExpression)
│   └── handler/
│       └── recipes.go                      # Admin bypass in Update/Delete; new Mine handler
├── cmd/
│   ├── lambda/main.go                      # Register GET /api/v1/recipes/mine route
│   └── server/main.go                      # Register GET /api/v1/recipes/mine route

frontend/src/
├── api/
│   ├── auth.js                             # Add getUserID()
│   └── client.js                           # Add getMyRecipes()
├── components/
│   └── RecipeCard.js                       # Accept currentUser prop; show edit/delete if owner/admin
├── pages/
│   └── MyRecipes.js                        # New page (same layout as RecipeList, no search bar)
└── main.js                                 # Add /my-recipes route; "My Recipes" nav link; pass currentUser to RecipeCard
```

**Structure Decision**: Web application (existing layout). No new directories needed — all changes fit within the existing `backend/internal/` and `frontend/src/` trees.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| DynamoDB Scan for ListByCreator | Consistent with existing store pattern; no infra change needed | GSI on creator_id would require Terraform change + data migration; disproportionate for current scale |
