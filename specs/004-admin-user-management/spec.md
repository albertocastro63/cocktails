# Feature Specification: Admin User Management Interface

**Feature Branch**: `004-admin-user-management`  
**Created**: 2026-05-12  
**Status**: Draft  
**Input**: User description: "Add admin interface for application. In this admin interface the admin can: List all users, Create new users, Edit and Delete users. Each user (aside from admin) has the following information associated to them: Username, Password, Name (First, Last), email. Only username and password are required."

## User Scenarios & Testing *(mandatory)*

### User Story 1 — User List (Priority: P1)

An administrator navigates to the admin section and sees a table of all registered users, showing each user's username, name, and email. This gives the admin full visibility into who has access to the application.

**Why this priority**: Before the admin can manage users, they must be able to see them. This is the entry point to all other admin actions and the minimum viable admin experience.

**Independent Test**: Navigate to the admin users page while logged in as an admin. Confirm a list of existing users is displayed with their username, name, and email columns. Confirm the page is not accessible when not logged in as admin.

**Acceptance Scenarios**:

1. **Given** an admin is logged in, **When** they navigate to the admin users page, **Then** a table of all non-admin users is displayed with columns for username, name, and email.
2. **Given** no non-admin users exist, **When** an admin views the users page, **Then** an empty state message is shown (e.g., "No users yet").
3. **Given** a non-admin user is logged in, **When** they attempt to access the admin users page, **Then** they are denied access with an appropriate error message.
4. **Given** an unauthenticated visitor, **When** they attempt to access the admin users page, **Then** they are redirected to the login page.

---

### User Story 2 — Create User (Priority: P2)

An administrator fills in a form to create a new user account, providing at minimum a username and password, and optionally a first name, last name, and email address. On success the new user appears in the user list.

**Why this priority**: Creating users is the core administrative action — without it the admin interface provides no operational value beyond read-only visibility.

**Independent Test**: From the admin users page, click "Add User", fill in username and password, submit the form. Confirm the new user appears in the list. Repeat with all optional fields. Confirm submitting without a username or password shows a validation error.

**Acceptance Scenarios**:

1. **Given** an admin is on the user creation form, **When** they submit a valid username and password, **Then** the user is created and appears in the user list.
2. **Given** an admin fills all fields (username, password, first name, last name, email), **When** they submit, **Then** all provided data is stored and visible in the user list.
3. **Given** an admin submits the form without a username, **When** the form is submitted, **Then** a validation error is shown and no user is created.
4. **Given** an admin submits the form without a password, **When** the form is submitted, **Then** a validation error is shown and no user is created.
5. **Given** an admin enters a username that already exists, **When** they submit, **Then** an error message indicates the username is taken.
6. **Given** an admin provides an email, **When** the email format is invalid, **Then** a validation error is shown.
7. **Given** an admin provides an email that is already in use by another user, **When** they submit, **Then** an error message indicates the email is already taken.

---

### User Story 3 — Edit User (Priority: P3)

An administrator selects an existing user and modifies their profile information — name, email, and optionally resets their password. Username cannot be changed (it is the stable identity key).

**Why this priority**: Editing user details is less frequently needed than creation, but essential for correcting mistakes or updating contact information.

**Independent Test**: From the user list, click Edit on a user. Change their first name and email. Save. Confirm the list reflects the updated values. Set a new password and confirm the user can log in with the new password.

**Acceptance Scenarios**:

1. **Given** an admin opens the edit form for a user, **When** they update the first name, last name, or email and save, **Then** the user's record is updated and the list reflects the new values.
2. **Given** an admin opens the edit form, **When** they leave the password field blank and save, **Then** the existing password is unchanged.
3. **Given** an admin enters a new password and saves, **When** the affected user logs in with the new password, **Then** login succeeds.
3a. **Given** an admin resets a user's password, **When** the affected user's existing session makes an authenticated request, **Then** that session is immediately rejected.
4. **Given** an admin provides an invalid email format, **When** they save, **Then** a validation error is shown and the record is not updated.
5. **Given** an admin changes a user's email to one already used by another user, **When** they save, **Then** an error message indicates the email is already taken and the record is not updated.
5. **Given** an admin views the edit form, **When** the form renders, **Then** the username field is read-only and cannot be modified.

---

### User Story 4 — Delete User (Priority: P4)

An administrator permanently removes a user account. A confirmation step prevents accidental deletion.

**Why this priority**: Deletion is needed but lower risk than creation or editing; it should be available but not prominent.

**Independent Test**: From the user list, click Delete on a user. Confirm a confirmation prompt appears. Confirm the deletion. Verify the user no longer appears in the list and can no longer log in.

