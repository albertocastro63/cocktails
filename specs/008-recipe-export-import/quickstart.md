# Quickstart: Recipe Export and Import

**Feature**: `008-recipe-export-import`  
**Date**: 2026-05-14

Integration scenarios for manual testing and integration test reference.

---

## Scenario 1: Download the Recipe JSON Schema (US1)

**Precondition**: logged in as admin.

1. Navigate to `#/admin/recipes`.
2. Click "Download Schema".
3. A file `recipe-schema.json` downloads to the device.
4. Open the file. Confirm:
   - `"$schema"` is present.
   - `"required": ["name"]` is present.
   - `"properties"` includes `name`, `ingredients`, `steps`, `properties`, `notes`.
   - `ingredients.items` has a `name` required field with optional `quantity` and `unit`.

**Expected API call**: `GET /api/v1/admin/schema` → 200, JSON Schema body.

---

## Scenario 2: Export All Recipes (US2)

**Precondition**: logged in as admin; at least two recipes exist.

1. Navigate to `#/admin/recipes`.
2. Click "Export Recipes".
3. A file `recipes-export.json` downloads.
4. Open the file. Confirm:
   - Top-level value is a JSON array.
   - Each element has at least a `name` field.
   - Fields `id`, `creator_id`, `created_at`, `updated_at` are absent.
   - All populated fields (ingredients, steps, properties, notes) appear correctly.

**Expected API call**: `GET /api/v1/admin/recipes/export` → 200, JSON array body.

---

## Scenario 3: Export Empty System (US2 edge case)

**Precondition**: logged in as admin; zero recipes exist.

1. Navigate to `#/admin/recipes`.
2. Click "Export Recipes".
3. A file downloads. Confirm the content is `[]`.

**Expected API call**: `GET /api/v1/admin/recipes/export` → 200, body `[]`.

---

## Scenario 4: Import Valid File (US3 — happy path)

**Precondition**: logged in as admin; the system has zero or partial recipes.

1. Prepare a valid JSON file:
   ```json
   [
     { "name": "Aperol Spritz", "steps": ["Mix", "Serve"] },
     { "name": "Mojito", "ingredients": [{ "name": "Rum", "quantity": "2", "unit": "oz" }] }
   ]
   ```
2. Navigate to `#/admin/recipes`.
3. Use the import control to select the file.
4. Click "Import".
5. Confirm success message: "2 recipes imported, 0 skipped".
6. Navigate to `#/recipes`. Confirm both recipes appear.

**Expected API call**: `POST /api/v1/admin/recipes/import` → 200, `{"imported": 2, "skipped": 0}`.

---

## Scenario 5: Import with Duplicate Names (US3 — skip duplicates)

**Precondition**: "Negroni" already exists; "Daiquiri" does not.

1. Import file:
   ```json
   [
     { "name": "Negroni" },
     { "name": "Daiquiri" }
   ]
   ```
2. Confirm success message: "1 recipes imported, 1 skipped".
3. Confirm "Negroni" is unchanged; "Daiquiri" now exists.

---

## Scenario 6: Round-trip — Export then Import (US2 + US3)

1. Create three recipes with all fields populated.
2. Export → download file.
3. Delete all three recipes.
4. Import the downloaded file.
5. Confirm all three recipes reappear with all fields intact.

**Expected result**: `{"imported": 3, "skipped": 0}`. Recipe list shows all three recipes with correct data.

---

## Scenario 7: Import Invalid JSON (US3 — format error)

1. Create a file containing `not json at all`.
2. Import it.
3. Confirm error message identifies the format problem.
4. Confirm zero recipes created.

**Expected API call**: `POST /api/v1/admin/recipes/import` → 400, `{"error": {"code": "BAD_REQUEST", "message": "import file must be a JSON array"}}`.

---

## Scenario 8: Import Schema Violation (US3 — missing name)

1. Import:
   ```json
   [
     { "steps": ["Pour and stir"] }
   ]
   ```
2. Confirm error message: "recipe at index 0: name is required".
3. Confirm zero recipes created.

---

## Scenario 9: Non-admin Access Denied

1. Log in as a non-admin user.
2. Navigate to `#/admin/recipes` directly.
3. Confirm "Access denied. Admin only." message is shown.
4. Confirm schema, export, and import controls are not rendered.

---

## Scenario 10: Schema Download Performance (SC-001)

1. Start a timer.
2. Click "Download Schema".
3. Stop timer when the file appears in downloads.
4. Confirm elapsed time < 2 seconds.

---

## Scenario 11: Export Performance — 500 Recipes (SC-002)

1. Seed the DB with 500 recipes (test fixture or bulk insert).
2. Start a timer.
3. Click "Export Recipes".
4. Stop timer when file appears.
5. Confirm elapsed time < 5 seconds. Confirm file contains exactly 500 objects.

---

## Scenario 12: Import Performance — 500 Recipes (SC-003)

1. Use the 500-recipe export file from Scenario 11 (delete all recipes first).
2. Start a timer.
3. Import the file.
4. Stop timer when success message appears.
5. Confirm elapsed time < 30 seconds. Confirm recipe list shows 500 recipes.
