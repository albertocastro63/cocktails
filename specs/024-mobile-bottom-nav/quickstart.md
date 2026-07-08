# Quickstart: Mobile Bottom Navigation

**Branch**: `024-mobile-bottom-nav` | **Date**: 2026-07-08

## Automated (Vitest — the primary gate)

```bash
cd frontend
npm test -- src/nav/destinations.test.js src/components/BottomNav.test.js
npm test -- --coverage        # confirm ≥ 75% maintained
```

These cover the unit contract (U1–U7: role sets, partition/overflow, active matcher) and the DOM contract (N3–N7, B2–B7 via jsdom: rendered items, More menu open/close, aria state, focus, active token, auth-state update).

## Manual device/responsive verification

Run the app (`npm run dev`) and use browser devtools device emulation.

1. **Desktop (≥ 768px)** — top nav identical to today; no bottom bar (N1, SC-004).
2. **Phone (< 768px), signed out** — bottom bar shows All, Sign in; slim brand header on top; no horizontal scroll (N2, N3).
3. **Phone, regular user** — bottom bar shows All, Mine, New, Sign out; scroll a long page and confirm the last content clears the bar and the bar stays fixed (N4, B1).
4. **Phone, admin** — bottom bar shows All, Mine, New, Users, More; open More → Manage + Sign out; select/Escape/outside-tap all close it (N5, B4, B5).
5. **Active state** — navigate to each destination; the matching item (or More, for admin `/admin/recipes`) shows the active token (B2, B3).
6. **Rotate / resize across 768px** — nav switches top⇄bottom with no reload, same page (B8).
7. **Keyboard** — on a phone, focus the recipe form / search input; the bar hides so it doesn't cover the field; blur restores it (B9).
8. **Safe area** — on a device/emulator with a home indicator, the bar's tap targets sit above the indicator (FR-014); verify `viewport-fit=cover` is present in `index.html`.
9. **Sign in/out without reload** — sign in on a phone; the bar immediately switches to the user set; sign out returns to the visitor set (B6, B7).

## Growth check (SC-005)

Temporarily add a dummy admin destination in `navDestinations` and confirm the bar stays at 5 slots (new item lands in More), nothing clips — then remove it. (Also covered by unit test U6.)
