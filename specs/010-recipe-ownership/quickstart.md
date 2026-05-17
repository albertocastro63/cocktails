# Quickstart: Recipe Ownership Integration Scenarios

These scenarios can be run locally with `STORE_BACKEND=sqlite` using the dev server.

## Scenario 1: Owner edits their own recipe

```
# Log in as user A → get token A
POST /api/v1/auth/login  { "username": "alice", "password": "..." }

# Create a recipe as user A
POST /api/v1/recipes  Authorization: Bearer <tokenA>
{ "name": "Margarita", "ingredients": [], "steps": [] }
→ 201 Created, recipe.id = "abc"

# Edit as owner — should succeed
PUT /api/v1/recipes/abc  Authorization: Bearer <tokenA>
{ "name": "Classic Margarita" }
→ 200 OK
```

## Scenario 2: Non-owner cannot edit

```
# Log in as user B → get token B
POST /api/v1/auth/login  { "username": "bob", "password": "..." }

# Attempt edit on user A's recipe
PUT /api/v1/recipes/abc  Authorization: Bearer <tokenB>
{ "name": "Hijacked Recipe" }
→ 403 Forbidden  { "error": { "code": "FORBIDDEN", ... } }
```

## Scenario 3: Admin can edit any recipe

```
# Log in as admin → get admin token
POST /api/v1/auth/login  { "username": "admin", "password": "..." }

# Edit user A's recipe as admin
PUT /api/v1/recipes/abc  Authorization: Bearer <adminToken>
{ "name": "Admin Edit" }
→ 200 OK
```

## Scenario 4: My Recipes listing

```
# As user A, create two recipes (abc, def)
# As user B, create one recipe (ghi)

# Get user A's recipes
GET /api/v1/recipes/mine  Authorization: Bearer <tokenA>
→ 200 OK  { "data": [abc, def], "total": 2 }

# Get user B's recipes
GET /api/v1/recipes/mine  Authorization: Bearer <tokenB>
→ 200 OK  { "data": [ghi], "total": 1 }
```

## Scenario 5: Empty My Recipes

```
# Log in as a user with no created recipes
GET /api/v1/recipes/mine  Authorization: Bearer <tokenNew>
→ 200 OK  { "data": [], "total": 0 }
```

## Scenario 6: My Recipes requires auth

```
GET /api/v1/recipes/mine  (no Authorization header)
→ 401 Unauthorized
```

## Frontend smoke test (manual)

1. Open app, sign in as Alice.
2. Verify "My Recipes" appears in the nav bar.
3. Click "My Recipes" → see only Alice's recipes, same card style as All Recipes.
4. Navigate to All Recipes → recipes Alice does not own show **no** edit/delete buttons.
5. Sign out → "My Recipes" link disappears from nav.
6. Sign in as admin → admin can see edit/delete on all recipe cards (or use the admin panel).
