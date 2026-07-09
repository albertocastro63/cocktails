# Implementation Plan: Compact Landing Header on Mobile

**Branch**: `025-mobile-landing-header` | **Date**: 2026-07-09 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/025-mobile-landing-header/spec.md`

## Summary

The landing (home) page hero uses fixed sizes (`py-16`, `text-4xl` title, `text-lg` subtitle) that eat a large share of a phone screen. Make those sizes **responsive**: smaller defaults for phones, with `md:` (≥ 768px) restoring today's exact values so tablets/desktop are pixel-for-pixel unchanged. Pure CSS-class change in one file; no JS logic, no data.

## Technical Context

**Language/Version**: JavaScript (ES modules), Vite build, Tailwind CSS  
**Primary Dependencies**: Vanilla DOM, Tailwind (utility classes), Vitest + jsdom (tests). No new dependency.  
**Storage**: N/A  
**Testing**: Vitest (jsdom) — class-contract assertions on the hero element  
**Target Platform**: Web; phones (< 768px) vs tablets/desktop (≥ 768px)  
**Project Type**: Frontend web app (`frontend/`)  
**Performance Goals**: No change (responsive utility classes only; no runtime cost)  
**Constraints**: Desktop/tablet header MUST be unchanged (SC-003); text content unchanged; no horizontal overflow; consistent with the 768px (`md`) breakpoint from feature 024  
**Scale/Scope**: One element block in `frontend/src/pages/Home.js`; one test file extended

### Current vs. planned classes (Home hero)

| Element | Today | Planned (mobile default → `md:` restores desktop) |
|---------|-------|---------------------------------------------------|
| hero wrapper | `... py-16 px-4` | `... py-4 md:py-16 px-4` |
| title `h1` | `text-4xl font-bold mb-3` | `text-xl md:text-4xl font-bold mb-0.5 md:mb-3` |
| subtitle `p` | `text-amber-400 text-lg mb-6` | `text-amber-400 text-sm md:text-lg mb-3 md:mb-6` |
| CTA `a` | (unchanged) | (unchanged) |

At `md` and above every value equals today's, so desktop/tablet render identically (SC-003). On phones the banner's vertical footprint drops by roughly 50% (quarter padding vs. today's `py-16`, smaller title/subtitle, tighter margins), meeting SC-001 and surfacing the CTA sooner (SC-002). Exact values are tunable while staying legible.

## Constitution Check

| Principle | Gate Question | Status |
|-----------|--------------|--------|
| I. Code Quality | Single-responsibility, within limits, no duplication? | ✅ One localized class edit in `Home()`; no logic added |
| II. Test-First | Failing tests written before implementation? | ✅ Extend `Home.test.js` first to assert the responsive header class contract (mobile-compact + `md:` desktop-restore) and unchanged title/subtitle text |
| III. UX Consistency | Design tokens + legibility + WCAG 2.1 AA? | ✅ Same stone/amber tokens and text; smaller but legible sizes; contrast unchanged (color pairs unchanged) |
| IV. Performance | Budgets met? | ✅ No runtime change; utility classes only |
| Quality Gates | Lint, coverage ≥ 75%, tests pass? | ✅ Test added; coverage maintained |

No violations. §II is satisfied by a class-contract test (the change is presentational; the testable contract is the responsive class set + unchanged text), consistent with feature 024's approach.

## Project Structure

### Documentation (this feature)

```text
specs/025-mobile-landing-header/
├── plan.md
├── research.md
├── data-model.md            # N/A entities — records the presentational contract
├── quickstart.md
├── contracts/
│   └── header-ui-contract.md
└── tasks.md                 # /speckit-tasks output (not created here)
```

### Source Code (repository root)

```text
frontend/src/pages/
├── Home.js         # EDIT — responsive classes on the hero wrapper, title, subtitle
└── Home.test.js    # EDIT — assert the responsive header class contract + unchanged text
```

**Structure Decision**: Frontend-only, single-file change. Desktop parity is guaranteed by making the `md:` variants equal to today's fixed values; only the mobile defaults shrink.

## Architecture

Responsive utility classes, no JavaScript branch:
- The hero wrapper, title, and subtitle carry mobile-first sizes with `md:` overrides equal to today's values.
- The phone/large-screen switch is a pure CSS media query (`md` = 768px) — resize/rotate updates it with no reload (FR-006), and preserves the current page.

## Phase 0: Research

See [research.md](research.md) — responsive-size approach and the desktop-parity guarantee. No `NEEDS CLARIFICATION` remain.

## Phase 1: Design

- [data-model.md](data-model.md) — no data entities; documents the presentational states.
- [contracts/header-ui-contract.md](contracts/header-ui-contract.md) — the class/behavior contract (acceptance assertions).
- [quickstart.md](quickstart.md) — Vitest + manual responsive verification.

## Complexity Tracking

No constitution violations; no entries required.
