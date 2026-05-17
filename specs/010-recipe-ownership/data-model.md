# Data Model: Recipe Ownership and Per-User Recipe Listing

## Existing Entities (no structural change)

### Recipe

The `Recipe` entity already has a `CreatorID` field. **No schema migration is required.**

| Field       | Type     | Notes                                                                 |
|-------------|----------|-----------------------------------------------------------------------|
| ID          | string   | UUID, primary key                                                     |
| Name        | string   | Required, non-empty                                                   |
| Ingredients | []Ingredient | Array, may be empty                                               |
| Steps       | []string | Array, may be empty                                                   |
| Properties  | map[string]string | Optional key-value pairs                                   |
| Notes       | string   | Free-form markdown, may be empty                                      |
| **CreatorID** | **string** | **User ID of the creator. Set at creation time, immutable. Empty string for legacy recipes.** |
| CreatedAt   | time.Time | Set at creation                                                      |
| UpdatedAt   | time.Time | Updated on each write                                                 |

**Ownership invariants**:
- `CreatorID` is written once at `Create` time from the authenticated user's JWT claims.
- `CreatorID` is never updated by `Update` — it is ignored if present in the request body.
- A recipe with `CreatorID == ""` is treated as unowned (legacy); only admins may modify it.

### User

No changes to the `User` entity. The `IsAdmin` flag drives administrator privilege.

---

## Store Interface Extension

One method is added to the `RecipeStore` interface:

```go
// ListByCreator returns the paginated list of recipes owned by creatorID.
// Recipes with empty CreatorID are never included.
ListByCreator(creatorID string, page, limit int) ([]*model.Recipe, int, error)
```

This method is implemented in both `store/sqlite` and `store/dynamo`.

---

## API Response Shape (no change)

All existing recipe endpoints already include `creator_id` in the JSON response body (from `model.Recipe.CreatorID`). The frontend uses this field to determine edit/delete button visibility.

The new `GET /api/v1/recipes/mine` endpoint returns the same envelope as `GET /api/v1/recipes`:

```json
{
  "data": [ /* Recipe objects */ ],
  "total": 12,
  "page": 1,
  "limit": 20
}
```
