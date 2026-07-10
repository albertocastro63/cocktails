# Feature Specification: Related Cocktails

**Feature Branch**: `028-related-cocktails`  
**Created**: 2026-07-10  
**Status**: Draft  
**Input**: User description: "Related cocktails: A new feature that allows a user that enters or edits a cocktail recipe to assign a list of related cocktails, for example the Negroni is related to the Left Hand and the Right Hand. Once a cocktail is related to another the opposite relation is also recorded, for example if the Negroni is related to the Right Hand then the Right Hand is related to the Negroni. The relation is not transitive... The list of related cocktails is shown only on the cocktail display page, not in the random cocktail displayed in the home page. The list has links to access the related cocktails and it is displayed at the bottom of the page. The interface to add new related cocktails consists of a text field to search for the cocktail name only, as text is entered a list of cocktails that can be selected is displayed below, one can then use arrows to select the cocktail to be added. This can be repeated as many times as necessary."

## Clarifications

### Session 2026-07-10

- Q: Who may create/remove a relation to cocktail B (given symmetric writes)? → A: Any user permitted to edit cocktail A — the reverse is written to B with no ownership check on B.
- Q: How should the add-related search match cocktail names? → A: Case-insensitive substring match, anywhere in the name.
- Q: How should the related-cocktails list be ordered? → A: Alphabetical by cocktail name (order derived at display; storage is an unordered set).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Relate cocktails while creating or editing a recipe (Priority: P1)

Someone editing a cocktail recipe wants to link it to other cocktails it resembles or pairs with (e.g., relate the Negroni to the Left Hand and the Right Hand). In the recipe form they type into a "related cocktails" search field, see matching cocktails appear as they type, pick one with the keyboard, and repeat for as many as they want. When they save, the links are recorded — and each related cocktail automatically gains the reverse link back.

**Why this priority**: This is the core capability the feature adds — without the ability to assign relations there is nothing to display. It delivers standalone value: an editor can curate connections between cocktails.

**Independent Test**: In the recipe edit form, search for and add one or more related cocktails, save, then confirm (via the data/API) that the edited cocktail lists them and that each of them lists the edited cocktail in return.

**Acceptance Scenarios**:

1. **Given** a recipe being edited, **When** the user types part of another cocktail's name into the related-cocktails field, **Then** a list of cocktails whose names match is shown below the field.
2. **Given** the suggestion list is shown, **When** the user moves the selection with the arrow keys and confirms a choice, **Then** that cocktail is added to the recipe's related list and the field is ready for another entry.
3. **Given** the user has added one or more related cocktails and saves the recipe, **When** the save completes, **Then** the relations are persisted and each related cocktail now also lists this cocktail (symmetric).
4. **Given** the current cocktail or an already-added cocktail, **When** the user searches, **Then** those are not offered as suggestions (no self-relation, no duplicates).

---

### User Story 2 - Discover related cocktails on the recipe page (Priority: P2)

A visitor viewing a cocktail's detail page wants to explore cocktails connected to it. At the bottom of the page they see a "Related cocktails" list; each entry is a link that takes them to that cocktail's page. The randomly featured cocktail on the home page does not show this list.

**Why this priority**: This is the visible payoff of the relations — it turns recorded links into a browsing path for visitors. It depends on relations existing (seedable for testing) but delivers the discovery value.

**Independent Test**: Given a cocktail that already has related cocktails, open its detail page and confirm a "Related cocktails" section appears at the bottom with a working link per relation; open the home page and confirm the featured random cocktail shows no such list.

**Acceptance Scenarios**:

1. **Given** a cocktail with at least one related cocktail, **When** its detail page is viewed, **Then** a "Related cocktails" list appears at the bottom of the page.
2. **Given** the related list is shown, **When** the user clicks a related cocktail, **Then** they navigate to that cocktail's detail page.
3. **Given** a cocktail with no related cocktails, **When** its detail page is viewed, **Then** no related-cocktails section is shown.
4. **Given** the home page's randomly displayed cocktail, **When** it is shown, **Then** no related-cocktails list appears for it.

---

### User Story 3 - Keep relations correct and tidy (Priority: P3)

An editor manages the relations over time: removing a link they no longer want, relying on relations staying non-transitive, and trusting that links to a deleted cocktail disappear. Removing a relation from one cocktail also removes it from the other.

**Why this priority**: Integrity and maintenance harden the feature and prevent confusing or stale data, but the feature is usable for a first release without every maintenance path polished.

**Independent Test**: Remove a related cocktail while editing and confirm it is gone from both cocktails; set A–B and B–C and confirm A is not related to C; delete a cocktail and confirm it no longer appears in any other cocktail's related list.

**Acceptance Scenarios**:

1. **Given** two related cocktails A and B, **When** the editor removes B from A's related list and saves, **Then** A no longer lists B and B no longer lists A.
2. **Given** A is related to B and B is related to C, **When** viewing A, **Then** C is not listed as related to A (relations are not transitive).
3. **Given** a cocktail that appears in another cocktail's related list, **When** it is deleted, **Then** it no longer appears in that (or any) related list.

---

### Edge Cases

