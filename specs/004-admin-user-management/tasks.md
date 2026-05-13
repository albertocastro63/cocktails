# Tasks: Admin User Management Interface

**Input**: Design documents from `specs/004-admin-user-management/`  
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/api.md ✅, contracts/ui.md ✅

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.  
**TDD**: Constitution Principle II requires tests to be written and confirmed failing before implementation.  
**No new packages**: All required libraries are already installed in both backend and frontend.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1–US4)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Extend the User model, store interface, database schema, JWT, and middleware. All user stories depend on this phase. No user story work can begin until this phase is complete.

**⚠️ CRITICAL**: Complete all foundational tasks before any Phase 3–6 work.

### Tests — Write First, Confirm Failing

- [X] T001 Add failing tests for extended SQLite UserStore methods in `backend/internal/store/sqlite/users_test.go`: "List returns all non-admin users", "List returns empty slice when no non-admin users exist", "Update persists first_name, last_name, email, token_version changes", "Update increments token_version on password change", "Delete removes user by ID", "Delete returns error for unknown ID", "GetByEmail returns user by email", "GetByEmail returns error when email not found", "Create stores first_name, last_name, email fields"
- [X] T002 Add failing tests for JWT TokenVersion in `backend/internal/auth/jwt_test.go`: "Issue embeds token_version in claims", "Parse returns token_version from valid token"
- [X] T003 Add failing tests for RequireAuthWithStore in `backend/internal/handler/middleware_test.go`: "rejects request when user ID not found in store", "rejects request when token_version does not match stored version", "passes request when user exists and token_version matches"

### Implementation

- [X] T004 Extend `backend/internal/model/model.go`: add `FirstName`, `LastName`, `Email string` and `TokenVersion int` fields to the `User` struct with JSON tags `first_name`, `last_name`, `email`, `token_version`; keep `PasswordHash` omitted from JSON
- [X] T005 Extend `backend/internal/store/store.go`: add `List() ([]*model.User, error)`, `Update(user *model.User) error`, `Delete(id string) error`, `GetByEmail(email string) (*model.User, error)` to the `UserStore` interface
- [X] T006 Add SQLite migration steps in `backend/internal/store/sqlite/store.go` `migrate()` function (using the existing idempotent ADD COLUMN pattern): (a) `ALTER TABLE users ADD COLUMN first_name TEXT NOT NULL DEFAULT ''`; (b) `ALTER TABLE users ADD COLUMN last_name TEXT NOT NULL DEFAULT ''`; (c) `ALTER TABLE users ADD COLUMN email TEXT NOT NULL DEFAULT ''`; (d) `ALTER TABLE users ADD COLUMN token_version INTEGER NOT NULL DEFAULT 0`; (e) `CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique ON users(email) WHERE email != ''`; (f) table-rebuild migration to make `recipes.creator_id` nullable with `ON DELETE SET NULL`: first run `db.QueryRow("SELECT notnull FROM pragma_table_info('recipes') WHERE name='creator_id'").Scan(&notnull)` — if `notnull == 0` the column is already nullable, skip; otherwise execute: `CREATE TABLE recipes_new (id TEXT PRIMARY KEY, name TEXT NOT NULL, ingredients TEXT NOT NULL DEFAULT '[]', steps TEXT NOT NULL DEFAULT '[]', properties TEXT NOT NULL DEFAULT '{}', notes TEXT NOT NULL DEFAULT '', creator_id TEXT REFERENCES users(id) ON DELETE SET NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL); INSERT OR IGNORE INTO recipes_new SELECT * FROM recipes; DROP TABLE recipes; ALTER TABLE recipes_new RENAME TO recipes;` — also drop and recreate the FTS virtual table (`recipes_fts`) after the rebuild since it references the old table
- [X] T007 Implement `List`, `Update`, `Delete`, `GetByEmail` in `backend/internal/store/sqlite/users.go`; update `Create` and `scanUser` to include `first_name`, `last_name`, `email`, `token_version` in all SQL queries and row scans
- [X] T008 [P] Implement `List`, `Update`, `Delete`, `GetByEmail` in `backend/internal/store/dynamo/users.go`; update existing methods to store and retrieve the four new fields
- [X] T009 Add `TokenVersion int` to `auth.Claims` with JSON tag `token_version` in `backend/internal/auth/jwt.go`; update `Issue(userID, username string, isAdmin bool, tokenVersion int)` to embed it in the claims
- [X] T010 Add `RequireAuthWithStore(us store.UserStore) func(http.Handler) http.Handler` to `backend/internal/handler/middleware.go`: after parsing the JWT, call `us.GetByID(claims.UserID)` — return 401 if user not found; return 401 if `user.TokenVersion != claims.TokenVersion`; keep existing `RequireAuth` unchanged for backward compatibility; update `backend/cmd/server/main.go` `buildHandler` to replace `handler.RequireAuth` with `handler.RequireAuthWithStore(userStore)` on all existing protected routes; in `backend/internal/handler/auth.go` `Login` method, after verifying the password, call `h.users.GetByID(user.ID)` to fetch the current `TokenVersion` and pass it to `auth.Issue(user.ID, user.Username, user.IsAdmin, user.TokenVersion)`; update `auth.Issue` signature in `backend/internal/auth/jwt.go` to accept `tokenVersion int` as a fourth parameter
- [X] T011 [P] Add `isAdmin() bool` to `frontend/src/api/auth.js`: decode the middle (payload) segment of the stored JWT with `atob`, parse as JSON, return `payload.is_admin === true`; return `false` if no token or decode fails
- [X] T012 [P] Extend `frontend/src/api/client.js`: update `createUser(data, token)` to accept `{ username, password, first_name, last_name, email }` object instead of positional args; add `listUsers(token)` → GET `/api/v1/admin/users`; add `getUser(id, token)` → GET `/api/v1/admin/users/${id}`; add `updateUser(id, data, token)` → PUT `/api/v1/admin/users/${id}`; add `deleteUser(id, token)` → DELETE `/api/v1/admin/users/${id}`

