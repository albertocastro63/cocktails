# Feature Specification: Base Spirit Designation for Recipe Ingredients

**Feature Branch**: `011-base-spirit`  
**Created**: 2026-05-17  
**Status**: Draft  
**Input**: User description: "Add base spirit property, this can be a checkmark on the ingredients, only one can be selected however and the selection can be cleared, there is no requirement for any drink to have a base spirit. In the list of ingredients shown in the popover highlight the base spirit, do the same in the recipie descriptions."

## Clarifications

### Session 2026-05-17

- Q: How should the new structured `is_base_spirit` flag relate to existing free-form `"Base spirit"` property entries in the recipe properties map? → A: Both coexist independently — existing `"Base spirit"` property labels remain untouched; authors may clean them up manually if desired but there is no automated migration or deprecation prompt.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Mark Base Spirit When Authoring a Recipe (Priority: P1)

A recipe author can designate one ingredient as the base spirit while creating or editing a recipe. The designation is a checkbox next to each ingredient; checking one automatically clears any previously checked ingredient. Unchecking the currently selected ingredient (or explicitly clearing it) leaves the recipe with no base spirit. The field is optional — no recipe is required to have one.

**Why this priority**: This is the data-entry foundation; without it, nothing can be stored or displayed. All other stories depend on it.

**Independent Test**: Create a recipe with three ingredients, mark the second as the base spirit, save, reopen — the second ingredient shows as base spirit. Then mark the third, save, reopen — only the third is marked. Then clear it, save, reopen — no ingredient is marked.

**Acceptance Scenarios**:

1. **Given** a recipe form with multiple ingredients, **When** the author checks the base spirit checkbox on ingredient B, **Then** ingredient B is marked and no other ingredient is marked.
2. **Given** ingredient B is already marked as base spirit, **When** the author checks the checkbox on ingredient C, **Then** ingredient C becomes the base spirit and ingredient B is automatically deselected.
3. **Given** ingredient B is marked as base spirit, **When** the author unchecks it, **Then** no ingredient is marked as base spirit.
4. **Given** a recipe with no base spirit designation, **When** the recipe is saved and reopened, **Then** no ingredient shows a base spirit marker.
5. **Given** a recipe with ingredient A marked as base spirit, **When** the recipe is saved, **Then** the designation persists correctly on subsequent loads.

---

### User Story 2 - Base Spirit Highlighted in Ingredient Hover Popover (Priority: P2)

When a viewer hovers over a recipe card to see the ingredient popover, the base spirit ingredient is visually distinguished from the other ingredients. Recipes without a base spirit show the ingredient list without any highlighting.

**Why this priority**: The popover is the primary quick-reference surface for ingredients (from feature 005); surfacing the base spirit there gives viewers instant context about the drink's character.

**Independent Test**: View the recipe list page, hover over a recipe that has a base spirit — the base spirit ingredient is visually distinct in the popover. Hover over a recipe without a base spirit — all ingredients appear uniform.

**Acceptance Scenarios**:

1. **Given** a recipe with a base spirit, **When** the viewer hovers over the recipe card, **Then** the base spirit ingredient is visually highlighted (e.g., bold, labelled, or accented) in the popover ingredient list.
2. **Given** a recipe with no base spirit, **When** the viewer hovers over the recipe card, **Then** all ingredients in the popover appear with equal visual weight.
3. **Given** a recipe where the base spirit was cleared by the author, **When** the viewer hovers over the card, **Then** no ingredient is highlighted.

---

### User Story 3 - Base Spirit Highlighted on Recipe Detail Page (Priority: P3)

On the full recipe detail page, the base spirit ingredient is visually distinguished in the ingredient list. The visual treatment should be consistent with the popover highlight so viewers recognise the same pattern.

**Why this priority**: The detail page is where viewers read the full recipe; the base spirit highlight reinforces the same information with more space available for a richer treatment.

**Independent Test**: Open the recipe detail page for a recipe with a base spirit — the base spirit ingredient is visually distinct. Open a recipe without one — all ingredients appear uniform.

