# Tasks: Markdown Notes Editor

**Input**: Design documents from `specs/003-notes-markdown/`  
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, contracts/ui.md ✅

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.  
**TDD**: Constitution Principle II requires tests to be written and confirmed failing before implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1, US2)

---

## Phase 1: Setup

**Purpose**: Add the two new npm dependencies required by both user stories.

- [X] T001 Add `marked` and `dompurify` to `dependencies` in `frontend/package.json` and run `npm install` to lock them

---

## Phase 2: Foundational (Blocking Prerequisite)

**Purpose**: The `renderMarkdown(text)` utility is the shared rendering pipeline used by both user stories. It must exist and be tested before any display or editor code is written.

**⚠️ CRITICAL**: Both user stories depend on this utility. No US1 or US2 implementation can proceed until T003 is complete.

- [X] T002 Add failing unit tests for `renderMarkdown` in `frontend/src/utils/markdown.test.js`: "returns empty string for empty input", "renders plain text as paragraph", "renders **bold** as `<strong>`", "renders `[text](url)` link with `target=_blank` and `rel=noopener noreferrer`", "strips `![alt](url)` image — no `<img>` in output", "strips `<script>alert(1)</script>` from output"
- [X] T003 Implement `renderMarkdown(text)` in `frontend/src/utils/markdown.js`: use `marked` with a custom renderer that returns `''` for image tokens and wraps links with `target="_blank" rel="noopener noreferrer"`; pipe the output through `DOMPurify.sanitize` with `{ FORBID_TAGS: ['img'] }` as defence-in-depth; export as a named export

**Checkpoint**: `renderMarkdown` tests pass. Both user story phases can now proceed.

---

## Phase 3: User Story 1 — Markdown Preview Toggle in Edit Form (Priority: P1) 🎯 MVP

**Goal**: The notes area in the recipe create/edit form has a "Preview" toggle button that switches between the plain-text textarea and a rendered markdown preview. Switching back restores the original text exactly. The form always submits raw markdown.

**Independent Test**: Open the recipe create form. Type `**bold**` in the notes textarea. Press "Preview". Confirm `<strong>bold</strong>` is rendered. Press "Edit". Confirm `**bold**` is restored unchanged in the textarea. Submit the form and confirm the saved value is `**bold**` (raw markdown, not HTML).

### Tests for User Story 1 — Write First, Confirm Failing

> **Write these tests FIRST. Run the test suite and confirm each new test fails before writing implementation code.**

- [X] T004 [P] [US1] Add failing tests for `MarkdownEditor` in `frontend/src/components/MarkdownEditor.test.js`: "renders a textarea with the given name attribute", "renders a Preview toggle button", "clicking Preview hides the textarea and shows rendered markdown", "clicking Edit after Preview restores the original markdown text in the textarea", "textarea value is accessible via its name attribute while in preview mode (form can read it)", "toggle button has aria-pressed='false' in edit mode and aria-pressed='true' in preview mode"
- [X] T005 [P] [US1] Update the notes tests in `frontend/src/pages/RecipeForm.test.js`: (1) update the existing "renders a textarea for notes" test to also assert a preview toggle button is present alongside the `textarea[name="notes"]`; (2) add a new failing test "submitting the form while notes editor is in preview mode sends raw markdown in the payload" — mock `createRecipe`, click Preview to enter preview mode, then submit the form, and assert `createRecipe` was called with the raw markdown string (not rendered HTML)

### Implementation for User Story 1

- [X] T006 [US1] Implement `MarkdownEditor` in `frontend/src/components/MarkdownEditor.js`: create a root `div.markdown-editor`; add a toolbar `div` containing a `button[type="button"]` labelled "Preview" with initial `aria-pressed="false"`; add `textarea[name][placeholder]` initialised with `value` prop (visible by default); add `div.markdown-preview` (hidden by default); on button click toggle state — in preview mode set `aria-pressed="true"`, hide textarea, show div with `innerHTML = renderMarkdown(textarea.value)`, change button label to "Edit"; on edit click set `aria-pressed="false"`, restore textarea visibility, clear preview div, change label back to "Preview"; accept `{ name, placeholder, value }` props; keep textarea present (hidden via `style.display`) in preview mode so form serialisation still works
- [X] T007 [US1] Update `frontend/src/pages/RecipeForm.js` to use `MarkdownEditor`: import `MarkdownEditor`; replace the `notesWrap`/`notesLbl`/`notesArea` block with a `let editorEl` variable set to `MarkdownEditor({ name: 'notes', placeholder: 'Personal notes, substitutions, tips…', value: '' })`; in the edit-mode prefill (inside the `getRecipe` `.then()` callback) replace the old `notesArea.value` assignment with a new `MarkdownEditor` call passing `value: recipe.notes || ''` and replacing `editorEl` in the DOM; ensure the form submit path continues to read `form.querySelector('[name="notes"]').value` unchanged

**Checkpoint**: All T004–T007 tests pass. Create/edit the notes field in the form, toggle preview and back, and submit. Saved value must be raw markdown.

---