**Checkpoint**: Run `cd backend && go test ./...` — all existing tests pass, T001–T003 tests now pass. Foundation is ready for user story phases.

---

## Phase 3: User Story 1 — User List (Priority: P1) 🎯 MVP

**Goal**: Admin can navigate to `/admin/users` and see a table of all non-admin users.

**Independent Test**: Log in as admin. Navigate to `#/admin/users`. Confirm a table appears listing all non-admin users with columns for username, name, and email. Log in as a non-admin user; confirm access is denied.

### Tests for User Story 1 — Write First, Confirm Failing

- [X] T013 [P] [US1] Add failing tests for `ListUsers` handler in `backend/internal/handler/admin_test.go`: "GET /admin/users returns 200 with array of non-admin users", "GET /admin/users returns empty array when no non-admin users", "GET /admin/users returns 401 when not authenticated", "GET /admin/users returns 403 when not admin"
- [X] T014 [P] [US1] Create `frontend/src/pages/AdminUserList.test.js` with failing tests: "renders loading state initially", "renders table with user rows when users load (shows username, name, email columns)", "renders empty state when no users", "renders error state on API failure", "renders Add User button that navigates to #/admin/users/new"
- [X] T014b [P] [US1] Add failing tests for admin route guard in `frontend/src/main.test.js`: "navigating to #/admin/users when not logged in renders Login component", "navigating to #/admin/users when logged in as non-admin renders access denied message", "Admin nav link is not rendered when isAdmin() returns false", "Admin nav link is rendered when isAdmin() returns true"

### Implementation for User Story 1

