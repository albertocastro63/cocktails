# Feature Specification: Cocktail Recipe Garnishes

**Feature Branch**: `019-cocktail-garnishes`  
**Created**: 2026-05-26  
**Status**: Draft  
**Input**: User description: "There should be a section for garnishes that should be displayed below the ingredients, these garnishes, if there is space, should appear in the preview that shows when you hover over the recipe card. The type should be in italics to differentiate it from the Base Spirit. These garnishes should be entered separately in the edit/create page, for example: 'Express orange oil over the cocktail', 'Use orange peel to garnish'."

## Clarifications

### Session 2026-05-26

- Q: Should garnish data be included in recipe exports so it round-trips correctly through import? → A: Yes — garnishes must be included in export/import so data is not silently lost.
- Q: What is the concrete threshold that determines when the hover preview is "full" and garnishes are hidden? → A: Reuse the existing preview's ingredient limit (from feature 005) — garnishes fill whatever space remains below that cap; no new threshold is defined.

### Session 2026-05-27

- Q: When garnishes are shown in the hover preview and there are more garnishes than remaining slots, should they be truncated with an ellipsis? → A: Yes — the combined total of ingredients and garnishes shown must not exceed MAX_VISIBLE (5); garnishes fill the remaining slots and an ellipsis is shown if any garnishes are hidden.
- Q: What is the canonical section order for both the recipe edit form and the recipe detail page? → A: Name → Ingredients → Garnishes → Steps → Notes → Properties (edit form and detail page must match).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Add Garnishes When Creating or Editing a Recipe (Priority: P1)

A recipe author can add one or more garnish entries to a recipe on the create/edit page. Garnishes are entered in a dedicated garnish section, separate from the ingredients section. Each entry is entered as a free-text description (e.g., "Express orange oil over the cocktail"). Authors can add multiple garnishes, remove individual ones, and save with any number of garnishes including zero.

**Why this priority**: This is the data-entry foundation; without it no garnish information can be stored or displayed. All other stories depend on it.

**Independent Test**: Create a recipe, add two garnish entries in the garnish section, save, reopen — both entries appear pre-populated in the garnish section.

**Acceptance Scenarios**:

1. **Given** a recipe being created or edited, **When** the author opens the garnish section, **Then** a control is available to add a new garnish entry.
2. **Given** the author types a garnish description and saves, **When** the recipe is reopened, **Then** the garnish text is preserved.
3. **Given** a recipe with multiple garnishes, **When** the author removes one, **Then** only that garnish is removed; the rest are unaffected.
4. **Given** a recipe with no garnishes, **When** saved and reopened, **Then** the recipe is valid and the garnish list is empty.
5. **Given** a garnish entry with no text, **When** the author tries to save, **Then** the blank entry is discarded and not persisted.

---

### User Story 2 - View Garnishes on the Recipe Detail Page (Priority: P2)

When a viewer opens a recipe detail page, a "Garnishes" section appears below the ingredients list showing all garnish entries. Garnish text is rendered in italics to visually distinguish it from the Base Spirit treatment used on ingredients. Recipes with no garnishes show no garnish section.

**Why this priority**: The detail page is the primary reading surface; garnishes must be visible there before the hover-preview optimisation is worth building.

**Independent Test**: Open the detail page for a recipe with garnishes — a "Garnishes" section appears below ingredients with each entry in italics. Open a recipe with no garnishes — no garnish section is visible.

**Acceptance Scenarios**:

1. **Given** a recipe with garnishes, **When** the viewer opens the detail page, **Then** a "Garnishes" section is shown below the ingredients section.
2. **Given** garnishes are displayed, **When** a viewer reads the list, **Then** each garnish entry is rendered in italics.
3. **Given** a recipe with no garnishes, **When** the viewer opens the detail page, **Then** no garnish section is displayed.
4. **Given** a recipe with both a base spirit and garnishes, **When** the viewer reads the detail page, **Then** the base spirit uses its own visual treatment and garnishes use italics — the two styles are visually distinct.

---

### User Story 3 - Garnishes in Hover Preview Card (Priority: P3)

When a viewer hovers over a recipe card, the hover preview shows garnishes below the ingredients if space allows. If the ingredient list fills the preview area, garnishes are omitted; the ingredient list is never truncated to make room for garnishes.

**Why this priority**: The hover preview is a quick-reference surface; garnishes add context when space permits but must not crowd out primary information.

**Independent Test**: Hover over a recipe with few ingredients and garnishes — a garnish section appears below ingredients in the preview. Hover over a recipe with many ingredients that fill the preview — no garnish section is shown and ingredient list is not truncated.

**Acceptance Scenarios**:

