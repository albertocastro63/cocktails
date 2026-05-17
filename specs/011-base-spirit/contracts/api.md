# API Contract: Base Spirit Designation

This feature makes no changes to API endpoints, HTTP methods, or response envelopes. The only change is to the `Ingredient` object shape within `POST /api/v1/recipes` and `PUT /api/v1/recipes/{id}`.

## Changed Object: Ingredient

### Before

```json
{
  "name": "Rye Whiskey",
  "quantity": "1.5",
  "unit": "oz"
}
```

### After

```json
{
  "name": "Rye Whiskey",
  "quantity": "1.5",
  "unit": "oz",
  "is_base_spirit": true
}
```

The `is_base_spirit` field is:
- **Optional** on input — omitting it is equivalent to `false`
- **Omitted on output** when `false` (Go `omitempty` behaviour)
- **Included on output** only when `true`

## Affected Endpoints

| Method | Path | Change |
|--------|------|--------|
| `POST` | `/api/v1/recipes` | `ingredients[*]` now accepts `is_base_spirit` boolean |
| `PUT` | `/api/v1/recipes/{id}` | `ingredients[*]` now accepts `is_base_spirit` boolean |
| `GET` | `/api/v1/recipes` | `ingredients[*]` may include `is_base_spirit: true` on one ingredient |
| `GET` | `/api/v1/recipes/{id}` | `ingredients[*]` may include `is_base_spirit: true` on one ingredient |
| `GET` | `/api/v1/recipes/mine` | `ingredients[*]` may include `is_base_spirit: true` on one ingredient |

## Example: Create Recipe with Base Spirit

**Request** `POST /api/v1/recipes`
```json
{
  "name": "Manhattan",
  "ingredients": [
    { "name": "Rye Whiskey", "quantity": "1.5", "unit": "oz", "is_base_spirit": true },
    { "name": "Sweet Vermouth", "quantity": "0.5", "unit": "oz" },
    { "name": "Angostura Bitters", "quantity": "2", "unit": "dashes" }
  ],
  "steps": ["Stir and serve up."],
  "notes": ""
}
```

**Response** `201 Created`
```json
{
  "data": {
    "id": "abc123",
    "name": "Manhattan",
    "ingredients": [
      { "name": "Rye Whiskey", "quantity": "1.5", "unit": "oz", "is_base_spirit": true },
      { "name": "Sweet Vermouth", "quantity": "0.5", "unit": "oz" },
      { "name": "Angostura Bitters", "quantity": "2", "unit": "dashes" }
    ],
    ...
  }
}
```

## Backward Compatibility

Existing clients that do not send `is_base_spirit` continue to work unchanged. Existing recipes in storage have no `is_base_spirit` attribute; they are returned without the field (all ingredients are treated as non-base-spirit by display logic).
