# Feature Specification: Mobile Bottom Navigation

**Feature Branch**: `024-mobile-bottom-nav`
**Created**: 2026-07-08
**Status**: Draft
**Input**: User description: "Mobile-friendly navigation: on phone-sized screens, move the site navigation to a fixed bottom navigation bar so the page renders properly on mobile. Currently the top navigation overflows/breaks the layout on phones after login (nav items wrap, content is cut off, horizontal overflow). The bottom navigation must scale gracefully as the number of links grows (e.g. admin users see more links than regular users), avoiding overflow or truncation regardless of link count. Desktop/tablet keeps the existing top navigation."

## Clarifications

### Session 2026-07-08

- Q: How should each destination appear in the bottom navigation bar? → A: Icon + short label (stacked), the standard mobile tab-bar pattern; overflow ("More") uses the same treatment. Icons are inline SVGs (no new dependency).
- Q: Below which viewport width should the bottom navigation replace the top navigation? → A: Below 768px (Tailwind `md`) shows the bottom nav (phones); 768px and wider keeps the existing top nav (tablets, desktop).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Logged-in user navigates from a phone (Priority: P1)

A signed-in user opens the site on a phone. Instead of a crowded top bar that wraps and pushes content off-screen, they see a slim top bar with the site name and a fixed navigation bar at the bottom of the screen with their available destinations (All Recipes, My Recipes, New Recipe, Sign Out). The page content renders at full width with no horizontal overflow, and the bottom bar stays visible while they scroll.

**Why this priority**: This is the reported defect — the current layout breaks for every signed-in phone user, which affects the core browsing experience. Fixing it for the regular-user link set delivers the primary value.

**Independent Test**: Can be fully tested by signing in as a regular user on a phone-sized viewport and verifying all navigation destinations are reachable from the bottom bar, the page has no horizontal scrolling, and no content is hidden behind the navigation.

**Acceptance Scenarios**:

1. **Given** a signed-in regular user on a phone-sized viewport, **When** any page loads, **Then** navigation appears as a bar fixed to the bottom of the screen and the top of the page shows only the site brand.
2. **Given** the bottom navigation is displayed, **When** the user scrolls the page content, **Then** the bottom bar remains visible and the last items of page content can still be scrolled fully into view (not hidden behind the bar).
3. **Given** a signed-in user on a phone-sized viewport, **When** they view any page, **Then** the page has no horizontal overflow and no navigation item is clipped or wrapped onto a second row.
4. **Given** the bottom navigation, **When** the user taps a destination, **Then** they navigate there and the bar indicates which destination is currently active.

---

### User Story 2 - Admin user with a longer menu (Priority: P2)

An admin signs in on a phone. They have more destinations than a regular user (including user management and all-recipes administration). The bottom navigation presents the most important destinations directly and gathers the remaining ones behind a clearly labeled overflow entry (e.g. "More"), so nothing is ever clipped, truncated, or pushed off-screen — regardless of how many links their role grants.

**Why this priority**: Admins are a smaller audience than regular users, but the design must prove it scales beyond the base link set; this is the explicit growth requirement.

**Independent Test**: Can be tested by signing in as an admin on a phone-sized viewport and verifying every admin destination is reachable from the bottom navigation without horizontal scrolling or clipped items.

**Acceptance Scenarios**:

1. **Given** a signed-in admin on a phone-sized viewport, **When** any page loads, **Then** all admin destinations are reachable from the bottom navigation and no item is clipped or overlaps another.
2. **Given** more destinations than fit comfortably in the bar, **When** the bar is rendered, **Then** the excess destinations are accessible through an overflow entry in the bar, and opening it shows the remaining destinations as tappable items.
3. **Given** the overflow menu is open, **When** the user taps outside it or selects a destination, **Then** the menu closes.
4. **Given** future roles or features add more links, **When** the number of destinations grows, **Then** the bar continues to show a bounded number of direct items plus the overflow entry, never shrinking items below a usable tap size.

---

### User Story 3 - Visitor and desktop experience unchanged (Priority: P3)

A visitor who is not signed in sees the appropriate reduced navigation (browse recipes, sign in) in the same mobile pattern. Users on desktop or tablet-sized screens continue to see the existing top navigation exactly as today.

**Why this priority**: Protects existing behavior for the majority screen sizes and completes coverage of all authentication states; it is a guard-rail rather than new value.

**Independent Test**: Can be tested by loading the site logged-out on a phone-sized viewport (bottom bar shows visitor links) and by loading it on a desktop-sized viewport (top navigation identical to current behavior, no bottom bar).

**Acceptance Scenarios**:

1. **Given** a signed-out visitor on a phone-sized viewport, **When** a page loads, **Then** the bottom bar shows only the destinations available to visitors.
2. **Given** any user on a desktop- or tablet-sized viewport, **When** a page loads, **Then** the existing top navigation is shown and no bottom navigation bar appears.
3. **Given** a user rotates their phone or resizes the window across the phone-size threshold, **When** the viewport crosses the threshold, **Then** the layout switches between bottom and top navigation without requiring a page reload and without losing the current page.

