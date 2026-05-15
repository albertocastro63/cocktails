# Research: Recipe Export and Import

**Feature**: `008-recipe-export-import`  
**Date**: 2026-05-14

---

## Decision 1: JSON Schema version for the schema document

**Decision**: JSON Schema Draft 7 (`http://json-schema.org/draft-07/schema#`)  
**Rationale**: Draft 7 is the most widely supported version across validators, editors, and tooling. It is stable, human-readable, and covers all required constructs (required fields, typed arrays, additionalProperties). Non-technical users can open it in any JSON-aware editor and understand the structure.  
**Alternatives considered**:
- Draft 2020-12: Newer but less tool support; overkill for this schema.
- OpenAPI Schema Object: Application-specific; harder for non-technical users to consume.
- YAML format: Slightly more readable but introduces an extra dependency for tools.

---

## Decision 2: Import file handling strategy (client-side read vs. multipart upload)

**Decision**: Client-side file read via `FileReader.readAsText()`, then POST the parsed JSON array as `application/json`  
**Rationale**: Consistent with the existing API pattern (all endpoints accept and return JSON). Avoids multipart parsing complexity in the Go handler. File size for ≤500 recipes is at most ~2 MB — well within what the browser can read synchronously and what Go's default request body limit handles. Enables clean JSON validation in the handler before any DB work begins.  
**Alternatives considered**:
- Multipart file upload (`multipart/form-data`): More HTTP-conventional for file uploads but adds `io.Reader` → JSON parsing complexity in the Go handler and changes the API contract.
- Server-side streaming: Unnecessary; 500 recipes × ~2 KB average = ~1 MB max, processed in memory in milliseconds.

---

## Decision 3: File download with JWT auth (schema and export)

**Decision**: `fetch()` the endpoint with the `Authorization` header; receive the response as a `Blob`; trigger download via a temporary `<a download>` element with `URL.createObjectURL`.  
**Rationale**: The bearer token cannot be passed via a plain `<a href>` or `window.location` since HTTP does not support custom headers that way. The Blob + object URL pattern is the standard browser approach for auth-gated file downloads. No cookies are used in this app.  
**Alternatives considered**:
- Signed one-time download token: Overcomplicated for an admin-only internal tool.
- Iframe/form POST with embedded token: Fragile and non-standard.

---

## Decision 4: Recipe export field set

**Decision**: Export only user-editable fields: `name`, `ingredients`, `steps`, `properties`, `notes`. Strip server-generated fields (`id`, `creator_id`, `created_at`, `updated_at`).  
**Rationale**: Server-generated fields have no meaning on import (the import handler assigns a new ID, creator, and timestamps). Including them would cause confusion and potential conflicts. The exported file is a clean, schema-conformant representation that can be re-imported without modification (FR-005).  
**Alternatives considered**:
- Export all fields including `id`: Would require the import handler to ignore or strip these fields, making the API surface confusing.

---

## Decision 5: Batch import atomicity — SQLite vs DynamoDB

**Decision**:  
- **SQLite** (primary): Use `sql.Tx` (BEGIN/COMMIT/ROLLBACK) — full ACID atomicity. All creates happen in one transaction; any unexpected error triggers automatic rollback.  
- **DynamoDB** (secondary): Sequential per-item writes with a compensation log for rollback on error. Track created IDs; if any write fails, delete all previously created items. This is best-effort all-or-nothing (not ACID), appropriate for a feature with a ≤500 recipe scope.  

**Rationale**: The spec requires all-or-nothing behavior on unexpected errors. SQLite natively supports this. DynamoDB `TransactWriteItems` is limited to 25 items per call, requiring 20 calls for 500 recipes — if call 15 fails, calls 1–14 have committed. Compensation (delete already-created items) is the practical DynamoDB pattern for this scope.  
**Alternatives considered**:
- DynamoDB `TransactWriteItems` in batches: Adds 20× the latency, partial atomicity only, more complex error handling — not worth it for ≤500 recipes.
- Ignore atomicity for DynamoDB: Violates the spec requirement.

---

## Decision 6: Duplicate name check during import

**Decision**: Use `ExistsByName(name)` (already in `RecipeStore`) within the import transaction (SQLite) or inline before each write (DynamoDB). No separate lookup pass needed.  
**Rationale**: Checking existence per-recipe inline is correct for SQLite (name check is inside the transaction, so concurrent imports do not create false positives). For DynamoDB, a pre-scan for all names would require a full table scan; inline check is equivalent and avoids an extra round-trip. Duplicate names are expected to be rare in normal usage.  
**Alternatives considered**:
- Pre-validate all names before starting any writes: Adds a full table scan pass; names could still change between the scan and the writes (race condition). No benefit for the target scale.

---

## Decision 7: New store interface methods

**Decision**: Add two methods to `RecipeStore`:
- `ListAll() ([]*model.Recipe, error)` — returns all recipes without pagination (for export)
- `ImportBatch(recipes []*model.Recipe, creatorID string) (created, skipped int, err error)` — encapsulates transaction/compensation logic per store backend

**Rationale**: Keeps transaction/compensation logic inside the store layer where it belongs, not in the HTTP handler. The handler only needs to call `ImportBatch` and map the result to an HTTP response. Both SQLite and DynamoDB implementations fulfill the interface contract without leaking store internals.  
**Alternatives considered**:
- Handle transaction in the handler by passing a `context` with a `sql.Tx`: Leaks SQLite specifics into the handler and breaks the store abstraction.
- Use only `Create` in a loop from the handler: Cannot provide ACID atomicity across multiple calls.

---

## Decision 8: Request body size limit for import

**Decision**: Apply `http.MaxBytesReader(w, r.Body, 10<<20)` (10 MB) to the import handler body.  
**Rationale**: 500 recipes × 20 KB worst-case per recipe = 10 MB. The limit prevents unbounded memory consumption for oversized uploads while allowing the expected maximum payload. The error message for an exceeded limit is surfaced as a validation error per FR-009.  
**Alternatives considered**:
- No limit: Risk of OOM on large accidental uploads.
- 1 MB: Too tight; large notes fields could push legitimate payloads close to the limit.

---

## Decision 9: Admin Recipes frontend page structure

**Decision**: New `AdminRecipes.js` page at route `#/admin/recipes`. The existing `AdminUserList.js` is unchanged. The nav admin link becomes two links: "Users" (`#/admin/users`) and "Recipes" (`#/admin/recipes`) when admin is logged in.  
**Rationale**: Separation of concerns — user management and recipe portability are distinct features. A single admin page mixing both would grow too large. Adding a new route follows the existing hash-routing pattern with zero refactoring to existing pages.  
**Alternatives considered**:
- Add controls to `AdminUserList.js`: Mixes user management with recipe operations; harder to test independently.
- Single `/admin` overview page linking to sub-sections: Adds a navigation level that doesn't exist in the current app — out of scope.
