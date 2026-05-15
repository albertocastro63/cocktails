# Feature Specification: Notes Rendered Markup Styling

**Feature Branch**: `007-notes-markup-render`  
**Created**: 2026-05-14  
**Status**: Draft  
**Input**: User description: "In the notes field when previewing render the HTML in a way that shows the different elements added with markup. Render the markup of the notes in a similar way in all the pages where it appears."

## Clarifications

### Session 2026-05-14

- Q: Should rendered markdown elements use the app's amber/stone accent colors or standard typography colors? → A: Standard typography colors (dark gray/near-black headings, neutral accents) that complement but do not echo the amber/stone palette.
- Q: Should the homepage show full notes content or truncate long notes? → A: Show full notes on the homepage, same as the detail page — no truncation.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Visually Styled Preview While Editing (Priority: P1)

A recipe author is composing notes using markdown (headings, bold text, bullet lists, etc.) in the recipe create or edit form. They press the "Preview" button to check how the notes will look. The preview renders the markdown with visible visual differentiation: headings appear larger, bold text is heavier, lists display with bullets or numbers, blockquotes are indented, and code is rendered in a distinct style.

**Why this priority**: The editor preview exists specifically to let authors verify their formatting before saving. If the preview shows plain, unstyled text, it fails its core purpose — the author cannot judge whether their formatting choices worked.

**Independent Test**: Open the recipe create or edit form. Type markdown such as `## Heading\n**bold** and a - list item\n> a blockquote` in the notes textarea. Press "Preview". Confirm that the heading is visually larger than body text, bold text is heavier, the list item has a bullet, and the blockquote is visually distinguished from body text.

**Acceptance Scenarios**:

1. **Given** the notes textarea contains a markdown heading (`# Title`), **When** the user presses "Preview", **Then** the heading renders visibly larger than surrounding text.
2. **Given** the notes textarea contains `**bold**` and `*italic*`, **When** the user presses "Preview", **Then** bold text appears heavier and italic text appears slanted.
3. **Given** the notes textarea contains a bulleted list (`- item`), **When** the user presses "Preview", **Then** list items display with visible bullets.
4. **Given** the notes textarea contains an ordered list (`1. item`), **When** the user presses "Preview", **Then** list items display with visible numbering.
5. **Given** the notes textarea contains a blockquote (`> note`), **When** the user presses "Preview", **Then** the blockquote is visually distinguished (e.g., indented or bordered).
6. **Given** the notes textarea contains inline code (`` `code` ``), **When** the user presses "Preview", **Then** the code is rendered in a distinct monospace or highlighted style.
7. **Given** the notes textarea contains only plain text (no markdown syntax), **When** the user presses "Preview", **Then** the text renders as normal body text without visual noise.

---

### User Story 2 - Consistently Styled Notes on Recipe Detail Page (Priority: P2)

A user viewing a recipe's detail page sees the notes section rendered with the same visual formatting as the editor preview — headings, bold text, lists, and other markdown elements are visually distinct and consistent in appearance with the preview.

**Why this priority**: If the author invested effort formatting their notes with markdown, readers should see the intended visual result. Unstyled output undermines the authoring experience already delivered.

**Independent Test**: Create a recipe with notes containing `## Tips\n**Shake well.** Use fresh ingredients:\n- Ice\n- Lime`. Navigate to the recipe detail page. Confirm the heading is visually larger, bold is heavier, and the list has bullets — matching what was shown in the editor preview.

**Acceptance Scenarios**:

1. **Given** a recipe has markdown notes with a heading, **When** a user views the recipe detail page, **Then** the heading is visually larger than body text.
2. **Given** a recipe has markdown notes with bold and italic text, **When** a user views the recipe detail page, **Then** bold and italic are rendered with the correct visual weight.
3. **Given** a recipe has markdown notes with a list, **When** a user views the recipe detail page, **Then** the list renders with bullets or numbers.
4. **Given** a recipe has notes with no markdown syntax, **When** a user views the recipe detail page, **Then** the text renders as plain body text.
5. **Given** a recipe has empty notes, **When** a user views the recipe detail page, **Then** no notes section is shown (existing behaviour preserved).

---

### User Story 3 - Consistently Styled Notes on Homepage (Priority: P3)

