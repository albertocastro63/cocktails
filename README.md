# Cocktails

A web app to store, browse, and search cocktail recipes. Every page load shows a randomly selected recipe on the homepage. Recipes support a flexible schema — any key-value properties can be attached (base spirit, style, garnish, occasion, etc.) and are fully searchable. For more details [read here](MOTIVATION.md) for my motivation and decisions.

Built with Go (net/http) on the backend and Vite + TailwindCSS on the frontend. Runs locally with SQLite; deployable to AWS via Lambda + API Gateway + S3 + CloudFront + DynamoDB. Live at [cocktails.albertomcastro.com](https://cocktails.albertomcastro.com).

## Features

### Recipes
- **Random recipe on homepage** — a different recipe is highlighted on every visit
- **Browse all recipes** — paginated grid with alphabetical (A→Z / Z→A) sort controls
- **Recipe detail view** — ingredients with quantities and units, ordered steps, garnishes, markdown notes, and arbitrary key-value properties
- **Ingredient hover popover** — hover a recipe card to preview its ingredients; if the recipe has garnishes and fewer than 5 ingredients, garnishes appear in italics below; the popover floats as a fixed overlay so it never expands the page layout

### Search
- **Full-text search** — searches across recipe name, ingredients, steps, and all property values
- **Multi-ingredient AND search** — combine ingredients with `and` or `+` (e.g. `rum + lime + mint`)
- **Base spirit filter** — type `base spirit is gin` (or `base spirit = gin`) in the search bar to filter by base spirit; alternative spellings `whisky` / `whiskey` are normalised automatically

### Authoring
- **Markdown notes** — the notes field accepts Markdown with a live preview toggle; rendered consistently on the detail page and the recipe form
- **Ingredients with base spirit** — mark one ingredient per recipe as the base spirit; it is highlighted in the popover and on the detail page
- **Garnishes** — add free-text garnish instructions per recipe (e.g. "Express orange oil over the cocktail"); displayed in italics below ingredients on the detail page and in the hover popover when space allows
- **Flexible properties** — attach arbitrary key-value pairs to any recipe; no schema changes required
- **Authenticated writes** — admin-created accounts only; only the recipe's creator or an admin can edit or delete it
- **My Recipes page** — shows all recipes created by the logged-in user, plus any they have favorited

### Favorites
- **Heart icon on recipe detail** — logged-in users can favorite any recipe they did not create; the heart turns red when active
- **Favorite indicators on recipe cards** — heart markers appear on favorited recipe cards on both the All Recipes and My Recipes pages

### Admin
- **User management** — admins can create, edit, and delete user accounts via the admin interface
- **Recipe export** — download all recipes as a JSON file conforming to the published schema
- **Recipe import** — upload a JSON file to bulk-import recipes; the schema is available from the admin page

### Infrastructure
- **Dual store** — SQLite with FTS5 for local development; DynamoDB for production; switched via `STORE_BACKEND` environment variable
- **AWS deployment** — Lambda + API Gateway + S3 + CloudFront + DynamoDB, provisioned with Terraform (serverless.tf modules); custom domain with HTTPS via ACM
- **CI pipeline** — GitHub Actions runs Go tests, frontend Vitest tests, and builds on every PR and push to main; failing checks block merging
- **Preview environments** — every open PR gets an isolated environment at `cocktails.albertomcastro.com/pr-{number}/`; the URL is posted as a PR comment after the `Deploy Preview` workflow completes; production deploys automatically on merge to main
- **Public read API** — all read endpoints require no authentication, suitable for external consumers

### Preview Environments

Each open pull request has its own isolated environment:

| Resource | Pattern | Example (PR 42) |
|---|---|---|
| Frontend URL | `cocktails.albertomcastro.com/pr-{number}/` | `cocktails.albertomcastro.com/pr-42/` |
| Backend Lambda | `cocktails-pr-{number}-api` | `cocktails-pr-42-api` |
| DynamoDB tables | `cocktails-pr-{number}-{recipes,users,favorites}` | `cocktails-pr-42-recipes` |

**How it works:**
1. Push to a PR branch → `Deploy Preview` workflow builds the Lambda + frontend and deploys to a PR-scoped environment; the preview URL is posted as a PR comment.
2. Subsequent pushes update the Lambda code; DynamoDB tables and seed data are preserved.
3. Merge or close the PR → `Teardown Preview` workflow removes all PR resources automatically.
4. Merge to `main` → `Deploy Production` workflow updates the production Lambda and frontend without any manual steps.

## Requirements

- Go 1.22+
- Node.js 24+

## Running Locally

### 1. Backend

```bash
cd backend
go mod download

export JWT_SECRET="change-me-to-a-long-random-string"
export ADMIN_BOOTSTRAP_PASSWORD="your-admin-password"  # creates admin user on first run

go run ./cmd/server
# Server listens on http://localhost:8080
```

| Variable | Default | Description |
|---|---|---|
| `JWT_SECRET` | — | **Required.** Secret for JWT signing |
| `PORT` | `8080` | HTTP listen port |
| `DB_PATH` | `cocktails.db` | SQLite database file |
| `STORE_BACKEND` | `sqlite` | Set to `dynamodb` to use DynamoDB instead |
| `ADMIN_BOOTSTRAP_PASSWORD` | — | If set and no users exist, creates an `admin` account on startup |

### 2. Frontend

In a second terminal:

```bash
cd frontend
npm install
npm run dev
# Opens on http://localhost:5173
```

API requests to `/api/*` are proxied to the Go backend automatically. No extra configuration needed for local development.

### 3. Create your first recipe

```bash
# Log in
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-admin-password"}' | jq -r .token)

# Create a recipe
curl -X POST http://localhost:8080/api/v1/recipes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Old Fashioned",
    "ingredients": [
      {"name": "rye whiskey", "quantity": "60", "unit": "ml", "is_base_spirit": true},
      {"name": "sugar syrup", "quantity": "5", "unit": "ml"},
      {"name": "Angostura bitters", "quantity": "2", "unit": "dashes"}
    ],
    "garnishes": ["Express orange oil over the glass", "Use orange peel to garnish"],
    "steps": ["Combine all ingredients in a mixing glass with ice", "Stir until cold", "Strain into a rocks glass over a large ice cube"],
    "notes": "Use a **2:1 sugar syrup** for a richer result.",
    "properties": {"style": "stirred", "glass": "rocks"}
  }'
```

## Running Tests

```bash
# Backend
cd backend
go test ./...

# Frontend
cd frontend
npm test
```

## Project Structure

```
cocktails/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go      # Local HTTP server (SQLite)
│   │   └── lambda/main.go      # AWS Lambda entry point (DynamoDB)
│   └── internal/
│       ├── handler/            # HTTP handlers (recipes, auth, admin, favorites)
│       ├── model/              # Domain types (Recipe, Ingredient, User)
│       ├── auth/               # JWT issue & parse
│       └── store/
│           ├── store.go        # RecipeStore / UserStore / FavoriteStore interfaces
│           ├── sqlite/         # SQLite + FTS5 implementation (local dev)
│           └── dynamo/         # DynamoDB implementation (AWS)
├── frontend/
│   └── src/
│       ├── pages/              # Home, RecipeList, RecipeDetail, RecipeForm,
│       │                       # MyRecipes, Login, AdminUserList, AdminUserForm,
│       │                       # AdminRecipes
│       ├── components/         # RecipeCard, SearchBar, SortButtonGroup,
│       │                       # IngredientList, PropertyTable, MarkdownEditor,
│       │                       # FavoriteButton, Footer, EmptyState
│       ├── utils/              # Markdown renderer
│       └── api/                # Fetch wrapper and auth helpers
├── infra/                      # Terraform configuration (Lambda, API Gateway,
│                               # DynamoDB, S3, CloudFront, ACM, Route 53)
└── specs/                      # Feature specifications (one directory per feature)
```

## AWS Deployment

The backend has two entry points that share identical handler code:

- `cmd/server` — standard `net/http` server for local use with SQLite
- `cmd/lambda` — wraps the same router with `httpadapter` for Lambda + API Gateway

```bash
# Build and deploy the Lambda
cd backend
GOOS=linux GOARCH=arm64 go build -o bootstrap ./cmd/lambda
zip -j function.zip bootstrap
aws lambda update-function-code --function-name cocktails-prod-api --zip-file fileb://function.zip

# Build and deploy the frontend
cd frontend
npm run build
aws s3 sync dist/ s3://<frontend-bucket> --delete
aws cloudfront create-invalidation --distribution-id <distribution-id> --paths "/*"
```

Set `STORE_BACKEND=dynamodb` and the `RECIPES_TABLE`, `USERS_TABLE`, and `FAVORITES_TABLE` environment variables on the Lambda to switch from SQLite to DynamoDB.

Infrastructure is managed with Terraform. See [`infra/`](infra/) for the full configuration and [`specs/009-aws-terraform-infra/`](specs/009-aws-terraform-infra/) for deployment details.
