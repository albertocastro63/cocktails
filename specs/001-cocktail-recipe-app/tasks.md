# Tasks: Cocktail Recipe App

**Input**: Design documents from `specs/001-cocktail-recipe-app/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/api.md

**Tests**: Included per Constitution Principle II (Test-First, NON-NEGOTIABLE). Write each test task and confirm it fails before starting the corresponding implementation task.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each increment.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no shared dependencies)
- **[Story]**: Which user story this task belongs to (US1–US5)

---

## Phase 1: Setup

**Purpose**: Initialize project structure, tooling, and configuration.

- [X] T001 Create backend/ and frontend/ directory structure per plan.md Project Structure section
- [X] T002 Initialize Go module in backend/ with dependencies: `golang-jwt/jwt/v5`, `modernc.org/sqlite`, `aws/aws-sdk-go-v2/service/dynamodb`, `awslabs/aws-lambda-go-api-proxy/httpadapter`, `aws/aws-lambda-go/lambda`
- [X] T003 [P] Initialize Vite project in frontend/ and install dependencies: `tailwindcss`, `chart.js`, `vitest`, `@vitest/coverage-v8`, `jsdom`, `@testing-library/dom`
- [X] T004 [P] Configure TailwindCSS in frontend/tailwind.config.js and frontend/src/index.css
- [X] T005 [P] Configure golangci-lint in backend/.golangci.yml (enable `errcheck`, `govet`, `staticcheck`, `gocyclo` max 10)
- [X] T006 [P] Configure Vitest with jsdom environment and coverage thresholds (≥80%) in frontend/vitest.config.js

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before any user story work begins.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T007 Define all domain types (Recipe, Ingredient, User) in backend/internal/model/model.go with JSON tags and all fields from data-model.md
- [X] T008 [P] Define RecipeStore and UserStore interfaces in backend/internal/store/store.go with all CRUD and search method signatures
- [X] T009 [P] Implement JWT issue, parse, and claims helpers in backend/internal/auth/jwt.go; secret loaded from `JWT_SECRET` env var
- [X] T010 Write failing tests for SQLite RecipeStore (Create, GetByID, List, Update, Delete, Search) in backend/internal/store/sqlite/recipes_test.go
- [X] T011 Implement SQLite schema (users table, recipes table, recipes_fts FTS5 virtual table with triggers) and full RecipeStore in backend/internal/store/sqlite/recipes.go (depends on T007, T008, T010)
- [X] T012 Write failing tests for SQLite UserStore (Create, GetByUsername, GetByID) in backend/internal/store/sqlite/users_test.go
- [X] T013 Implement SQLite UserStore with bcrypt password hashing in backend/internal/store/sqlite/users.go (depends on T007, T008, T012)
- [X] T014 Write failing tests for JWT auth middleware (valid token, missing token, expired token, admin-only gate) in backend/internal/handler/middleware_test.go
- [X] T015 Implement JWT auth middleware and admin-only middleware in backend/internal/handler/middleware.go (depends on T009, T014)
- [X] T016 [P] Implement shared error response helpers and JSON write helpers in backend/internal/handler/helpers.go
- [X] T017 Implement admin bootstrap: on first run, if `ADMIN_BOOTSTRAP_PASSWORD` env var is set and no users exist, create initial `admin` user with `is_admin=true` in backend/cmd/server/main.go
- [X] T018 Wire local HTTP server entry point (router registration, store factory, env config) in backend/cmd/server/main.go (depends on T011, T013, T015, T016, T017)
- [X] T019 Wire AWS Lambda entry point using httpadapter wrapping the same router in backend/cmd/lambda/main.go (depends on T018)
- [X] T020 [P] Create frontend API client with base fetch wrapper, base URL from `VITE_API_BASE_URL` env var, and JSON/error handling in frontend/src/api/client.js
- [X] T021 [P] Configure Vite dev server proxy (`/api` → `http://localhost:8080`) and env variable defaults in frontend/vite.config.js

**Checkpoint**: Backend server starts locally, serves routes (even if handlers return 501), and passes linting. Frontend dev server starts. All store tests pass.

---

## Phase 3: User Story 1 — Browse & Discover Recipes (Priority: P1) 🎯 MVP

**Goal**: Homepage shows a randomly selected recipe on every load; users can navigate to a full recipe list.

