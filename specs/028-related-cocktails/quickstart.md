# Quickstart: Related Cocktails

How to exercise and verify the feature. Maps to the spec's Success Criteria.

## Backend (store + API)

```bash
cd backend
go test ./internal/store/... -run 'Related|SetRelated|Delete' -v   # symmetry, non-transitivity, dedupe, self, delete cleanup
go test ./internal/handler/ -run 'Related|Names|Create|Update|Delete' -v
go test ./... -p 1                                                  # full suite, no regressions
```

Manual (local server, SQLite):

```bash
JWT_SECRET=x STORE_BACKEND=sqlite ADMIN_BOOTSTRAP_PASSWORD=admin123 go run ./cmd/server &
TOKEN=$(curl -s -X POST localhost:8080/api/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | jq -r .token)

# create two recipes, relate A -> B, then confirm B lists A (symmetry, SC-001)
A=$(curl -s -X POST localhost:8080/api/v1/recipes -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"name":"Negroni","ingredients":[],"steps":[]}' | jq -r .data.id)
B=$(curl -s -X POST localhost:8080/api/v1/recipes -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"name":"Left Hand","ingredients":[],"steps":[]}' | jq -r .data.id)
curl -s -X PUT localhost:8080/api/v1/recipes/$A -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d "{\"related_ids\":[\"$B\"]}" >/dev/null
curl -s localhost:8080/api/v1/recipes/$B | jq '.related'   # -> [{ id: A, name: "Negroni" }]

# names endpoint for the picker
curl -s localhost:8080/api/v1/recipes/names | jq '.[0]'     # -> { id, name }

# delete A, confirm it disappears from B (SC-006)
curl -s -X DELETE localhost:8080/api/v1/recipes/$A -H "Authorization: Bearer $TOKEN" >/dev/null
curl -s localhost:8080/api/v1/recipes/$B | jq '.related'   # -> []
```

## Frontend

```bash
cd frontend
npm test -- RelatedCocktailPicker RecipeForm RecipeDetail
npm run build
```

Manual: open a recipe in edit mode → in "Related cocktails", type a name, arrow-select, add a chip, save. Open the related cocktail's detail page → confirm the reverse link appears at the bottom. Open the home page → confirm the random cocktail shows **no** related section.

## Success-criteria checklist

- [ ] SC-001 — relate A→B, both detail pages list the other (symmetric).
- [ ] SC-002 — A–B and B–C: A's related never contains C (non-transitive).
- [ ] SC-003 — detail page shows related links at the bottom, alphabetical, all resolve.
- [ ] SC-004 — home random cocktail shows no related section.
- [ ] SC-005 — add a relation by keyboard (type → arrow → Enter) in < 10s.
- [ ] SC-006 — removing a relation clears both sides; deleting a recipe removes it everywhere.
- [ ] SC-007 — no duplicate relation, no self-relation.
