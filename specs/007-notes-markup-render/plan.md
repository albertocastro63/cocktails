# Implementation Plan: Notes Rendered Markup Styling

**Branch**: `007-notes-markup-render` | **Date**: 2026-05-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/007-notes-markup-render/spec.md`

## Summary

The notes field is currently rendered as markdown HTML in three locations (editor preview, recipe detail page, homepage), but the rendered HTML has no typography styling — headings, lists, bold, and blockquotes all appear as unstyled text. This feature adds consistent prose typography styling to all three surfaces using the Tailwind CSS Typography plugin with stone-palette colors, aligning with the existing design system without introducing amber/stone as prose foreground colors.

## Technical Context

**Language/Version**: JavaScript (ES2022 modules, vanilla DOM)
**Primary Dependencies**: Tailwind CSS v3, marked v18, DOMPurify v3, Vitest v4
**Storage**: N/A (display-only feature; no data model or persistence changes)
**Testing**: Vitest with jsdom, @testing-library/dom
**Target Platform**: Web browser (SPA with hash routing)
**Project Type**: Web application (frontend-only change)
**Performance Goals**: No regression to TTI ≤ 3 s; CSS plugin adds zero runtime cost
**Constraints**: No API changes; no data model changes; all existing tests must continue to pass
**Scale/Scope**: 3 source files changed, 3 test files updated, 1 package install, 1 config update

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Are functions single-responsibility and below complexity limits? | ✅ |
| II. Test-First | Are failing tests written before implementation begins? | ✅ |
| III. UX Consistency | Do all UI surfaces follow the design language and handle loading/empty/error states? | ✅ |
| IV. Performance | Do API responses meet p95 ≤ 200 ms read / ≤ 500 ms write and TTI ≤ 3 s? | ✅ |
| Quality Gates | Do all CI checks (lint, coverage ≥ 80%, benchmarks) pass? | ✅ |

All gates pass. No violations.

**Post-design re-check**: All principles remain satisfied — changes are confined to CSS class strings in 3 files and a dev-dependency install. No logic changes, no API changes, no new abstractions.

## Project Structure

### Documentation (this feature)

```text
specs/007-notes-markup-render/
├── plan.md              ← this file
├── research.md          ← Phase 0 output
├── data-model.md        ← Phase 1 output
└── tasks.md             ← Phase 2 output (/speckit-tasks)
```

### Source Code (files touched)

```text
frontend/
├── package.json                              ← add @tailwindcss/typography devDep
├── tailwind.config.js                        ← add typography plugin
└── src/
    ├── components/
    │   ├── MarkdownEditor.js                 ← update preview div class
    │   └── MarkdownEditor.test.js            ← add prose-class assertion test
    └── pages/
        ├── RecipeDetail.js                   ← update notes container class
        ├── RecipeDetail.test.js              ← add prose-class assertion test
        ├── Home.js                           ← update notes container class
        └── Home.test.js                      ← add prose-class assertion test
```

## Complexity Tracking

No violations — not applicable.
