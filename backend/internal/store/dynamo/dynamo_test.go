package dynamo_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"github.com/almc/cocktails/internal/model"
	dynstore "github.com/almc/cocktails/internal/store/dynamo"
)

const (
	testRecipesTable = "test-recipes"
	testUsersTable   = "test-users"
)

func newTestClient(t *testing.T) *dynamodb.Client {
	t.Helper()
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		t.Skip("DYNAMODB_ENDPOINT not set; skipping DynamoDB tests")
	}
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithBaseEndpoint(endpoint),
	)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	return dynamodb.NewFromConfig(cfg)
}

func createTable(t *testing.T, client *dynamodb.Client, name string, gsi bool) {
	t.Helper()
	input := &dynamodb.CreateTableInput{
		TableName: aws.String(name),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		BillingMode: types.BillingModePayPerRequest,
	}
	if gsi {
		input.AttributeDefinitions = append(input.AttributeDefinitions,
			types.AttributeDefinition{AttributeName: aws.String("username"), AttributeType: types.ScalarAttributeTypeS},
		)
		input.GlobalSecondaryIndexes = []types.GlobalSecondaryIndex{{
			IndexName: aws.String("username-index"),
			KeySchema: []types.KeySchemaElement{
				{AttributeName: aws.String("username"), KeyType: types.KeyTypeHash},
			},
			Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
		}}
	}
	_, err := client.CreateTable(context.Background(), input)
	if err != nil {
		t.Logf("create table %s: %v (may already exist)", name, err)
	}
	t.Cleanup(func() {
		client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{TableName: aws.String(name)})
	})
}

