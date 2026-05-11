# Implementation Plan: Recipe Notes and Full Homepage Display

**Branch**: `002-recipe-notes-homepage` | **Date**: 2026-05-10 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/002-recipe-notes-homepage/spec.md`

## Summary

Adds an optional `notes` plain-text field to the `Recipe` entity. Notes are stored, returned in all recipe API responses, and excluded from full-text search. The recipe create/edit form gains a notes textarea. The homepage is updated to display the full recipe detail (ingredients, steps, properties, notes) rather than a summary card.

This is a purely additive change: no existing endpoints, fields, or behaviors are modified. The existing SQLite schema is migrated with an idempotent `ALTER TABLE` statement.

## Technical Context

**Language/Version**: Go 1.22+ (backend); Node.js 20+ / Vite 5 (frontend) — same as feature 001  
**Primary Dependencies**: Same as feature 001; no new dependencies required  
**Storage**: SQLite (local) — schema migration via `ALTER TABLE ... ADD COLUMN`; DynamoDB (AWS) — additive attribute, no table changes required  
**Testing**: `go test ./...` (backend); Vitest (frontend unit tests) — TDD: failing tests first  
**Target Platform**: Same as feature 001 (local + AWS Lambda)  
**Performance Goals**: Homepage load ≤ 500ms (SC-004); no regression on existing benchmarks  
**Constraints**: Notes excluded from FTS index (SQLite) and from `matchesQuery` (DynamoDB); partial-update pattern must preserve notes when omitted  
**Scale/Scope**: Same as feature 001

## Constitution Check

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Are functions single-responsibility and below complexity limits? | ✅ Each change is isolated to a single layer; scan/marshal helpers remain under 40 lines |
| II. Test-First | Are failing tests written before implementation begins? | ✅ Tasks require failing tests first for each changed function and new UI component |
| III. UX Consistency | Do all UI surfaces follow the design language and handle loading/empty/error states? | ✅ Notes section hidden when empty; homepage full-detail follows existing RecipeDetail pattern |
| IV. Performance | Do API responses meet p95 ≤ 200ms read / ≤ 500ms write and TTI ≤ 3s? | ✅ Notes is a single text column with no indexing overhead; homepage reads same endpoint as before |
| Quality Gates | Do all CI checks (lint, coverage ≥ 80%, benchmarks) pass? | ✅ All changed packages must maintain ≥ 80% coverage; existing benchmarks must not regress |

## Project Structure

### Documentation (this feature)

```text
specs/002-recipe-notes-homepage/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0: decisions and rationale
├── data-model.md        # Phase 1: Recipe entity delta
├── contracts/
│   └── api.md           # Phase 1: API contract delta (notes field addition)
└── tasks.md             # Phase 2 output (created by /speckit-tasks)
```

### Source Code — Files Modified by This Feature

```text
backend/
└── internal/
    ├── model/
    │   └── model.go                  # Add Notes string field to Recipe struct
    ├── store/
    │   ├── sqlite/
    │   │   ├── store.go              # Add ALTER TABLE migration for notes column
    │   │   └── recipes.go            # Add notes to all SQL SELECT/INSERT/UPDATE;
    │   │                             #   scanRecipe/scanRecipes scan notes column;
    │   │                             #   upsertFTS does NOT include notes
    │   └── dynamo/
    │       └── recipes.go            # Add Notes to recipeItem; update toItem/unmarshalRecipe;
    │                                 #   matchesQuery does NOT check notes
    └── handler/
        └── recipes.go                # Add Notes *string to Create/Update body structs;
                                      #   partial-update preserves notes when omitted

frontend/
└── src/
    ├── pages/
    │   ├── Home.js                   # Replace RecipeCard with full recipe detail display
    │   │                             #   (ingredients, steps, properties, notes)
    │   ├── RecipeDetail.js           # Add notes section (after properties)
    │   └── RecipeForm.js             # Add notes textarea field
    └── [tests for above pages]
```

**No new files, packages, or dependencies are required.** All changes are confined to existing files.

## Complexity Tracking

No Constitution violations. No additional complexity justification required.

## Phase 0 Artifacts

- [research.md](research.md) — 4 decisions documented: SQLite migration strategy, FTS exclusion, partial-update pattern, homepage display approach.

## Phase 1 Artifacts

- [data-model.md](data-model.md) — Recipe entity delta: `notes` field addition with storage and search behavior.
- [contracts/api.md](contracts/api.md) — API contract delta: `notes` field added to recipe object schema.
