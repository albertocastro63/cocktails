# Implementation Plan: Site Footer and Ingredient Popup Layout Fix

**Branch**: `017-footer-popup-fix` | **Date**: 2026-05-25 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/017-footer-popup-fix/spec.md`

## Summary

Add a site-wide footer (separator line + dynamic copyright notice) to every page, and fix the ingredient hover popup so it renders as a true overlay—appended to `document.body` with JS-calculated coordinates—eliminating the current layout shift that expands the page when the popup appears.

## Technical Context

**Language/Version**: JavaScript (ES modules, no framework)
**Primary Dependencies**: Vitest (testing), Tailwind CSS (styling), Vite (bundler)
**Storage**: N/A — purely presentational change, no backend involvement
**Testing**: Vitest with jsdom
**Target Platform**: Web browser (SPA, hash-based routing)
**Project Type**: Web application (frontend only for this feature)
**Performance Goals**: Zero layout shift (CLS = 0) on popup open; footer adds no perceptible render time
**Constraints**: Must follow amber/stone Tailwind design system; footer must be constrained to `max-w-4xl` content width matching existing pages; popup must not use `position: fixed` (spec requires it scrolls with content)
**Scale/Scope**: 10 pages in the SPA, one new component, one modified component, one modified entry point

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Are functions single-responsibility and below complexity limits? | ✅ |
| II. Test-First | Are failing tests written before implementation begins? | ✅ — Footer.test.js and RecipeCard popup tests written first |
| III. UX Consistency | Do all UI surfaces follow the design language and handle loading/empty/error states? | ✅ — Footer uses amber/stone tokens; popup is display-only |
| IV. Performance | Do API responses meet p95 ≤ 200 ms read / ≤ 500 ms write and TTI ≤ 3 s? | ✅ — No API changes; DOM ops are negligible |
| Quality Gates | Do all CI checks (lint, coverage ≥ 75%, benchmarks) pass? | ✅ — New tests maintain coverage |

## Project Structure

### Documentation (this feature)

```text
specs/017-footer-popup-fix/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```text
frontend/
├── src/
│   ├── components/
│   │   ├── Footer.js              # NEW — footer component
│   │   ├── Footer.test.js         # NEW — unit tests for Footer
│   │   ├── RecipeCard.js          # MODIFIED — popup portalled to body
│   │   └── RecipeCard.test.js     # MODIFIED — popup overlay assertions
│   └── main.js                    # MODIFIED — append Footer in renderPage()
```

**Structure Decision**: Single-project frontend. No backend changes. No new directories—Footer.js follows the existing `components/` convention.

## Complexity Tracking

> No constitution violations. No complexity justifications required.
