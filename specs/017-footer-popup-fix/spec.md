# Feature Specification: Site Footer and Ingredient Popup Layout Fix

**Feature Branch**: `017-footer-popup-fix`
**Created**: 2026-05-25
**Status**: Draft
**Input**: User description: "Add a footer to each page, make it so that when the popup with ingredients is displayed the page does not increase in size and the user has to scroll down to see the popup. The footer should consist of a line that spans the width of the display area (not the whole screen) and a copyright notice below."

## Clarifications

### Session 2026-05-25

- Q: Should the copyright year in the footer be dynamic (always shows the current calendar year) or static (hardcoded at build time)? → A: Dynamic — always reflects the current calendar year automatically.
- Q: Does the footer appear on the login page and other auth/entry screens, or only on main content pages? → A: All pages — footer appears everywhere including login and any error pages.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Site-Wide Footer (Priority: P1)

Every page in the application displays a consistent footer at the bottom of its content. The footer contains a horizontal separator line — constrained to the same width as the main content area, not the full browser window — followed by a copyright notice beneath it. This gives the application a polished, finished appearance and establishes consistent branding across all pages.

**Why this priority**: A footer is visible on every page and is a quick win for perceived quality. It can be implemented and verified independently of the popup fix.

**Independent Test**: Navigate to any page (recipe list, recipe detail, login, admin) and confirm a footer with a separator line and copyright text is visible at the bottom of the content.

**Acceptance Scenarios**:

1. **Given** a user visits any page, **When** they scroll to the bottom, **Then** they see a horizontal separator line followed by a copyright notice, both within the content area width.
2. **Given** a narrow viewport, **When** the user scrolls to the bottom, **Then** the separator line spans the full width of the content area without breaking layout.
3. **Given** a wide viewport, **When** the user scrolls to the bottom, **Then** the separator line does not extend beyond the content area's maximum width and does not touch the browser edges.
4. **Given** any page, **When** the user views the footer, **Then** the copyright notice is legible and positioned below the separator line.

---

### User Story 2 — Ingredient Popup as Non-Expanding Overlay (Priority: P1)

When a user views the recipe list and triggers the ingredient popup on a recipe card, the popup appears as a visual overlay. The total height of the page does not change and no content below the card shifts downward. The popup is positioned near its triggering card and the user can scroll the page normally while the popup is visible.

**Why this priority**: The current behaviour — expanding the page height when the popup opens — is disorienting and breaks the browsing experience. Users lose their scroll context and the layout jumps unexpectedly.

**Independent Test**: On the recipe list page, trigger the ingredient popup on any recipe card. Confirm the vertical scroll position, total page height, and position of all other cards remain identical before and after the popup appears.

**Acceptance Scenarios**:

1. **Given** a user is browsing the recipe list, **When** they trigger the ingredient popup on a card, **Then** the page total height does not increase and no surrounding content moves.
2. **Given** the ingredient popup is visible, **When** the user scrolls the page, **Then** the page scrolls normally and the popup moves with the page content (it is not fixed to the screen).
3. **Given** the ingredient popup is open, **When** the user moves their cursor away or clicks elsewhere, **Then** the popup closes and the page layout is unchanged.
4. **Given** a recipe card near the bottom of the current viewport, **When** the ingredient popup opens, **Then** the popup renders in its natural position without forcing the page to grow; the user may scroll to see it fully.

---

### Edge Cases

- What if a page's content is shorter than the viewport height? The footer appears below the content in the natural document flow; it is not pinned to the bottom of the viewport.
- What if the popup opens on a card at the very bottom of the page? The popup renders at its natural position; the page does not expand, but the user may need to scroll to see the popup fully.
- What if multiple popups are triggered in quick succession? Only one popup is visible at a time; opening a new one closes the previous one without any layout shift.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The application MUST display a footer at the bottom of every page's content, including the login page and any error pages.
- **FR-002**: The footer MUST contain a horizontal separator line that spans the full width of the content display area.
- **FR-003**: The separator line MUST NOT extend beyond the content area's maximum width; on wide viewports it must be bounded the same way as the main content, not the full screen width.
- **FR-004**: The footer MUST display a copyright notice immediately below the separator line.
- **FR-005**: The ingredient popup on the recipe list MUST appear as an overlay that does not alter the total height of the page or shift any surrounding content.
- **FR-006**: When the ingredient popup is visible, all other recipe cards and page elements MUST remain in their original positions.
- **FR-007**: The ingredient popup MUST close when the user moves their cursor away from the triggering element or clicks elsewhere on the page.
- **FR-008**: The footer styling MUST be visually consistent with the existing design language of the application.
- **FR-009**: The copyright year in the footer MUST always reflect the current calendar year without requiring a code or content change when the year rolls over.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The footer is present and visible on 100% of pages in the application.
- **SC-002**: On any viewport width, the footer separator line stays within the content area boundaries — the line's rendered width matches the content container's width, not the viewport width.
- **SC-003**: Opening the ingredient popup produces zero change in the page's total scroll height and zero positional shift in any other recipe card.
- **SC-004**: The copyright notice is legible against the page background.

## Assumptions

- The copyright year in the footer is dynamic — it always reflects the current calendar year and does not require a code change to update annually. The full notice reads "© [current year] Cocktails"; the wording can be adjusted without requiring a new specification.
- The footer appears in the normal document flow at the bottom of each page's content — it is not sticky or fixed to the viewport bottom.
- The ingredient popup fix applies to the recipe list page popover introduced in feature 005; no other popups are in scope.
- The footer follows the existing amber/stone visual design system used across the application.
- No backend or data model changes are required; this is a purely presentational change.
