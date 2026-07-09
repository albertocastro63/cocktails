# Feature Specification: Compact Landing Header on Mobile

**Feature Branch**: `025-mobile-landing-header`  
**Created**: 2026-07-09  
**Status**: Draft  
**Input**: User description: "Mobile-friendly design of landing page: on phone-sized screens, the landing page has a header that uses too much of the real estate of the small screen. I would like the header (that displays Cocktail Recipes/Discover your next favorite drink) to be reduced in size. Maintain the design as-is for all the other screen sizes."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Visitor sees a compact landing header on a phone (Priority: P1)

Someone opens the landing (home) page on a phone. Today the header banner — the "Cocktail Recipes" title and "Discover your next favorite drink" subtitle — takes up a large portion of the small screen, pushing the main call-to-action and content down. With this change, on phone-sized screens the header is noticeably smaller so more of the page (including the primary action) is visible without scrolling, while the same header on larger screens is unchanged.

**Why this priority**: This is the whole feature and the reported problem — the header wastes scarce vertical space on phones, degrading the first impression and requiring extra scrolling to reach the content.

**Independent Test**: Load the landing page on a phone-sized viewport and confirm the header banner occupies noticeably less vertical space than before and the primary call-to-action is reachable with little or no scrolling; load it on a desktop-sized viewport and confirm the header looks exactly as it does today.

**Acceptance Scenarios**:

1. **Given** a visitor on a phone-sized viewport, **When** the landing page loads, **Then** the header banner (title + subtitle) occupies a noticeably smaller vertical area than the current design, and no header content is clipped or overflows horizontally.
2. **Given** a visitor on a phone-sized viewport, **When** the landing page loads, **Then** the primary call-to-action (the "All Recipes" action) is visible within, or immediately below, the first screenful — reachable with minimal scrolling.
3. **Given** a visitor on a tablet- or desktop-sized viewport, **When** the landing page loads, **Then** the header renders exactly as it does today (size, spacing, and content unchanged).
4. **Given** the compact mobile header, **When** it is displayed, **Then** it still shows the same title and subtitle text, remains legible, and keeps the site's visual style.

---

### Edge Cases

- On very narrow or short phones (including landscape orientation), the header must remain legible and must not clip, wrap awkwardly, or cause horizontal overflow.
- The header title and subtitle text are unchanged — only their presentation size/spacing is reduced on phones; text is not truncated or hidden.
- Crossing the phone/large-screen boundary (rotate or resize) switches between the compact and full header without a page reload.
- Dynamic type / larger system font settings must not break the compact header layout.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: On phone-sized viewports, the landing page header banner MUST occupy meaningfully less vertical space than the current design (reduced height/padding and title/subtitle sizing).
- **FR-002**: The compact header MUST still display the same title ("Cocktail Recipes") and subtitle ("Discover your next favorite drink") text, with no truncation or removal of content.
- **FR-003**: On viewports larger than phone size (tablets and desktop), the header MUST remain exactly as it is today — unchanged size, spacing, styling, and content.
- **FR-004**: The compact header MUST NOT clip content, wrap awkwardly, or cause horizontal overflow on phone-sized viewports, including narrow and short (landscape) screens.
- **FR-005**: The header MUST retain the site's existing visual style (colors, typography family, call-to-action) — only the size/spacing changes on phones.
- **FR-006**: Switching across the phone/large-screen boundary (rotation or window resize) MUST update the header presentation without a page reload.
- **FR-007**: The change MUST be limited to the landing (home) page header; no other page or component is altered.

### Key Entities

*Not applicable — this is a presentational change with no new data.*

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On a typical phone screen, the landing header banner's vertical height is reduced by at least ~50% compared to the current design (the header should be clearly slim, not just marginally smaller).
- **SC-002**: On a typical phone screen, the primary call-to-action ("All Recipes") is visible without scrolling (or reachable with a single short scroll), where it currently requires more scrolling.
- **SC-003**: On tablet and desktop screens, the landing header is pixel-for-pixel unchanged from the current release.
- **SC-004**: 100% of phone, tablet, and desktop states render the landing page with no horizontal overflow and no clipped or truncated header text.

## Assumptions

- "Phone-sized" means a viewport width below 768px, consistent with the breakpoint established by the mobile bottom-navigation feature (024); tablets and desktop (≥ 768px) keep the current header.
- The header's text content is unchanged; only its size and spacing are reduced on phones.
- Scope is limited to the landing (home) page hero/header; the top/bottom navigation and all other pages are out of scope.
- "Reduced by ~40%" is a target to make the outcome measurable; the exact reduction is a design decision during planning, provided the header is clearly more compact and the acceptance scenarios pass.