**Independent Test**: Start the app with at least one recipe seeded in the DB. Load the homepage and verify a recipe is displayed. Refresh and confirm the displayed recipe may vary. Navigate to the recipe list and confirm all recipes appear.

### Tests for User Story 1

> **Write these tests FIRST, confirm they fail before any implementation.**

- [X] T022 Write failing tests for `GET /api/v1/recipes` handler (returns list, 200, empty array when no recipes) in backend/internal/handler/recipes_test.go
- [X] T023 Write failing test for `GET /api/v1/recipes/random` handler (returns one recipe, 204 when DB empty) in backend/internal/handler/recipes_test.go
- [X] T024 [P] Write failing test for EmptyState component (renders message prop) in frontend/src/components/EmptyState.test.js
- [X] T025 [P] Write failing test for RecipeCard component (renders name, ingredient count) in frontend/src/components/RecipeCard.test.js
- [X] T026 Write failing test for Home page (renders RecipeCard on data, renders EmptyState on 204) in frontend/src/pages/Home.test.js
- [X] T027 Write failing test for RecipeList page (renders list of RecipeCards) in frontend/src/pages/RecipeList.test.js

### Implementation for User Story 1

- [X] T028 Implement `GET /api/v1/recipes` handler (list all, ordered by `created_at` desc, paginated) in backend/internal/handler/recipes.go (depends on T022)
- [X] T029 Implement `GET /api/v1/recipes/random` handler (random selection, 204 when no recipes, satisfies FR-001 and FR-011) in backend/internal/handler/recipes.go (depends on T023)
- [X] T030 [P] Implement EmptyState component in frontend/src/components/EmptyState.js (depends on T024)
- [X] T031 [P] Implement RecipeCard component (name, ingredient count, link placeholder) in frontend/src/components/RecipeCard.js (depends on T025)
- [X] T032 Implement Home page: fetches `/api/v1/recipes/random`, renders RecipeCard, shows EmptyState on 204, loading state during fetch in frontend/src/pages/Home.js (depends on T026, T030, T031)
- [X] T033 Implement RecipeList page: fetches `/api/v1/recipes`, renders RecipeCard list, shows EmptyState when empty, loading state during fetch in frontend/src/pages/RecipeList.js (depends on T027, T030, T031)
- [X] T034 Set up client-side router (hash or history) and app shell with navigation links (Home, All Recipes) in frontend/src/main.js (depends on T032, T033)
- [X] T035 Add recipe list and random recipe API call functions in frontend/src/api/client.js (depends on T020)

**Checkpoint**: Homepage loads, shows a recipe card. Refreshing may show a different recipe. Navigating to /recipes shows all recipes in a list. Empty DB shows a friendly message on both pages.

---

## Phase 4: User Story 2 — Search Recipes (Priority: P2)

**Goal**: Users can search across all recipe fields (name, ingredients, steps, properties) from the recipe list page.

**Independent Test**: Seed recipes with varied ingredients and properties. Type "lime" in the search bar and confirm only recipes containing "lime" anywhere appear. Search for a word in a step and confirm matching recipes are returned. Clear search to see all recipes.

### Tests for User Story 2

> **Write these tests FIRST, confirm they fail before any implementation.**

- [X] T036 Write failing tests for `GET /api/v1/recipes?q=` handler (search returns matching recipes only, empty string returns all) in backend/internal/handler/recipes_test.go
- [X] T037 Write failing tests for SQLite FTS5 search method in backend/internal/store/sqlite/recipes_test.go (search by ingredient name, by step text, by property value)
- [X] T038 [P] Write failing test for SearchBar component (renders input, calls onSearch prop on change) in frontend/src/components/SearchBar.test.js
- [X] T039 Write failing test for RecipeList search behavior (search triggers API call with q param) in frontend/src/pages/RecipeList.test.js

### Implementation for User Story 2

- [X] T040 Implement SQLite FTS5 search method in RecipeStore: rebuild `search_text` from name + ingredients + steps + property values on each write; query FTS5 on search in backend/internal/store/sqlite/recipes.go (depends on T037)
- [X] T041 Update `GET /api/v1/recipes` handler to pass `q` query param to store search method in backend/internal/handler/recipes.go (depends on T036, T040)
- [X] T042 [P] Implement SearchBar component with input, debounce (300ms), and clear button in frontend/src/components/SearchBar.js (depends on T038)
- [X] T043 Update RecipeList page to render SearchBar and pass `q` to API call; show "no results" EmptyState when search returns empty array in frontend/src/pages/RecipeList.js (depends on T039, T042)