**Acceptance Scenarios**:

1. **Given** a recipe with a base spirit, **When** the viewer opens the recipe detail page, **Then** the base spirit ingredient is visually highlighted in the ingredient list.
2. **Given** a recipe with no base spirit, **When** the viewer opens the recipe detail page, **Then** all ingredients appear with equal visual weight.
3. **Given** the highlight style used in the popover, **When** the same recipe is viewed on the detail page, **Then** the visual treatment is consistent (same pattern, same vocabulary).

---

### Edge Cases

- A recipe with only one ingredient: the author may or may not mark it as the base spirit; both states are valid.
- If an ingredient marked as base spirit is deleted from the recipe during editing, the base spirit designation is automatically removed — no ingredient becomes marked.
- Legacy recipes created before this feature was introduced have no base spirit data; they display as if no base spirit is set (all ingredients uniform).
- Recipes may also have a free-form `"Base spirit"` entry in their properties map (entered as plain text before this feature existed). This text property and the new structured `is_base_spirit` flag are independent; neither affects the other. Authors may choose to remove the old text entry manually, but the system does not prompt or enforce this.
- The base spirit checkbox is only accessible to logged-in users with edit rights (owner or admin); read-only viewers see only the display highlight, not the control.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The ingredient model MUST support an optional base spirit flag; at most one ingredient per recipe may have this flag set at any time.
- **FR-002**: The recipe authoring form MUST present a base spirit toggle (e.g., checkbox or radio-style control) for each ingredient row.
- **FR-003**: Selecting the base spirit toggle on any ingredient MUST automatically deselect it on all other ingredients in the same recipe (single-select behaviour).
- **FR-004**: The base spirit selection MUST be clearable so that a recipe has no base spirit designated.
- **FR-005**: The base spirit designation is optional; the system MUST accept and save recipes with no base spirit set.
- **FR-006**: The ingredient hover popover MUST visually distinguish the base spirit ingredient from non-base-spirit ingredients when one is designated.
- **FR-007**: The recipe detail page MUST visually distinguish the base spirit ingredient in the full ingredient list when one is designated.
- **FR-008**: The visual highlight for base spirit MUST be consistent between the popover and the detail page (same visual vocabulary).
- **FR-009**: If the ingredient currently designated as base spirit is removed during editing, the base spirit designation MUST be cleared automatically.
- **FR-010**: Legacy recipes with no base spirit data MUST display all ingredients with uniform visual weight (no spurious highlight).

### Key Entities

- **Ingredient**: Extended with an optional `is_base_spirit` boolean (default false/absent). Remains a sub-document of Recipe — no independent identity.
- **Recipe**: At most one ingredient in its ingredient list may have `is_base_spirit = true`. The recipe is valid with zero or one base spirit designations.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A recipe author can designate, change, and clear the base spirit within a single edit session without page reload, in under 5 seconds of interaction time.
- **SC-002**: 100% of recipes displayed in the card popover and detail page correctly reflect the stored base spirit state — highlighted when set, uniform when absent.
- **SC-003**: No existing recipe data is altered by the introduction of this feature; all legacy recipes continue to display and edit correctly.
- **SC-004**: The base spirit toggle is absent or disabled on recipe cards viewed by non-editors, ensuring read-only viewers cannot trigger the control.

## Assumptions

- The existing ingredient structure is a list of sub-documents within a recipe (not a separate entity with its own ID); the base spirit flag is therefore stored inline.
- "Highlight" means a visually distinguishable treatment (bold text, an accent colour, a label, or a small icon) — the exact style is left to design/implementation; the spec requires consistency between surfaces, not a specific colour.
- The hover popover from feature 005 already iterates the ingredients list; this feature adds a conditional highlight to that iteration.
- Only the recipe owner and admins see the base spirit toggle control in the edit form; the highlight on popover and detail page is visible to all viewers.
- No search or filter by base spirit is in scope for this feature.
- Some existing recipes carry a free-form `"Base spirit"` property (plain text). This is independent of the new structured flag — neither overrides nor replaces the other, and no automated migration is performed.
