# Data Model: Mobile Bottom Navigation

**Branch**: `024-mobile-bottom-nav` | **Date**: 2026-07-08

No persisted/domain data. This models the in-memory **navigation destination** structure and the deterministic rules that turn it into the two bars, so the logic is directly unit-testable.

## Entity: Navigation destination

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Stable key (e.g. `all-recipes`, `my-recipes`, `new-recipe`, `users`, `admin-recipes`, `sign-out`, `sign-in`) |
| `label` | string | Short text shown under the icon (e.g. "All", "Mine", "New", "Users", "Sign out") and full text in top nav / overflow |
| `icon` | string (SVG) | Inline SVG markup for the bottom bar / overflow |
| `href` | string \| null | Hash target (e.g. `#/recipes`); null when it is an action |
| `action` | function \| null | For non-navigation items (e.g. Sign Out clears token then routes) |
| `priority` | number | Lower = prefer a direct slot when overflow is needed |
| `match` | string \| RegExp | Route pattern used to mark the item active (compared to `getPath()`) |
| `roles` | set | Which states show it: `visitor`, `user`, `admin` |

## Per-state destination sets (source of truth)

`navDestinations(state)` returns, in display order:

| State | Destinations (id) |
|-------|-------------------|
| visitor | `all-recipes`, `sign-in` |
| user | `all-recipes`, `my-recipes`, `new-recipe`, `sign-out` |
| admin | `all-recipes`, `my-recipes`, `new-recipe`, `users`, `admin-recipes`, `sign-out` |

## Priority table (direct-slot preference when overflow is needed)

| id | priority | label (mobile) | match |
|----|----------|----------------|-------|
| `all-recipes` | 1 | All | `/recipes` (browse) |
| `my-recipes` | 2 | Mine | `/my-recipes` |
| `new-recipe` | 3 | New | `/recipes/new` |
| `sign-in` | 3 | Sign in | `/login` |
| `users` | 4 | Users | `/admin/users` |
| `admin-recipes` | 5 | Manage | `/admin/recipes` |
| `sign-out` | 6 | Sign out | (action; never "active") |

## Partition rule: `partitionNav(list, maxSlots = 5) → { direct, overflow }`

- If `list.length ≤ maxSlots`: `direct = list`, `overflow = []` (no More entry).
- Else: `direct = the first (maxSlots - 1) items by priority`, `overflow = the rest`; the bar renders `direct` + a synthetic **More** slot (`id = "more"`) that opens a menu listing `overflow`.
- `direct.length + (overflow.length ? 1 : 0) ≤ maxSlots` always holds.

### Worked results (today's sets)

| State | Count | Direct | Overflow (More) |
|-------|-------|--------|-----------------|
| visitor | 2 | All, Sign in | — |
| user | 4 | All, Mine, New, Sign out | — |
| admin | 6 | All, Mine, New, Users | Manage, Sign out |

### Growth check (SC-005)

Adding a hypothetical 7th destination (priority 7) to admin keeps `direct = All, Mine, New, Users` and grows `overflow` to `{Manage, Sign out, New-thing}` — bar stays at 5 slots, nothing clipped.

## Home reachability

Home (`/`) is **not** a bottom-bar destination. On mobile it is reached via the slim brand header, which is an anchor to `#/` (mirrors the desktop brand). On the Home route, no bottom-bar item is active (All Recipes is active only on `/recipes`).

## Active-state rule (FR-009)

Given `path = getPath()`:
- A **direct** item is active when its `match` equals `path` (exact) or, for prefix matches like `/recipes/new`, when `path` starts with it.
- The **More** button is active when the active route belongs to a destination in `overflow`.
- Action items (`sign-out`) are never active.

## State transitions

```
auth change (login/logout)  → navDestinations(newState) recomputed on next renderPage() → bars update (no reload)
route change                → nav re-rendered; overflow menu (if open) is torn down; active item recomputed
viewport crosses 768px      → CSS media query flips top/bottom visibility (no JS, no reload)
input focus (mobile)        → bar hidden; blur → bar restored
```