- [X] T015 [US1] Add `ListUsers` handler to `backend/internal/handler/admin.go`: call `h.users.List()`, filter is already done in List (non-admin only), return JSON array; update `AdminHandler` struct to include `users store.UserStore` (already has it); register `GET /api/v1/admin/users` with `RequireAuthWithStore` + `RequireAdmin` in `backend/cmd/server/main.go`
- [X] T016 [US1] Create `frontend/src/pages/AdminUserList.js`: `div.max-w-4xl` root; on mount call `listUsers(getToken())` and render loading → table (columns: username, full name, email, actions with Edit `<a>` and Delete button placeholder) or empty state "No users yet." or error "Failed to load users."; Add User button navigates to `#/admin/users/new`
- [X] T017 [US1] Update `frontend/src/main.js`: add route `{ pattern: /^\/admin\/users$/, factory: () => AdminUserList() }`; in `buildNav()` add `<a href="#/admin/users">Admin</a>` link visible only when `isAdmin()` is true; add admin route guard — if path matches `/^\/admin/` and `!isLoggedIn()`, redirect to `#/login`; if logged in but `!isAdmin()`, show "Access denied" paragraph

**Checkpoint**: All T013–T014 tests pass. Navigate to `#/admin/users` as admin to see user list.

---

## Phase 4: User Story 2 — Create User (Priority: P2)

**Goal**: Admin can open a form, fill in username + password (required) and optional name/email, and create a new user who then appears in the list.

**Independent Test**: From `#/admin/users`, click Add User. Fill in username and password only. Submit. Confirm new user appears in the list. Submit again with all fields. Confirm email conflict and username conflict errors appear when relevant.

### Tests for User Story 2 — Write First, Confirm Failing

- [X] T018 [P] [US2] Add failing tests for extended `CreateUser` handler in `backend/internal/handler/admin_test.go`: "POST /admin/users with first_name/last_name/email stores all fields", "POST /admin/users returns 409 EMAIL_CONFLICT when email already in use", "POST /admin/users with no email succeeds when another user also has no email"
- [X] T019 [P] [US2] Create `frontend/src/pages/AdminUserForm.test.js` with failing tests (create mode): "renders username, password, first_name, last_name, email inputs", "submitting without username shows validation error and does not call API", "submitting without password shows validation error and does not call API", "successful create calls createUser with all provided fields and navigates to #/admin/users", "API error 409 USERNAME_CONFLICT is shown in error paragraph", "API error 409 EMAIL_CONFLICT is shown in error paragraph"

### Implementation for User Story 2

- [X] T020 [US2] Update `CreateUser` in `backend/internal/handler/admin.go` to decode and store `first_name`, `last_name`, `email` from request body; validate email format (simple regex or `strings.Contains(email, "@")`) when non-empty; call `h.users.GetByEmail(email)` and return 409 `EMAIL_CONFLICT` if another user already has that email; pass new fields to `model.User` before calling `h.users.Create`
- [X] T021 [US2] Create `frontend/src/pages/AdminUserForm.js` (create mode only for this task): form with `input[name=username]`, `input[name=password, type=password]`, `input[name=first_name]`, `input[name=last_name]`, `input[name=email, type=email]`, error paragraph, submit button "Create User"; client-side validate username and password non-empty before calling `createUser({ username, password, first_name, last_name, email }, getToken())`; on success navigate to `#/admin/users`; on API error show message in error paragraph; add route `{ pattern: /^\/admin\/users\/new$/, factory: () => AdminUserForm({ onSave: () => navigate('#/admin/users') }) }` in `frontend/src/main.js`

**Checkpoint**: All T018–T019 tests pass. Create users through the form and confirm they appear in the list.

---

## Phase 5: User Story 3 — Edit User (Priority: P3)

**Goal**: Admin can click Edit on a user, modify name/email/password, and save changes. Username is read-only. Blank password preserves existing.

**Independent Test**: Click Edit on a user. Confirm the form shows current name and email, and that the username field is read-only. Update first name and email. Save. Confirm list reflects changes. Enter a new password and confirm the user's old session is invalidated.

### Tests for User Story 3 — Write First, Confirm Failing

