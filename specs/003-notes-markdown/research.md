# Research: Markdown Notes Editor

**Branch**: `003-notes-markdown` | **Date**: 2026-05-11

---

## Decision 1: Markdown Parsing Library

**Decision**: Use **marked** (v9+) as the markdown-to-HTML parser.

**Rationale**: The project uses Vite + npm (ES modules). `marked` is the most widely adopted browser-side CommonMark parser: ~30 kB minified+gzipped, zero transitive dependencies, tree-shakeable, actively maintained, and has a well-documented custom renderer API needed for image suppression and link target injection.

**Alternatives considered**:
- `markdown-it` — equally capable, slightly larger (~43 kB), more extensible but more setup overhead for this use case.
- `micromark` — lower-level, requires assembling a pipeline; overkill for a notes field.
- `showdown` — older, less maintained, not CommonMark spec-compliant.

---

## Decision 2: HTML Sanitisation Library

**Decision**: Use **DOMPurify** (v3+) for sanitising marked's HTML output before DOM injection.

**Rationale**: DOMPurify is the industry-standard client-side XSS sanitiser. It maintains an allow-list of safe HTML tags/attributes and actively strips any dangerous constructs. It runs natively in the browser without a DOM emulation layer and is trivially configured to forbid specific tags (e.g., `img`) and enforce attribute rules on `a` tags.

**Alternatives considered**:
- Manual allow-list filtering — error-prone, unmaintainable, not security-audited.
- `sanitize-html` — Node.js-oriented, requires a heavier bundling setup; no meaningful advantage over DOMPurify in the browser.

---

## Decision 3: Image Suppression and Link Target Strategy

**Decision**: Configure `marked` with a custom renderer that (a) returns an empty string for image tokens and (b) appends `target="_blank" rel="noopener noreferrer"` to all anchor tags. Apply DOMPurify as a final pass with `FORBID_TAGS: ['img']` as defence-in-depth.

**Rationale**: Blocking images at the parser level prevents the `![alt](url)` markdown syntax from emitting any HTML at all — cleaner than stripping after the fact. Adding `target="_blank" rel="noopener noreferrer"` to links at render time avoids unexpected same-tab navigation; `rel="noopener noreferrer"` is required to prevent reverse tabnapping. DOMPurify as a second line of defence handles any edge cases the custom renderer misses (e.g., raw HTML passthrough).

**Alternatives considered**:
- CSS `pointer-events: none` on images — does not prevent hotlink requests from firing.
- Stripping images only in DOMPurify — possible but requires extra configuration and leaves raw image markup in the marked output momentarily.

---

## Decision 4: Component Architecture

**Decision**: Extract a `MarkdownEditor` component (`frontend/src/components/MarkdownEditor.js`) for the form edit/preview toggle, and a `renderMarkdown(text)` utility function (`frontend/src/utils/markdown.js`) shared by all read-only display sites.

**Rationale**: The toggle state (edit vs. preview) is self-contained UI logic that does not belong in RecipeForm. Extracting it as a component makes it independently testable and reusable if other markdown fields are added later. A shared `renderMarkdown` utility avoids duplicating the marked + DOMPurify pipeline across RecipeDetail, Home, and any future read-only consumer.

**Alternatives considered**:
- Inline the toggle logic directly in `RecipeForm.js` — would exceed the 40-line function limit (Constitution Principle I) and is not independently testable.
- A single unified `MarkdownField` component handling both edit and display modes — more complex, harder to test the read-only path in isolation.

---

## Decision 5: Form Submission — Raw Markdown Preservation

**Decision**: The `MarkdownEditor` component always maintains the raw markdown string internally. On form submit, the component exposes the value via a standard named `<textarea name="notes">` that is kept in sync even while in preview mode (hidden but present in the DOM), so existing form serialisation code in `RecipeForm.js` requires no changes.

**Rationale**: The current `RecipeForm.js` reads values with `form.querySelector('[name="notes"]').value`. Keeping the textarea present (but hidden) during preview means zero changes to the form submission path and zero risk of accidentally saving rendered HTML.

**Alternatives considered**:
- Exposing a `getValue()` method on the component — requires RecipeForm to know about the component's API, increasing coupling.
- Using a separate hidden input to shadow the textarea — adds unnecessary DOM complexity.
