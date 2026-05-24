# Data Model: Recipe Favorites

**Feature**: 014-recipe-favorites
**Date**: 2026-05-22

---

## New Entity: Favorite

A Favorite represents a user's saved interest in a recipe they did not create.

### Go Model (`internal/model/model.go` addition)

```go
type Favorite struct {
    UserID    string    `json:"user_id"`
    RecipeID  string    `json:"recipe_id"`
    CreatedAt time.Time `json:"created_at"`
}
```

### DynamoDB Table: `cocktails-favorites`

| Attribute | Type | Key Role |
|-----------|------|----------|
| `user_id` | String | Partition Key (PK) |
| `recipe_id` | String | Sort Key (SK) |
| `created_at` | String (RFC3339Nano) | — |

**Access patterns:**

| Operation | DynamoDB call |
|-----------|---------------|
| Add favorite | `PutItem(PK=user_id, SK=recipe_id)` |
| Remove favorite | `DeleteItem(PK=user_id, SK=recipe_id)` |
| Check if favorited | `GetItem(PK=user_id, SK=recipe_id)` → item exists? |
| List user's favorites | `Query(PK=user_id)` |
| Count favorites for recipe (P3) | `Scan(filter recipe_id=X)` or GSI (deferred) |

### SQLite Table: `favorites` (local dev)

```sql
CREATE TABLE IF NOT EXISTS favorites (
    user_id    TEXT NOT NULL,
    recipe_id  TEXT NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (user_id, recipe_id)
);
```

---

## Store Interface Addition

```go
// internal/store/store.go addition

type FavoriteStore interface {
    Add(userID, recipeID string) error
    Remove(userID, recipeID string) error
    IsFavorite(userID, recipeID string) (bool, error)
    ListByUser(userID string) ([]*model.Favorite, error)
    CountByRecipe(recipeID string) (int, error) // P3 — stub returns 0 for now
}
```

---

## Modified Entity: Recipe (API response only)

The `model.Recipe` struct is NOT modified. The API response for `GET /recipes/favorites` wraps recipes with an `is_favorite: true` marker via the handler layer, not the model.

```json
// GET /api/v1/recipes/favorites response shape
{
  "data": [
    {
      "id": "...",
      "name": "...",
      "ingredients": [...],
      "creator_id": "...",
      "is_favorite": true
    }
  ],
  "total": 5,
  "page": 1,
  "limit": 20
}
```

The `is_favorite` field is added at the handler layer as an envelope field — no model change.

---

## Frontend State

### `RecipeDetail.js`

| Variable | Type | Lifecycle |
|----------|------|-----------|
| `isFavorited` | `boolean` | Fetched on load via `GET /recipes/{id}/favorite`; toggled on button click |

### `MyRecipes.js`

| Variable | Type | Lifecycle |
|----------|------|-----------|
| `createdRecipes` | `Recipe[]` | Fetched from `GET /recipes/mine` |
| `favoritedRecipes` | `Recipe[]` | Fetched from `GET /recipes/favorites` |
| `unifiedList` | `{recipe, isFavorite}[]` | Merged in browser; created recipes take precedence over favorited |

---

## Entity Relationships

```
User ──< Favorite >── Recipe
(1)       (many)      (many)

• One user can have many favorites
• One recipe can be favorited by many users
• A user CANNOT favorite their own recipe (enforced at handler layer)
• Uniqueness: (user_id, recipe_id) pair is unique
```

---

## Infrastructure Change (Terraform)

A new DynamoDB table resource is required in `infra/main.tf`:

```hcl
module "favorites_table" {
  source  = "terraform-aws-modules/dynamodb-table/aws"
  version = "~> 5.0"

  name         = "${var.project_name}-favorites"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "user_id"
  range_key    = "recipe_id"

  attributes = [
    { name = "user_id",   type = "S" },
    { name = "recipe_id", type = "S" },
  ]

  point_in_time_recovery_enabled = true
  server_side_encryption_enabled = true
}
```

Lambda IAM policy must be extended to allow `GetItem`, `PutItem`, `DeleteItem`, `Query` on the new table.
