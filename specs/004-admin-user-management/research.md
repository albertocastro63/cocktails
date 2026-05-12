# Research: Admin User Management

**Feature**: 004-admin-user-management  
**Date**: 2026-05-12

---

## Decision 1: Session Invalidation Mechanism

**Decision**: Token versioning — add `token_version INT` to the `users` table and embed it in the JWT payload. `RequireAuth` (with UserStore injected) fetches the user by ID on every authenticated request and rejects the request if the user is deleted or the stored `token_version` doesn't match the token's version.

**Rationale**: JWTs are stateless by design. Immediate invalidation requires a server-side check. Options considered:
- **Blacklist table**: Stores revoked JTI values. Clean but adds a table and requires periodic cleanup.
- **Token versioning** (chosen): No new table. Deletion is handled naturally (user not found → 401). Password change increments the version, invalidating all outstanding tokens for that user. Acceptable for this app's scale since user count is small and the extra DB read on each authenticated request is negligible.

**Alternatives considered**: Blacklist table, short-lived tokens with refresh (over-engineered for this use case), no invalidation (rejected — spec requires immediate invalidation).

---

## Decision 2: Recipe Orphaning on User Delete (Database-Level)

**Decision**: Add a SQLite migration to recreate the `recipes` table with `creator_id TEXT REFERENCES users(id) ON DELETE SET NULL` (nullable, with cascade-null FK). The `recipes.creator_id` column currently has `NOT NULL` and no cascade rule, which would prevent user deletion while recipes exist.

**Rationale**: Relying on application-level "set creator_id to empty string then delete" is fragile (FK constraint blocks it with `PRAGMA foreign_keys=ON`). The SQLite table-rebuild migration pattern is already used in the codebase (`ALTER TABLE recipes ADD COLUMN notes`). Changing the FK to ON DELETE SET NULL with a nullable column cleanly implements the "orphan" semantics at the data layer.

**Migration approach** (additive, idempotent):
1. Check if `creator_id` is already nullable (via `PRAGMA table_info(recipes)` — if `notnull = 0`, skip).
2. If not nullable: create `recipes_new` with the new schema, copy data, drop old table, rename.

**Alternatives considered**: Application-level FK disable (fragile, non-atomic), soft-delete sentinel ID (requires special-casing everywhere), no orphaning/block deletion instead (rejected — spec requires orphaning).

---

## Decision 3: UserStore Interface Extensions

**Decision**: Add four methods to the `UserStore` interface in `store.go`:
- `List() ([]*model.User, error)` — returns all non-admin users
- `Update(user *model.User) error` — updates first_name, last_name, email, password_hash, token_version
- `Delete(id string) error` — deletes the user (after recipes are orphaned by FK cascade)
- `GetByEmail(email string) (*model.User, error)` — for email uniqueness checks

**Rationale**: The current interface (Create, GetByID, GetByUsername, Count) covers only the bootstrap and login paths. All four methods map directly to spec requirements. `List` returns only non-admin users (the admin interface doesn't manage admin accounts — FR-003 / clarification Q3).

**DynamoDB**: The DynamoDB `UserStore` must also implement these four methods. Implementation is straightforward (GSI on email for `GetByEmail`; full table scan for `List` given small expected user count). DynamoDB implementation is in scope but lower priority; SQLite is the primary store.

---

## Decision 4: New User Profile Fields (Backend Model)

**Decision**: Add `FirstName`, `LastName`, `Email` to `model.User` as optional string fields (empty string = not provided). Add `TokenVersion int` for session invalidation.

**Rationale**: The spec defines these as optional (only username + password required). Using empty string rather than pointer (`*string`) keeps the model simple and consistent with the existing codebase style. Email uniqueness is enforced at the DB level with a partial unique index: `CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique ON users(email) WHERE email != ''` — this allows multiple users to have no email while still enforcing uniqueness among those who do.

**Alternatives considered**: `*string` pointers for optional fields (more idiomatic Go null semantics but adds complexity throughout), separate `UserProfile` table (over-engineered for this feature scope).

---

## Decision 5: Frontend Admin UI Pattern

**Decision**: Two new page components using the existing vanilla-JS pattern:
- `AdminUserList.js` — table of users with Edit/Delete actions and an Add User button
- `AdminUserForm.js` — shared create/edit form (receives optional `id` prop; fetches user if editing)

**Rationale**: Mirrors the existing RecipeList/RecipeForm pair. No new libraries or patterns introduced. `isAdmin()` helper added to `api/auth.js` by decoding the stored JWT payload (base64 URL decode of the middle segment) to expose the `is_admin` claim — used to conditionally render the Admin nav link and guard the route.

**Nav addition**: An "Admin" link in `buildNav()` visible only when `isAdmin()` returns true, linking to `#/admin/users`.

**Alternatives considered**: Separate admin subdomain (overkill), embedding user management in a settings page (doesn't match spec's "admin section" framing).

---

## Decision 6: RequireAuth Middleware Refactor

**Decision**: Add a `RequireAuthWithStore(us store.UserStore) func(http.Handler) http.Handler` constructor that replaces the current `RequireAuth` wrapper for protected routes. The existing `RequireAuth` function is kept for backward compatibility but the new one is used for all routes that require token-version validation.

**Rationale**: The current `RequireAuth` is a plain function with no DB access. Injecting `UserStore` via a constructor (closure) follows the existing handler pattern (`NewAdminHandler(us)`, `NewAuthHandler(us)`) and avoids global state.

**Impact**: `main.go`'s `buildHandler` passes `userStore` to the new middleware constructor. All existing routes that use `RequireAuth` are updated to use `RequireAuthWithStore`.