**Checkpoint**: Search bar on recipe list page filters results in real time. Searching by ingredient, base spirit, style, garnish, or step text all return correct results. Empty search restores full list.

---

## Phase 5: User Story 3 — View Recipe Details (Priority: P3)

**Goal**: Clicking a recipe shows the full detail view: all ingredients with quantities, all ordered preparation steps, and all flexible properties.

**Independent Test**: Open any recipe detail page. Confirm all ingredients (name, quantity, unit), all steps, and all stored properties are displayed. A recipe with custom properties (e.g., `occasion: Brunch`) should show that field even if no other recipe has it.

### Tests for User Story 3

> **Write these tests FIRST, confirm they fail before any implementation.**

- [X] T044 Write failing test for `GET /api/v1/recipes/{id}` handler (returns full recipe, 404 for unknown ID) in backend/internal/handler/recipes_test.go
- [X] T045 [P] Write failing test for IngredientList component (renders each ingredient with quantity and unit) in frontend/src/components/IngredientList.test.js
- [X] T046 [P] Write failing test for PropertyTable component (renders all arbitrary key-value pairs) in frontend/src/components/PropertyTable.test.js
- [X] T047 Write failing test for RecipeDetail page (renders full recipe data, all properties visible) in frontend/src/pages/RecipeDetail.test.js

### Implementation for User Story 3

- [X] T048 Implement `GET /api/v1/recipes/{id}` handler in backend/internal/handler/recipes.go (depends on T044)
- [X] T049 [P] Implement IngredientList component (renders ingredient rows with quantity + unit) in frontend/src/components/IngredientList.js (depends on T045)
- [X] T050 [P] Implement PropertyTable component (renders all key-value property pairs regardless of key name, satisfies FR-005) in frontend/src/components/PropertyTable.js (depends on T046)
- [X] T051 Implement RecipeDetail page: fetches `/api/v1/recipes/{id}`, renders IngredientList, steps list, PropertyTable, loading and error states in frontend/src/pages/RecipeDetail.js (depends on T047, T049, T050)
- [X] T052 Update RecipeCard component to link to `/recipes/{id}` detail page in frontend/src/components/RecipeCard.js
- [X] T053 Add recipe detail route to router and recipe detail API call in frontend/src/main.js and frontend/src/api/client.js (depends on T051)

**Checkpoint**: Clicking any recipe card opens a detail page showing complete ingredients, steps, and all stored properties. Unknown recipe IDs show an error state.

---

## Phase 6: User Story 4 — Add and Edit Recipes (Priority: P4)

**Goal**: Authenticated users can log in, create new recipes with flexible properties, edit existing recipes, and delete only their own recipes.

**Independent Test**: Log in with a valid account. Create a recipe with name, two ingredients, two steps, and a custom property. Verify it appears in the list and detail view. Edit the recipe to add a new property. Verify the change persists. Attempt to delete a recipe created by a different user and confirm it is rejected.

### Tests for User Story 4

> **Write these tests FIRST, confirm they fail before any implementation.**

- [X] T054 Write failing tests for `POST /api/v1/auth/login` handler (valid credentials returns JWT, invalid returns 401) in backend/internal/handler/auth_test.go
- [X] T055 Write failing tests for `POST /api/v1/admin/users` handler (admin creates user, non-admin gets 403, duplicate username gets 409) in backend/internal/handler/admin_test.go
- [X] T056 Write failing tests for `POST /api/v1/recipes` handler (creates recipe, sets creator_id from JWT, warns on duplicate name) in backend/internal/handler/recipes_test.go
- [X] T057 Write failing tests for `PUT /api/v1/recipes/{id}` handler (updates recipe, 401 if no JWT, 403 if authenticated user is not the creator, partial update preserves unset fields) in backend/internal/handler/recipes_test.go
- [X] T058 Write failing tests for `DELETE /api/v1/recipes/{id}` handler (deletes recipe, 403 if not creator, 404 if not found) in backend/internal/handler/recipes_test.go
- [X] T059 [P] Write failing test for Login page (renders form, stores JWT on success, shows error on failure) in frontend/src/pages/Login.test.js
- [X] T060 [P] Write failing test for RecipeForm page (renders fields for name, ingredients, steps, properties; submits to correct endpoint) in frontend/src/pages/RecipeForm.test.js

