# Feature Specification: Backend Logging

**Feature Branch**: `027-backend-logging`  
**Created**: 2026-07-10  
**Status**: Draft  
**Input**: User description: "Logging on the backend: The Lambdas that manage the backend infrastructure lack proper logging and that makes it very difficult to debug potential problems. What is needed is a logging system that would output error logging in production, and informational or debugging logging in preview environments. These logging should be controlled by a Lambda environment variable that is set appropriately for each environment and that, if needed, can be changed on the AWS console. All the main actions that the backend performs such as logging in and out, adding a recipe, editing a recipe, getting data to display a recipe or a list of ingredients, searches, etc., should be covered in this logging. Logs will be available in the Lambda logging in CloudWatch."

## Clarifications

### Session 2026-07-10

- Q: What format should each backend log entry use? → A: Structured JSON — one JSON object per line.
- Q: What should the production log level include by default? → A: Warnings + errors (WARN and above); preview defaults to debug.
- Q: At what level are normal successful actions logged? → A: Successful state-changing actions (writes) at INFO; successful reads/searches at DEBUG.
- Q: Should per-request correlation be required, and how? → A: Required (MUST); reuse the AWS Lambda / API Gateway request ID.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Control log verbosity per environment (Priority: P1)

An operator (developer maintaining the app) wants each environment to emit the right amount of logging without changing code: production should stay quiet and record only problems, while preview environments should record detailed step-by-step activity to help debug. The verbosity is governed by a single setting that differs per environment and can be adjusted from the AWS console when a problem needs deeper investigation.

**Why this priority**: This is the control mechanism that makes the whole feature usable and safe. Without per-environment level control, production would either be flooded with noise (cost, signal loss) or previews would be too quiet to debug. It is the foundation every other story builds on.

**Independent Test**: Set the level setting to "errors only" and confirm a successful request produces no informational lines but a failing request produces an error line; then change the setting to the verbose level (no redeploy) and confirm the same successful request now produces informational/debug lines. Fully testable on a single environment by toggling the setting.

**Acceptance Scenarios**:

1. **Given** the production environment configured for warning-level logging (WARN and above), **When** a normal successful action occurs, **Then** no informational or debug entries are written, and only warnings/errors (if any) appear.
2. **Given** a preview environment configured for debug-level logging, **When** a normal successful action occurs, **Then** informational and debug entries describing the action appear in CloudWatch.
3. **Given** a running environment, **When** an operator changes the level setting in the AWS console, **Then** subsequent requests honor the new level without a code change or redeploy.
4. **Given** the level setting is missing or set to an unrecognized value, **When** the backend runs, **Then** it falls back to a safe default (errors only) and continues serving requests normally.

---

### User Story 2 - All main backend actions are logged (Priority: P2)

An operator investigating a reported problem wants to open CloudWatch and find a log entry for whatever the user was doing — signing in or out, adding or editing a recipe, viewing a recipe or ingredient list, or searching — with enough context (who, what, and the outcome) to understand what happened.

**Why this priority**: Level control (P1) is only valuable if the important actions actually emit log entries. This story delivers the coverage that makes logs useful for diagnosis. It depends on P1 for the level machinery but delivers the day-to-day debugging value.

**Independent Test**: Exercise each main action (login, logout, create/edit/delete recipe, view recipe, list ingredients, search, favorite/unfavorite, password recovery) against a preview environment set to the verbose level, then confirm each produced a corresponding log entry identifying the action, the actor, the target, and the outcome.

**Acceptance Scenarios**:

1. **Given** the verbose level, **When** a user signs in successfully, **Then** a log entry records a successful login for that user identifier.
2. **Given** any level, **When** a user's sign-in fails, **Then** an error-level entry records the failed authentication attempt (without revealing the submitted password).
3. **Given** the verbose level, **When** a recipe is created or edited, **Then** a log entry records the action, the acting user, and the affected recipe identifier.
4. **Given** the verbose level, **When** a user views a recipe, lists ingredients, or runs a search, **Then** a log entry records the read/search action and its key parameters and result count.
5. **Given** any level, **When** an action fails due to a validation, authorization, or dependency error, **Then** an error-level entry records the action and the failure reason.

