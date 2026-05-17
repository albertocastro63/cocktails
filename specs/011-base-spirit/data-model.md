# Data Model: Base Spirit Designation

## Changed Entities

### Ingredient (sub-document of Recipe)

**Current fields**: `name`, `quantity`, `unit`

**New field**:

| Field          | Type    | Required | Default | Constraints                                    |
|----------------|---------|----------|---------|------------------------------------------------|
| `is_base_spirit` | boolean | No       | false   | At most one ingredient per recipe may be `true` |

**Serialisation**:
- JSON: `"is_base_spirit": true` — field omitted when `false` (`omitempty`)
- DynamoDB: attribute `is_base_spirit` (boolean) — omitted when `false` (`omitempty`)
- SQLite: ingredients column is a JSON blob; transparent via `encoding/json`

**Backward compatibility**: Existing records have no `is_base_spirit` attribute. On read, the field deserialises to `false` in all stores. No migration required.

### Recipe (unchanged structurally)

The Recipe entity gains no new top-level fields. The `ingredients` array now carries richer sub-documents. The invariant "at most one `is_base_spirit=true` per recipe" is enforced by the UI, not the database.

## State Transitions

```
Ingredient (no base spirit)
        │ author checks checkbox
        ▼
Ingredient (is_base_spirit = true)   ←──────┐
        │ author checks a different          │
        │ ingredient OR unchecks this one    │ author checks this one
        ▼                                   │
Ingredient (no base spirit) ────────────────┘
```

At recipe save, the full ingredient list (with whichever `is_base_spirit` state is current) is sent to the API and persisted atomically as part of the recipe document.

## Validation Rules

- `is_base_spirit` is optional; absence is equivalent to `false`.
- No server-side uniqueness check; the single-select constraint is a UI concern (see research.md Decision 2).
- If the flagged ingredient is deleted during editing, the flag disappears with it — no orphan state possible.
