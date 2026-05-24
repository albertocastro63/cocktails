package dynamo

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/almc/cocktails/internal/model"
)

type FavoriteStore struct {
	client    *dynamodb.Client
	tableName string
}

func NewFavoriteStore(client *dynamodb.Client, tableName string) *FavoriteStore {
	return &FavoriteStore{client: client, tableName: tableName}
}

type favItem struct {
	UserID    string `dynamodbav:"user_id"`
	RecipeID  string `dynamodbav:"recipe_id"`
	CreatedAt string `dynamodbav:"created_at"`
}

func (s *FavoriteStore) Add(userID, recipeID string) error {
	item := favItem{
		UserID:    userID,
		RecipeID:  recipeID,
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}
	_, err = s.client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      av,
	})
	return err
}

func (s *FavoriteStore) Remove(userID, recipeID string) error {
	_, err := s.client.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"user_id":   &types.AttributeValueMemberS{Value: userID},
			"recipe_id": &types.AttributeValueMemberS{Value: recipeID},
		},
	})
	return err
}

func (s *FavoriteStore) IsFavorite(userID, recipeID string) (bool, error) {
	out, err := s.client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"user_id":   &types.AttributeValueMemberS{Value: userID},
			"recipe_id": &types.AttributeValueMemberS{Value: recipeID},
		},
	})
	if err != nil {
		return false, err
	}
	return out.Item != nil, nil
}

func (s *FavoriteStore) ListByUser(userID string) ([]*model.Favorite, error) {
	out, err := s.client.Query(context.Background(), &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		KeyConditionExpression: aws.String("user_id = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: userID},
		},
		ScanIndexForward: aws.Bool(false),
	})
	if err != nil {
		return nil, err
	}

	favs := make([]*model.Favorite, 0, len(out.Items))
	for _, item := range out.Items {
		var fi favItem
		if err := attributevalue.UnmarshalMap(item, &fi); err != nil {
			return nil, err
		}
		t, _ := time.Parse(time.RFC3339Nano, fi.CreatedAt)
		favs = append(favs, &model.Favorite{
			UserID:    fi.UserID,
			RecipeID:  fi.RecipeID,
			CreatedAt: t,
		})
	}
	return favs, nil
}

// P3: replace with Query on recipe_id GSI when favorite count UI is implemented
func (s *FavoriteStore) CountByRecipe(recipeID string) (int, error) {
	return 0, nil
}
