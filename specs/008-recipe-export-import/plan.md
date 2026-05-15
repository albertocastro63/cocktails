# Implementation Plan: Recipe Export and Import

**Branch**: `008-recipe-export-import` | **Date**: 2026-05-14 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/008-recipe-export-import/spec.md`

## Summary

Add three admin-only operations to the cocktail app: (1) download a JSON Schema document describing the recipe structure, (2) export all recipes as a downloadable JSON file, (3) import recipes from a JSON file with duplicate-skip and all-or-nothing error handling. Backend: three new Go HTTP endpoints protected by `RequireAdmin`. Store: two new `RecipeStore` interface methods (`ListAll`, `ImportBatch`) with SQLite (ACID) and DynamoDB (best-effort) implementations. Frontend: a new `AdminRecipes.js` page with three controls and a `fetchBlob` download helper.

---

## Technical Context

**Language/Version**: Go 1.25.0 (backend), JavaScript ES2020 (frontend)  
**Primary Dependencies**: standard `net/http` (backend), `modernc.org/sqlite`, `aws-sdk-go-v2/dynamodb`, Tailwind CSS, Vitest+jsdom (frontend tests)  
**Storage**: SQLite (default) / DynamoDB (cloud)  
**Testing**: `go test` (backend, table-driven integration tests with real in-process SQLite), Vitest+jsdom (frontend)  
**Target Platform**: Linux server (backend), modern browser (frontend)  
**Project Type**: Web service + SPA  
**Performance Goals**: Schema download < 2 s (SC-001); export 500 recipes < 5 s (SC-002); import 500 recipes < 30 s (SC-003)  
**Constraints**: Import body ≤ 10 MB; no streaming required for ≤ 500 recipes  
**Scale/Scope**: Up to 500 recipes per import/export batch

---

## Constitution Check

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Are functions single-responsibility and below complexity limits? | ✅ Each handler method is focused; `ImportBatch` is split into validate → check-duplicate → create-in-tx |
| II. Test-First | Are failing tests written before implementation begins? | ✅ TDD enforced per tasks.md — each implementation task is preceded by a failing test task |
| III. UX Consistency | Do all UI surfaces follow the design language and handle loading/empty/error states? | ✅ `AdminRecipes.js` uses the amber/stone design system; all three actions show loading/success/error states |
| IV. Performance | Do API responses meet p95 ≤ 200 ms read / ≤ 500 ms write? | ✅ Schema (static) < 10 ms; export (all recipes from SQLite) < 500 ms for 500 records; import uses a single transaction |
| Quality Gates | Do all CI checks (lint, coverage ≥ 80%, benchmarks) pass? | ✅ New handler and store methods covered by integration tests |

---

## Project Structure

### Documentation (this feature)

```text
specs/008-recipe-export-import/
├── plan.md              # This file
├── research.md          # Decisions made in Phase 0
├── data-model.md        # Entity and schema reference
├── quickstart.md        # Integration test scenarios
├── contracts/
│   └── api.md           # API endpoint contracts
└── tasks.md             # Implementation tasks (from /speckit-tasks)
```

### Source Code

```text
backend/
├── internal/
│   ├── handler/
│   │   ├── admin_recipes.go        NEW — ExportSchema, ExportRecipes, ImportRecipes handlers
│   │   └── admin_recipes_test.go   NEW — integration tests for all three handlers
│   └── store/
│       ├── store.go                EDIT — add ListAll, ImportBatch to RecipeStore interface
│       ├── sqlite/
│       │   └── recipes.go          EDIT — implement ListAll, ImportBatch (sql.Tx)
│       └── dynamo/
│           └── recipes.go          EDIT — implement ListAll, ImportBatch (best-effort)
└── cmd/server/
    └── main.go                     EDIT — register three new admin recipe routes

