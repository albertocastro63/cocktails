# Feature Specification: Local DynamoDB Emulator (replace SQLite)

**Feature Branch**: `029-local-dynamodb-emulator`  
**Created**: 2026-07-14  
**Status**: Draft  
**Input**: User description: "Local DynamoDB instead of SQLite: For local development and testing I would like to change to a DynamoDB emulator container. This means to remove the SQLite database and all its dependencies for local development. While the DynamoDB emulator is not the same as DynamoDB in AWS, it resembles more closely the behavior of the production code than SQLite."

## Clarifications

### Session 2026-07-14

- Q: The `username-index` GSI on the users table is required by login but is not declared in the production Terraform — reconcile it here, or keep this feature strictly local/test? → A: Reconcile it — add the missing `username-index` GSI to the production Terraform (`infra/main.tf` users table) as part of this feature (see FR-012).
- Q: On a fresh local emulator, seed sample content or start empty? → A: Empty tables only, plus the existing optional admin bootstrap; no automatic recipe seeding.

## User Scenarios & Testing *(mandatory)*

The "users" of this feature are the project's developers and contributors. The value is **fidelity**: local development and automated tests should exercise the same storage model that runs in production, so behavior differences and storage-specific bugs surface before deployment instead of after.

### User Story 1 - Run the app locally against a production-like store (Priority: P1)

As a developer, I can start the local backend and have it read and write through a locally running data store that behaves like the production database, so what I see locally reflects how the app behaves in production.

**Why this priority**: This is the core of the request and the largest source of value. Today the local store and the production store are different technologies, so a feature can pass locally yet misbehave in production (or vice versa). Aligning the local store removes that entire class of "works on my machine" surprises.

**Independent Test**: Start the local environment, exercise the app end-to-end (create/list/update/delete recipes, log in, favorite a recipe, admin user management), and confirm every feature works with data served entirely by the local emulator — no SQLite involved.

**Acceptance Scenarios**:

1. **Given** a clean checkout and the documented local setup steps, **When** a developer starts the local environment, **Then** the backend serves all features using the local emulator as its only data store.
2. **Given** the local environment is running, **When** the developer creates data (e.g., a recipe) and then reads it back, **Then** the data persists and is returned for the lifetime of the running emulator.
3. **Given** the emulator is not running, **When** the developer starts the backend, **Then** it fails fast with a clear message explaining that the local data store must be started first.

---

### User Story 2 - Automated tests run against the emulator (Priority: P2)

As a developer, I can run the automated test suite against the local emulator so tests validate the same storage behavior used in production, and CI and local test runs use the same approach.

**Why this priority**: Tests are the safety net. Running them against the emulator catches storage-specific issues (query/filter semantics, key/index behavior, conditional writes) that a different local database would hide. CI already runs an emulator, so this also unifies local and CI test environments.

**Independent Test**: Run the full test suite locally with the emulator available and confirm it passes; run it with the emulator unavailable and confirm it fails with a clear, actionable message rather than silently falling back to another store.

**Acceptance Scenarios**:

1. **Given** the emulator is available, **When** the developer runs the full test suite, **Then** all tests execute against the emulator and pass.
2. **Given** the behaviors previously verified by the SQLite-backed tests, **When** the suite runs, **Then** those same behaviors are still verified (no loss of coverage).
3. **Given** two tests run in sequence, **When** they use the emulator, **Then** neither test's data leaks into the other (isolated, repeatable runs).

---

### User Story 3 - SQLite fully removed (Priority: P3)

As a maintainer, I want the SQLite store, its tests, and its library dependencies removed so the project has a single storage model to reason about and a lighter build.

**Why this priority**: Depends on Stories 1 and 2 being in place first (the emulator must fully cover dev and test needs before SQLite can go). Once done, it eliminates duplicated store logic, removes a dependency, and prevents anyone from accidentally relying on the divergent local store again.

**Independent Test**: Search the codebase and dependency manifest for any SQLite reference and confirm none remain; build and run the app and tests successfully without the SQLite library present.

**Acceptance Scenarios**:

1. **Given** the change is complete, **When** the codebase and dependency manifest are searched for SQLite, **Then** no references remain.
2. **Given** SQLite is removed, **When** the app is built and started locally, **Then** it runs successfully against the emulator with no missing-dependency or configuration errors.
3. **Given** a developer selects a storage backend, **When** they start the app, **Then** the only non-production option is the emulator (there is no SQLite option to choose).

---

### Edge Cases

