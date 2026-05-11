# Feature Specification: Recipe Notes and Full Homepage Display

**Feature Branch**: `002-recipe-notes-homepage`  
**Created**: 2026-05-10  
**Status**: Draft  

## Clarifications

### Session 2026-05-10

- Q: Should recipe notes be readable by anyone who can access the recipe, or only by the recipe's creator? → A: Publicly readable — anyone who can read the recipe (authenticated or not) sees the notes field.
- Q: Should the recipe create/edit form include a notes input field so authors can enter notes through the UI? → A: Yes — the create/edit recipe form gains a notes textarea so authors can add/edit notes through the UI.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Add and Edit Notes on a Recipe (Priority: P1)

As a recipe author, I want to add notes to any recipe so I can record personal observations, substitutions tried, or reminders that are not meant to be part of the searchable recipe content.

**Why this priority**: Notes are a net-new data field on the Recipe entity. Establishing this field is a prerequisite for displaying it. It is the core data change that unlocks both features.

**Independent Test**: Can be fully tested by creating a recipe with notes, retrieving it, verifying the notes field is present and contains the correct value, and confirming a search query does not surface recipes via their notes content.

**Acceptance Scenarios**:

1. **Given** an authenticated recipe author, **When** they create a new recipe and include a notes value, **Then** the notes are saved and returned in the recipe detail response.
2. **Given** an authenticated recipe author, **When** they update an existing recipe and set the notes field, **Then** the updated notes are persisted and visible when retrieving the recipe.
3. **Given** an existing recipe with notes containing the word "secret", **When** a user searches for "secret", **Then** the recipe is NOT returned in search results (notes are excluded from search).
4. **Given** an unauthenticated visitor, **When** they retrieve a recipe by ID, **Then** the notes field is included in the response (notes are readable by anyone who can read the recipe).
5. **Given** an authenticated author, **When** they update a recipe and omit the notes field, **Then** the existing notes value is preserved (partial update behaviour).
6. **Given** an authenticated author, **When** they set notes to an empty string, **Then** the notes field is saved as empty and the previous value is replaced.

---

### User Story 2 - Full Recipe Display on the Homepage (Priority: P2)

As a visitor to the app, I want to see the complete details of the randomly selected featured recipe on the homepage so I can immediately read the full recipe without navigating elsewhere.

**Why this priority**: This is a display enhancement that depends on the recipe data structure (including notes from US1). It delivers the most visible user-facing improvement — the homepage currently shows only the title of the featured recipe.

**Independent Test**: Can be fully tested by loading the homepage and verifying that the featured recipe's ingredients, steps, properties, and notes are all visible on the page without any additional navigation.

**Acceptance Scenarios**:

1. **Given** at least one recipe exists, **When** a visitor loads the homepage, **Then** the featured recipe's name, all ingredients (with quantities and units), all ordered steps, all properties, and notes are displayed.
2. **Given** a featured recipe has no notes, **When** the homepage loads, **Then** the notes section is either hidden or shown as empty — the page does not error or show a broken placeholder.
3. **Given** a featured recipe has no properties, **When** the homepage loads, **Then** the properties section is hidden or shown as empty without errors.
4. **Given** no recipes exist in the system, **When** a visitor loads the homepage, **Then** an appropriate empty state message is displayed (existing behaviour preserved).

---

### Edge Cases

- What happens when a recipe note is very long (e.g., thousands of characters)? The system stores it faithfully; no length limit is imposed unless explicitly specified.
- What happens when a recipe is updated by a non-creator who somehow submits a notes field? The existing 403 Forbidden logic applies — only the creator may update the recipe at all.
- What happens if the notes field contains special characters or multi-line text? The field is treated as plain text and stored/returned as-is.
- What happens if a recipe with notes is deleted? The notes are deleted along with the recipe (cascade).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The Recipe entity MUST support an optional `notes` text field that stores free-form text authored by the recipe creator.
- **FR-002**: The `notes` field MUST be returned as part of every recipe response (create, update, get by ID, list, random), when present.
- **FR-003**: The `notes` field MUST be excluded from full-text search indexing so that searches do not match recipes based on notes content.
- **FR-004**: When creating a recipe, the author MAY include a `notes` value; if omitted the field defaults to empty/absent.
- **FR-005**: When updating a recipe, the author MAY include a `notes` value; if omitted the existing notes value MUST be preserved (consistent with the existing partial-update behaviour for all recipe fields).
- **FR-006**: Only the recipe creator may write or modify the `notes` field, consistent with the existing creator-only edit restriction.
- **FR-007**: The homepage MUST display the full detail of the randomly selected recipe, including: name, all ingredients with quantities and units, all ordered steps, all key-value properties, and notes (when present).
- **FR-008**: When the featured recipe has no notes, the homepage MUST NOT display a broken or placeholder notes area — it is either hidden or shown as empty gracefully.
- **FR-009**: The existing public API contract for recipe responses MUST be maintained; `notes` is additive and must not break existing consumers that ignore unknown fields.
- **FR-010**: The recipe creation and editing form MUST include a notes textarea so that authenticated authors can add or edit notes through the user interface without requiring direct API access.

### Key Entities

- **Recipe**: Gains a new optional `notes` field (plain text, no length constraint). All other existing fields are unchanged. The `notes` field is read-publicly, write-creator-only, and excluded from search indexing.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A recipe created with notes is retrievable with the exact notes value intact — 100% fidelity, no truncation or transformation.
- **SC-002**: A keyword present only in a recipe's notes does not appear in search results — notes-only searches return 0 matching recipes.
- **SC-003**: The homepage displays all recipe fields for the featured recipe in a single page load — no additional navigation or interaction required.
- **SC-004**: Homepage load time with full recipe detail is within 500ms under normal conditions (same baseline as the existing homepage).
- **SC-005**: Existing recipe search and browse behaviour is unaffected — all previously passing acceptance tests continue to pass.

## Assumptions

- Notes are visible to any user who can read a recipe (public read, same as other recipe fields). There is no private/owner-only visibility for notes.
- There is no maximum length enforced on the notes field by the system; very long notes are stored as-is.
- Notes are plain text only; no markdown rendering or formatting is required for this feature.
- The homepage currently shows only the recipe name/title for the featured recipe — this is the baseline being extended to show full detail.
- The partial-update behaviour already established for ingredients, steps, and properties applies equally to the new notes field.
- The external API contract (`/api/v1/recipes` and related endpoints) is additive — adding `notes` to the response does not constitute a breaking change.
