package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/almc/cocktails/internal/handler"
	"github.com/almc/cocktails/internal/logging"
	"github.com/almc/cocktails/internal/model"
	"github.com/almc/cocktails/internal/store"
	dynstore "github.com/almc/cocktails/internal/store/dynamo"
	"github.com/google/uuid"
)

func main() {
	logging.SetDefault(logging.New(logging.LevelFromEnv()))

	ctx := context.Background()
	client, err := dynstore.NewClient(ctx)
	if err != nil {
		log.Fatalf("dynamodb client: %v", err)
	}

	tables := dynstore.TableNames{
		Recipes:   envOr("RECIPES_TABLE", "cocktails-recipes"),
		Users:     envOr("USERS_TABLE", "cocktails-users"),
		Favorites: envOr("FAVORITES_TABLE", "cocktails-favorites"),
	}

	// Local dev/tests target the emulator via DYNAMODB_ENDPOINT; there we
	// provision the schema on startup. Production tables are Terraform-managed,
	// so EnsureSchema is skipped (and never has permission) without an endpoint.
	if os.Getenv("DYNAMODB_ENDPOINT") != "" {
		if err := dynstore.EnsureSchema(ctx, client, tables); err != nil {
			log.Fatalf("provision local schema — is the emulator running? "+
				"start it with `docker compose up -d dynamodb-local`: %v", err)
		}
	}

	var recipeStore store.RecipeStore = dynstore.NewRecipeStore(client, tables.Recipes)
	var userStore store.UserStore = dynstore.NewUserStore(client, tables.Users)
	var favoriteStore store.FavoriteStore = dynstore.NewFavoriteStore(client, tables.Favorites)

	bootstrapAdmin(userStore)

	h := buildHandler(recipeStore, userStore, favoriteStore)
	// Outermost: bind a request-scoped logger (rid+req) and recover panics.
	h = handler.RequestLogger(handler.Recover(h))

	port := envOr("PORT", "8080")
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, h))
}

func buildHandler(rs store.RecipeStore, us store.UserStore, fs store.FavoriteStore) http.Handler {
	recipes := handler.NewRecipeHandler(rs)
	authH := handler.NewAuthHandler(us)
	adminH := handler.NewAdminHandler(us)
	favH := handler.NewFavoriteHandler(fs, rs)

	mux := http.NewServeMux()

	requireAuth := handler.RequireAuthWithStore(us)

	mux.HandleFunc("GET /api/v1/recipes", recipes.List)
	mux.HandleFunc("GET /api/v1/recipes/random", recipes.Random)
	mux.HandleFunc("GET /api/v1/recipes/names", recipes.Names)
	mux.Handle("GET /api/v1/recipes/mine", requireAuth(http.HandlerFunc(recipes.Mine)))
	// Favorites routes must be registered before the {id} wildcard
	mux.Handle("GET /api/v1/recipes/favorites", requireAuth(http.HandlerFunc(favH.List)))
	mux.Handle("PUT /api/v1/recipes/{id}/favorite", requireAuth(http.HandlerFunc(favH.Add)))
	mux.Handle("DELETE /api/v1/recipes/{id}/favorite", requireAuth(http.HandlerFunc(favH.Remove)))
	mux.Handle("GET /api/v1/recipes/{id}/favorite", requireAuth(http.HandlerFunc(favH.Check)))
	mux.HandleFunc("GET /api/v1/recipes/{id}", recipes.GetByID)
	mux.Handle("POST /api/v1/recipes", requireAuth(http.HandlerFunc(recipes.Create)))
	mux.Handle("PUT /api/v1/recipes/{id}", requireAuth(http.HandlerFunc(recipes.Update)))
	mux.Handle("DELETE /api/v1/recipes/{id}", requireAuth(http.HandlerFunc(recipes.Delete)))
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.Handle("GET /api/v1/admin/users",
		requireAuth(handler.RequireAdmin(http.HandlerFunc(adminH.ListUsers))))
	mux.Handle("POST /api/v1/admin/users",
		requireAuth(handler.RequireAdmin(http.HandlerFunc(adminH.CreateUser))))
	mux.Handle("GET /api/v1/admin/users/{id}",
		requireAuth(handler.RequireAdmin(http.HandlerFunc(adminH.GetUser))))
	mux.Handle("PUT /api/v1/admin/users/{id}",
		requireAuth(handler.RequireAdmin(http.HandlerFunc(adminH.UpdateUser))))
	mux.Handle("DELETE /api/v1/admin/users/{id}",
		requireAuth(handler.RequireAdmin(http.HandlerFunc(adminH.DeleteUser))))

	adminRecipesH := handler.NewAdminRecipeHandler(rs)
	mux.Handle("GET /api/v1/admin/schema",
		requireAuth(handler.RequireAdmin(http.HandlerFunc(adminRecipesH.ExportSchema))))
	mux.Handle("GET /api/v1/admin/recipes/export",
		requireAuth(handler.RequireAdmin(http.HandlerFunc(adminRecipesH.ExportRecipes))))
	mux.Handle("POST /api/v1/admin/recipes/import",
		requireAuth(handler.RequireAdmin(http.HandlerFunc(adminRecipesH.ImportRecipes))))

	return handler.CORSMiddleware(mux)
}

func bootstrapAdmin(us store.UserStore) {
	password := os.Getenv("ADMIN_BOOTSTRAP_PASSWORD")
	if password == "" {
		return
	}
	count, err := us.Count()
	if err != nil || count > 0 {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("bootstrap: hash password: %v", err)
		return
	}
	admin := &model.User{
		ID:           uuid.NewString(),
		Username:     "admin",
		PasswordHash: string(hash),
		IsAdmin:      true,
		CreatedAt:    time.Now().UTC(),
	}
	if err := us.Create(admin); err != nil {
		log.Printf("bootstrap: create admin: %v", err)
		return
	}
	log.Println("bootstrap: admin user created")
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