---

### User Story 3 - Logs are consistent, searchable, and safe (Priority: P3)

An operator wants log entries to share a consistent structure with common fields (time, level, action, outcome, and context) so they can filter and group them in CloudWatch, and wants to be confident that no secrets or sensitive credentials ever appear in the logs.

**Why this priority**: Consistency and safety turn raw log output into a dependable tool and prevent the logging system from becoming a security liability. It refines the value delivered by P1/P2 rather than being required for a first usable slice.

**Independent Test**: Trigger a variety of actions and failures, then confirm (a) every entry carries the common fields and can be filtered by level and by action, (b) entries for one request can be grouped by a shared request context, and (c) no entry contains a password, session token, or reset token at any level.

**Acceptance Scenarios**:

1. **Given** logs from multiple actions, **When** an operator filters CloudWatch by level, **Then** only entries at or above that level are returned.
2. **Given** logs from a single request that produced several entries, **When** an operator filters by that request's context, **Then** all of its entries are grouped together.
3. **Given** any action at any level (including debug), **When** its log entries are inspected, **Then** no password, session token, or password-reset token value appears.

---

### Edge Cases

- **Unset or invalid level setting**: The backend must not crash; it falls back to the safe default (errors only) and records that the fallback was applied.
- **Level change under live traffic**: Changing the setting affects subsequent invocations; in-flight requests may complete under the previous level. No requests should fail because of the change.
- **Logging failure**: If writing a log entry fails, the request it accompanies must still succeed — logging is best-effort and never breaks request handling.
- **High-volume verbose logging**: Debug level on a busy preview may produce large volumes; this is acceptable for previews but must never be the production default (cost/noise control).
- **Sensitive inputs**: Actions that receive secrets (login, password reset) must log the event and outcome without ever including the secret value, even at debug level.
- **Unexpected/uncaught errors**: Panics or unhandled failures must produce an error-level entry with enough context to locate the failing action.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The backend MUST emit log entries to the Lambda's standard log destination so they are available in CloudWatch.
- **FR-002**: A single environment-level setting MUST control the minimum severity that is emitted, with standard ordered levels (debug < informational < warning < error) such that selecting a level includes all higher-severity entries.
- **FR-003**: The production environment MUST default to warning-level logging (warnings and errors); preview environments MUST default to debug-level logging (all levels) for debugging.
- **FR-004**: The level setting MUST be adjustable from the AWS console (as a Lambda environment variable) and MUST take effect for subsequent requests without requiring a code change or redeploy.
- **FR-005**: If the level setting is absent or set to an unrecognized value, the backend MUST fall back to a safe default of error-level only and continue operating normally.
- **FR-006**: The backend MUST log every main action it performs, including at minimum: sign-in (success and failure), recipe creation, recipe editing, recipe deletion, recipe retrieval, ingredient-list retrieval, searches (including by ingredient and by base spirit), adding/removing a favorite, and password-recovery request and reset. Sign-out is client-side in the stateless-JWT model and is logged only where the server actually participates in session invalidation (e.g., the token-version bump on password reset); it is otherwise not applicable.
- **FR-007**: Each action log entry MUST record the action name, the outcome (success or failure), the acting user identifier when the action is authenticated, and the affected resource identifier when applicable.
- **FR-008**: Handled failures (validation errors, authorization/authentication failures, and dependency/store errors) MUST be logged at error level in every environment, including production, with the failure reason.
- **FR-009**: Log entries MUST NOT contain secrets or sensitive credentials — specifically passwords, session tokens, and password-reset tokens — at any level; where a user must be identified, a stable non-sensitive identifier (e.g., user ID) MUST be used in preference to raw personal data.
- **FR-010**: Log entries MUST be emitted as structured JSON, one JSON object per line, with common fields (at minimum: timestamp, level, action, message, and contextual key/value details) to support field-based filtering and searching in CloudWatch Logs Insights.
- **FR-011**: Every request-scoped log entry MUST include a request-correlation identifier that reuses the platform-provided request ID (AWS Lambda / API Gateway request ID), plus request context (HTTP method and path), so that all entries produced while handling one request can be grouped. Startup/initialization logs that occur outside of request handling (e.g., the configuration-fallback notice) are exempt from the correlation-id requirement.
- **FR-012**: Logging MUST be best-effort with respect to request correctness: a failure to write a log entry MUST NOT cause the associated request to fail.
- **FR-013**: The same logging mechanism and configuration approach MUST apply to both the production Lambda and the preview Lambdas, each carrying its environment-appropriate level.
- **FR-014**: Read/search actions MUST record the key request parameters and a result summary (e.g., result count) at debug level to aid diagnosis without dumping full payloads.
- **FR-015**: Log severity MUST follow this mapping: successful state-changing actions (sign-in, sign-out where applicable per FR-006, recipe create/edit/delete, favorite add/remove, password-recovery request/reset) at informational level; successful read/search actions at debug level; recoverable anomalies (e.g., rate-limit rejections, configuration fallback applied) at warning level; and handled failures at error level.

