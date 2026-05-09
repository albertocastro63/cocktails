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
