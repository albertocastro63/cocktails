# API Contracts: Admin User Management

**Feature**: 004-admin-user-management  
**Date**: 2026-05-12  
**Auth**: All endpoints require `Authorization: Bearer <token>` with `is_admin: true`

---

## GET /api/v1/admin/users

List all non-admin users.

**Auth**: RequireAuth + RequireAdmin

**Response 200**:
```json
[
  {
    "id": "uuid",
    "username": "alice",
    "first_name": "Alice",
    "last_name": "Smith",
    "email": "alice@example.com",
    "is_admin": false,
    "created_at": "2026-05-12T10:00:00Z"
  }
]
```

**Response 200 (no users)**:
```json
[]
```

**Error responses**: 401 UNAUTHORIZED, 403 FORBIDDEN

---

## POST /api/v1/admin/users

Create a new non-admin user. (Existing endpoint — extended with new optional fields.)

**Auth**: RequireAuth + RequireAdmin

**Request body**:
```json
{
  "username":   "alice",       // required
  "password":   "secret",      // required
  "first_name": "Alice",       // optional
  "last_name":  "Smith",       // optional
  "email":      "alice@example.com" // optional
}
```

**Response 201**: Created user object (password_hash omitted).

**Error responses**:
| Status | Code       | Condition                        |
|--------|------------|----------------------------------|
| 400    | BAD_REQUEST   | Missing username or password  |
| 409    | CONFLICT      | Username already exists       |
| 409    | EMAIL_CONFLICT | Email already in use         |
| 401    | UNAUTHORIZED  | Not authenticated             |
| 403    | FORBIDDEN     | Not admin                     |

---

## PUT /api/v1/admin/users/{id}

Update an existing non-admin user's profile. Username cannot be changed.

**Auth**: RequireAuth + RequireAdmin

**Request body** (all fields optional):
```json
{
  "password":   "newpassword", // optional; if provided and non-empty, replaces existing; increments token_version
  "first_name": "Alice",       // optional
  "last_name":  "Smith",       // optional
  "email":      "alice@example.com" // optional
}
```

**Response 200**: Updated user object.

**Error responses**:
| Status | Code          | Condition                        |
|--------|---------------|----------------------------------|
| 400    | BAD_REQUEST   | Invalid email format             |
| 404    | NOT_FOUND     | User ID not found                |
| 409    | EMAIL_CONFLICT | Email already in use by another user |
| 403    | FORBIDDEN     | Attempting to edit an admin account |
| 401    | UNAUTHORIZED  | Not authenticated                |
| 403    | FORBIDDEN     | Not admin                        |

---

## DELETE /api/v1/admin/users/{id}

Permanently delete a non-admin user. Their recipes are orphaned (creator_id set to NULL by FK cascade).

**Auth**: RequireAuth + RequireAdmin

**Response 204**: No content.

**Error responses**:
| Status | Code       | Condition                                     |
|--------|------------|-----------------------------------------------|
| 404    | NOT_FOUND  | User ID not found                             |
| 403    | FORBIDDEN  | Attempting to delete an admin account         |
| 401    | UNAUTHORIZED | Not authenticated                           |
| 403    | FORBIDDEN  | Not admin                                     |

---

## Modified: RequireAuth middleware

After this feature, `RequireAuth` fetches the user by ID on every authenticated request and validates `token_version`. This affects all existing protected endpoints:

- `POST /api/v1/recipes`
- `PUT /api/v1/recipes/{id}`
- `DELETE /api/v1/recipes/{id}`
- All `POST/PUT/DELETE /api/v1/admin/users*`

**New 401 conditions** (in addition to missing/invalid token):
- User ID in token no longer exists in the database (user was deleted)
- `token_version` in token does not match the stored version (password was reset)
