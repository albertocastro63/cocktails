<!--
  SYNC IMPACT REPORT
  Version change: (unversioned template) → 1.0.0
  Ratification date: 2026-05-07

  Principles added (initial ratification — 4 principles):
  - I. Code Quality (NON-NEGOTIABLE)
  - II. Test-First Development (NON-NEGOTIABLE)
  - III. User Experience Consistency
  - IV. Performance Requirements

  Sections added:
  - Quality Gates & Review Standards
  - Development Workflow
  - Governance

  Template updates:
  - .specify/templates/plan-template.md — Constitution Check gates filled with principle names ✅
  - .specify/templates/spec-template.md — Aligned with principles; no structural changes required ✅
  - .specify/templates/tasks-template.md — Test-First and performance tasks reflected in structure ✅

  Deferred items: None
-->

# Cocktails Constitution

## Core Principles

### I. Code Quality (NON-NEGOTIABLE)

All code MUST be clean, readable, and maintainable at every commit. Non-negotiable rules:

- Functions MUST have a single, clear responsibility; any function exceeding 40 lines MUST be refactored.
- Duplication MUST be eliminated through abstraction when identical logic appears in three or more places.
- All public interfaces MUST be self-documenting through naming; comments MUST explain WHY, never WHAT.
- Cyclomatic complexity MUST NOT exceed 10 per function.
- Dead code, commented-out blocks, and unused imports MUST NOT be committed.
- Linting and formatting tools MUST pass with zero warnings before any code is merged.

### II. Test-First Development (NON-NEGOTIABLE)

Tests MUST be written and confirmed failing before implementation begins. Non-negotiable rules:

- The TDD cycle is mandatory: write failing test → implement minimal code → refactor.
- Unit test coverage MUST be ≥ 80% for all business logic modules.
- Integration tests MUST cover all API contracts and inter-service boundaries.
- Tests MUST be fully deterministic — flaky or order-dependent tests are treated as bugs.
- Test names MUST describe the scenario in plain language; Given/When/Then format is preferred.
- No user story is considered complete until all its acceptance scenarios pass in CI.

### III. User Experience Consistency

The application MUST deliver a predictable, coherent experience across all surfaces. Non-negotiable rules:

- All user-facing interactions MUST follow a consistent visual and interaction language
  (design tokens, component library, or a documented pattern guide).
- Error messages MUST be human-readable and actionable; stack traces and internal codes
  MUST NEVER be exposed to end users.
- Loading states, empty states, and error states MUST be handled for every data-driven UI element.
- Navigation patterns MUST be consistent — equivalent actions MUST use equivalent affordances.
- All interactive elements MUST meet WCAG 2.1 AA accessibility standards.

### IV. Performance Requirements

The application MUST stay within defined performance budgets at every release. Non-negotiable baselines:

- API response time: p95 MUST be ≤ 200 ms for read operations and ≤ 500 ms for write operations.
- Page/screen initial load: Time-to-Interactive MUST be ≤ 3 seconds on a median mobile connection.
- Database queries MUST be reviewed for N+1 patterns before merging; execution plans MUST confirm
  index usage for queries over large tables.
- Memory and CPU usage MUST NOT degrade more than 10% per release; regressions MUST be
  investigated and resolved before shipping.
- Performance benchmarks MUST run as part of CI for all critical user paths.

## Quality Gates & Review Standards

All pull requests MUST pass every gate below before merge:

- Automated linting and formatting checks pass with zero warnings.
- All tests pass (unit, integration, contract) in CI.
- Code coverage does not decrease below the established threshold (≥ 80% business logic).
- No unresolved review comments on the PR.
- Performance benchmarks show no regression on critical paths.
- Constitution Check in plan.md is completed and flagged violations are documented.

## Development Workflow

Features MUST follow the Specify workflow in order:

1. **Specification** (`/speckit-specify`) — define user stories and acceptance criteria.
2. **Clarification** (`/speckit-clarify`) — resolve ambiguities before planning begins.
3. **Planning** (`/speckit-plan`) — produce research, data model, and design artifacts.
4. **Tasks** (`/speckit-tasks`) — generate dependency-ordered implementation tasks.
5. **Implementation** (`/speckit-implement`) — execute tasks with per-task commits.
6. **Analysis** (`/speckit-analyze`) — validate consistency and quality before opening a PR.

Each phase MUST be committed before advancing to the next. Feature branches MUST follow the
naming convention `###-feature-name` (sequential number prefix, hyphenated description).

## Governance

This constitution supersedes all other project conventions and informal agreements. Any practice
not addressed here defers to community best practices for the technology in use.

**Amendment procedure**:

1. Document the rationale — what changes and why.
2. Assess impact on existing code, tests, and downstream templates.
3. Obtain team consensus (or project lead approval for solo projects).
4. Bump the version according to the versioning policy below.
5. Propagate changes to all dependent templates and artifacts; update the Sync Impact Report.

**Versioning policy**:

- MAJOR: removal or backward-incompatible redefinition of an existing principle.
- MINOR: new principle or section added, or materially expanded guidance.
- PATCH: clarifications, wording fixes, or non-semantic refinements.

**Compliance**: All PRs and reviews MUST verify alignment with this constitution. Principle
violations MUST be corrected or formally justified in the Complexity Tracking section of the
relevant plan.md before merge.

**Version**: 1.0.0 | **Ratified**: 2026-05-07 | **Last Amended**: 2026-05-07
