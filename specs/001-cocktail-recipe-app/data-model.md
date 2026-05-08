# Data Model: Cocktail Recipe App

**Phase**: 1 | **Date**: 2026-05-08 | **Branch**: `001-cocktail-recipe-app`

## Entities

---

### Recipe

The central entity. Holds core fields plus an unbounded map of additional properties.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| `id` | string (UUID v4) | Required, unique, immutable | Generated on creation |
| `name` | string | Required, non-empty | Duplicate names allowed; warn on save (FR-017) |
| `ingredients` | []Ingredient | Required, min 0 items | Ordered list |
| `steps` | []string | Required, min 0 items | Ordered preparation instructions |
| `properties` | map[string]string | Optional, unbounded | Arbitrary key-value metadata (base spirit, style, garnish, etc.) |
| `creator_id` | string (UUID) | Required | Foreign key → User.id; recorded at creation (FR-015) |
| `created_at` | timestamp (UTC) | Required, immutable | Set on insert |
| `updated_at` | timestamp (UTC) | Required | Updated on every modification |

**Validation rules**:
- `name` must be non-empty after trimming whitespace.
- `ingredients` and `steps` may be empty arrays (not null); an empty recipe is valid (saved with a warning if both are empty).
- `properties` keys must be non-empty strings; values may be empty strings.
- `creator_id` must reference a valid User.

**State transitions**: No draft state. A recipe is always published immediately on creation (FR-018). The only lifecycle transitions are: created → updated (any number of times) → deleted.

---

### Ingredient

An embedded sub-entity within Recipe. Not stored independently.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| `name` | string | Required, non-empty | e.g., "lime juice" |
| `quantity` | string | Required, non-empty | e.g., "30", "1/2" |
| `unit` | string | Optional | e.g., "ml", "oz", "dash", "sprig"; empty string if not applicable |

**Notes**: Quantity is a string (not a number) to accommodate fractions and mixed expressions (e.g., "1 1/2"). Unit is optional for items like "a lime wedge" where no measure applies.

---

### User

A contributor account. Created by an administrator only (FR-013).

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| `id` | string (UUID v4) | Required, unique, immutable | Generated on creation |
| `username` | string | Required, unique, non-empty | Case-insensitive uniqueness check |
| `password_hash` | string | Required | bcrypt hash; never returned in API responses |
| `is_admin` | bool | Required, default false | Admin users can create other user accounts |
| `created_at` | timestamp (UTC) | Required, immutable | Set on insert |

**Validation rules**:
- `username` must be 3–50 characters, alphanumeric plus hyphens and underscores.
- Plaintext passwords must be ≥ 8 characters; hashed with bcrypt (cost ≥ 12) before storage.
- `password_hash` is never included in any API response.

---

## Relationships

```
User ──< Recipe          (one User creates many Recipes)
Recipe ──< Ingredient    (one Recipe contains many Ingredients, embedded)
Recipe ──< Property      (one Recipe has many Properties, stored as a map)
```

---

## Storage Schema

### SQLite (local development)

```sql
CREATE TABLE users (
    id           TEXT PRIMARY KEY,
    username     TEXT NOT NULL UNIQUE COLLATE NOCASE,
    password_hash TEXT NOT NULL,
    is_admin     INTEGER NOT NULL DEFAULT 0,
    created_at   TEXT NOT NULL
);

CREATE TABLE recipes (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    ingredients TEXT NOT NULL,   -- JSON array of {name, quantity, unit}
    steps       TEXT NOT NULL,   -- JSON array of strings
    properties  TEXT NOT NULL,   -- JSON object (map[string]string)
    creator_id  TEXT NOT NULL REFERENCES users(id),
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

-- FTS5 index for full-text search across all searchable text
CREATE VIRTUAL TABLE recipes_fts USING fts5(
    recipe_id UNINDEXED,
    search_text,               -- concatenated: name + ingredient names + steps + property values
    content='',
    contentless_delete=1
);
```

The `search_text` field is rebuilt on every insert/update by the application layer (concatenating all text fields and all property values). FTS5 tokenizer: `unicode61` (handles accented characters).

### DynamoDB (AWS production)

**Table**: `cocktails-recipes`

| Attribute | Type | Role |
|-----------|------|------|
| `PK` | String | `RECIPE#<id>` — primary key |
| `name` | String | |
| `ingredients` | List | Each item: Map with `name`, `quantity`, `unit` |
| `steps` | List | Strings |
| `properties` | Map | Arbitrary string→string attributes |
| `creator_id` | String | |
| `created_at` | String (ISO 8601) | |
| `updated_at` | String (ISO 8601) | |

**Table**: `cocktails-users`

| Attribute | Type | Role |
|-----------|------|------|
| `PK` | String | `USER#<id>` — primary key |
| `username` | String | GSI partition key (for login lookup) |
| `password_hash` | String | |
| `is_admin` | Boolean | |
| `created_at` | String (ISO 8601) | |

**GSI**: `username-index` on `cocktails-users` with partition key `username` — used for login lookups.

---

## Search Index Design

The search query string is matched against a concatenated text representation of each recipe:

```
searchable_text = name
               + " " + join(ingredient.name for each ingredient)
               + " " + join(steps)
               + " " + join(property_values)
```

- **SQLite**: `recipes_fts` is queried with `MATCH '<query>*'` for prefix matching.
- **DynamoDB**: `Scan` with `FilterExpression` using `contains(search_text, :q)`. The `search_text` attribute is stored as a denormalized string alongside each record.

Both strategies ensure that properties added after initial launch are automatically included in search (FR-003) because the search text is derived from the live `properties` map at write time.

---

## Notes on Flexibility

The `properties` map satisfies FR-006 (flexible schema) — no schema migration is required to add a new property type to a recipe. The application never enumerates or validates property keys; it stores and returns whatever the client provides. The search index flattens all property values into the searchable text, so new properties are immediately findable (FR-003).