## Phase 4: User Story 2 — Rendered Markdown in Read-Only Views (Priority: P2)

**Goal**: Notes displayed on the recipe detail page and the homepage featured recipe are rendered as formatted HTML rather than raw markdown syntax characters.

**Independent Test**: Create a recipe with notes `## Tips\n- Shake well`. Open the recipe detail page. Confirm an `<h2>` element containing "Tips" and an `<li>` containing "Shake well" appear in the notes section. Open the homepage (with that recipe as the featured recipe). Confirm the same rendered output is present.

### Tests for User Story 2 — Write First, Confirm Failing

> **Write these tests FIRST. Run the test suite and confirm each new test fails before writing implementation code.**

- [X] T008 [P] [US2] Update the notes test in `frontend/src/pages/RecipeDetail.test.js`: change "renders notes when recipe has notes" to pass notes containing `**bold text**` and assert `document.body.querySelector('strong')` is not null (rendered HTML, not raw textContent match)
- [X] T009 [P] [US2] Update the notes test in `frontend/src/pages/Home.test.js`: change "renders notes when recipe has notes" to pass notes containing `**bold text**` and assert `document.body.querySelector('strong')` is not null

### Implementation for User Story 2

- [X] T010 [P] [US2] Update `frontend/src/pages/RecipeDetail.js` to render notes via `renderMarkdown`: import `renderMarkdown` from `../utils/markdown.js`; in the notes section replace the `<p>` element and `p.textContent = recipe.notes` with a `<div>` element and `div.innerHTML = renderMarkdown(recipe.notes)`; remove the `whitespace-pre-wrap` class (no longer needed); keep the `if (recipe.notes)` guard unchanged
- [X] T011 [P] [US2] Update `frontend/src/pages/Home.js` to render notes via `renderMarkdown`: import `renderMarkdown` from `../utils/markdown.js`; in the notes section replace the `<p>` element and `p.textContent = recipe.notes` with a `<div>` element and `div.innerHTML = renderMarkdown(recipe.notes)`; remove the `whitespace-pre-wrap` class; keep the `if (recipe.notes)` guard unchanged

**Checkpoint**: All T008–T011 tests pass. Notes on the detail page and homepage display formatted HTML.

---

## Phase 5: Polish & Cross-Cutting Concerns

- [X] T012 [P] Run `cd frontend && npm test -- --run` and confirm all tests pass with zero failures and no regressions in existing tests
- [X] T013 [P] Run `cd frontend && npm run build` and confirm zero build errors (verifies marked and dompurify are bundled correctly and tree-shaking produces no import warnings)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — install packages first
- **Phase 2 (Foundational)**: Depends on Phase 1 (packages must be installed before import) — BLOCKS both user stories
- **Phase 3 (US1)**: Depends on Phase 2 (T003) completion
- **Phase 4 (US2)**: Depends on Phase 2 (T003) completion; T008–T011 can run in parallel with Phase 3 tasks (different files)
- **Phase 5 (Polish)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Starts after T003. T004 and T005 (tests) can run in parallel. T006 and T007 (impl) must run sequentially: T006 before T007.
- **US2 (P2)**: Starts after T003. T008 and T009 (tests) can run in parallel. T010 and T011 (impl) can run in parallel — different files.

### Within Each User Story

- All test tasks MUST be written and confirmed failing before the corresponding implementation tasks begin

---

## Parallel Example: User Story 1

```
# Test tasks (write first, confirm failing):
T004: MarkdownEditor.test.js
T005: RecipeForm.test.js (update)

# Implementation (after tests fail):
T006: MarkdownEditor.js       ← must complete before T007
T007: RecipeForm.js           ← depends on T006
```

## Parallel Example: User Story 2

```
# Test tasks (write first, confirm failing):
T008: RecipeDetail.test.js (update)    ← parallel
T009: Home.test.js (update)            ← parallel

# Implementation (after tests fail):
T010: RecipeDetail.js                  ← parallel
T011: Home.js                          ← parallel
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Install packages
2. Complete Phase 2: Implement `renderMarkdown` utility (T002 → T003)
3. Complete Phase 3: Implement `MarkdownEditor` and wire into `RecipeForm` (T004–T007)
4. **STOP and VALIDATE**: Toggle works, form saves raw markdown, all tests pass
5. Ship US1 if ready

### Incremental Delivery

1. Phase 1 + 2 → `renderMarkdown` utility ready
2. Phase 3 (US1) → Recipe form has markdown preview → Demo/ship
3. Phase 4 (US2) → Read-only views render markdown → Demo/ship
4. Phase 5 → Full suite green, build clean

---

## Notes

- No backend changes required — all work is in `frontend/src/`
- The existing form submission path (`form.querySelector('[name="notes"]').value`) requires no changes if the textarea is kept in the DOM during preview mode (hidden via `style.display = 'none'`)
- DOMPurify must run in a browser environment (jsdom in tests satisfies this)
- `marked` custom renderer: override `image()` to return `''`; override `link()` to inject `target="_blank" rel="noopener noreferrer"`
