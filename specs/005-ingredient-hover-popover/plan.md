# Implementation Plan: Ingredient Hover Popover

**Branch**: `005-ingredient-hover-popover` | **Date**: 2026-05-12 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/005-ingredient-hover-popover/spec.md`

## Summary

When a user hovers over a recipe tile on the recipe list page, a popover appears listing the recipe's ingredient names. If there are more than 5 ingredients, only the first 5 are shown followed by an ellipsis; the popover hides when the mouse leaves the tile. The implementation modifies `RecipeCard.js` to attach `mouseenter`/`mouseleave` event listeners and render an absolutely-positioned overlay with ingredient names from the data already present on the recipe object — no API changes required.

## Technical Context

**Language/Version**: JavaScript (ES Modules, no framework)  
**Primary Dependencies**: Tailwind CSS (utility classes), Vite (build/dev server)  
**Storage**: N/A — popover uses data already loaded with the recipe list  
**Testing**: Vitest with jsdom  
**Target Platform**: Web browser, desktop (mouse hover; touch devices unaffected)  
**Project Type**: Web application (frontend SPA)  
**Performance Goals**: Popover appears within one animation frame of mouseenter — no async work involved  
**Constraints**: No additional API calls; ingredient data must already be present in the recipe object  
**Scale/Scope**: Single component modification (`RecipeCard.js`) and its test file

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Are functions single-responsibility and below complexity limits? | ✅ |
| II. Test-First | Are failing tests written before implementation begins? | ✅ |
| III. UX Consistency | Do all UI surfaces follow the design language and handle loading/empty/error states? | ✅ |
| IV. Performance | Do API responses meet p95 ≤ 200 ms read / ≤ 500 ms write and TTI ≤ 3 s? | ✅ |
| Quality Gates | Do all CI checks (lint, coverage ≥ 80%, benchmarks) pass? | ✅ |

**Notes**:
- Principle I: Popover rendering extracted to a private helper function inside `RecipeCard.js`; all functions stay well under 40 lines.
- Principle II: Tests for all new behaviour written first (see tasks).
- Principle III: Empty-ingredients state handled; popover follows existing Tailwind design language.
- Principle IV: No new API calls; no impact on TTI or server response times.

## Project Structure

### Documentation (this feature)

```text
specs/005-ingredient-hover-popover/
├── plan.md         # This file
├── research.md     # Phase 0 output
├── quickstart.md   # Phase 1 output
└── tasks.md        # Phase 2 output (/speckit-tasks — not created here)
```

*No `data-model.md` — this feature introduces no new data entities; ingredient data is already part of the recipe object.*  
*No `contracts/` — this feature requires no API or interface changes.*

### Source Code (repository root)

```text
frontend/
├── src/
│   └── components/
│       ├── RecipeCard.js        ← modified (add popover helper + event listeners)
│       └── RecipeCard.test.js   ← extended (new popover tests)
└── (all other files unchanged)
```

**Structure Decision**: Web application layout. Only the `RecipeCard` component and its test file are touched. No new files are required.