frontend/src/
├── api/
│   └── client.js                   EDIT — add downloadRecipeSchema, exportRecipes, importRecipes
├── pages/
│   ├── AdminRecipes.js             NEW — admin recipe portability page
│   └── AdminRecipes.test.js        NEW — Vitest tests for AdminRecipes page
└── main.js                         EDIT — add /admin/recipes route; update nav admin links
```

---

## Phase 0: Research (complete)

See `research.md`. All unknowns resolved:

- JSON Schema Draft 7 for the schema document
- Client-side file read → JSON POST for import
- `fetch` + Blob + object URL for schema/export downloads
- `ListAll` + `ImportBatch` store methods
- SQLite: `sql.Tx` for atomic import
- DynamoDB: sequential writes + compensation for best-effort all-or-nothing

---

## Phase 1: Design (complete)

### Backend handler: `admin_recipes.go`

Three handler methods on a new `AdminRecipeHandler` struct:

**`ExportSchema`**  
- Responds with `Content-Type: application/json`, `Content-Disposition: attachment; filename="recipe-schema.json"`
- Body: the embedded JSON Schema constant (see `data-model.md`)

**`ExportRecipes`**  
- Calls `store.ListAll()` → encodes result as RecipeExportRecord array (stripping server-generated fields)
- Responds with `Content-Type: application/json`, `Content-Disposition: attachment; filename="recipes-export.json"`
- Empty collection → `[]`

**`ImportRecipes`**  
- Applies `http.MaxBytesReader` (10 MB limit)
- Decodes JSON body: must be a `[]json.RawMessage`
- Validates each element: `name` required, field types checked
- Calls `store.ImportBatch(recipes, claimsFromContext.UserID)`
- Returns `{"imported": N, "skipped": M}` on success
- Returns 400 with specific message on validation failure
- Returns 500 (all rolled back) on unexpected store error

### Store extensions

**`store.RecipeStore` interface** (add to `backend/internal/store/store.go`):
```go
ListAll() ([]*model.Recipe, error)
ImportBatch(recipes []*model.Recipe, creatorID string) (created, skipped int, err error)
```

**SQLite `ListAll`**: `SELECT ... FROM recipes ORDER BY created_at DESC` (no LIMIT)

**SQLite `ImportBatch`**:
1. `tx, err := s.db.Begin()`
2. For each recipe: `SELECT COUNT(*) ... WHERE name = ?` within tx
3. If exists → `skipped++`; else → insert + FTS upsert within tx
4. `tx.Commit()` on success; `tx.Rollback()` on any error (deferred)

**DynamoDB `ListAll`**: paginated `Scan` following `LastEvaluatedKey` until exhausted

**DynamoDB `ImportBatch`**: sequential `ExistsByName` + `Create` per recipe; on error, delete all created IDs for compensation

### Frontend: `AdminRecipes.js`

Page structure (Tailwind amber/stone design):

```
<div class="max-w-4xl mx-auto px-4 py-8">
  <h1> Admin · Recipes </h1>

  <!-- Section 1: Schema -->
  <section>
    <h2> Recipe Schema </h2>
    <p> Download the JSON Schema that defines the recipe format. </p>
    <button data-download-schema> Download Schema </button>
  </section>

  <!-- Section 2: Export -->
  <section>
    <h2> Export Recipes </h2>
    <p> Download all recipes as a JSON file. </p>
    <button data-export-recipes> Export Recipes </button>
  </section>

  <!-- Section 3: Import -->
  <section>
    <h2> Import Recipes </h2>
    <p> Select a JSON file conforming to the recipe schema. </p>
    <input type="file" accept=".json" data-import-file />
    <button data-import-submit> Import </button>
    <div data-import-status></div>
  </section>
</div>
```

**Import flow**:
1. User selects file → FileReader.readAsText → JSON.parse → basic array check
2. On submit: show "Importing…", call `importRecipes(parsed, token)`
3. On success: show "N recipes imported, M skipped"
4. On error: show error message from server

**Download flow** (schema and export):
1. Show "Downloading…"
2. `fetchBlob('/api/v1/admin/schema', token)`
3. `triggerDownload(blob, 'recipe-schema.json')`

### Route and nav updates (`main.js`)

Add route:
```js
{ pattern: /^\/admin\/recipes$/, factory: () => AdminRecipes() }
```

Update nav: when `isAdmin()`, render two links — "Users" → `#/admin/users`, "Recipes" → `#/admin/recipes`.

---

## Complexity Tracking

No constitution violations. All changes are additive (new handler, new store methods, new page). Existing code is unchanged except for the store interface and server router registration.

The DynamoDB `ImportBatch` uses compensation rather than native transactions — this is documented in `research.md` (Decision 5) and acceptable for the ≤500 recipe scope.