- **Self-relation**: A cocktail cannot be related to itself; it is never offered as a suggestion for its own list.
- **Duplicate relation**: Adding an already-related cocktail has no effect (no duplicate entry); already-related cocktails are not offered as suggestions.
- **Deleted counterpart**: If a related cocktail is deleted, it disappears from every other cocktail's related list (no broken links).
- **Empty state**: A cocktail with no relations shows no related-cocktails section on its detail page.
- **Editing a counterpart you don't own**: Because relations are symmetric, saving a relation on cocktail A updates cocktail B's related list even if the editor does not own B — this is intended.
- **No matches**: If the search text matches no cocktail names, the suggestion list is empty (and communicates that nothing was found).
- **Large lists**: A cocktail may have many related cocktails; the list remains readable and the add-search remains responsive.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A user creating or editing a recipe MUST be able to add one or more related cocktails to it.
- **FR-002**: Relations MUST be symmetric — relating cocktail A to cocktail B MUST also make B related to A, with no additional action.
- **FR-003**: Relations MUST NOT be transitive — if A is related to B and B is related to C, A MUST NOT be considered related to C unless a user sets that relation explicitly.
- **FR-004**: The add interface MUST provide a text field that searches existing cocktails by name; as the user types, cocktails whose names match MUST be shown as a selectable list below the field.
- **FR-005**: Search suggestions MUST match on cocktail name only (not ingredients, notes, or other fields), using a case-insensitive substring match (the typed text may appear anywhere in the name).
- **FR-006**: The user MUST be able to move the selection through the suggestion list with the arrow keys and confirm a selection to add that cocktail, and MUST be able to repeat this to add multiple related cocktails.
- **FR-007**: Suggestions MUST exclude the current cocktail and any already-related cocktails, preventing self-relations and duplicates.
- **FR-008**: The system MUST prevent duplicate relations and self-relations even if attempted through other means.
- **FR-009**: The recipe detail page MUST display the list of related cocktails at the bottom of the page whenever at least one related cocktail exists.
- **FR-010**: Each entry in the related list MUST display the related cocktail's name and MUST be a link that navigates to that cocktail's detail page.
- **FR-011**: The related-cocktails list MUST NOT be displayed for the randomly featured cocktail on the home page.
- **FR-012**: When a cocktail has no related cocktails, the detail page MUST NOT show a related-cocktails section.
- **FR-013**: A user editing a recipe MUST be able to remove a related cocktail; removal MUST also remove the reverse relation from the counterpart cocktail.
- **FR-014**: When a cocktail is deleted, it MUST be removed from the related list of every other cocktail.
- **FR-015**: Relations MUST be persisted when the recipe is saved and MUST survive subsequent viewing and editing.
- **FR-016**: Any user permitted to edit cocktail A MUST be able to set and remove A's relations to any cocktail B; because relations are symmetric, this MUST write the reverse relation onto B with **no ownership check on B** (any recipe editor may relate to any cocktail).
- **FR-017**: The related-cocktails list on the detail page MUST be displayed in alphabetical order by cocktail name; ordering is derived at display time and the stored relation set is unordered.

### Key Entities *(include if feature involves data)*

- **Cocktail Relation**: An undirected, symmetric association between two **distinct** cocktails, identified by the unordered pair of their cocktail identifiers. At most one relation exists per pair (no duplicates). Non-transitive: the existence of A–B and B–C implies nothing about A–C. A relation ceases to exist if either cocktail is deleted.
- **Cocktail (Recipe)** *(existing)*: Gains an associated set of related cocktails, surfaced on its detail page and editable through its create/edit form.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After a user relates cocktail A to cocktail B and saves, both A's and B's detail pages list the other — 100% of relations are symmetric.
- **SC-002**: Given A related to B and B related to C, A's related list never contains C (relations are non-transitive) — verified across representative cases.
- **SC-003**: On a recipe detail page with relations, the related-cocktails list appears at the bottom and every link navigates to the correct cocktail's page (100% of links resolve).
- **SC-004**: The home page's randomly featured cocktail never displays a related-cocktails list.
- **SC-005**: A user can add a related cocktail using only the keyboard (type, arrow to select, confirm) in under 10 seconds, with matching suggestions appearing as they type.
- **SC-006**: Removing a relation removes it from both cocktails, and deleting a cocktail removes it from every other cocktail's related list — no stale or one-sided relations remain.
- **SC-007**: The system never records a duplicate relation or a cocktail related to itself.

## Assumptions

- **Editing model**: Relations are managed as part of the recipe create/edit form and are persisted when the recipe is saved (not as a separate, independent action).
- **Symmetric writes across ownership** (resolved): Any recipe editor may relate cocktail A to any cocktail B; the reverse relation is written onto B with no ownership check on B (FR-016).
- **Suggestion matching** (resolved): Name search is a case-insensitive **substring** match anywhere in the name (FR-005); only existing cocktails can be related (a relation cannot be created to a name that has no cocktail).
- **Ordering** (resolved): The related-cocktails list is displayed alphabetically by cocktail name; the stored relation set is unordered (FR-017).
- **No hard cap**: There is no fixed maximum number of related cocktails per cocktail, within reason.
- **Accessibility**: The search-and-select control is a keyboard-operable, screen-reader-friendly combobox meeting WCAG 2.1 AA (per project standards).
- **Existing surfaces reused**: The related list links to the existing recipe detail pages; no new page types are introduced.

## Out of Scope

- Automatically suggesting relations (e.g., inferring related cocktails from shared ingredients or base spirit).
- Relation types or labels (e.g., "variation of", "same base spirit") — all relations are a single, undirected "related" link.
- Ranking or weighting related cocktails by relevance.
- Showing related cocktails anywhere other than the recipe detail page (e.g., in list/search results or the home random feature).
- Bulk import/export of relations.
