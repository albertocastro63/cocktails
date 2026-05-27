# Research: Cocktail Recipe Garnishes

**Feature**: 019-cocktail-garnishes  
**Date**: 2026-05-26  
**Status**: Complete — no unknowns remain

## Decision 1: Garnish Storage Format

**Decision**: Store `garnishes` as a `[]string` field on the `Recipe` model (same pattern as `steps`).  
**Rationale**: Steps are already stored as an ordered `[]string` both in the model and in both storage backends. Garnishes have identical characteristics — ordered, free-text, variable count — so reusing the same pattern minimises implementation surface and keeps the two backends in sync.  
**Alternatives considered**: Separate `Garnish` entity with its own table — rejected because garnishes have no independent identity and a join would add latency with no benefit; `properties` map entry — rejected because it loses ordering and provides no structure for display logic.

## Decision 2: SQLite Column Migration

**Decision**: Add a `garnishes TEXT NOT NULL DEFAULT '[]'` column using the idempotent `ALTER TABLE` pattern already established in `store.go` (see `notes` column and profile columns).  
**Rationale**: The codebase already handles `"duplicate column name"` errors by ignoring them. This zero-downtime approach works for SQLite and requires no rebuild of the recipes table.  
**Alternatives considered**: Rebuilding the table (like `rebuildRecipesTable`) — rejected as unnecessarily destructive for a simple additive change.

## Decision 3: DynamoDB Migration

**Decision**: No migration needed. Add `Garnishes []string` to `recipeItem` in `dynamo/recipes.go`. Existing items without the field will deserialise to an empty/nil slice, which is treated the same as an empty garnish list.  
**Rationale**: DynamoDB is schema-less. Existing recipes never had garnishes, so the absent field maps cleanly to the zero value for a slice. This is the same behaviour as other optional fields (`Notes`, `Properties`).

## Decision 4: Hover Preview Space Rule

**Decision**: Show garnishes in the popover only when `ingredients.length < MAX_VISIBLE` (currently 5). When garnishes are shown, display all garnishes (no secondary cap).  
**Rationale**: The clarification specified reusing the existing ingredient cap without defining a new threshold. The cleanest rule: "if the ingredient list was not truncated (no '…' appended), show garnishes." This maps directly to `ingredients.length < MAX_VISIBLE`. Showing all garnishes (not just `MAX_VISIBLE - ingredients.length`) is acceptable because the popover already expands to fit its content.  
**Alternatives considered**: Capping total items (ingredients + garnishes) to MAX_VISIBLE — rejected because it creates confusing partial garnish lists; always showing garnishes — rejected because it violates the spec's "if there is space" requirement.

## Decision 5: Export/Import Schema

**Decision**: Add `"garnishes"` as an optional `array of strings` property to `recipeSchema` in `admin_recipes.go`. Add `Garnishes []string` to `recipeExport`. Update `ExportRecipes` and `ImportRecipes` accordingly.  
**Rationale**: The spec (FR-010, via clarification) requires garnishes to round-trip through export/import. The existing schema already has `additionalProperties: false`, so the field must be explicitly declared or imports with garnishes will be silently stripped.  
**Alternatives considered**: Storing garnishes as a `properties` map entry for backward compatibility — rejected because it loses ordering and does not survive the `additionalProperties: false` schema gate cleanly.
