# Research: Recipe Notes and Full Homepage Display

**Branch**: `002-recipe-notes-homepage` | **Date**: 2026-05-10

## Decision 1: SQLite Schema Migration Strategy

**Decision**: Add `notes` column via `ALTER TABLE recipes ADD COLUMN notes TEXT NOT NULL DEFAULT ''` inside the existing `migrate()` function, with the specific "duplicate column name" error silently ignored to make the migration idempotent.

**Rationale**: The project uses a hand-written `migrate()` function (no migration library). SQLite's `ALTER TABLE ADD COLUMN` fails if the column already exists, so a fresh database and an upgraded existing database must both work without error. Ignoring the duplicate-column error is the simplest idempotent approach within the existing pattern. Using `IF NOT EXISTS` syntax requires SQLite ≥ 3.37.0; while `modernc.org/sqlite` bundles a recent enough version, the error-ignore approach is more portable and explicit.

**Alternatives considered**:
- Versioned migration table (e.g., `schema_migrations`) — introduces a new mechanism not present in the project; overkill for a single column addition.
- `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` — valid in SQLite 3.37+, but less obvious to readers unfamiliar with that version gate.

---

## Decision 2: FTS Exclusion of Notes

**Decision**: In `sqlite/recipes.go`, the `upsertFTS` function builds `searchText` by concatenating name, ingredient names, steps, and property values. The `notes` field is simply **not appended** to this string. In `dynamo/recipes.go`, the `matchesQuery` function checks name, ingredients, steps, and properties — notes are not added to this check.

**Rationale**: The spec is explicit that notes must not be searchable (FR-003). Exclusion by omission (never adding the field to the search index) is safer and simpler than exclusion by filtering — there is no risk of accidental inclusion through a future refactor that copies all fields.

**Alternatives considered**:
- Separate FTS column with `content=''` for non-indexed notes — unnecessary complexity; the FTS table already works by selective inclusion.

---

## Decision 3: Partial Update Pattern for Notes

**Decision**: In the `Update` handler's body struct, `Notes` is declared as `*string` (pointer to string), consistent with the existing `Name *string` pattern. If `Notes` is `nil` in the decoded request body (field omitted by caller), the existing notes value from `GetByID` is preserved. If `Notes` points to an empty string, the notes are cleared.

**Rationale**: The existing partial-update pattern already uses `*string` for `Name` to distinguish "not provided" from "set to empty". Applying the same pattern to `Notes` ensures consistent behaviour across all optional fields and aligns with FR-005.

**Alternatives considered**:
- Always-required notes field — breaks existing clients that don't send notes.
- JSON `null` vs omitted distinction via a custom unmarshaler — more complex than the pointer pattern already in use.

---

## Decision 4: Homepage Full Recipe Display

**Decision**: `Home.js` is updated to render a full recipe detail view inline, reusing the existing `IngredientList` and `PropertyTable` components. The display pattern mirrors `RecipeDetail.js` sections (title → ingredients → steps → properties → notes), but without the edit/delete controls (which are creator-only and do not belong on the homepage). No new shared component is extracted.

**Rationale**: The homepage is the only place that currently shows a summary card; `RecipeDetail.js` already has the full display pattern. Since the spec requires "display the whole featured recipe" (FR-007), reusing existing components while keeping homepage-specific layout in `Home.js` avoids premature abstraction. The two pages have different surrounding context (homepage vs detail navigation, presence of controls), so a shared component would need conditional logic that adds complexity without benefit.

**Alternatives considered**:
- Extract a `RecipeFullDetail` component shared by Home and RecipeDetail — would require conditional edit/delete controls, adding complexity for no immediate gain; deferred until a third consumer appears.
- Use `RecipeDetail` component directly in Home — not possible; RecipeDetail fetches by ID, while Home uses the random endpoint.
