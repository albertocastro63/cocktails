# Research: Base Spirit Designation

## Decision 1: Single-Select UX Control

**Decision**: Use a styled checkbox per ingredient row with JavaScript mutual-exclusion logic.

**Rationale**: HTML `<input type="radio">` enforces single-select natively but cannot be deselected by clicking again, violating FR-004 (clearable). A checkbox per row with a `click` handler that (a) clears all other rows' base-spirit checkboxes and (b) allows toggling off the currently checked box satisfies FR-003 and FR-004 with minimal code.

**Alternatives considered**:
- Native radio inputs: single-select is free but clearing requires a hidden "none" radio option — awkward UX.
- A separate dropdown listing ingredient names: decoupled from the row list, harder to read at a glance.

## Decision 2: Backend Validation of Single-Select Constraint

**Decision**: No server-side enforcement of the single-ingredient constraint; frontend is the sole enforcer.

**Rationale**: The base spirit flag is not a security- or billing-sensitive field. Adding a server-side check (count ingredients with `is_base_spirit=true`, reject if >1) adds complexity for no tangible benefit — a malformed API call from a custom client could set two flags, but the display code handles this gracefully by highlighting the first match found. Constitution Principle I (single responsibility, avoid unnecessary complexity) supports leaving enforcement at the UI layer.

**Alternatives considered**:
- Backend validation in the `Create`/`Update` handlers: would require iterating `Ingredients` in the handler, adding a validation function, and writing extra test cases. Rejected as over-engineering.

## Decision 3: Data Storage — Inline Boolean on Ingredient

**Decision**: Add `IsBaseSpirit bool` to `model.Ingredient` with `json:"is_base_spirit,omitempty"` and the corresponding `dynamodbav:"is_base_spirit,omitempty"` in the DynamoDB `ingItem` struct.

**Rationale**: Ingredients are stored inline within the Recipe document (JSON blob in SQLite, list attribute in DynamoDB). The flag belongs with the ingredient it annotates — no join or separate lookup needed. `omitempty` ensures legacy records (where the field is absent) deserialize cleanly to `false`.

**Alternatives considered**:
- Store a single `base_spirit_index int` on the Recipe: avoids changing the Ingredient struct but breaks if ingredient order changes during editing. Fragile.
- Store `base_spirit_name string` on the Recipe: free-text match is brittle if the ingredient name is edited. Rejected.

## Decision 4: Visual Highlight Treatment

**Decision**: Bold ingredient name + amber accent text label "(base spirit)" appended inline — consistent across popover and detail page.

**Rationale**: The existing design system uses `text-amber-700` / `font-semibold` as the established accent vocabulary (seen in section headings, edit controls). Using the same tokens requires no new CSS and is immediately recognisable to users already familiar with the UI.

**Alternatives considered**:
- Icon (star, flame): requires SVG or emoji; adds visual noise in the compact popover.
- Background highlight on the row: more disruptive to the ingredient list layout.

## Decision 5: Export/Import Compatibility

**Decision**: `is_base_spirit` is automatically included in the JSON export format via the model struct; no additional contract changes are needed.

**Rationale**: Feature 008 export uses Go's `encoding/json` on `model.Recipe`, which serialises all exported struct fields. With `omitempty`, the field is omitted for `false`/absent values (clean output for legacy recipes). On import, the JSON decoder populates the field if present and leaves it `false` if absent — round-trip safe by design.
