# Research: Compact Landing Header on Mobile

**Branch**: `025-mobile-landing-header` | **Date**: 2026-07-09

---

## Decision 1 — Mobile-first responsive classes with `md:` overrides equal to today's values

**Decision**: Give the hero wrapper, title, and subtitle smaller default (phone) sizes and add `md:` variants that exactly match the current fixed values (`py-16`, `text-4xl`, `mb-3`, `text-lg`, `mb-6`).

**Rationale**: Because every `md:` variant equals today's value, the header at ≥ 768px is pixel-for-pixel unchanged (SC-003) with zero risk to the desktop/tablet design, while phones get a smaller banner. The switch is a CSS media query, so rotate/resize updates it with no reload (FR-006), consistent with the 768px breakpoint established by feature 024.

**Alternatives considered**:
- *JavaScript that swaps classes based on `matchMedia`* — rejected; unnecessary code and reflow for something CSS does natively.
- *A separate mobile hero element* — rejected; duplicates content and risks desktop drift.

---

## Decision 2 — Reduce padding + title/subtitle sizes (not remove content)

**Decision**: Halve the hero vertical padding (`py-16` → `py-8` on phones) and step down the title (`text-4xl` → `text-2xl`) and subtitle (`text-lg` → `text-base`), with tighter bottom margins. Keep the exact title/subtitle text and the CTA.

**Rationale**: This targets the ~40% vertical reduction (SC-001) from the biggest contributors (padding + type scale) while keeping the header legible and on-brand (FR-002, FR-005). Content is unchanged — only presentation shrinks.

**Alternatives considered**:
- *Hide the subtitle on phones* — rejected; the spec requires the same text to remain visible.
- *Only shrink padding* — insufficient to reach the target; type scale is a large contributor on small screens.

---

## Resolved unknowns

| Unknown | Resolution |
|---------|-----------|
| Breakpoint | `md` = 768px (matches feature 024) |
| Desktop parity mechanism | `md:` overrides equal to current fixed values → provably unchanged |
| Reduction target | ~35–40% via halved padding + smaller title/subtitle + tighter margins; tunable |
| Content changes | None — same title/subtitle text and CTA |