### Key Entities *(include if feature involves data)*

- **Log Entry**: A single record of something the backend did or observed. Attributes: timestamp, severity level, action name, outcome, human-readable message, actor identifier (when authenticated), target resource identifier (when applicable), request-correlation context, and additional contextual key/values. Explicitly excludes secrets and sensitive credentials.
- **Log Level Setting**: The per-environment configuration value (a Lambda environment variable) that selects the minimum severity emitted. Has a per-environment default (production = warning; preview = debug) and can be overridden from the AWS console.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For each of the enumerated main actions (FR-006), exercising it produces at least one CloudWatch log entry that identifies the action, the actor (when authenticated), the target (when applicable), and the outcome — 100% coverage of the enumerated actions.
- **SC-002**: In production (warning level), a routine successful request produces zero informational/debug entries, while a failing request produces an error entry — verified by triggering one success and one failure and observing only the failure (warnings surface only when a recoverable anomaly occurs).
- **SC-003**: After an operator changes the level setting in the AWS console, subsequent requests honor the new level within one minute, with no code change or redeploy.
- **SC-004**: Across all levels including debug, no log entry contains a password, session token, or password-reset token — verified by inspection of entries produced by the authentication and password-recovery actions.
- **SC-005**: An operator can locate every log entry belonging to a single failing request in under two minutes by filtering CloudWatch on level and request-correlation context.
- **SC-006**: A missing or invalid level setting never prevents the backend from serving requests — verified by removing/mangling the setting and confirming requests still succeed while logging falls back to error-only.

## Assumptions

- **Log format** (resolved): Entries are structured JSON, one object per line (see FR-010) — chosen for CloudWatch Logs Insights field filtering.
- **Level names**: Standard, ordered severity levels — debug, informational (info), warning (warn), and error.
- **Setting name and defaults** (resolved): A single environment variable (working name `LOG_LEVEL`) controls verbosity; defaults are production = `warn` (warnings + errors) and preview = `debug`. The final variable name is confirmed during planning/infra; the fallback when the setting is missing/invalid is error-only (FR-005).
- **Environments**: "Production" is the live Lambda; "preview" refers to the per-PR preview Lambdas. Both already write to CloudWatch with existing log groups and retention (from prior infrastructure work); this feature reuses them and does not change retention.
- **Scope — backend only**: This feature covers the Go backend Lambda(s). Frontend/browser logging is out of scope.
- **CloudWatch only**: Logs are consumed via CloudWatch (console / Logs Insights). No external log-aggregation service, dashboards, metrics, or alerting are in scope — those may be separate future features.
- **Existing ad-hoc logs**: A few handlers already emit occasional log lines (e.g., password-recovery failures). Those are folded into the consistent mechanism rather than kept as one-offs.

## Out of Scope

- Metrics, dashboards, tracing, and alarms/alerting based on logs.
- Shipping logs to any destination other than CloudWatch.
- Frontend or client-side logging.
- Changing CloudWatch log-group retention or storage configuration.
- Audit-grade tamper-evident logging or long-term compliance archival.
