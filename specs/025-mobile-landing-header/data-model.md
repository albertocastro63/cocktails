# Data Model: Compact Landing Header on Mobile

**Branch**: `025-mobile-landing-header` | **Date**: 2026-07-09

No data entities — this is a presentational change. The only "model" is the header's two presentational states, driven purely by viewport width:

| State | Viewport | Header presentation |
|-------|----------|---------------------|
| Compact | width < 768px (phones) | Reduced hero padding, smaller title/subtitle, tighter margins; same title/subtitle text and CTA |
| Full (today) | width ≥ 768px (tablets, desktop) | Exactly the current header — unchanged padding, sizes, spacing, content |

Transition: crossing 768px (rotate/resize) flips between states via CSS media query — no reload, page preserved.
