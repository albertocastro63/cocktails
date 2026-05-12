# Data Model: Admin User Management

**Feature**: 004-admin-user-management  
**Date**: 2026-05-12

---

## Entity: User (modified)

Existing entity extended with new optional profile fields and a session token version.

| Field          | Type    | Required | Constraints                                     | Notes                              |
|----------------|---------|----------|-------------------------------------------------|------------------------------------|
| id             | string  | yes      | UUID, primary key                               | Unchanged                          |
| username       | string  | yes      | Unique, case-insensitive                        | Immutable after creation           |
| password_hash  | string  | yes      | bcrypt hash, never exposed in API responses     | Unchanged                          |
| is_admin       | bool    | yes      | Default false                                   | Set only via bootstrap             |
| first_name     | string  | no       | Empty string = not provided                     | New field                          |
| last_name      | string  | no       | Empty string = not provided                     | New field                          |
| email          | string  | no       | Valid format when non-empty; unique when non-empty | New field; partial unique index  |
| token_version  | int     | yes      | Default 0; incremented on password change       | New field; used for session invalidation |
| created_at     | time    | yes      | UTC, set on creation                            | Unchanged                          |

### Validation Rules

- `username`: non-empty after trimming whitespace
- `password`: non-empty on creation; non-empty when provided on update (empty = preserve existing)
- `email`: must match standard email format when non-empty; must be unique across all users when non-empty
- `first_name`, `last_name`: no format constraint; stored as-is

### State Transitions

```
token_version: 0 (creation)
             ↑ +1 on each password change (invalidates all active sessions)
```

### Database Changes (SQLite)

Additive migration (idempotent — uses duplicate-column-name guard pattern):

```sql
ALTER TABLE users ADD COLUMN first_name    TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN last_name     TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN email         TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN token_version INTEGER NOT NULL DEFAULT 0;

-- Partial unique index: enforces uniqueness only for non-empty emails
CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique ON users(email) WHERE email != '';
```

---

## Entity: Recipe (modified)

One field changes: `creator_id` becomes nullable to support user orphaning.

| Field       | Type    | Required | Constraints                              | Notes                          |
|-------------|---------|----------|------------------------------------------|--------------------------------|
| creator_id  | string  | no       | FK → users(id) ON DELETE SET NULL        | Was NOT NULL; now nullable     |

### Database Changes (SQLite)

Table rebuild migration (required because SQLite cannot ALTER a FK constraint):

```sql
-- Only run if creator_id is still NOT NULL (check via PRAGMA table_info)
CREATE TABLE IF NOT EXISTS recipes_new (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    ingredients TEXT NOT NULL DEFAULT '[]',
    steps       TEXT NOT NULL DEFAULT '[]',
    properties  TEXT NOT NULL DEFAULT '{}',
    notes       TEXT NOT NULL DEFAULT '',
    creator_id  TEXT REFERENCES users(id) ON DELETE SET NULL,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);
INSERT OR IGNORE INTO recipes_new SELECT * FROM recipes;
DROP TABLE recipes;
ALTER TABLE recipes_new RENAME TO recipes;
```

---

## UserStore Interface Changes

```
Added to store.UserStore:
  List()                    → ([]*model.User, error)    — all non-admin users
  Update(*model.User)       → error                     — profile + token_version
  Delete(id string)         → error                     — removes user record
  GetByEmail(email string)  → (*model.User, error)      — for uniqueness check
```

---

## Relationships

```
users (1) ──── (0..*) recipes
               creator_id NULL = orphaned recipe (user was deleted)
```
