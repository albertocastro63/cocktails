# Implementation Plan: Related Cocktails

**Branch**: `028-related-cocktails` | **Date**: 2026-07-10 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/028-related-cocktails/spec.md`

## Summary

Add bidirectional, non-transitive "related cocktails" to recipes. A recipe stores a set of related recipe IDs; when a recipe is saved, the system reconciles the symmetric counterpart (adding/removing this recipe's ID on each related recipe) so relations are always two-sided. The recipe detail page renders a "Related cocktails" section at the bottom with alphabetical links (never on the home random cocktail). The recipe form gains a keyboard-operable type-ahead combobox that searches cocktail names (case-insensitive substring) and adds selections as removable chips. Deleting a recipe removes it from every counterpart's list. No new dependencies.

## Technical Context

**Language/Version**: Go 1.25 (backend), ES modules + Vite (frontend)  
**Primary Dependencies**: existing only — `net/http`, aws-sdk-go-v2 (DynamoDB), `modernc.org/sqlite` (local/tests) on the backend; vanilla JS + Vitest/jsdom on the frontend. No new modules.  
**Storage**: Recipe gains a `related_ids` string set. DynamoDB: new attribute on the recipes table item (no new table/GSI). SQLite: idempotent `ALTER TABLE recipes ADD COLUMN related_ids TEXT NOT NULL DEFAULT '[]'` (JSON-encoded, matching the existing `garnishes`/`steps` pattern).  
**Testing**: `go test` (SQLite-backed store tests exercise the symmetric reconciliation + delete cleanup; handler tests via `httptest`); Vitest + jsdom for the combobox and detail rendering.  
**Target Platform**: AWS Lambda behind API Gateway + CloudFront; SPA on S3/CloudFront.  
**Project Type**: Web application (Go backend + vanilla-JS SPA).  
**Performance Goals**: Type-ahead is client-side over a small in-memory name list (no per-keystroke requests); relation reconciliation and delete cleanup are bounded by a recipe's relation count. Meets p95 ≤ 200 ms read / ≤ 500 ms write.  
**Constraints**: Relations must stay symmetric and non-transitive; no self/duplicate relations; a deleted recipe must vanish from all lists (FR-014); the related section must never render on the home random cocktail (FR-011).  
**Scale/Scope**: Low-traffic personal app, hundreds of recipes; a recipe has at most a handful of relations.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Single-responsibility, within complexity limits? | ✅ Reconciliation is one focused store method (normalize → diff → write self → write counterparts); handlers stay thin. |
| II. Test-First | Failing tests before implementation? | ✅ Store tests (symmetry, non-transitivity, dedupe, self-exclusion, delete cleanup) and combobox/detail tests written first. |
| III. UX Consistency | Design language + loading/empty/error states; WCAG AA? | ✅ Combobox is keyboard-operable (arrow/enter/escape) with ARIA listbox roles; related section hidden when empty; reuses existing card/link styling. |
| IV. Performance | p95 ≤ 200 ms read / ≤ 500 ms write, TTI ≤ 3 s? | ✅ Client-side type-ahead over a cached name list; bounded reconciliation writes. Detail enrichment resolves a few names (small N); noted below. |
| Quality Gates | Lint, coverage ≥ 75%, benchmarks pass? | ✅ New store/handler/UI logic is unit-testable; targets ≥ 75% (constitution v2.0.0). |

> Coverage threshold is **≥ 75%** (constitution v2.0.0), not the 80% in the stock template.

**Performance note (N+1)**: The detail endpoint resolves related IDs → names for display. At this scale it is a handful of point reads; if a recipe ever has many relations, batch the lookup. Documented, not optimized now.

## Project Structure

### Documentation (this feature)

```text
specs/028-related-cocktails/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (API contracts)
└── tasks.md             # Phase 2 (/speckit-tasks — not created here)
```

### Source Code (repository root)

```text
backend/
├── internal/model/model.go          # Recipe: + RelatedIDs []string (stored); + Related []RelatedRef (transient, read-only)
├── internal/store/store.go          # RecipeStore: + SetRelated(id, relatedIDs); Delete cleans counterparts
├── internal/store/dynamo/recipes.go # related_ids attribute; toItem/unmarshal; SetRelated; Delete cleanup; name projection
├── internal/store/sqlite/
│   ├── store.go                     # idempotent ALTER TABLE add related_ids column
│   └── recipes.go                   # related_ids read/write; SetRelated; Delete cleanup
├── internal/handler/recipes.go      # Create/Update accept related_ids + reconcile; GetByID enriches related; Names handler
└── cmd/{lambda,server}/main.go      # register GET /api/v1/recipes/names

frontend/src/
├── api/client.js                    # getRecipeNames(); include related_ids in create/update payloads
├── components/RelatedCocktailPicker.js  # NEW keyboard combobox (search names, chips)
├── pages/RecipeForm.js              # embed the picker; send related_ids on save
└── pages/RecipeDetail.js            # render "Related cocktails" section (bottom, alphabetical, links)
```

**Structure Decision**: Reuse the existing web-service + SPA layout. Relations live as a `related_ids` set on the recipe item (no new table/GSI) — symmetric integrity is maintained by a single store method `SetRelated` that reconciles both sides on save, and by extending `Delete` to strip the recipe from its counterparts. A new lightweight `GET /api/v1/recipes/names` feeds the client-side type-ahead. The picker is a small, self-contained frontend component so it can be unit-tested and reused.

## Complexity Tracking

> No constitution violations. Section intentionally empty.
