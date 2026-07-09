# Feature Specification: Password Recovery

**Feature Branch**: `026-password-recovery`  
**Created**: 2026-07-09  
**Status**: Draft  
**Input**: User description: "Password recovery feature: A user that has provided their email should be able to reset the login password using the standard mechanism of an email with a link to change the password. The link expires in 15 minutes and leads to a page where the user enters the new password twice; the password must be at least 12 characters and include numbers, lower- and upper-case letters, and symbols. The email should have a similar overall design to the website."

## Clarifications

### Session 2026-07-09

- Q: On the "Forgot password?" request page, should the response reveal whether the email is registered? → A: Neutral — always show the same confirmation whether or not the email is registered (no account enumeration); an email is sent only when the account exists.
- Q: Must the new password be different from the account's current password? → A: No — accept any password that meets the complexity rules; the new password is not compared against the previous one.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Request a reset link by email (Priority: P1)

A user who cannot sign in because they forgot their password goes to the sign-in page, chooses "Forgot password?", and enters the email address on their account. The system emails them a link to reset their password. To protect privacy, the on-screen confirmation is the same whether or not an account exists for that email, so the page never reveals which emails are registered.

**Why this priority**: This is the entry point of the whole feature — without a way to request the link, nothing else can happen. It also carries the most important security property (no account enumeration).

**Independent Test**: Submit a registered email on the "Forgot password?" page and confirm a reset email arrives; submit an unregistered email and confirm the on-screen message is identical and no email is sent.

**Acceptance Scenarios**:

1. **Given** the sign-in page, **When** the user selects "Forgot password?" and submits the email on their account, **Then** they see a neutral confirmation ("If an account exists for that email, a reset link has been sent") and receive a password-reset email.
2. **Given** the request page, **When** a user submits an email that is not registered, **Then** they see the exact same neutral confirmation and no email is sent.
3. **Given** a reset email, **When** the user opens it, **Then** it is styled consistently with the website (brand, colors) and contains a single link to set a new password, and never contains a password.

---

### User Story 2 - Set a new password via the emailed link (Priority: P1)

The user opens the link from the email within 15 minutes and lands on a page to choose a new password. They enter the new password twice; it must be strong (at least 12 characters with upper- and lower-case letters, a number, and a symbol) and both entries must match. On success, their password is changed and they can sign in with it.

**Why this priority**: This is the payoff of the flow — actually changing the password. Together with US1 it forms the minimum viable feature.

**Independent Test**: Follow a valid link, enter a matching strong password twice, submit, and confirm the user can sign in with the new password (and not the old one).

**Acceptance Scenarios**:

1. **Given** a valid, unexpired link, **When** the user enters a matching password that meets all complexity rules and submits, **Then** the password is updated, a success message is shown, and the user can sign in with the new password.
2. **Given** the reset page, **When** the two password entries do not match, **Then** submission is rejected with a clear "passwords do not match" message and the password is not changed.
3. **Given** the reset page, **When** the entered password does not meet a complexity rule (too short, or missing an upper/lower/number/symbol), **Then** submission is rejected with a clear message stating the unmet requirement(s).
4. **Given** a successful reset, **When** it completes, **Then** the link no longer works and any existing signed-in sessions for that account are signed out.

---

### User Story 3 - Expired, used, or invalid links (Priority: P2)

If the user opens a link that has expired (older than 15 minutes), has already been used, or is otherwise invalid, they get a clear explanation and an easy way to request a fresh link.

**Why this priority**: Protects the security model (time-limited, single-use links) and prevents user confusion, but it is a guard-rail around the core P1 flow.

**Independent Test**: Open an expired link, a used link, and a tampered/invalid link, and confirm each shows a clear message with a link to request a new reset — and none allows changing the password.

**Acceptance Scenarios**:

1. **Given** a link older than 15 minutes, **When** the user opens it, **Then** they see an "expired link" message and a way to request a new one, and cannot change the password.
2. **Given** a link that was already used successfully, **When** the user opens it again, **Then** they see an "invalid/used link" message and cannot change the password.
3. **Given** a link with a tampered or unrecognized token, **When** the user opens it, **Then** they see the same generic "invalid link" message (revealing nothing about why).

---

### Edge Cases