1. **Given** a recipe with garnishes and a short ingredient list, **When** the viewer hovers over the recipe card, **Then** garnishes appear below the ingredients in the hover preview.
2. **Given** a recipe with garnishes and an ingredient list that reaches or exceeds the preview's existing ingredient cap, **When** the viewer hovers, **Then** garnishes are not shown and the ingredient list is not truncated.
3. **Given** a recipe with no garnishes, **When** the viewer hovers over the card, **Then** no garnish section appears in the preview.
4. **Given** garnishes are shown in the hover preview, **Then** garnish entries are rendered in italics, consistent with the detail page.

---

### Edge Cases

- A recipe with no garnishes is fully valid; no empty placeholder or prompt should appear in any read-only view.
- Blank or whitespace-only garnish entries must not be saved.
- Recipes created before this feature was introduced have no garnish data; they display as if no garnishes are set, with no visible change to existing views.
- The garnish edit controls are only accessible to users with edit rights (owner or admin); read-only viewers see only the displayed garnish list.
- If a recipe has garnishes but the hover preview is already fully occupied by ingredients (ingredients.length >= MAX_VISIBLE), garnishes are silently omitted — no truncation indicator is shown for garnishes in that case (the ingredient ellipsis already signals truncation).
- If garnishes partially fill the remaining slots (e.g., 4 ingredients + 2 garnishes with MAX_VISIBLE=5), only the fitting garnishes are shown and an ellipsis is appended to the garnish list.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The recipe data model MUST support an ordered list of garnish entries; zero garnishes is a valid state.
- **FR-002**: The recipe create and edit pages MUST include a dedicated garnish section, visually and functionally separate from the ingredients section. The section order on the create/edit form MUST be: Name → Ingredients → Garnishes → Steps → Notes → Properties.
- **FR-003**: Each garnish entry MUST be a free-text description field; the system MUST NOT persist entries with empty or whitespace-only text.
- **FR-004**: The recipe detail page MUST display a "Garnishes" section when at least one garnish is present. The section order on the detail page MUST be: Ingredients → Garnishes → Steps → Notes → Properties.
- **FR-005**: Garnish entries on the detail page MUST be rendered in italics; each garnish is a single text field and the entire entry is italicized.
- **FR-006**: The hover preview MUST display garnishes below ingredients when space is available; the combined total of displayed ingredients and garnishes MUST NOT exceed MAX_VISIBLE (5); garnishes fill the remaining slots and an ellipsis MUST be shown if any garnishes are hidden; when ingredients reach MAX_VISIBLE, garnishes MUST be omitted without truncating the ingredient list.
- **FR-007**: Garnish entries in the hover preview MUST be rendered in italics, consistent with the detail page treatment.
- **FR-008**: Recipes with no garnishes MUST NOT display a garnish section in any read-only view (detail page or hover preview).
- **FR-009**: Legacy recipes with no garnish data MUST continue to display and edit correctly without requiring garnish data.
- **FR-010**: Garnish data MUST be included in recipe exports and correctly restored on import, so that a full export/import round-trip preserves all garnish entries.

### Key Entities

- **Garnish**: A text entry attached to a recipe describing a garnish instruction or preparation step (e.g., "Express orange oil over the cocktail"). Stored as an ordered list within a recipe; no independent identity outside a recipe.
- **Recipe**: Extended with an optional ordered list of garnish entries. A recipe with an empty or absent garnish list is fully valid.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A recipe author can add, edit, and remove garnish entries within a single edit session without page reload, completing each garnish action in under 10 seconds.
- **SC-002**: 100% of recipes with garnishes display the garnish section correctly on the detail page and (where space permits) in the hover preview.
- **SC-003**: No existing recipe data is altered by this feature; all legacy recipes continue to display and save correctly with no garnish section visible.
- **SC-004**: The garnish section is absent from all read-only views for recipes with no garnishes, ensuring no empty state is presented to viewers.
- **SC-005**: Italic rendering of garnish entries is consistent across the detail page and hover preview, verifiable by visual inspection across all recipes with garnishes.

## Assumptions

- Garnishes are ordered; the display order matches the order in which they were added. No drag-to-reorder interface is required for v1 — order is preserved but not explicitly managed by the author.
- "Space" in the hover preview is determined by the existing ingredient cap (MAX_VISIBLE = 5) established in feature 005. Garnishes fill any remaining slots below that cap; if all slots are occupied by ingredients, garnishes are omitted. If garnishes exceed the remaining slots, they are truncated and an ellipsis is shown. No new threshold is defined — the existing limit is reused as a combined cap across ingredients and garnishes.
- The garnish section label is "Garnishes" on both the detail page and hover preview.
- Only recipe owners and admins can add or modify garnishes; all authenticated and anonymous viewers can see them.
- There is no upper limit enforced on the number of garnishes per recipe; in practice, recipes rarely exceed 5 garnish entries.
- Garnishes are not searchable or filterable in v1 — they are purely a display/authoring enhancement.
