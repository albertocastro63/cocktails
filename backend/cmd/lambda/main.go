package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"golang.org/x/crypto/bcrypt"

	"github.com/almc/cocktails/internal/email"
	"github.com/almc/cocktails/internal/handler"
	"github.com/almc/cocktails/internal/model"
	"github.com/almc/cocktails/internal/store"
	dynstore "github.com/almc/cocktails/internal/store/dynamo"
	sqstore "github.com/almc/cocktails/internal/store/sqlite"
	"github.com/google/uuid"
)

func main() {
	recipeStore, userStore, favoriteStore := openStore()
	bootstrapAdmin(userStore)
	h := buildHandler(recipeStore, userStore, favoriteStore)
	if prefix := os.Getenv("STRIP_PATH_PREFIX"); prefix != "" {
		h = http.StripPrefix(prefix, h)
	}
	lambda.Start(httpadapter.NewV2(h).ProxyWithContext)
}

func openStore() (store.RecipeStore, store.UserStore, store.FavoriteStore) {
	backend := os.Getenv("STORE_BACKEND")
	if backend == "dynamodb" {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			log.Fatalf("load aws config: %v", err)
		}
		client := dynamodb.NewFromConfig(cfg)
		recipesTable := os.Getenv("RECIPES_TABLE")
		usersTable := os.Getenv("USERS_TABLE")
		favoritesTable := os.Getenv("FAVORITES_TABLE")
		if recipesTable == "" || usersTable == "" || favoritesTable == "" {
			log.Fatal("RECIPES_TABLE, USERS_TABLE and FAVORITES_TABLE must be set when STORE_BACKEND=dynamodb")
		}
		return dynstore.NewRecipeStore(client, recipesTable),
			dynstore.NewUserStore(client, usersTable),
			dynstore.NewFavoriteStore(client, favoritesTable)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/tmp/cocktails.db"
	}
	rs, us, fs, err := sqstore.OpenAll(dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	return rs, us, fs
}

// newEmailSender returns an SES-backed sender when MAIL_FROM is configured
// (production/Lambda), otherwise a no-op stub (local/dev) so the app runs
// without email configured.
func newEmailSender() email.Sender {
	from := os.Getenv("MAIL_FROM")
	if from == "" {
		return &email.StubSender{}
	}
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("load aws config for SES: %v", err)
	}
	return email.NewSESSender(sesv2.NewFromConfig(cfg), from)
}

func appBaseURL() string {
	if v := os.Getenv("APP_BASE_URL"); v != "" {
		return v
	}
	return "https://cocktails.albertomcastro.com"
}

func buildHandler(rs store.RecipeStore, us store.UserStore, fs store.FavoriteStore) http.Handler {
	recipes := handler.NewRecipeHandler(rs)
	authH := handler.NewAuthHandler(us)
	adminH := handler.NewAdminHandler(us)
	favH := handler.NewFavoriteHandler(fs, rs)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/recipes", recipes.List)
	mux.HandleFunc("GET /api/v1/recipes/random", recipes.Random)
	mux.Handle("GET /api/v1/recipes/mine", handler.RequireAuth(http.HandlerFunc(recipes.Mine)))
	// Favorites routes must be registered before the {id} wildcard
	mux.Handle("GET /api/v1/recipes/favorites", handler.RequireAuth(http.HandlerFunc(favH.List)))
	mux.Handle("PUT /api/v1/recipes/{id}/favorite", handler.RequireAuth(http.HandlerFunc(favH.Add)))
	mux.Handle("DELETE /api/v1/recipes/{id}/favorite", handler.RequireAuth(http.HandlerFunc(favH.Remove)))
	mux.Handle("GET /api/v1/recipes/{id}/favorite", handler.RequireAuth(http.HandlerFunc(favH.Check)))
	mux.HandleFunc("GET /api/v1/recipes/{id}", recipes.GetByID)
	mux.Handle("POST /api/v1/recipes", handler.RequireAuth(http.HandlerFunc(recipes.Create)))
	mux.Handle("PUT /api/v1/recipes/{id}", handler.RequireAuth(http.HandlerFunc(recipes.Update)))
	mux.Handle("DELETE /api/v1/recipes/{id}", handler.RequireAuth(http.HandlerFunc(recipes.Delete)))
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)

	resetH := handler.NewPasswordResetHandler(us, newEmailSender(), appBaseURL())
	mux.HandleFunc("POST /api/v1/auth/forgot-password", resetH.Forgot)
	mux.HandleFunc("POST /api/v1/auth/reset-password", resetH.Reset)
	mux.Handle("GET /api/v1/admin/users",
		handler.RequireAuth(handler.RequireAdmin(http.HandlerFunc(adminH.ListUsers))))
	mux.Handle("POST /api/v1/admin/users",
		handler.RequireAuth(handler.RequireAdmin(http.HandlerFunc(adminH.CreateUser))))
	mux.Handle("GET /api/v1/admin/users/{id}",
		handler.RequireAuth(handler.RequireAdmin(http.HandlerFunc(adminH.GetUser))))
	mux.Handle("PUT /api/v1/admin/users/{id}",
		handler.RequireAuth(handler.RequireAdmin(http.HandlerFunc(adminH.UpdateUser))))
	mux.Handle("DELETE /api/v1/admin/users/{id}",
		handler.RequireAuth(handler.RequireAdmin(http.HandlerFunc(adminH.DeleteUser))))

	adminRecipesH := handler.NewAdminRecipeHandler(rs)
	mux.Handle("GET /api/v1/admin/schema",
		handler.RequireAuth(handler.RequireAdmin(http.HandlerFunc(adminRecipesH.ExportSchema))))
	mux.Handle("GET /api/v1/admin/recipes/export",
		handler.RequireAuth(handler.RequireAdmin(http.HandlerFunc(adminRecipesH.ExportRecipes))))
	mux.Handle("POST /api/v1/admin/recipes/import",
		handler.RequireAuth(handler.RequireAdmin(http.HandlerFunc(adminRecipesH.ImportRecipes))))

	return mux
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
	}
}
