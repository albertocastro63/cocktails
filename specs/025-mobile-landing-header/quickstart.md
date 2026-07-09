# Quickstart: Compact Landing Header on Mobile

**Branch**: `025-mobile-landing-header` | **Date**: 2026-07-09

## Automated (Vitest — primary gate)

```bash
cd frontend
npm test -- src/pages/Home.test.js
npm test -- --coverage     # confirm ≥ 75% maintained
```

Covers the class contract (H1–H7): the hero wrapper, title, and subtitle carry the compact mobile classes plus the `md:` overrides equal to today's values, and the title/subtitle text and CTA are unchanged.

## Manual responsive verification

Run `npm run dev` and use browser device emulation.

1. **Phone (< 768px)** — the header banner is clearly shorter than before; the "All Recipes" CTA is visible with little or no scrolling; no horizontal overflow (V1, SC-001/SC-002).
2. **Desktop/tablet (≥ 768px)** — the header looks exactly as it does today: same padding, title size, subtitle size, spacing (V2, SC-003). Compare side-by-side with production if unsure.
3. **Resize/rotate across 768px** — the header switches between compact and full with no reload and the page is preserved (V3).
4. **Narrow / landscape phone** — the title and subtitle stay legible, on one screen width, with no clipping or awkward wrap (V4).
5. **Content** — the title still reads "Cocktail Recipes" and the subtitle "Discover your next favorite drink" (H4/H5).
