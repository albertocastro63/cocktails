package dynamo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// TableNames carries the three table names EnsureSchema provisions. Tests pass
// uniquely-suffixed names for isolation; the local server passes the conventional
// names from the environment.
type TableNames struct {
	Recipes   string
	Users     string
	Favorites string
}

func (n TableNames) validate() error {
	if n.Recipes == "" || n.Users == "" || n.Favorites == "" {
		return errors.New("dynamo: recipes, users and favorites table names must all be set")
	}
	return nil
}

// EnsureSchema creates the recipes, users and favorites tables (with their GSIs)
// if they do not already exist, then waits until each is ACTIVE. It is the single
// source of truth for the local/test schema and matches the production key/index
// shape. Idempotent: an existing table is treated as success. Intended for local
// and test environments (DynamoDB Local) — production tables are Terraform-managed.
func EnsureSchema(ctx context.Context, client *dynamodb.Client, names TableNames) error {
	if err := names.validate(); err != nil {
		return err
	}
	for _, in := range []*dynamodb.CreateTableInput{
		recipesTableInput(names.Recipes),
		usersTableInput(names.Users),
		favoritesTableInput(names.Favorites),
	} {
		if err := ensureTable(ctx, client, in); err != nil {
			return err
		}
	}
	return nil
}

func ensureTable(ctx context.Context, client *dynamodb.Client, in *dynamodb.CreateTableInput) error {
	name := aws.ToString(in.TableName)
	_, err := client.CreateTable(ctx, in)
	if err != nil {
		var inUse *types.ResourceInUseException
		if errors.As(err, &inUse) {
			return nil // already exists — idempotent
		}
		return fmt.Errorf("create table %s: %w", name, err)
	}
	if err := dynamodb.NewTableExistsWaiter(client).Wait(ctx,
		&dynamodb.DescribeTableInput{TableName: in.TableName}, 2*time.Minute); err != nil {
		return fmt.Errorf("wait for table %s to become active: %w", name, err)
	}
	return nil
}

func recipesTableInput(name string) *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName:   aws.String(name),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
	}
}

func usersTableInput(name string) *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName:   aws.String(name),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("username"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{{
			IndexName: aws.String("username-index"),
			KeySchema: []types.KeySchemaElement{
				{AttributeName: aws.String("username"), KeyType: types.KeyTypeHash},
			},
			Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
		}},
	}
}

func favoritesTableInput(name string) *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName:   aws.String(name),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("user_id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("recipe_id"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("user_id"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("recipe_id"), KeyType: types.KeyTypeRange},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{{
			IndexName: aws.String("recipe_id-index"),
			KeySchema: []types.KeySchemaElement{
				{AttributeName: aws.String("recipe_id"), KeyType: types.KeyTypeHash},
			},
			Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
		}},
	}
}
