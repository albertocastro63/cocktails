# Feature Specification: Markdown Notes Editor

**Feature Branch**: `003-notes-markdown`  
**Created**: 2026-05-11  
**Status**: Draft  
**Input**: User description: "Modify the notes text area to accept markdown and to parse it. During editing the markdown is input in plain text and a button is displayed on the top right of the notes area that makes the markdown to be rendered in the notes area. When the button is pressed again the plain text input is replaced. Whenever the notes field is displayed otherwise (non editing mode) the markdown is rendered."

## Clarifications

### Session 2026-05-11

- Q: Should the markdown renderer support embedded images and external links? → A: Render links as clickable anchors (opening in a new tab); block image tags entirely.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Markdown Preview While Editing (Priority: P1)

A recipe creator is writing or editing the notes field on a recipe. They type markdown syntax (headings, lists, bold/italic text) in the plain-text textarea. A "Preview" button in the top-right corner of the notes area lets them toggle between the raw markdown input and a rendered preview without leaving the form. Clicking "Preview" replaces the textarea with a rendered view of the markdown. Clicking "Edit" (the same button in toggled state) restores the textarea with the original markdown text intact.

**Why this priority**: This is the core authoring experience — without it, users have no way to verify their markdown looks correct before saving. It is also a prerequisite for confident use of markdown formatting.

**Independent Test**: Open the recipe create or edit form. Type markdown (e.g., `**bold** and a - list item`) in the notes field. Press the Preview button. Confirm rendered HTML (bold text, bullet) appears. Press Edit. Confirm the original markdown text is restored in the textarea unchanged.

**Acceptance Scenarios**:

1. **Given** the recipe form is open, **When** the user types markdown into the notes textarea, **Then** the raw markdown text is visible in the textarea.
2. **Given** the notes textarea contains markdown text, **When** the user presses the Preview button, **Then** the textarea is replaced with a rendered HTML view of the markdown.
3. **Given** the preview is active, **When** the user presses the Edit button (toggled label), **Then** the rendered preview is replaced by the textarea containing the original unmodified markdown text.
4. **Given** the user is in preview mode, **When** the form is submitted, **Then** the raw markdown text (not rendered HTML) is saved to the recipe.
5. **Given** the notes field is empty, **When** the user presses Preview, **Then** an empty preview area is shown without error.

---

### User Story 2 - Rendered Markdown in Read-Only Views (Priority: P2)

Whenever the notes field is displayed outside of the edit form — on the recipe detail page, on the homepage featured recipe — the stored markdown is rendered as formatted HTML rather than shown as raw text.

**Why this priority**: This delivers the end-user value of markdown: readers see nicely formatted notes (headings, lists, bold/italic) without seeing raw syntax characters.

**Independent Test**: Create a recipe with markdown notes (e.g., `## Tips\n- Shake well`). Navigate to the recipe detail page and to the homepage (if that recipe is featured). Confirm the notes display rendered HTML (a heading, a bullet list) and no raw markdown syntax is visible.

**Acceptance Scenarios**:

1. **Given** a recipe has notes containing markdown, **When** a user views the recipe detail page, **Then** the notes are rendered as formatted content (headings, lists, emphasis, etc.).
2. **Given** a recipe is featured on the homepage, **When** a user views the homepage, **Then** the notes section shows rendered markdown, not raw syntax.
3. **Given** a recipe has notes with no markdown syntax (plain text), **When** displayed in any read-only view, **Then** the text is shown without alteration.
4. **Given** a recipe has empty notes, **When** displayed in any read-only view, **Then** the notes section is hidden (existing behaviour preserved).

---

### Edge Cases

- What happens when notes contain potentially unsafe HTML (e.g., `<script>` tags embedded in markdown)? Rendered output must be sanitised — no executable scripts or injected HTML. Image tags are stripped entirely; links open in a new tab.
- What happens when the user switches to preview, then changes their mind and edits, and then saves — is the latest textarea content (not a stale copy) saved?
- What happens when notes contain markdown that produces very long rendered output (e.g., a large table)? The layout should not break.
- What happens when the user is in preview mode and navigates away without saving? No data loss beyond the normal unsaved-form behaviour.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The notes textarea in the recipe create and edit form MUST display a toggle button in the top-right corner of the notes area. The button MUST communicate its pressed state accessibly (WCAG 2.1 AA), reflecting whether the editor is currently in preview mode.
- **FR-002**: The toggle button MUST switch the notes area between "edit" mode (plain-text textarea) and "preview" mode (rendered markdown display).
- **FR-003**: In preview mode, the notes area MUST render the current textarea content as formatted markdown.
- **FR-004**: Switching from preview back to edit mode MUST restore the exact markdown text that was in the textarea before preview was activated.
- **FR-005**: When the form is submitted, the system MUST save the raw markdown text, regardless of whether the user is in edit or preview mode at the time of submission.
- **FR-006**: The recipe detail page MUST render the notes field as formatted markdown when notes are non-empty.
- **FR-007**: The homepage featured recipe display MUST render the notes field as formatted markdown when notes are non-empty.
- **FR-008**: Markdown rendering MUST sanitise output to prevent injection of executable scripts or unsafe HTML. Links MUST be rendered as clickable anchors opening in a new tab. Image tags MUST be stripped from rendered output.
- **FR-009**: The notes field storage format is unchanged — raw markdown text is stored; rendering is a display-only concern.
- **FR-010**: All existing notes behaviour (hidden when empty, plain-text fallback for notes without markdown syntax) MUST be preserved.

### Key Entities

- **Recipe Notes**: A plain-text field on a recipe containing markdown-formatted content. Stored as raw markdown; rendered on display. No change to data model or API contract.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can toggle between edit and preview mode in the notes area with a single button press, with the transition completing instantly (no perceptible delay).
- **SC-002**: Rendered markdown in read-only views accurately reflects the stored markdown — 100% of common formatting elements (headings, bold, italic, lists, links) are rendered correctly; image tags are stripped and links open in a new tab.
- **SC-003**: No raw markdown syntax characters (e.g., `**`, `##`, `-`) are visible in read-only views when the notes contain valid markdown.
- **SC-004**: Switching from preview back to edit mode preserves the original markdown text with 100% fidelity (no character loss or alteration).
- **SC-005**: Rendered output contains no executable scripts or injected HTML — unsafe content is stripped before display.

## Assumptions

- The notes field already exists on recipes (introduced in feature 002). This feature modifies its display behaviour only; no data model or API changes are required.
- The markdown subset to support is standard CommonMark (headings, paragraphs, bold, italic, lists, links, code spans, blockquotes). Extended flavours (tables, footnotes) are out of scope for this iteration.
- Markdown rendering and sanitisation are handled entirely on the client side; the server stores and returns raw markdown text unchanged.
- The recipe creator is the primary author of notes, but all users (including unauthenticated) can view rendered notes on the detail page and homepage.
- Mobile layout is in scope — the preview/edit toggle must be usable on small screens.