- [X] T022 [P] [US3] Add failing tests for `GetUser` and `UpdateUser` handlers in `backend/internal/handler/admin_test.go`: "GET /admin/users/{id} returns 200 with user", "GET /admin/users/{id} returns 404 for unknown ID", "GET /admin/users/{id} returns 403 for admin user", "PUT /admin/users/{id} updates first_name, last_name, email", "PUT /admin/users/{id} with non-empty password updates hash and increments token_version", "PUT /admin/users/{id} with blank password preserves existing password and token_version", "PUT /admin/users/{id} returns 409 EMAIL_CONFLICT when email taken by another user", "PUT /admin/users/{id} returns 403 for admin user", "PUT /admin/users/{id} returns 404 for unknown ID"
- [X] T023 [P] [US3] Add failing tests for `AdminUserForm` edit mode in `frontend/src/pages/AdminUserForm.test.js`: "edit mode pre-fills first_name, last_name, email from fetched user", "edit mode renders username as read-only (not an editable input)", "password input has placeholder 'Leave blank to keep existing'", "successful edit calls updateUser with correct payload and navigates", "blank password field in edit mode does not include password in payload"

### Implementation for User Story 3

- [X] T024 [US3] Add `GetUser` and `UpdateUser` handlers to `backend/internal/handler/admin.go`: `GetUser` — extract `{id}` from URL path, call `h.users.GetByID(id)`, return 404 if not found, return 403 if user is admin; `UpdateUser` — decode body `{ first_name, last_name, email, password }`, reject if target user is admin (403), check email uniqueness (skip if email unchanged), update fields, if password non-empty hash it and increment `TokenVersion`, call `h.users.Update`; register `GET /api/v1/admin/users/{id}` and `PUT /api/v1/admin/users/{id}` in `backend/cmd/server/main.go`
- [X] T025 [US3] Add edit mode to `frontend/src/pages/AdminUserForm.js`: when `id` prop is provided, call `getUser(id, getToken())` on mount and pre-fill `first_name`, `last_name`, `email`; render username as a `<p>` (read-only, not an `<input>`); change password placeholder to "Leave blank to keep existing"; on submit omit password from payload if field is empty; call `updateUser(id, payload, getToken())` on submit; change heading to "Edit User" and submit button to "Save Changes"; add routes `{ pattern: /^\/admin\/users\/([^/]+)\/edit$/, factory: (m) => AdminUserForm({ id: m[1], onSave: () => navigate('#/admin/users') }) }` in `frontend/src/main.js`

**Checkpoint**: All T022–T023 tests pass. Edit a user and confirm changes persist. Change password and confirm old session is rejected on next request.

---

## Phase 6: User Story 4 — Delete User (Priority: P4)

**Goal**: Admin can delete a non-admin user after confirming. The user's recipes are orphaned. Their session is immediately invalidated.

**Independent Test**: Click Delete on a user. Confirm a confirmation prompt appears. Confirm deletion. Verify the user no longer appears in the list and cannot log in. Verify their recipes still exist with no creator.

### Tests for User Story 4 — Write First, Confirm Failing

- [X] T026 [P] [US4] Add failing tests for `DeleteUser` handler in `backend/internal/handler/admin_test.go`: "DELETE /admin/users/{id} returns 204 and removes user", "DELETE /admin/users/{id} returns 403 when target is admin", "DELETE /admin/users/{id} returns 404 for unknown ID", "DELETE /admin/users/{id} orphans user's recipes (creator_id becomes NULL)", "RequireAuthWithStore returns 401 after user is deleted (simulates session invalidation)"
- [X] T027 [P] [US4] Add failing tests for delete behaviour in `frontend/src/pages/AdminUserList.test.js`: "clicking Delete shows confirm dialog", "confirming delete calls deleteUser and refreshes list", "cancelling delete dialog does not call deleteUser", "delete failure shows inline error message"

### Implementation for User Story 4

- [X] T028 [US4] Add `DeleteUser` handler to `backend/internal/handler/admin.go`: extract `{id}`, call `h.users.GetByID(id)` — return 404 if not found; return 403 if target is admin; call `h.users.Delete(id)` (SQLite FK ON DELETE SET NULL handles recipe orphaning automatically); return 204; register `DELETE /api/v1/admin/users/{id}` with `RequireAuthWithStore` + `RequireAdmin` in `backend/cmd/server/main.go`
- [X] T029 [US4] Wire delete button behaviour in `frontend/src/pages/AdminUserList.js`: in each row's Delete button click handler, call `confirm("Delete user {username}?")` — if confirmed, call `deleteUser(id, getToken())`, on success re-fetch and re-render list; on failure show inline error text next to the button

