# Header UI Contract: Compact Landing Header on Mobile

**Branch**: `025-mobile-landing-header` | **Date**: 2026-07-09

The observable contract is the hero element's responsive classes and content. Rows H* are jsdom-checkable (class presence); rows V* are viewport-dependent and confirmed manually per `quickstart.md`.

## Class contract (jsdom-testable)

| # | Element | MUST include | Meaning | Maps to |
|---|---------|--------------|---------|---------|
| H1 | hero wrapper | `py-4` and `md:py-16` | compact padding on phones, today's padding at md+ | FR-001, FR-003 |
| H2 | title `h1` | `text-xl` and `md:text-4xl` | smaller title on phones, today's at md+ | FR-001, FR-003 |
| H3 | subtitle `p` | `text-sm` and `md:text-lg` | smaller subtitle on phones, today's at md+ | FR-001, FR-003 |
| H4 | title `h1` | text = "Cocktail Recipes" | content unchanged | FR-002 |
| H5 | subtitle `p` | text = "Discover your next favorite drink" | content unchanged | FR-002 |
| H6 | CTA `a` | href `#/recipes`, text "All Recipes" | CTA target/label unchanged | FR-005 |
| H7 | title/subtitle | keep `text-white` / `text-amber-400` tokens | visual style unchanged | FR-005 |
| H8 | hero inner | `flex` and `md:block` | phone: text + CTA on one row; md+: stacked (today) | FR-009 |
| H9 | CTA `a` | `px-4 py-1.5 text-sm` and `md:px-6 md:py-2 md:text-base` | smaller CTA on phones, today's size at md+ | FR-010 |
| H10 | mobile brand header (nav) | on home route → `hidden` (no `flex`); on other routes → `flex md:hidden` | landing hides the "Cocktails" brand on phones | FR-008 |

## Viewport contract (manual)

| # | Viewport | Expected | Maps to |
|---|----------|----------|---------|
| V1 | < 768px | hero banner noticeably shorter than today (~35–40% less height); no horizontal overflow; CTA reachable with little/no scroll | SC-001, SC-002, SC-004 |
| V2 | ≥ 768px | header pixel-for-pixel identical to current release | SC-003 |
| V3 | resize/rotate across 768px | switches compact⇄full with no reload, page preserved | FR-006 |
| V4 | narrow/short (landscape) phone | legible, no clip/awkward wrap/overflow | FR-004 |