A user viewing the homepage sees the featured recipe's notes section rendered with the same visual formatting as the recipe detail page and the editor preview.

**Why this priority**: Visual consistency across all views reinforces trust and polish. The homepage should present formatted notes identically to how they appear on the detail page.

**Independent Test**: Ensure a recipe with markdown notes is the featured recipe on the homepage. Navigate to the homepage and confirm the notes render with the same visual formatting as observed on the detail page.

**Acceptance Scenarios**:

1. **Given** a recipe with markdown notes is featured on the homepage, **When** a user views the homepage, **Then** the notes are rendered with the same visual styling as on the recipe detail page.
2. **Given** the featured recipe has empty notes, **When** a user views the homepage, **Then** no notes section is shown.

---

### Edge Cases

- What happens when notes contain very deeply nested markdown (e.g., heading inside a blockquote)? Rendered output should display correctly without layout breakage.
- What happens when notes contain a long unbroken word or URL? Text should wrap without overflowing its container.
- What happens when notes contain a markdown table? The table should render with visible rows/columns and not break surrounding layout.
- What happens when notes contain sanitised-out elements (e.g., `<script>`, `<img>`)? Those elements must remain stripped; no error is shown to the user.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The editor notes preview area MUST render each distinct markdown element (headings, bold, italic, unordered lists, ordered lists, blockquotes, inline code, horizontal rules) with visually distinct styling that differentiates it from plain body text. Styling MUST use standard typography colors (dark gray/near-black headings, neutral accents) that complement the app's amber/stone design system without using amber or stone as foreground colors on markdown elements.
- **FR-002**: The recipe detail page notes section MUST render markdown elements with the same visual styling as the editor preview.
- **FR-003**: The homepage featured recipe notes section MUST render markdown elements with the same visual styling as the recipe detail page. Notes MUST be displayed in full on the homepage regardless of length — no truncation is applied.
- **FR-004**: Rendered notes in all locations MUST maintain existing sanitisation rules: no executable scripts, no image tags, links open in a new tab.
- **FR-005**: The visual styling applied to rendered notes MUST be consistent across the editor preview, the recipe detail page, and the homepage — a user should not perceive a difference in formatting between these surfaces.
- **FR-006**: Rendered notes MUST NOT overflow their container width regardless of content length (long words, URLs, or tables).
- **FR-007**: All existing notes behaviours MUST be preserved: notes are hidden when empty; plain-text notes without markdown syntax display as normal body text; the raw markdown text is always what is stored.

### Key Entities

- **Rendered Notes**: The visual representation of a recipe's markdown notes field. Display-only — no change to the stored data format or data model.
- **Markup Elements**: Markdown constructs that produce distinct HTML elements: headings (h1–h3), emphasis (strong, em), lists (ul, ol, li), blockquotes, inline code, code blocks, horizontal rules, and anchors.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of supported markdown element types (headings, bold, italic, unordered lists, ordered lists, blockquotes, inline code) are visually distinguishable from plain body text in the editor preview.
- **SC-002**: Visual appearance of rendered notes on the recipe detail page is indistinguishable from the editor preview for all supported markdown element types.
- **SC-003**: Visual appearance of rendered notes on the homepage is indistinguishable from the recipe detail page for all supported markdown element types.
- **SC-004**: No layout overflow is introduced by any markdown content in any rendered notes location.
- **SC-005**: All existing sanitisation rules are preserved — verified by confirming `<script>` and `<img>` tags are absent from rendered output across all surfaces.

## Assumptions

- The notes field in the recipe create/edit form uses a preview toggle (implemented in feature 003); this feature does not change the toggle mechanism.
- Notes appear in three locations: the editor preview, the recipe detail page, and the homepage. No other pages currently display notes.
- Styling is a display-only concern; the stored notes data format (raw markdown text) is unchanged.
- The visual design of rendered notes complements the amber/stone design system using standard typography colors (dark gray/near-black headings, neutral accents); amber and stone are not used as foreground colors on markdown elements.
- Markdown support covers the element types already handled by the existing markdown renderer: headings, emphasis, lists, blockquotes, inline code, horizontal rules, and anchors. No new markdown features are added by this feature.
