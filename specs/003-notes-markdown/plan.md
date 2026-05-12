# Implementation Plan: Markdown Notes Editor

**Branch**: `003-notes-markdown` | **Date**: 2026-05-11 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/003-notes-markdown/spec.md`

## Summary

Adds markdown authoring and rendering support to the recipe notes field. No backend or data model changes are required — notes are already stored as raw text. The implementation is entirely frontend:

1. A new `renderMarkdown(text)` utility (marked + DOMPurify) shared by all read-only display sites.
2. A new `MarkdownEditor` component encapsulating the edit/preview toggle, used inside `RecipeForm`.
3. Updated `RecipeDetail` and `Home` pages to render notes as HTML instead of plain text.

New npm dependencies: `marked` (markdown parser) and `dompurify` (HTML sanitiser).

## Technical Context

**Language/Version**: Node.js 20+ / Vite 8 (frontend only — no backend changes)  
**Primary Dependencies**: `marked` v9+, `dompurify` v3+ (new); existing Vitest, jsdom, Tailwind unchanged  
**Storage**: N/A — no data model or API changes  
**Testing**: Vitest + jsdom (existing); TDD — failing tests written before each implementation unit  
**Target Platform**: Browser (same as existing frontend)  
**Performance Goals**: Preview toggle transition instantaneous (SC-001); markdown rendering adds < 5 ms to page render  
**Constraints**: Images stripped; links open in new tab with `rel="noopener noreferrer"`; DOMPurify as defence-in-depth XSS layer  
**Scale/Scope**: Frontend only; 4 new/modified source files, 4 new/modified test files

## Constitution Check

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Are functions single-responsibility and below complexity limits? | ✅ `renderMarkdown` is a single pure function; `MarkdownEditor` encapsulates toggle state; each stays well under 40 lines |
| II. Test-First | Are failing tests written before implementation begins? | ✅ Tasks require failing tests first for `renderMarkdown`, `MarkdownEditor`, and each updated page |
| III. UX Consistency | Do all UI surfaces follow the design language and handle loading/empty/error states? | ✅ Toggle uses existing button styling; empty notes remain hidden; rendering failure falls back to raw text |
| IV. Performance | Do API responses meet p95 ≤ 200ms / TTI ≤ 3s? | ✅ No API changes; marked + DOMPurify add < 5ms rendering cost; no TTI regression expected |
| Quality Gates | Do all CI checks (lint, coverage ≥ 80%, benchmarks) pass? | ✅ New utility and component must have ≥ 80% coverage; existing tests must not regress |

## Project Structure

### Documentation (this feature)

```text
specs/003-notes-markdown/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0: 5 decisions documented
├── contracts/
│   └── ui.md            # Phase 1: UI component contracts
└── tasks.md             # Phase 2 output (created by /speckit-tasks)
```

### Source Code — New and Modified Files

```text
frontend/
├── package.json                              # Add marked, dompurify dependencies
└── src/
    ├── utils/
    │   ├── markdown.js                       # NEW: renderMarkdown(text) → sanitised HTML string
    │   └── markdown.test.js                  # NEW: unit tests for renderMarkdown
    ├── components/
    │   ├── MarkdownEditor.js                 # NEW: edit/preview toggle component
    │   └── MarkdownEditor.test.js            # NEW: component tests
    └── pages/
        ├── RecipeForm.js                     # MODIFIED: replace notes textarea with MarkdownEditor
        ├── RecipeForm.test.js                # MODIFIED: update notes tests for MarkdownEditor
        ├── RecipeDetail.js                   # MODIFIED: replace notes p.textContent with renderMarkdown
        ├── RecipeDetail.test.js              # MODIFIED: update notes test for rendered output
        ├── Home.js                           # MODIFIED: replace notes p.textContent with renderMarkdown
        └── Home.test.js                      # MODIFIED: update notes test for rendered output
```

**No backend files are modified. No new pages or routes are added.**

## Complexity Tracking

No Constitution violations. No additional complexity justification required.

## Phase 0 Artifacts

- [research.md](research.md) — 5 decisions: markdown library (marked), sanitiser (DOMPurify), image/link strategy, component architecture, form submission preservation.

## Phase 1 Artifacts

- [contracts/ui.md](contracts/ui.md) — UI component contracts for `MarkdownEditor` and `renderMarkdown`.