### Implementation for User Story 4

- [X] T061 Implement `POST /api/v1/auth/login` handler in backend/internal/handler/auth.go (depends on T054)
- [X] T062 Implement `POST /api/v1/admin/users` handler with admin middleware gate in backend/internal/handler/admin.go (depends on T055)
- [X] T063 Implement `POST /api/v1/recipes` handler (auth required, creator_id from JWT claim, duplicate name warning per FR-017, immediate publish per FR-018) in backend/internal/handler/recipes.go (depends on T056)
- [X] T064 Implement `PUT /api/v1/recipes/{id}` handler (auth required, creator-only per FR-014, partial update) in backend/internal/handler/recipes.go (depends on T057)
- [X] T065 Implement `DELETE /api/v1/recipes/{id}` handler (auth required, creator-only per FR-014) in backend/internal/handler/recipes.go (depends on T058)
- [X] T066 [P] Implement Login page (form, POST to `/api/v1/auth/login`, store JWT in memory/sessionStorage, redirect to home on success) in frontend/src/pages/Login.js (depends on T059)
- [X] T067 Implement RecipeForm page (create/edit form with: name field, dynamic ingredient row add/remove, ordered steps add/remove, dynamic property key-value add/remove, submits POST or PUT) in frontend/src/pages/RecipeForm.js (depends on T060, T066)
- [X] T068 Add auth-gated routes (login page, create recipe, edit recipe) to router; redirect to login when JWT missing in frontend/src/main.js (depends on T066, T067)
- [X] T069 Add create, update, and delete recipe API call functions; attach `Authorization: Bearer` header from stored JWT in frontend/src/api/client.js
- [X] T070 Add "Create Recipe" button (visible when logged in) to RecipeList page and "Edit / Delete" controls to RecipeDetail page (visible to creator only) in frontend/src/pages/RecipeList.js and frontend/src/pages/RecipeDetail.js

**Checkpoint**: Full recipe management works end-to-end. Login, create, edit, delete all function correctly. Creator-only enforcement prevents unauthorized deletions. Duplicate name shows a warning without blocking save.

---

## Phase 7: User Story 5 — External Service Data Access (Priority: P5)

**Goal**: The public read API is accessible without authentication and returns structured data matching the contract in contracts/api.md.

**Independent Test**: Use curl or a script to request `GET /api/v1/recipes`, `GET /api/v1/recipes/random`, and `GET /api/v1/recipes/{id}` without any Authorization header. Verify all responses are well-formed JSON matching the contract shape, all flexible properties are included, and search via `?q=` works identically to the UI.

### Tests for User Story 5

> **Write these tests FIRST, confirm they fail before any implementation.**

- [X] T071 Write integration tests verifying all public endpoints (`GET /recipes`, `GET /recipes/random`, `GET /recipes/{id}`, `GET /recipes?q=`) return correct JSON structure without auth token in backend/internal/handler/integration_test.go
- [X] T072 Write integration test verifying flexible properties are fully included in API responses (no field omission) in backend/internal/handler/integration_test.go

### Implementation for User Story 5

- [X] T073 Add development CORS middleware (allow all origins, applicable only when `ENV=development`) in backend/internal/handler/middleware.go so browser-based external consumers work locally
- [X] T074 Add curl usage examples for all public endpoints to specs/001-cocktail-recipe-app/contracts/api.md

**Checkpoint**: External services can query the full recipe dataset and search results via the public API without credentials. All flexible properties appear in responses.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: DynamoDB implementation, loading/error states, code quality, and deployment validation.

