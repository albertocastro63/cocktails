# Implementation Plan: Visual Redesign

**Branch**: `006-visual-redesign` | **Date**: 2026-05-14 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/006-visual-redesign/spec.md`

## Summary

All user-facing pages are updated with a modern-minimal visual design: a warm stone-neutral base (`stone-50` backgrounds, `stone-900`/`stone-700` text), an amber accent color (`amber-500`) applied to buttons, card accents, section labels, and focus rings, and a dark stone navigation bar. A hero band is added to the home page. No routes, data models, or API contracts change — this is a pure visual update across 12 existing source files.

## Technical Context

**Language/Version**: JavaScript (ES Modules, no framework)  
**Primary Dependencies**: Tailwind CSS v3 (utility-first; `stone` and `amber` palettes already built in), Vite  
**Storage**: N/A  
**Testing**: Vitest with jsdom — tests assert presence of key CSS class tokens on rendered DOM elements  
**Target Platform**: Web browser, desktop + mobile (responsive)  
**Project Type**: Web application (frontend SPA)  
**Performance Goals**: No new assets added; TTI unaffected  
**Constraints**: No external fonts, no new npm packages, no changes to routing or backend  
**Scale/Scope**: 12 source files modified (components + pages + nav + CSS); no new files created

## Constitution Check

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Are functions single-responsibility and below complexity limits? | ✅ |
| II. Test-First | Are failing tests written before implementation begins? | ✅ |
| III. UX Consistency | Do all UI surfaces follow the design language and handle loading/empty/error states? | ✅ |
| IV. Performance | Do API responses meet p95 ≤ 200 ms read / ≤ 500 ms write and TTI ≤ 3 s? | ✅ |
| Quality Gates | Do all CI checks (lint, coverage ≥ 80%, benchmarks) pass? | ✅ |

**Notes**:
- Principle I: Changes are class-string replacements; no function logic changes. All functions stay under 40 lines.
- Principle II: Tests are updated first to assert new class tokens (they fail on old classes), then components are updated to make them pass.
- Principle III: EmptyState, loading text, and error messages are all restyled as part of FR-008.
- Principle IV: No new network requests; no impact on TTI.

## Design System

| Token | Value | Usage |
|---|---|---|
| Base background | `bg-stone-50` | All page wrappers, body |
| Card background | `bg-white` | Cards, modals, forms |
| Accent primary | `amber-500` | Buttons, card left border, CTA links |
| Accent hover | `amber-600` | Button hover, link hover |
| Accent light | `amber-100` / `amber-50` | Input focus bg, badge bg |
| Heading text | `text-stone-900` | All h1/h2 |
| Body text | `text-stone-700` | Paragraphs, list items |
| Muted text | `text-stone-500` | Captions, secondary labels |
| Section label | `text-amber-700 text-sm font-semibold uppercase tracking-widest` | Section headings (Ingredients, Steps…) |
| Nav background | `bg-stone-900` | Navigation bar |
| Nav text | `text-stone-100 hover:text-amber-400` | Nav links |
| Border | `border-stone-200` | Card and input borders |
| Focus ring | `focus:ring-amber-400` | All inputs |
| Border radius | `rounded-2xl` (cards, buttons, inputs) | Consistent 16px radius language |

## Project Structure

### Documentation (this feature)

```text
specs/006-visual-redesign/
├── plan.md         # This file
├── research.md     # Phase 0 output
├── quickstart.md   # Phase 1 output
└── tasks.md        # Phase 2 output (/speckit-tasks — not created here)
```

*No `data-model.md` — no data entities changed.*  
*No `contracts/` — no API or interface changes.*

### Source Code (repository root)

```text
frontend/
├── src/
│   ├── index.css                          ← add body bg-stone-50 base style
│   ├── main.js                            ← nav bar redesign (dark stone + amber)
│   ├── components/
│   │   ├── RecipeCard.js                  ← card redesign (rounded-2xl, amber left border)
│   │   ├── RecipeCard.test.js             ← update class assertions
│   │   ├── EmptyState.js                  ← updated empty state style
│   │   ├── EmptyState.test.js             ← update class assertions
│   │   ├── SearchBar.js                   ← amber focus ring, rounded-xl input
│   │   ├── SearchBar.test.js              ← update class assertions
│   │   ├── IngredientList.js              ← stone text colors, amber accent
│   │   └── IngredientList.test.js         ← update class assertions
│   └── pages/
│       ├── Home.js                        ← hero band, updated headings/typography
│       ├── RecipeList.js                  ← heading + layout style
│       ├── RecipeDetail.js                ← section labels, button styles
│       ├── Login.js                       ← form card, amber button + focus
│       └── RecipeForm.js                  ← input/button styles
└── (all other files unchanged)
```

**Structure Decision**: Web application layout. All changes are confined to existing files. No new files required.
