package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"golang.org/x/crypto/bcrypt"

	"github.com/almc/cocktails/internal/handler"
	"github.com/almc/cocktails/internal/model"
	"github.com/almc/cocktails/internal/store"
	sqstore "github.com/almc/cocktails/internal/store/sqlite"
	"github.com/google/uuid"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/tmp/cocktails.db"
	}

	recipeStore, userStore, err := sqstore.Open(dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}

	bootstrapAdmin(userStore)

	h := buildHandler(recipeStore, userStore)
	lambda.Start(httpadapter.New(h).ProxyWithContext)
}

func buildHandler(rs store.RecipeStore, us store.UserStore) http.Handler {
	recipes := handler.NewRecipeHandler(rs)
	authH := handler.NewAuthHandler(us)
	adminH := handler.NewAdminHandler(us)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/recipes", recipes.List)
	mux.HandleFunc("GET /api/v1/recipes/random", recipes.Random)
	mux.HandleFunc("GET /api/v1/recipes/{id}", recipes.GetByID)
	mux.Handle("POST /api/v1/recipes", handler.RequireAuth(http.HandlerFunc(recipes.Create)))
	mux.Handle("PUT /api/v1/recipes/{id}", handler.RequireAuth(http.HandlerFunc(recipes.Update)))
	mux.Handle("DELETE /api/v1/recipes/{id}", handler.RequireAuth(http.HandlerFunc(recipes.Delete)))
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.Handle("POST /api/v1/admin/users",
		handler.RequireAuth(handler.RequireAdmin(http.HandlerFunc(adminH.CreateUser))))

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
