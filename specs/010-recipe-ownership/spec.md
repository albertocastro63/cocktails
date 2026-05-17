# Feature Specification: Recipe Ownership and Per-User Recipe Listing

**Feature Branch**: `010-recipe-ownership`  
**Created**: 2026-05-15  
**Status**: Draft  

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Recipe Edit and Delete Restricted to Owner (Priority: P1)

A logged-in user who created a recipe can edit or delete it. Any other logged-in user who did not create the recipe cannot edit or delete it — attempts are rejected. Unauthenticated users cannot edit or delete any recipe.

**Why this priority**: This is the core ownership enforcement. Without it, any authenticated user can overwrite or remove another user's recipes, which is the fundamental problem being solved.

**Independent Test**: Can be fully tested by logging in as two different users, having user A create a recipe, then verifying user B's edit/delete attempts are rejected while user A's succeed.

**Acceptance Scenarios**:

1. **Given** a logged-in user who created a recipe, **When** they edit that recipe, **Then** the changes are saved successfully.
2. **Given** a logged-in user who created a recipe, **When** they delete that recipe, **Then** the recipe is removed successfully.
3. **Given** a logged-in user who did NOT create a recipe, **When** they view that recipe, **Then** edit/delete controls are not shown to them.
4. **Given** a logged-in user who did NOT create a recipe, **When** they attempt to edit or delete it via the API directly, **Then** they receive a "forbidden" error and the recipe is unchanged.
5. **Given** an unauthenticated request to edit or delete any recipe, **When** the request is made, **Then** it is rejected with an authentication error.

---

### User Story 2 - Administrator Can Edit or Delete Any Recipe (Priority: P2)

An administrator can edit or delete any recipe regardless of who created it, providing a moderation capability.

**Why this priority**: Administrators need a safety valve — the ability to correct errors or remove inappropriate content across all users. This is secondary to ownership enforcement because it extends rather than replaces the base model.

**Independent Test**: Can be fully tested by logging in as admin and attempting to edit/delete a recipe created by a regular user; both operations should succeed.

**Acceptance Scenarios**:

1. **Given** an administrator, **When** they edit a recipe created by any user, **Then** the changes are saved successfully.
2. **Given** an administrator, **When** they delete a recipe created by any user, **Then** the recipe is removed successfully.

---

### User Story 3 - Per-User Recipe Listing (Priority: P3)

Any logged-in user (including administrators) can view a list of only the recipes they have personally created. This listing uses the same visual style and layout as the main all-recipes page.

**Why this priority**: This is a convenience and discoverability feature. Users can find their own recipes without scanning the global list. It delivers value only after ownership enforcement (P1) makes "created by me" a meaningful filter.

**Independent Test**: Can be fully tested by logging in as user A, clicking "My Recipes" in the nav bar, and verifying only user A's recipes appear styled like the main page.

**Acceptance Scenarios**:

1. **Given** a logged-in user who has created recipes, **When** they navigate to their personal recipe list, **Then** only recipes created by that user are shown.
2. **Given** a logged-in user who has not created any recipes, **When** they navigate to their personal recipe list, **Then** an empty state is shown (no recipes).
3. **Given** an administrator who has created some recipes, **When** they navigate to their personal recipe list, **Then** only recipes created by that administrator are shown (not all recipes in the system).
4. **Given** a logged-in user viewing their personal recipe list, **When** the page renders, **Then** each recipe card matches the same visual style as the main all-recipes page.
5. **Given** a logged-in user viewing their personal recipe list, **When** the page renders, **Then** every recipe card shows edit/delete controls (the viewer owns all listed recipes).

---

### Edge Cases

- What happens when a recipe has no `owner` field (recipes created before this feature)? They should be treated as unowned — only administrators can edit or delete them.
- How does the system handle a concurrent edit where ownership is transferred? Ownership is fixed at creation time and cannot be transferred.
- What happens when the owner account is deleted? The recipe remains but becomes effectively unowned (only admins can modify it).
- How does the system handle an unauthorized edit attempt by a non-admin? Returns HTTP 403 Forbidden with a clear error message; no partial writes occur.

