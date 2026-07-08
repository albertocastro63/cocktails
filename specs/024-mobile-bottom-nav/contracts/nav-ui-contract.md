# Navigation UI Contract: Mobile Bottom Navigation

**Branch**: `024-mobile-bottom-nav` | **Date**: 2026-07-08

Observable contract for the navigation across viewport size and auth state. Each row is an acceptance assertion (most are Vitest/jsdom-checkable; viewport-dependent ones are asserted by class presence and confirmed manually per `quickstart.md`).

## Rendering by viewport & state

| # | Viewport | State | Expected | Maps to |
|---|----------|-------|----------|---------|
| N1 | ≥ 768px | any | Existing top nav visible (`hidden md:flex` wrapper shown); no bottom bar visible (`md:hidden`) | FR-010, SC-004 |
| N2 | < 768px | any | Bottom bar visible (fixed, `md:hidden` shows); top nav hidden; slim brand header shown | FR-001, FR-002 |
| N3 | < 768px | visitor | Bottom bar direct items = All, Sign in (no More) | FR-005 |
| N4 | < 768px | user | Bottom bar direct items = All, Mine, New, Sign out (no More) | FR-005/006 |
| N5 | < 768px | admin | Bottom bar direct items = All, Mine, New, Users + **More**; More menu = Manage, Sign out | FR-005/006/007 |
| N6 | < 768px | any | Every bar item and every overflow item ≥ 44×44px | FR-008, SC-003 |
| N7 | < 768px | any | Each direct item shows an icon **and** a text label | FR-008a |

## Behavior

| # | Given | When | Then | Maps to |
|---|-------|------|------|---------|
| B1 | bottom bar shown | user scrolls | bar stays fixed at bottom; page end scrollable clear of bar (content has bottom padding) | FR-003, FR-004 |
| B2 | current route matches a direct item | render | that item has the active token; others do not | FR-009 |
| B3 | current route lives in overflow (admin on `/admin/recipes`) | render | the **More** button carries the active token | FR-009 |
| B4 | More menu closed | tap More | menu opens, `aria-expanded=true`, focus on first item | FR-006, §III |
| B5 | More menu open | select an item / press Escape / pointer-down outside / navigate | menu closes (`aria-expanded=false`); Escape returns focus to More | FR-013, §III |
| B6 | signed in on mobile | tap Sign out (direct or in More) | token cleared, routed home, nav reflects visitor set | FR-012 |
| B7 | signed out→in (no reload) | auth changes and a render occurs | bottom bar updates to the new destination set | FR-011 edge, FR-005 |
| B8 | viewport crosses 768px (resize/rotate) | media query flips | top⇄bottom switch with no reload, hash/page preserved | FR-011 |
| B9 | text input focused on mobile | focusin | bottom bar hidden so it doesn't cover the field; restored on blur | Edge: keyboard |
| B10 | any | render | no horizontal overflow; no nav item wraps to a second row, for visitor/user/admin | SC-001, FR-007 |

## Unit-level contract (pure functions)

| # | Function | Assertion |
|---|----------|-----------|
| U1 | `navDestinations('visitor')` | ids = [all-recipes, sign-in] |
| U2 | `navDestinations('user')` | ids = [all-recipes, my-recipes, new-recipe, sign-out] |
| U3 | `navDestinations('admin')` | ids include users + admin-recipes |
| U4 | `partitionNav(list≤5)` | direct = list, overflow = [] |
| U5 | `partitionNav(admin 6)` | direct = 4 items, overflow = 2, More synthesized |
| U6 | `partitionNav(7)` | direct still 4, overflow = 3 (bounded, SC-005) |
| U7 | active matcher | exact + prefix (`/recipes/new`) match correctly; actions never active |