func TestDynamo_RecipeStore(t *testing.T) {
	client := newTestClient(t)
	table := testRecipesTable + "-" + uuid.NewString()[:8]
	createTable(t, client, table, false)
	rs := dynstore.NewRecipeStore(client, table)

	recipe := &model.Recipe{
		ID:          uuid.NewString(),
		Name:        "Mojito",
		Ingredients: []model.Ingredient{{Name: "rum", Quantity: "50", Unit: "ml"}},
		Steps:       []string{"muddle", "shake"},
		Properties:  map[string]string{"style": "refreshing"},
		CreatorID:   "u1",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := rs.Create(recipe); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := rs.GetByID(recipe.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Mojito" {
		t.Errorf("name: got %q want Mojito", got.Name)
	}
	if got.Properties["style"] != "refreshing" {
		t.Errorf("property: got %q", got.Properties["style"])
	}

	results, _, err := rs.Search("refreshing", 1, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Error("search by property value returned nothing")
	}

	recipe.Name = "Updated Mojito"
	if err := rs.Update(recipe); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if err := rs.Delete(recipe.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := rs.GetByID(recipe.ID); err == nil {
		t.Error("expected error after delete")
	}
}

func TestDynamo_RecipeStore_Extended(t *testing.T) {
	client := newTestClient(t)
	table := testRecipesTable + "-ext-" + uuid.NewString()[:8]
	createTable(t, client, table, false)
	rs := dynstore.NewRecipeStore(client, table)

	r1 := &model.Recipe{ID: uuid.NewString(), Name: "Mojito", CreatorID: "u1", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	r2 := &model.Recipe{ID: uuid.NewString(), Name: "Daiquiri", CreatorID: "u1", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	r3 := &model.Recipe{ID: uuid.NewString(), Name: "Old Fashioned", CreatorID: "u2", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	for _, r := range []*model.Recipe{r1, r2, r3} {
		if err := rs.Create(r); err != nil {
			t.Fatalf("Create %q: %v", r.Name, err)
		}
	}

	recipes, total, err := rs.List(1, 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total < 3 || len(recipes) < 3 {
		t.Errorf("List: got %d/%d, want >=3", len(recipes), total)
	}

	exists, err := rs.ExistsByName("Mojito")
	if err != nil {
		t.Fatalf("ExistsByName (exists): %v", err)
	}
	if !exists {
		t.Error("expected ExistsByName=true for Mojito")
	}
	exists, err = rs.ExistsByName("NonExistent")
	if err != nil {
		t.Fatalf("ExistsByName (not found): %v", err)
	}
	if exists {
		t.Error("expected ExistsByName=false for NonExistent")
	}

	mine, mineTotal, err := rs.ListByCreator("u1", 1, 10)
	if err != nil {
		t.Fatalf("ListByCreator: %v", err)
	}
	if mineTotal != 2 || len(mine) != 2 {
		t.Errorf("ListByCreator: got %d/%d, want 2", len(mine), mineTotal)
	}

	// Out-of-range page returns empty slice
	emptyPage, _, err := rs.List(999, 10)
	if err != nil {
		t.Fatalf("List (out-of-range page): %v", err)
	}
	if len(emptyPage) != 0 {
		t.Errorf("expected empty page, got %d", len(emptyPage))
	}

	all, err := rs.ListAll()
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) < 3 {
		t.Errorf("ListAll: expected >=3, got %d", len(all))
	}

	rnd, err := rs.Random()
	if err != nil {
		t.Fatalf("Random: %v", err)
	}
	if rnd == nil {
		t.Error("Random returned nil with items present")
	}

	imported, skipped, err := rs.ImportBatch([]*model.Recipe{
		{ID: uuid.NewString(), Name: "Mojito", CreatorID: "u1"},   // duplicate name → skip
		{ID: uuid.NewString(), Name: "Negroni", CreatorID: "u2"},  // new → import
	}, "u3")
	if err != nil {
		t.Fatalf("ImportBatch: %v", err)
	}
	if imported != 1 || skipped != 1 {
		t.Errorf("ImportBatch: imported=%d skipped=%d, want 1/1", imported, skipped)
	}
}

func TestDynamo_UserStore_Extended(t *testing.T) {
	client := newTestClient(t)
	table := testUsersTable + "-ext-" + uuid.NewString()[:8]
	createTable(t, client, table, true)
	us := dynstore.NewUserStore(client, table)

	user := &model.User{
		ID:           uuid.NewString(),
		Username:     "bob-" + uuid.NewString()[:8],
		PasswordHash: "$2a$12$placeholder",
		IsAdmin:      false,
		CreatedAt:    time.Now().UTC(),
	}
	if err := us.Create(user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := us.GetByID(user.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Username != user.Username {
		t.Errorf("GetByID: username mismatch, got %q", got.Username)
	}

	users, err := us.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(users) < 1 {
		t.Error("List: expected at least 1 user")
	}

	// Update only touches first_name, last_name, email, password_hash, token_version
	user.FirstName = "Bobby"
	user.PasswordHash = "$2a$12$updated"
	user.TokenVersion = 2
	if err := us.Update(user); err != nil {
		t.Fatalf("Update: %v", err)
	}
	updated, err := us.GetByID(user.ID)
	if err != nil {
		t.Fatalf("GetByID after Update: %v", err)
	}
	if updated.FirstName != "Bobby" {
		t.Errorf("expected FirstName=Bobby after Update, got %q", updated.FirstName)
	}
	if updated.TokenVersion != 2 {
		t.Errorf("expected TokenVersion=2 after Update, got %d", updated.TokenVersion)
	}

	if err := us.Delete(user.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := us.GetByID(user.ID); err == nil {
		t.Error("expected error after Delete")
	}
}

func TestDynamo_FavoriteStore(t *testing.T) {
	client := newTestClient(t)
	table := "test-favorites-" + uuid.NewString()[:8]

	_, err := client.CreateTable(context.Background(), &dynamodb.CreateTableInput{
		TableName: aws.String(table),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("user_id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("recipe_id"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("user_id"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("recipe_id"), KeyType: types.KeyTypeRange},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Logf("create favorites table: %v (may already exist)", err)
	}
	t.Cleanup(func() {
		client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{TableName: aws.String(table)}) //nolint:errcheck
	})

	fs := dynstore.NewFavoriteStore(client, table)

	userID := uuid.NewString()
	recipeID := uuid.NewString()

	if err := fs.Add(userID, recipeID); err != nil {
		t.Fatalf("Add: %v", err)
	}

	isFav, err := fs.IsFavorite(userID, recipeID)
	if err != nil {
		t.Fatalf("IsFavorite: %v", err)
	}
	if !isFav {
		t.Error("expected IsFavorite=true after Add")
	}

	isFav, err = fs.IsFavorite("other-user", recipeID)
	if err != nil {
		t.Fatalf("IsFavorite (other user): %v", err)
	}
	if isFav {
		t.Error("expected IsFavorite=false for a different user")
	}

	favs, err := fs.ListByUser(userID)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(favs) != 1 || favs[0].RecipeID != recipeID {
		t.Errorf("ListByUser: got %d items, recipeID=%q", len(favs), func() string {
			if len(favs) > 0 {
				return favs[0].RecipeID
			}
			return ""
		}())
	}

	count, err := fs.CountByRecipe(recipeID)
	if err != nil {
		t.Fatalf("CountByRecipe: %v", err)
	}
	if count != 0 {
		t.Errorf("CountByRecipe: expected 0 (stub), got %d", count)
	}

	if err := fs.Remove(userID, recipeID); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	isFav, err = fs.IsFavorite(userID, recipeID)
	if err != nil {
		t.Fatalf("IsFavorite after Remove: %v", err)
	}
	if isFav {
		t.Error("expected IsFavorite=false after Remove")
	}

	favs, err = fs.ListByUser(userID)
	if err != nil {
		t.Fatalf("ListByUser after Remove: %v", err)
	}
	if len(favs) != 0 {
		t.Errorf("expected 0 favorites after Remove, got %d", len(favs))
	}
}

func TestDynamo_UserStore(t *testing.T) {
	client := newTestClient(t)
	table := testUsersTable + "-" + uuid.NewString()[:8]
	createTable(t, client, table, true)
	us := dynstore.NewUserStore(client, table)

	user := &model.User{
		ID:           uuid.NewString(),
		Username:     "alice-" + uuid.NewString()[:8],
		PasswordHash: "$2a$12$placeholder",
		IsAdmin:      false,
		CreatedAt:    time.Now().UTC(),
	}

	if err := us.Create(user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := us.GetByUsername(user.Username)
	if err != nil {
		t.Fatalf("GetByUsername: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("id mismatch")
	}

	n, err := us.Count()
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n < 1 {
		t.Errorf("expected count ≥ 1, got %d", n)
	}
}