- **Emulator not started**: The backend and the test suite must fail fast with a clear, actionable message ("start the local data store") rather than hanging, crashing obscurely, or silently using a different store.
- **Ephemeral data**: When the emulator container is stopped/removed, its data is gone. This is expected for local dev; the workflow must not assume long-term persistence.
- **First-run / empty store**: On a fresh emulator with no tables, the required tables and access paths are provisioned automatically so the app works without manual setup.
- **Port already in use**: If the emulator's local port is occupied, the developer gets a clear failure and guidance rather than a confusing runtime error.
- **Container runtime absent**: If no container runtime is installed, local dev/testing cannot proceed; this prerequisite is documented up front.
- **Local vs CI drift**: The local emulator approach must match what CI already uses so results are consistent across both.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The local development environment MUST provide a DynamoDB-compatible data store that runs as a local container and behaves like the production database.
- **FR-002**: When run locally, the backend MUST use the emulator as its data store through the same storage code path used in production (differing only in endpoint/credentials, not in store implementation).
- **FR-003**: The required tables and access paths (including any secondary lookup used by existing features, e.g., finding a user by username) MUST be provisioned automatically in the emulator so every existing feature works locally without manual database setup.
- **FR-004**: Developers MUST be able to start the complete local environment (emulator plus backend) by following a single, documented set of steps.
- **FR-005**: The automated test suite MUST run against the emulator and MUST preserve the behaviors currently verified by the SQLite-backed tests (no reduction in behavioral coverage).
- **FR-006**: Test runs MUST be isolated and repeatable — no shared state leaks between tests, and re-running the suite yields the same result.
- **FR-007**: The SQLite store implementation, its dedicated tests, and its library dependencies MUST be removed from the codebase and dependency manifest.
- **FR-008**: The application MUST NOT offer or default to a SQLite backend; the only non-production storage option is the emulator.
- **FR-009**: When the emulator is unreachable, the backend and the tests MUST fail with a clear, actionable message and MUST NOT silently fall back to any other store.
- **FR-010**: The local emulator approach MUST be consistent with the environment CI already uses for tests, so local and CI runs behave the same way.
- **FR-011**: Contributor/local-setup documentation MUST be updated to describe the emulator-based workflow and remove references to SQLite.
- **FR-012**: The production infrastructure definition MUST be reconciled to declare the `username-index` secondary lookup on the users table that the login path requires, so the production schema definition matches the schema the code depends on and the local provisioning creates.

### Key Entities *(include if feature involves data)*

- **Local data store**: The emulator-backed equivalent of the production storage. It holds the same logical collections the app already uses — recipes, users, and favorites — including any secondary access paths those features depend on. It is provisioned automatically and is ephemeral for local use.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: From a clean checkout, a developer can reach a running local app backed by the emulator in under 10 minutes by following the documented steps.
- **SC-002**: 100% of application features that work in production also work in the local emulator environment; none require SQLite-specific behavior.
- **SC-003**: The full automated test suite passes against the emulator with no reliance on SQLite.
- **SC-004**: Zero SQLite references remain in the codebase or the dependency manifest (0 occurrences).
- **SC-005**: Every behavior previously verified by a SQLite-backed test remains verified by a test that runs against the emulator.
- **SC-006**: Local development and CI use the same emulator approach — a single storage model across all non-production environments.
- **SC-007**: The build's dependency footprint is reduced: the SQLite library and its transitive dependencies no longer appear.

## Assumptions

- Developers have, or can install, a local container runtime; this becomes a documented prerequisite for local development and testing.
- The emulator used locally is the same DynamoDB-compatible container already used in the CI test environment, so local mirrors CI.
- Local emulator data is ephemeral by default (reset when the container is removed); persistence across sessions is not a requirement for local development.
- Production storage, behavior, and credentials are unchanged **except** for one reconciliation: declaring the already-required `username-index` secondary lookup on the users table in the infrastructure definition (FR-012). Production already uses the real managed database.
- Behaviors currently exercised only by the SQLite store tests will be ported to run against the emulator rather than dropped.
- This is a developer-experience/infrastructure change with no new end-user-facing functionality.

## Out of Scope

- Any change to the production storage service or its configuration, **except** declaring the missing `username-index` secondary lookup on the users table in the infrastructure definition (FR-012) — no data migration or table replacement, just bringing the definition in line with the running schema.
- Migrating existing local SQLite database files; local data is disposable and is not carried over.
- Seeding sample content (e.g., recipes) into the local emulator; local tables start empty apart from the existing optional admin bootstrap.
- Guaranteeing byte-for-byte parity between the emulator and the managed cloud database — the goal is materially closer fidelity than SQLite, not perfect equivalence.
