# Feature Specification: Recipe Export and Import

**Feature Branch**: `008-recipe-export-import`
**Created**: 2026-05-14
**Status**: Draft
**Input**: User description: "Add export and import features. Define a JSON schema to represent each recipe, only the name of the cocktail is a required field, allow the admin user to download that schema in the admin page. In the admin page add a button to export into JSON all the recipes and download that file. The file is arranged as a JSON array of objects that conform with the JSON schema above. In the admin page allow to import files that conform with the schema above."

## Clarifications

### Session 2026-05-14

- Q: When importing a recipe whose name already exists in the system, what should happen? → A: Skip duplicates — recipes whose name matches an existing recipe are not created; the success message reports both the number imported and the number skipped.
- Q: If an unexpected system error occurs after some recipes have already been created during an import, should those already-created recipes be retained or rolled back? → A: All-or-nothing — if any unexpected error occurs during creation, no recipes from the import batch are retained; the admin sees an error message and can retry.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Download the Recipe JSON Schema (Priority: P1)

An admin user visits the admin page and wants to understand the structure of the recipe data format before preparing an import file. They click a "Download Schema" action and receive a document that precisely defines all allowed fields, their types, and which are required. They can use this document to validate files before importing or to build third-party tooling.

**Why this priority**: Without the schema, users cannot reliably prepare import files. It is the reference contract for the other two operations.

**Independent Test**: Log in as an admin. Navigate to the admin page. Locate the schema download action. Trigger the download. Open the downloaded file and confirm it is a well-formed document describing the recipe structure: `name` is marked as required; `ingredients`, `steps`, `properties`, and `notes` are present as optional fields with their types and shapes described.

**Acceptance Scenarios**:

1. **Given** an admin user is on the admin page, **When** they trigger the schema download, **Then** a file is downloaded to their device immediately.
2. **Given** the downloaded schema file, **When** inspected, **Then** `name` (string) is the only required field.
3. **Given** the downloaded schema file, **When** inspected, **Then** optional fields `ingredients` (list of items with name, quantity, and unit), `steps` (ordered list of text), `properties` (named value pairs), and `notes` (text) are all described.
4. **Given** a non-admin user, **When** they access the admin page, **Then** the schema download action is not available to them.

---

### User Story 2 — Export All Recipes to a JSON File (Priority: P2)

An admin user wants to back up all recipes or migrate them to another system. They click an "Export Recipes" button on the admin page. The system collects every recipe and immediately downloads a single JSON file containing all recipes as an array. The file conforms to the schema defined in User Story 1 and can be re-imported using the import tool.

**Why this priority**: Export is the primary data portability action and the natural source file for the import flow. It must be complete and self-consistent with the schema.

**Independent Test**: Log in as an admin. Ensure at least two recipes exist. Click "Export Recipes". Confirm a file downloads. Open the file and verify it is a JSON array where each object has at minimum a `name` field and any other populated recipe fields. Verify the exported file can be successfully re-imported using the import tool (User Story 3).

**Acceptance Scenarios**:

1. **Given** an admin user is on the admin page, **When** they click "Export Recipes", **Then** a file download begins immediately.
2. **Given** the downloaded export file, **When** inspected, **Then** it is a JSON array of recipe objects, each conforming to the schema.
3. **Given** the system has zero recipes, **When** an admin exports, **Then** the downloaded file contains an empty array (not an error).
4. **Given** a recipe with all fields populated (name, ingredients, steps, properties, notes), **When** exported, **Then** all fields appear in the export file.
5. **Given** a recipe with only a name, **When** exported, **Then** only the name appears (or omitted optional fields are absent/null).
6. **Given** a non-admin user, **When** they access the admin page, **Then** the export action is not available to them.

---

### User Story 3 — Import Recipes from a JSON File (Priority: P3)

An admin user has a JSON file — either exported from this system or manually prepared using the schema — and wants to load its recipes into the system. They use the import control on the admin page to select and submit the file. The system validates the file against the schema and creates the recipes. The admin sees a clear success message telling them how many recipes were imported, or a clear error message if the file is invalid.

**Why this priority**: Import completes the round-trip data portability story. It depends on the schema (P1) and is naturally tested using an export file (P2).

**Independent Test**: Export recipes (User Story 2). Delete all recipes. Use the import tool to upload the exported file. Confirm the same recipes reappear in the recipe list with all their fields intact.

**Acceptance Scenarios**:

1. **Given** an admin user on the admin page, **When** they select a valid JSON file and submit it via the import control, **Then** the recipes in the file are created in the system.
2. **Given** a successfully imported file, **When** the admin views the recipe list, **Then** all imported recipes appear with their name, ingredients, steps, properties, and notes intact.
3. **Given** a valid file is submitted, **When** import completes, **Then** the admin sees a success message stating how many recipes were imported (e.g., "5 recipes imported").
4. **Given** a file that is not valid JSON, **When** the admin tries to import it, **Then** a clear error message is shown and no recipes are created.
5. **Given** a file that is valid JSON but does not conform to the schema (e.g., a recipe object missing the required `name` field), **When** the admin tries to import it, **Then** a clear error message identifies the validation problem and no recipes are created.
6. **Given** an import file containing a recipe whose name already exists in the system, **When** the admin imports the file, **Then** that recipe is skipped and the success message reports both the count of created recipes and the count of skipped duplicates.
7. **Given** a non-admin user, **When** they access the admin page, **Then** the import control is not available to them.

