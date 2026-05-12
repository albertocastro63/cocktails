package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/almc/cocktails/internal/handler"
	"github.com/almc/cocktails/internal/model"
	"github.com/almc/cocktails/internal/store"
	dynstore "github.com/almc/cocktails/internal/store/dynamo"
	sqstore "github.com/almc/cocktails/internal/store/sqlite"
	"github.com/google/uuid"
)

func main() {
	var recipeStore store.RecipeStore
	var userStore store.UserStore

	switch envOr("STORE_BACKEND", "sqlite") {
	case "dynamodb":
		cfg, err := awscfg.LoadDefaultConfig(context.Background())
		if err != nil {
			log.Fatalf("load aws config: %v", err)
		}
		client := dynamodb.NewFromConfig(cfg)
		recipeStore = dynstore.NewRecipeStore(client, envOr("RECIPES_TABLE", "cocktails-recipes"))
		userStore = dynstore.NewUserStore(client, envOr("USERS_TABLE", "cocktails-users"))
	default:
		dbPath := envOr("DB_PATH", "cocktails.db")
		rs, us, err := sqstore.Open(dbPath)
		if err != nil {
			log.Fatalf("open store: %v", err)
		}
		recipeStore = rs
		userStore = us
	}

	bootstrapAdmin(userStore)

	h := buildHandler(recipeStore, userStore)

	port := envOr("PORT", "8080")
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, h))
}

func buildHandler(rs store.RecipeStore, us store.UserStore) http.Handler {
	recipes := handler.NewRecipeHandler(rs)
	authH := handler.NewAuthHandler(us)
	adminH := handler.NewAdminHandler(us)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/recipes", recipes.List)
	mux.HandleFunc("GET /api/v1/recipes/random", recipes.Random)
	mux.HandleFunc("GET /api/v1/recipes/{id}", recipes.GetByID)
	requireAuth := handler.RequireAuthWithStore(us)
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