**Checkpoint**: All T026–T027 tests pass. Delete a user, confirm they are removed from list, confirm their recipes remain.

---

## Phase 7: Polish & Cross-Cutting Concerns

- [X] T030 [P] Run `cd backend && go test ./...` and confirm all tests pass with zero failures and no regressions
- [X] T031 [P] Run `cd frontend && npm test -- --run` and confirm all 54+ tests pass with zero failures
- [X] T032 [P] Run `cd backend && go build ./...` and confirm zero build errors
- [X] T033 [P] Run `cd frontend && npm run build` and confirm zero build errors
- [X] T034 [P] Run `cd backend && golangci-lint run` (if available) and confirm zero lint warnings

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 2 (Foundational)**: No dependencies — start immediately; BLOCKS all user stories
- **Phase 3 (US1)**: Depends on Phase 2 complete
- **Phase 4 (US2)**: Depends on Phase 2 complete; can run in parallel with Phase 3
- **Phase 5 (US3)**: Depends on Phase 2 complete; T025 depends on T021 (AdminUserForm base)
- **Phase 6 (US4)**: Depends on Phase 2 complete; T029 depends on T016 (AdminUserList base)
- **Phase 7 (Polish)**: Depends on all phases complete

### Within Each Phase

- Test tasks MUST be written and confirmed failing before implementation tasks begin
- Within foundational: T004 → T005 → T006 → T007 (sequential); T008, T009, T010, T011, T012 can run in parallel after T005
- Within US1: T013/T014 [P] (parallel tests) → T015 → T016 → T017
- Within US2: T018/T019 [P] (parallel tests) → T020 → T021
- Within US3: T022/T023 [P] (parallel tests) → T024 → T025
- Within US4: T026/T027 [P] (parallel tests) → T028 → T029

---

## Parallel Execution Examples

### Foundational

```
# After T004, T005:
T008: dynamo/users.go   ← parallel with T007
T009: jwt.go            ← parallel with T007
T011: auth.js           ← parallel with T007
T012: client.js         ← parallel with T007
```

### User Stories (after Phase 2 complete)

```
# Test tasks within each story (parallel):
T013 + T014  (US1 tests)
T018 + T019  (US2 tests)
T022 + T023  (US3 tests)
T026 + T027  (US4 tests)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2 (Foundational) — model, store, migration, JWT, middleware, frontend helpers
2. Complete Phase 3 (US1) — list endpoint + AdminUserList page + admin nav link
3. **STOP and VALIDATE**: Admin can view user list; access control works; empty state works
4. Ship US1 if ready

### Incremental Delivery

1. Phase 2 → Foundation ready (no visible UI change yet)
2. Phase 3 (US1) → Admin can see users → Demo/ship
3. Phase 4 (US2) → Admin can create users → Demo/ship
4. Phase 5 (US3) → Admin can edit users + session invalidation on password reset → Demo/ship
5. Phase 6 (US4) → Admin can delete users + immediate session revocation → Demo/ship
6. Phase 7 → Full suite green, builds clean

---

## Notes

- `RequireAuthWithStore` replaces `RequireAuth` on all protected routes — this is a breaking change to the middleware signature; existing recipe routes must be updated in T010
- The SQLite recipes table rebuild (T006) is the trickiest migration — test it against a copy of the real DB before deploying
- DynamoDB `GetByEmail` requires a GSI on the `email` attribute; the DynamoDB implementation (T008) must include this GSI setup or use a scan (acceptable for small user counts)
- Frontend `isAdmin()` decodes the JWT client-side for UI purposes only — actual admin enforcement is always server-side
- `AdminUserForm` is shared between create (US2) and edit (US3) — implement create mode first (T021), then add edit mode on top (T025)
