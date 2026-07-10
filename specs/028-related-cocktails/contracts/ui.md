# Contract: Related Cocktails UI

## Related Cocktails picker (recipe form)

A keyboard-operable combobox embedded in the recipe create/edit form.

- **Input**: text field, `role="combobox"`, labelled "Related cocktails", placeholder e.g. "Search cocktails by name…".
- **Suggestions**: as the user types, a `listbox` shows cocktails whose name contains the text (case-insensitive substring). Excludes the current recipe and already-selected cocktails. Empty text → no list (or nothing); no matches → an accessible "no matches" note.
- **Keyboard**: `ArrowDown`/`ArrowUp` move the highlighted option (`aria-activedescendant`); `Enter` adds the highlighted option; `Escape` closes the list. Fully operable without a mouse (SC-005).
- **Selected relations**: rendered as a list of removable **chips** (name + a keyboard-focusable "remove" control). Removing a chip drops that relation.
- **Persistence**: the current chip set is submitted as `related_ids` when the recipe form is saved (create or update). No separate save action.
- **States**: loading (names fetch), empty (no chips), error (names fetch failed → the picker degrades gracefully and does not block saving the rest of the form).

## Related Cocktails section (recipe detail page)

- Rendered **only** on the recipe detail page, at the **bottom**, after the recipe body.
- Heading: "Related cocktails".
- Shows each related cocktail as a link (`#/recipes/{id}`) labelled by name, in **alphabetical order**.
- **Hidden entirely** when the recipe has no relations (no empty heading).
- **Never rendered** on the home page's random cocktail (FR-011).

## Acceptance mapping

| UI behavior | Spec |
|-------------|------|
| Type-ahead by name, arrow-select, repeatable | FR-004, FR-005, FR-006, SC-005 |
| Excludes self + already-related from suggestions | FR-007 |
| Chips removable; removal drops relation | FR-013 |
| Detail section at bottom, alphabetical links | FR-009, FR-010, FR-017, SC-003 |
| Hidden when empty | FR-012 |
| Absent on home random | FR-011, SC-004 |