---

### Edge Cases

- What happens when the import file is extremely large (e.g., thousands of recipes)? The system should either complete successfully or provide a clear error if resource limits are reached; it must not silently truncate the data.
- What happens when an imported recipe's `name` already exists in the system? The recipe is skipped (not created, not overwritten). The success message reports the number of skipped recipes alongside the number created.
- What happens when the exported file is re-imported and some recipes have been added or deleted in between? The import should only apply to the recipes in the file, not touch others.
- What happens when `ingredients` contains items with a `name` but missing `quantity` or `unit`? These optional sub-fields should be accepted without error.
- What happens when an unexpected system error occurs after some recipes from the import batch have already been created? The entire import is rolled back — no recipes from that batch are retained — and the admin is shown a clear error message and may retry.
- What happens when the user selects a non-JSON file (e.g., CSV) for import? A clear format error is shown before any processing.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The admin page MUST provide a clearly labelled action to download the recipe schema document. The schema document MUST describe all allowed recipe fields, their types, which are required (`name` only), and the shape of nested structures (`ingredients`, `steps`, `properties`, `notes`).
- **FR-002**: The recipe schema MUST define `name` (text string) as the only required field.
- **FR-003**: The recipe schema MUST define the following optional fields: `ingredients` (an ordered list of items each having a `name` text field and optional `quantity` and `unit` text fields), `steps` (an ordered list of text strings), `properties` (a collection of named text-value pairs), and `notes` (a text string).
- **FR-004**: The admin page MUST provide an "Export Recipes" action that, when triggered, immediately downloads a single file containing all recipes as an array of objects conforming to the recipe schema.
- **FR-005**: The exported file MUST be human-readable and machine-parseable, usable as a valid import file without modification.
- **FR-006**: The admin page MUST provide an import control that allows an admin user to select a local file and submit it for import.
- **FR-007**: On import, the system MUST validate the submitted file: it must be a valid structured data file AND conform to the recipe schema (all recipe objects must have a `name`; other fields must match their defined types if present).
- **FR-008**: If the submitted file passes validation, the system MUST attempt to create a recipe for each object in the array. Recipes whose `name` matches an existing recipe MUST be skipped (not created, not overwritten). The system MUST display a success message stating both the number of recipes created and the number skipped due to name conflicts (e.g., "3 recipes imported, 2 skipped (name already exists)"). If an unexpected system error occurs during creation, the entire import MUST be rolled back — no recipes from the batch are retained — and a clear error message is shown; the admin may retry.
- **FR-009**: If the submitted file fails validation (invalid format or schema violation), or if an unexpected error occurs during import, the system MUST display a human-readable error message describing the problem and MUST NOT retain any recipes from that import attempt.
- **FR-010**: All three actions (schema download, export, import) MUST be accessible only to authenticated admin users. Non-admin users MUST NOT see or be able to invoke these actions.

### Key Entities

- **Recipe Schema Document**: A formal, downloadable document that defines the structure of a recipe for import/export purposes. Describes field names, types, and required status. Downloadable as a file.
- **Recipe Export File**: A structured data file containing an array of recipe objects, each conforming to the Recipe Schema Document. Produced by the export action.
- **Import Result**: The outcome of an import operation — either a count of successfully created recipes or a description of the validation error that prevented import.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An admin user can download the schema document in under 2 seconds from clicking the action.
- **SC-002**: An admin user can export all recipes and receive the downloaded file in under 5 seconds for a collection of up to 500 recipes.
- **SC-003**: An admin user can import a valid export file containing up to 500 recipes, see a success message, and find all imported recipes in the recipe list — all within 30 seconds of submitting the file.
- **SC-004**: 100% of recipe objects in a valid export file are recreated faithfully on import (no field data lost or corrupted).
- **SC-005**: 100% of invalid import files (bad format or schema violations) are rejected with a user-readable error; zero recipes are created from a rejected file.
- **SC-006**: All three admin actions are inaccessible to non-admin users — verified by confirming the actions are absent for non-admin sessions.

## Assumptions

- The admin page already exists and is accessible only to authenticated admin users; this feature adds new sections/controls to that existing page.
- The recipe data model (name, ingredients, steps, properties, notes) is the canonical model already in use by the application; the schema document describes this existing model.
- Import creates new recipes; it does not update or merge existing ones. A recipe whose name matches an existing recipe is skipped; the success message reports both the count created and the count skipped (see Clarifications).
- Export includes all recipes in the system regardless of pagination; there is no filtering or selection in the initial scope.
- The schema document format is a widely-recognised, structured description format that non-technical users can read and that tools can validate against.
- The import file size is expected to be modest (up to 500 recipes); no special large-file streaming is required in the initial scope.