**Acceptance Scenarios**:

1. **Given** an admin clicks Delete on a user, **When** a confirmation prompt is shown and the admin confirms, **Then** the user is permanently removed from the system and their recipes remain in the system with no creator assigned.
2. **Given** a confirmation prompt is shown, **When** the admin cancels, **Then** no change is made and the user remains in the list.
3. **Given** a user is deleted, **When** that user attempts to log in, **Then** login is rejected.
4. **Given** a user is deleted while they have an active session, **When** they make any authenticated request, **Then** their session is rejected immediately.
5. **Given** an admin attempts to delete themselves or another admin account, **When** the action is attempted, **Then** it is blocked with an explanatory message.

---

### Edge Cases

- What happens when an admin tries to delete their own account? (Blocked — FR-010.)
- What happens if two admins simultaneously edit the same user? (Last write wins; no conflict detection in v1.)
- How should the system handle an empty password field on the edit form (preserve existing vs. clear)? (Preserve existing — FR-008.)
- What happens when the user list contains a very large number of accounts? (Full list displayed; pagination deferred to v2.)
- What happens to a deleted user's recipes? (Orphaned — no creator assigned, recipes remain visible.)

## Clarifications

### Session 2026-05-12

- Q: When a user is deleted, what should happen to the recipes they created? → A: Orphan their recipes — recipes remain in the system, unowned (no creator).
- Q: Should a deleted or password-changed user's active session be invalidated immediately? → A: Yes — invalidate immediately on both delete and password change.
- Q: Can an admin grant admin privileges to a new or existing user through this interface? → A: No — this interface manages regular users only; admin accounts are created outside this UI.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide an admin-only user management section accessible from the main navigation when logged in as admin.
- **FR-002**: The system MUST display a list of all non-admin users showing username, first name, last name, and email.
- **FR-003**: The system MUST allow an admin to create a new user with a username and password (required) and optional first name, last name, and email. All users created through this interface are non-admin; admin account creation is outside the scope of this UI.
- **FR-004**: The system MUST validate that username and password are present before creating a user.
- **FR-005**: The system MUST reject creation when the provided username already exists and display a descriptive error.
- **FR-006**: The system MUST allow an admin to edit an existing user's first name, last name, email, and password.
- **FR-007**: The system MUST NOT allow the username to be changed after account creation.
- **FR-008**: When the password field is left blank during an edit, the system MUST preserve the user's existing password.
- **FR-009**: The system MUST allow an admin to delete a non-admin user after confirming the action; upon deletion the user's recipes are orphaned (retained in the system with no assigned creator).
- **FR-010**: The system MUST prevent deletion of admin accounts.
- **FR-011**: The system MUST deny access to all admin user management pages and actions for non-admin and unauthenticated users.
- **FR-012**: The system MUST validate email format when an email address is provided.
- **FR-013**: The system MUST reject creation or update when the provided email address is already in use by another user, and display a descriptive error.
- **FR-014**: When a user is deleted, the system MUST immediately invalidate all of that user's active sessions so they can no longer authenticate with existing tokens.
- **FR-015**: When an admin resets a user's password, the system MUST immediately invalidate all of that user's active sessions, requiring them to log in with the new password.

### Key Entities

- **User**: Represents an application account. Attributes: username (unique, required), password (required, stored securely), first name (optional), last name (optional), email (optional, valid format when provided), admin flag, creation timestamp.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An admin can create a new user in under 60 seconds from opening the admin page.
- **SC-002**: The user list loads and displays all existing users within 2 seconds under normal conditions.
- **SC-003**: All admin user management actions (create, edit, delete) are inaccessible to non-admin and unauthenticated users — zero unauthorised actions succeed.
- **SC-004**: Validation errors for missing required fields are shown immediately on form submission without a page reload.
- **SC-005**: A deleted user's session is invalidated immediately — any in-flight request using their token after deletion is rejected.
- **SC-006**: After an admin resets a user's password, the user's existing session is invalidated immediately and they must log in again with the new password.

## Assumptions

- The existing admin authentication mechanism (login + JWT) is reused without modification.
- Admin status is determined by the existing `is_admin` flag on the user record; no new role system is introduced. Admin accounts can only be created outside this UI (e.g., bootstrap or direct server operation).
- Password strength rules are not enforced beyond "non-empty" (consistent with existing behaviour).
- The admin user management section is accessible via a link visible only when logged in as admin.
- Pagination or search is out of scope for v1; the full user list is displayed.
- Email addresses must be unique across all users when provided; two users cannot share the same email.
- The admin's own account cannot be edited or deleted through this interface to prevent self-lockout.