- [X] T075 [P] Implement DynamoDB RecipeStore (all RecipeStore interface methods using AWS SDK v2; search via Scan + FilterExpression) in backend/internal/store/dynamo/recipes.go
- [X] T076 [P] Implement DynamoDB UserStore (Create, GetByUsername, GetByID using GSI) in backend/internal/store/dynamo/users.go
- [X] T077 Write tests for DynamoDB RecipeStore and UserStore using DynamoDB local (docker or local endpoint) in backend/internal/store/dynamo/recipes_test.go and backend/internal/store/dynamo/users_test.go
- [X] T078 [P] Add store factory to backend/cmd/server/main.go: select SQLite or DynamoDB based on `STORE_BACKEND` env var (`sqlite` default, `dynamodb` for AWS)
- [X] T079 [P] Audit all frontend pages for consistent loading states (spinner) and error states (user-friendly message, no stack traces) in frontend/src/pages/
- [X] T080 Run `golangci-lint run ./...` and `go vet ./...` in backend/ and resolve all warnings
- [X] T081 Run `npm run test -- --coverage` in frontend/ and confirm ≥80% coverage on all components and pages
- [X] T082 Validate local quickstart end-to-end: fresh clone, follow specs/001-cocktail-recipe-app/quickstart.md, confirm all 5 user stories work
- [X] T083 Write Go benchmark for `GET /api/v1/recipes?q=` handler targeting p95 ≤ 200ms under 50 concurrent requests in backend/internal/handler/benchmarks_test.go
- [X] T084 [P] Add Lighthouse CI script targeting homepage TTI ≤ 3s; add `npm run bench` script in frontend/package.json that runs Lighthouse in headless mode against the built app
- [X] T085 Run `go test -coverprofile=coverage.out ./internal/... && go tool cover -func=coverage.out` in backend/ and confirm coverage ≥ 80% for all packages under internal/; fail if any package is below threshold

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 completion — **blocks all user stories**
- **Phases 3–7 (User Stories)**: All depend on Phase 2 completion; may proceed in priority order
- **Phase 8 (Polish)**: Depends on all desired user stories complete

### User Story Dependencies

- **US1 (P1)**: Requires Phase 2 complete. No dependency on other stories.
- **US2 (P2)**: Requires Phase 2 complete. Extends US1 (adds search to RecipeList page).
- **US3 (P3)**: Requires Phase 2 complete. Extends US1 (adds detail link from RecipeCard).
- **US4 (P4)**: Requires Phase 2 complete. Login/write endpoints independent; UI builds on US1/US3 pages.
- **US5 (P5)**: Requires US1–US4 complete. Integration tests validate the full public API surface.

### Within Each User Story

1. Write failing tests first — confirm they fail
2. Implement to make tests pass
3. Models/store methods before handlers
4. Handlers before frontend pages
5. Components before pages that compose them

### Parallel Opportunities

- T003, T004, T005, T006 (Phase 1 frontend config) can run in parallel
- T008, T009 (Phase 2 interfaces + JWT) can start in parallel with T010 (store tests)
- T024, T025 (US1 component tests) can run in parallel
- T030, T031 (US1 component implementations) can run in parallel after their tests
- T045, T046 (US3 component tests) can run in parallel
- T049, T050 (US3 component implementations) can run in parallel
- T075, T076 (Phase 8 DynamoDB stores) can run in parallel

---

## Parallel Example: User Story 1

```bash
# Run in parallel — component tests:
Task T024: "Write failing test for EmptyState in frontend/src/components/EmptyState.test.js"
Task T025: "Write failing test for RecipeCard in frontend/src/components/RecipeCard.test.js"

# Run in parallel after tests pass — component implementations:
Task T030: "Implement EmptyState component in frontend/src/components/EmptyState.js"
Task T031: "Implement RecipeCard component in frontend/src/components/RecipeCard.js"

# Run sequentially — pages depend on components:
Task T032: "Implement Home page in frontend/src/pages/Home.js"
Task T033: "Implement RecipeList page in frontend/src/pages/RecipeList.js"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks everything)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Load homepage, browse recipe list, confirm empty states work
5. Deploy locally and demo if ready

### Incremental Delivery

1. Setup + Foundational → server starts, stores work
2. US1 → homepage + recipe list (MVP — deployable)
3. US2 → search (immediately useful)
4. US3 → detail view (completes the read experience)
5. US4 → create/edit/delete (full management)
6. US5 → external API validated
7. Polish → DynamoDB + deployment readiness

---

## Notes

- `[P]` tasks operate on different files with no shared dependencies — safe to run in parallel
- `[Story]` label maps each task to a specific user story for traceability
- Each user story is independently completable and testable
- Constitution Principle II is non-negotiable: all tests must be written and confirmed failing before implementation begins
- Commit after each completed task or logical group
- Stop at any phase checkpoint to validate the story independently before advancing