---

### Edge Cases

- What happens when the on-screen keyboard is open (e.g. while typing in the recipe form or search)? The bottom bar must not cover or fight with form inputs; content being edited must remain visible.
- What happens on phones with rounded corners / home-indicator areas? The bar's tap targets must sit inside the device's safe display area.
- What happens on a phone in landscape orientation (short viewport height)? Navigation must remain usable and content must retain reasonable vertical space.
- What happens when the user signs in or out without a full page reload? The bottom bar must immediately reflect the new set of destinations.
- What happens with very long destination labels or a translated/renamed link? Labels may shorten visually but each item must remain identifiable and tappable.
- What happens when the overflow menu is open and the user navigates back / changes route? The menu must close and not block the page.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: On phone-sized viewports (width below 768px), the system MUST present site navigation as a bar fixed to the bottom of the screen, replacing the current top navigation links.
- **FR-002**: On phone-sized viewports, the top of the page MUST show at most a slim brand/header area that does not wrap, clip, or cause horizontal overflow.
- **FR-003**: The bottom navigation MUST remain visible and in a fixed position while the user scrolls page content.
- **FR-004**: Page content MUST never be obscured by the bottom navigation; the end of any scrollable page must be fully viewable.
- **FR-005**: The bottom navigation MUST show the destinations appropriate to the user's state: visitor, signed-in regular user, or signed-in admin.
- **FR-006**: The bottom navigation MUST display a bounded number of direct items; when the user's destinations exceed that bound, the remaining destinations MUST be reachable through an overflow entry ("More") within the bar.
- **FR-007**: The navigation MUST accommodate any number of destinations without horizontal overflow, item clipping, label overlap, or multi-row wrapping — including future growth beyond today's admin link count.
- **FR-008**: Every navigation item, including items inside the overflow menu, MUST meet a comfortable minimum tap-target size for touch use.
- **FR-008a**: Each direct bottom-bar item MUST present a recognizable icon with a short text label (stacked, standard tab-bar style); the overflow ("More") entry MUST use the same icon+label treatment. Icons are inline SVGs, requiring no new dependency.
- **FR-009**: The bottom navigation MUST indicate which destination is currently active.
- **FR-010**: On viewports 768px wide and larger (tablets and desktop), the system MUST continue to show the existing top navigation and MUST NOT show the bottom bar.
- **FR-011**: When the viewport crosses the phone-size threshold (rotation or window resize), the navigation MUST switch modes without a page reload and without losing the user's current page.
- **FR-012**: Sign Out MUST remain reachable from the mobile navigation (directly or via the overflow menu) and behave identically to the desktop action.
- **FR-013**: The overflow menu MUST close when the user selects a destination, taps outside it, or navigates.
- **FR-014**: The bottom navigation MUST respect device safe areas (e.g. home-indicator region) so no tap target is partially unusable.

### Key Entities

- **Navigation destination**: A labeled link or action available to the current user (e.g. All Recipes, My Recipes, New Recipe, Users, Sign Out); has a label, an icon (inline SVG) for the bottom bar, a target or action, a visibility rule based on authentication state and role, and a priority that determines whether it appears as a direct item or inside the overflow menu.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On phone-sized screens, 100% of pages render with zero horizontal overflow and zero navigation items clipped or wrapped, for visitor, regular-user, and admin states.
- **SC-002**: Every destination available to a user's role is reachable from the mobile navigation in at most 2 taps (1 tap for direct items, 2 via the overflow menu).
- **SC-003**: All mobile navigation tap targets measure at least the platform-recommended minimum touch size (≈44×44 points equivalent).
- **SC-004**: Desktop and tablet users see no change: the existing top navigation renders identically to the current release.
- **SC-005**: Adding a hypothetical extra destination (simulating future growth) causes no layout breakage — the new destination lands in the bar or overflow menu per its priority.

## Assumptions

- "Phone-sized" means a viewport width below 768px (Tailwind `md`). At 768px and above (tablets and desktops) the current top navigation is kept.
- The overflow ("More") pattern is the chosen strategy for scaling beyond the direct-item bound, in line with common mobile tab-bar conventions; a horizontally scrolling bar was considered and rejected as a default because it hides items with no visual cue.
- Up to 5 direct items (including the "More" entry when present) is the working bound for the bar; item priority ordering (which links are direct vs. overflow) will be settled during planning, with primary browsing/creation actions favored for direct slots.
- The current set of destinations is: visitor — All Recipes, Sign In; regular user — All Recipes, My Recipes, New Recipe, Sign Out; admin — the regular set plus Users and admin Recipes management.
- No new destinations are added by this feature; it reorganizes the presentation of existing ones.
- The existing top navigation's visual design on desktop/tablet is out of scope and remains untouched.
