# UI Contracts: Admin User Management

**Feature**: 004-admin-user-management  
**Date**: 2026-05-12

---

## Helper: isAdmin()

**File**: `frontend/src/api/auth.js`

```
isAdmin() → boolean
```

Decodes the stored JWT payload (base64url, middle segment) and returns `is_admin`. Returns `false` if no token is stored or decoding fails.

Used by: `main.js` (nav rendering, route guard).

---

## API Client Extensions

**File**: `frontend/src/api/client.js`

```
listUsers(token)              → Promise<User[]>
createUser(data, token)       → Promise<User>   (existing, extended with first_name/last_name/email)
updateUser(id, data, token)   → Promise<User>
deleteUser(id, token)         → Promise<null>
```

`User` shape returned by all four:
```json
{ "id", "username", "first_name", "last_name", "email", "is_admin", "created_at" }
```

---

## Page: AdminUserList

**File**: `frontend/src/pages/AdminUserList.js`  
**Route**: `#/admin/users`  
**Props**: none  
**Auth guard**: redirects to `#/login` if not logged in; shows "Access denied" if logged in but not admin.

### DOM structure

```
div.max-w-4xl
  h1 "Users"
  button[data-add-user] "+ Add User"   ← navigates to #/admin/users/new
  div[data-content]
    ── Loading state: text "Loading…"
    ── Empty state: text "No users yet."
    ── Populated state:
       table
         thead > tr > th[username, name, email, actions]
         tbody > tr* (one per user)
           td username
           td full name (first + last, or "—" if empty)
           td email (or "—" if empty)
           td
             a[href=#/admin/users/{id}/edit] "Edit"
             button[data-delete-user={id}] "Delete"
    ── Error state: p.text-red-600 "Failed to load users."
```

### Behaviour

- On mount: calls `listUsers(token)`, renders table or empty/error state.
- Delete button: shows `confirm("Delete user {username}?")`. On confirm, calls `deleteUser(id, token)`, re-renders list on success, shows inline error on failure.
- Edit and Add links navigate via `location.hash`.

### Tests (vitest + jsdom)

- renders loading state initially
- renders table with user rows when users load
- renders empty state when no users
- renders error state on API failure
- clicking Add User navigates to #/admin/users/new
- clicking Delete shows confirm dialog; on confirm calls deleteUser and refreshes list
- clicking Delete and cancelling confirm does not call deleteUser

---

## Page: AdminUserForm

**File**: `frontend/src/pages/AdminUserForm.js`  
**Routes**: `#/admin/users/new` (create), `#/admin/users/{id}/edit` (edit)  
**Props**: `{ id?: string, onSave?: (user) => void }`

### DOM structure

```
div.max-w-2xl
  h1 "New User" | "Edit User"
  form
    input[name=username]   ← read-only in edit mode
    input[name=first_name]
    input[name=last_name]
    input[name=email]
    input[name=password, type=password]
      ── placeholder "Leave blank to keep existing" in edit mode
    p.text-red-600[hidden]  ← validation / API error message
    button[type=submit] "Create User" | "Save Changes"
```

### Behaviour

- **Create mode** (no `id`): all fields editable; submits `createUser({username, password, first_name, last_name, email}, token)`.
- **Edit mode** (`id` provided): fetches user via `getUser(id, token)` (new client function); pre-fills name/email; username rendered as read-only text; empty password = preserve existing.
- On success: calls `onSave(user)` if provided, otherwise navigates to `#/admin/users`.
- Validation (client-side): username non-empty (create only), password non-empty (create only). Server errors (409 USERNAME_CONFLICT, 409 EMAIL_CONFLICT) displayed in the error paragraph.

### Tests (vitest + jsdom)

- renders username input (create) / read-only username (edit)
- renders all optional fields (first_name, last_name, email, password)
- password placeholder says "Leave blank to keep existing" in edit mode
- submitting without username (create) shows validation error, does not call API
- submitting without password (create) shows validation error, does not call API
- successful create calls createUser with correct payload and navigates
- in edit mode, pre-fills fields from fetched user
- successful edit calls updateUser with correct payload
- API error (409 USERNAME_CONFLICT) is shown in the error paragraph

---

## Navigation update

**File**: `frontend/src/main.js`

- `buildNav()`: add `<a href="#/admin/users">Admin</a>` link visible only when `isAdmin()` returns true.
- `routes`: add entries for `#/admin/users` and `#/admin/users/new` and `#/admin/users/:id/edit`.
- Route guard: `#/admin/users*` paths require `isAdmin()`; redirect to `#/login` if not authenticated, show "Access denied" if authenticated but not admin.