- Requesting a new reset while a prior link is still valid invalidates the earlier link (only the most recent link works).
- Rapid repeated requests for the same email are rate-limited to prevent inbox flooding/abuse; the neutral confirmation is still shown.
- A user who is currently signed in can still use the flow; completing it signs their other sessions out.
- The old password keeps working until a reset is completed (requesting a link does not lock the account).
- Email delivery failure does not change the on-screen neutral confirmation (to preserve non-enumeration); users can request again.
- The reset page is reached without signing in (the link itself is the authorization) and requires no other credentials.
- Submitting the reset form after the link has expired mid-session is rejected with the expired-link message.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The sign-in page MUST offer a "Forgot password?" entry point that leads to a page for requesting a reset.
- **FR-002**: A user MUST be able to request a password reset by submitting the email address associated with their account.
- **FR-003**: When a reset is requested, the system MUST send a reset email only if an account exists for that email, but MUST return the same neutral confirmation regardless of whether an account exists (no account enumeration).
- **FR-004**: The reset email MUST contain a single, unique link to a set-new-password page, and MUST NOT contain the password or any credential.
- **FR-005**: The reset link MUST expire 15 minutes after it is issued and MUST NOT work after expiry.
- **FR-006**: The reset link MUST be single-use (invalidated after a successful reset) and MUST be invalidated when a newer reset is requested for the same account.
- **FR-007**: The set-new-password page MUST require the new password to be entered twice, and MUST reject the request if the two entries do not match.
- **FR-008**: The new password MUST be at least 12 characters and contain at least one upper-case letter, one lower-case letter, one number, and one symbol; the system MUST reject passwords that fail any rule with a clear message. The new password is NOT required to differ from the account's current password (no comparison against the previous password is performed).
- **FR-009**: On a successful reset, the system MUST update the account's password so the user can sign in with it, and the old password MUST no longer work.
- **FR-010**: On a successful reset, the system MUST invalidate the account's existing signed-in sessions (the user must sign in again).
- **FR-011**: Expired, already-used, or invalid links MUST show a clear message and a path to request a new link, without allowing a password change and without disclosing why the link is invalid.
- **FR-012**: Reset requests MUST be rate-limited to prevent abuse (e.g., repeated emails to the same address in a short period).
- **FR-013**: The reset email MUST visually match the website's overall design (branding, colors, tone).
- **FR-014**: Complexity and match errors MUST be presented to the user clearly enough to correct the input before the reset succeeds.

### Key Entities

- **Password reset request**: Represents a pending reset for one account. Attributes: the associated account, a unique secret token embedded in the link, an issued time and a 15-minute expiry, and a used/consumed state. Only the most recent request per account is valid; it becomes invalid on use, on expiry, or when superseded.
- **Account**: The existing user account (has a username used to sign in and an email used to receive the reset). Its stored password and its "active sessions" validity are what a successful reset updates.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user who requests a reset for a registered email receives the email within ~2 minutes under normal conditions.
- **SC-002**: 100% of reset links stop working after 15 minutes or after one successful use (whichever comes first).
- **SC-003**: 100% of attempted new passwords that fail any complexity rule are rejected with a clear reason, and 100% that pass are accepted.
- **SC-004**: A user can complete the entire flow (request → open email → set new password → sign in) in under 5 minutes.
- **SC-005**: The request step reveals no information about whether an email is registered — the response is identical for registered and unregistered emails (verified by observed responses and timing).
- **SC-006**: After a successful reset, sessions that were active before the reset can no longer access the account.

## Assumptions

- **No account enumeration**: the request step always returns the same neutral confirmation; this is the intended security default even though it means a user who mistypes their email still sees "email sent."
- **Session invalidation**: a successful reset invalidates existing sessions for the account (the account model already supports a per-account session/token validity marker).
- **Single active link**: only the most recent reset link for an account is valid; older ones are superseded.
- **Symbols**: "symbol" means a common special/punctuation character (e.g., `! @ # $ % ^ & * ( ) - _ = +` and similar).
- **Reset by email, sign-in by username**: the user requests the reset with their email; after resetting, they sign in with their existing username and the new password.
- **Email capability**: the system can send transactional email (a new capability for this application); the specific email provider is an implementation decision.
- **Rate limiting**: a reasonable default limit applies to reset requests per email address over a short window; exact thresholds are a planning detail.
- **Reset page is public**: the set-new-password page is reachable directly from the emailed link and requires no prior sign-in; the token in the link is the authorization.
- **Scope**: this feature covers self-service password reset for existing accounts; it does not add account registration, email verification, or multi-factor authentication.
