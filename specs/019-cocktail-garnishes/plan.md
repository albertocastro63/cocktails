# Implementation Plan: Cocktail Recipe Garnishes

**Branch**: `019-cocktail-garnishes` | **Date**: 2026-05-26 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/019-cocktail-garnishes/spec.md`

## Summary

Add an ordered list of free-text garnish entries to the cocktail recipe data model. Garnishes are displayed in italics below the ingredients section on the recipe detail page, and in the hover-preview popover when the ingredient list does not fill the existing `MAX_VISIBLE` cap. Authors add/edit/remove garnishes via a dedicated section in the recipe form. Garnishes are included in export/import round-trips.

## Technical Context

**Language/Version**: Go 1.25 (backend), Vanilla JavaScript ES modules (frontend)  
**Primary Dependencies**: Backend — AWS SDK v2, DynamoDB, SQLite (`modernc.org/sqlite`), `golang-jwt/jwt`, `google/uuid`. Frontend — Vite, Vitest, Tailwind CSS  
**Storage**: Two backends — SQLite (local/dev, JSON-blob columns per field) and DynamoDB (production/Lambda, native list attributes). Both must be updated.  
**Testing**: Backend — Go `testing` package, table-driven tests. Frontend — Vitest with JSDOM  
**Target Platform**: Web SPA (Vite frontend) + Go HTTP backend (Lambda + local server)  
**Performance Goals**: p95 ≤ 200 ms read / ≤ 500 ms write (from constitution). Garnishes are fetched inline with the recipe — no extra requests.  
**Constraints**: SQLite migration is additive and idempotent (ALTER TABLE pattern already established). DynamoDB is schema-less, no migration needed.  
**Scale/Scope**: Single-tenant/small-multi-user cocktail app; garnish list per recipe rarely exceeds 5 entries.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Are functions single-responsibility and below complexity limits? | ✅ Each new function (garnish row factory, garnish popover renderer, garnish detail renderer) is single-purpose and under 40 lines |
| II. Test-First | Are failing tests written before implementation begins? | ✅ TDD cycle enforced; tasks require failing tests first |
| III. UX Consistency | Do all UI surfaces follow the design language and handle loading/empty/error states? | ✅ Garnish section follows existing Ingredients/Steps/Properties pattern; hidden when empty |
| IV. Performance | Do API responses meet p95 ≤ 200 ms read / ≤ 500 ms write and TTI ≤ 3 s? | ✅ Garnishes are inline in the recipe document; no additional API calls introduced |
| Quality Gates | Do all CI checks (lint, coverage ≥ 75%, benchmarks) pass? | ✅ No violations expected; new code follows existing patterns |

## Project Structure

### Documentation (this feature)

```text
specs/019-cocktail-garnishes/
├── plan.md              ← this file
├── research.md          ← Phase 0 output
├── data-model.md        ← Phase 1 output
└── tasks.md             ← Phase 2 output (/speckit-tasks — not yet created)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── model/
│   │   └── model.go            ← add Garnishes []string to Recipe
│   ├── handler/
│   │   ├── recipes.go          ← add garnishes to Create/Update bodies
│   │   └── admin_recipes.go    ← add garnishes to recipeExport, schema, import
│   └── store/
│       ├── sqlite/
│       │   ├── store.go        ← add garnishes column migration
│       │   └── recipes.go      ← add garnishes to all queries and scan functions
│       └── dynamo/
│           └── recipes.go      ← add Garnishes field to recipeItem; update toItem/unmarshalRecipe

frontend/src/
├── components/
│   └── RecipeCard.js           ← update buildIngredientPopover to show garnishes
├── pages/
│   ├── RecipeDetail.js         ← add Garnishes section below Ingredients
│   └── RecipeForm.js           ← add garnishesSection dynamic section
```

**Structure Decision**: Web application (Option 2 from template). Backend and frontend are separate top-level directories. No new files are required — all changes extend existing files.

## Complexity Tracking

No constitution violations requiring justification.
