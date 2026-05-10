# Cocktails

A web app to store, browse, and search cocktail recipes. Every page load shows a randomly selected recipe on the homepage. Recipes support a flexible schema — any key-value properties can be attached (base spirit, style, garnish, occasion, etc.) and are fully searchable.

Built with Go (net/http) on the backend and Vite + TailwindCSS on the frontend. Runs locally with SQLite; deployable to AWS via Lambda + API Gateway + S3 + CloudFront + DynamoDB.

## Features

- **Random recipe on homepage** — refreshes on every load
- **Browse & search** — full-text search across name, ingredients, steps, and all properties
- **Flexible properties** — attach arbitrary key-value pairs to any recipe; no schema changes needed
- **Recipe detail view** — ingredients with quantities, ordered steps, all properties
- **Authenticated writes** — admin-created accounts; only the recipe creator can edit or delete
- **Public read API** — all read endpoints require no authentication, suitable for external consumers

## Requirements

- Go 1.22+
- Node.js 20+

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
    "name": "Mojito",
    "ingredients": [
      {"name": "white rum", "quantity": "50", "unit": "ml"},
      {"name": "lime juice", "quantity": "25", "unit": "ml"},
      {"name": "mint leaves", "quantity": "10", "unit": "leaves"},
      {"name": "sugar syrup", "quantity": "15", "unit": "ml"},
      {"name": "soda water", "quantity": "75", "unit": "ml"}
    ],
    "steps": ["Muddle mint with sugar syrup", "Add rum and lime juice", "Fill with ice", "Top with soda water"],
    "properties": {"base_spirit": "rum", "style": "refreshing", "garnish": "lime wedge and mint sprig"}
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
│   │   ├── server/main.go     # Local HTTP server
│   │   └── lambda/main.go     # AWS Lambda entry point
│   └── internal/
│       ├── handler/           # HTTP handlers
│       ├── model/             # Domain types (Recipe, Ingredient, User)
│       ├── auth/              # JWT issue & parse
│       └── store/
│           ├── store.go       # RecipeStore / UserStore interfaces
│           ├── sqlite/        # SQLite + FTS5 (local dev)
│           └── dynamo/        # DynamoDB (AWS)
└── frontend/
    └── src/
        ├── pages/             # Home, RecipeList, RecipeDetail, Login, RecipeForm
        ├── components/        # RecipeCard, SearchBar, IngredientList, PropertyTable, EmptyState
        └── api/               # Fetch wrapper and auth helpers
```

## AWS Deployment

The backend has two entry points that share identical handler code:

- `cmd/server` — standard `net/http` server for local use
- `cmd/lambda` — wraps the same router with `httpadapter` for Lambda

Set `STORE_BACKEND=dynamodb` to switch from SQLite to DynamoDB. The frontend builds to a static bundle (`npm run build`) suitable for S3 + CloudFront.

See [`specs/001-cocktail-recipe-app/quickstart.md`](specs/001-cocktail-recipe-app/quickstart.md) for detailed AWS deployment steps.
