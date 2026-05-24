# Quickstart: Multi-Ingredient Search

**Feature**: 015-multi-ingredient-search
**Date**: 2026-05-24

## Integration Test Scenarios

Seed the following recipes before each scenario (or use the existing seed data):

| ID | Name | Ingredients |
|----|------|-------------|
| r1 | Gin Fizz | Gin, Lemon Juice, Sugar, Soda Water |
| r2 | Daiquiri | Rum, Lime Juice, Sugar |
| r3 | Bee's Knees | Gin, Lemon Juice, Honey |
| r4 | Whiskey Sour | Whiskey, Lemon Juice, Sugar, Egg White |

### Scenario 1 — Two-ingredient AND (natural language)

```
GET /api/v1/recipes?q=gin+and+lemon+juice
```

Expected: recipes r1, r3 (both contain Gin AND Lemon Juice). r2 and r4 excluded.

### Scenario 2 — Two-ingredient AND (+ delimiter, no spaces)

```
GET /api/v1/recipes?q=gin%2Blemon+juice
```

Expected: same as Scenario 1 — r1 and r3.

### Scenario 3 — Three-ingredient AND

```
GET /api/v1/recipes?q=gin+%2B+lemon+juice+%2B+sugar
```

Expected: r1 only (has Gin, Lemon Juice, AND Sugar). r3 excluded (no Sugar).

### Scenario 4 — No matches

```
GET /api/v1/recipes?q=rum+and+gin
```

Expected: empty `data` array, `total: 0`. No recipe contains both Rum and Gin.

### Scenario 5 — Single term (existing behaviour preserved)

```
GET /api/v1/recipes?q=lemon
```

Expected: r1, r3, r4 (all have some form of lemon in ingredients). Existing FTS/substring logic applies.

### Scenario 6 — Empty query (existing behaviour preserved)

```
GET /api/v1/recipes
```

Expected: all four recipes returned.

### Scenario 7 — Case insensitivity

```
GET /api/v1/recipes?q=GIN+AND+LEMON+JUICE
```

Expected: r1, r3 (same as Scenario 1).

### Scenario 8 — Token with extra whitespace

```
GET /api/v1/recipes?q=gin+%2B++lemon+juice
```

Expected: r1, r3 (whitespace tolerance).

## Frontend Manual Verification

1. Open the all-recipes page.
2. Verify the hint text `Tip: use "and" or "+" to search multiple ingredients — e.g. "gin and lemon"` is visible below the search bar.
3. Type `gin and lemon juice` — grid updates to show only Gin Fizz and Bee's Knees.
4. Clear the bar — all recipes reload.
5. Type `gin + lemon juice + sugar` — grid shows only Gin Fizz.
6. Type `rum and gin` — empty state with "No recipes match your search."