## Clarifications

### Session 2026-05-17

- Q: Should edit/delete buttons be hidden from non-owners in the UI, or shown to everyone with enforcement only at the API level? → A: Hide buttons — edit/delete controls only appear for the recipe owner and admins; API list responses must expose owner context so the client can determine visibility.
- Q: Where is the UI entry point for the "My Recipes" listing? → A: Nav bar link — a "My Recipes" link in the main navigation header, visible only when logged in.
- Q: On the "My Recipes" page, should each recipe card show edit/delete controls? → A: Always show — edit/delete controls appear on every card in the "My Recipes" listing since the viewer owns all recipes there.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST store the creator's user ID on each recipe at creation time.
- **FR-002**: System MUST allow a recipe's owner to edit that recipe.
- **FR-003**: System MUST allow a recipe's owner to delete that recipe.
- **FR-004**: System MUST reject edit requests for a recipe from any authenticated user who is not the recipe's owner and is not an administrator, returning a 403 Forbidden response.
- **FR-005**: System MUST reject delete requests for a recipe from any authenticated user who is not the recipe's owner and is not an administrator, returning a 403 Forbidden response.
- **FR-006**: System MUST allow an administrator to edit any recipe regardless of who created it.
- **FR-007**: System MUST allow an administrator to delete any recipe regardless of who created it.
- **FR-008**: System MUST provide an endpoint that returns only the recipes created by the currently authenticated user.
- **FR-009**: Recipes created before this feature was introduced (i.e., with no stored owner) MUST be treated as unowned; only administrators may edit or delete them.
- **FR-010**: The per-user recipe listing page MUST use the same visual design, card layout, and interaction patterns as the main all-recipes listing page.
- **FR-011**: The system MUST hide edit/delete controls from users who are neither the recipe's owner nor an administrator; these controls MUST be visible only to the owner and to administrators.
- **FR-012**: Recipe list API responses MUST include sufficient owner context (owner identifier) so the client can determine edit/delete control visibility without a separate request.
- **FR-013**: The application MUST display a "My Recipes" link in the main navigation header that is visible only to authenticated users.
- **FR-014**: The "My Recipes" listing MUST display edit/delete controls on every recipe card, as the authenticated user is always the owner of all listed recipes.

### Key Entities

- **Recipe**: Existing entity extended with a new `ownerID` attribute — the user ID of the user who created the recipe. `ownerID` is immutable after creation.
- **User**: Existing entity. The `isAdmin` flag determines administrator privilege for ownership bypass.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of edit and delete requests made by non-owners (non-admin) are rejected with a 403 response — no unauthorized modifications reach the data store.
- **SC-002**: 100% of edit and delete requests made by the recipe's owner or by an administrator succeed (no false rejections).
- **SC-003**: The per-user recipe listing returns only recipes whose `ownerID` matches the requesting user's ID, with zero leakage of other users' recipes.
- **SC-004**: The per-user recipe listing page is visually indistinguishable in layout and style from the main all-recipes page (same card dimensions, typography, spacing, and interaction states).
- **SC-005**: Recipes with no `ownerID` (legacy recipes) are inaccessible for modification by non-administrators — they behave as if owned by the system.

## Assumptions

- Authentication is already implemented; this feature builds on top of existing JWT-based auth and the `isAdmin` flag on users.
- Recipe creation (`POST /api/v1/recipes`) already requires authentication; the `ownerID` is derived from the authenticated user's JWT claims, not from the request body.
- Mobile support is out of scope for v1 of this feature; the per-user listing is a web-only page.
- The visual design system (amber/stone palette, card components) from the existing all-recipes page is available for reuse in the per-user listing without modification.
- Legacy recipes (those without an `ownerID`) exist in the data store and must be handled gracefully; a backfill migration is out of scope.
- Ownership cannot be transferred between users; `ownerID` is set once at creation and is immutable.
