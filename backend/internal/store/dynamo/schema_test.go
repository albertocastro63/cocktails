package dynamo_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"

	dynstore "github.com/almc/cocktails/internal/store/dynamo"
)

func TestDynamo_EnsureSchema_CreatesExpectedShape(t *testing.T) {
	client := testClient(t)
	s := uuid.NewString()[:8]
	names := dynstore.TableNames{
		Recipes:   "schema-recipes-" + s,
		Users:     "schema-users-" + s,
		Favorites: "schema-favorites-" + s,
	}
	t.Cleanup(func() {
		for _, n := range []string{names.Recipes, names.Users, names.Favorites} {
			_, _ = client.DeleteTable(context.Background(),
				&dynamodb.DeleteTableInput{TableName: aws.String(n)})
		}
	})

	if err := dynstore.EnsureSchema(context.Background(), client, names); err != nil {
		t.Fatalf("EnsureSchema: %v", err)
	}

	// users table must expose the username-index GSI (login depends on it).
	desc, err := client.DescribeTable(context.Background(),
		&dynamodb.DescribeTableInput{TableName: aws.String(names.Users)})
	if err != nil {
		t.Fatalf("DescribeTable(users): %v", err)
	}
	if !hasGSI(desc, "username-index") {
		t.Errorf("users table missing username-index GSI; got %v", gsiNames(desc))
	}

	// favorites table must expose the recipe_id-index GSI.
	descFav, err := client.DescribeTable(context.Background(),
		&dynamodb.DescribeTableInput{TableName: aws.String(names.Favorites)})
	if err != nil {
		t.Fatalf("DescribeTable(favorites): %v", err)
	}
	if !hasGSI(descFav, "recipe_id-index") {
		t.Errorf("favorites table missing recipe_id-index GSI; got %v", gsiNames(descFav))
	}
}

func TestDynamo_EnsureSchema_Idempotent(t *testing.T) {
	client := testClient(t)
	s := uuid.NewString()[:8]
	names := dynstore.TableNames{
		Recipes:   "idem-recipes-" + s,
		Users:     "idem-users-" + s,
		Favorites: "idem-favorites-" + s,
	}
	t.Cleanup(func() {
		for _, n := range []string{names.Recipes, names.Users, names.Favorites} {
			_, _ = client.DeleteTable(context.Background(),
				&dynamodb.DeleteTableInput{TableName: aws.String(n)})
		}
	})

	if err := dynstore.EnsureSchema(context.Background(), client, names); err != nil {
		t.Fatalf("first EnsureSchema: %v", err)
	}
	// Second call against existing tables must be a no-op, not an error.
	if err := dynstore.EnsureSchema(context.Background(), client, names); err != nil {
		t.Fatalf("second EnsureSchema (idempotent) returned error: %v", err)
	}
}

func TestDynamo_EnsureSchema_RejectsEmptyNames(t *testing.T) {
	// No emulator needed: validation happens before any client call.
	err := dynstore.EnsureSchema(context.Background(), nil, dynstore.TableNames{Recipes: "r"})
	if err == nil {
		t.Fatal("expected error when table names are incomplete")
	}
}

func hasGSI(desc *dynamodb.DescribeTableOutput, name string) bool {
	for _, g := range desc.Table.GlobalSecondaryIndexes {
		if aws.ToString(g.IndexName) == name {
			return true
		}
	}
	return false
}

func gsiNames(desc *dynamodb.DescribeTableOutput) []string {
	var out []string
	for _, g := range desc.Table.GlobalSecondaryIndexes {
		out = append(out, aws.ToString(g.IndexName))
	}
	return out
}
