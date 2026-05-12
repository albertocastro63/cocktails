package dynamo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/almc/cocktails/internal/model"
	"github.com/almc/cocktails/internal/store"
)

type UserStore struct {
	client    *dynamodb.Client
	tableName string
}

func NewUserStore(client *dynamodb.Client, tableName string) *UserStore {
	return &UserStore{client: client, tableName: tableName}
}

type userItem struct {
	ID           string `dynamodbav:"id"`
	Username     string `dynamodbav:"username"`
	PasswordHash string `dynamodbav:"password_hash"`
	IsAdmin      bool   `dynamodbav:"is_admin"`
	FirstName    string `dynamodbav:"first_name"`
	LastName     string `dynamodbav:"last_name"`
	Email        string `dynamodbav:"email"`
	TokenVersion int    `dynamodbav:"token_version"`
	CreatedAt    string `dynamodbav:"created_at"`
}

func (s *UserStore) Create(u *model.User) error {
	item := userItem{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		IsAdmin:      u.IsAdmin,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Email:        u.Email,
		TokenVersion: u.TokenVersion,
		CreatedAt:    u.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}
	_, err = s.client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName:           aws.String(s.tableName),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	})
	if err != nil {
		var cfe *types.ConditionalCheckFailedException
		if errors.As(err, &cfe) {
			return store.ErrDuplicate
		}
		if strings.Contains(err.Error(), "ConditionalCheckFailedException") {
			return store.ErrDuplicate
		}
		return err
	}
	return nil
}

func (s *UserStore) GetByID(id string) (*model.User, error) {
	out, err := s.client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, err
	}
	if out.Item == nil {
		return nil, errors.New("user not found")
	}
	return unmarshalUser(out.Item)
}

func (s *UserStore) GetByUsername(username string) (*model.User, error) {
	out, err := s.client.Query(context.Background(), &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		IndexName:              aws.String("username-index"),
		KeyConditionExpression: aws.String("username = :u"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":u": &types.AttributeValueMemberS{Value: username},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}
	if len(out.Items) == 0 {
		return nil, errors.New("user not found")
	}
	return unmarshalUser(out.Items[0])
}

func (s *UserStore) Count() (int, error) {
	out, err := s.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName: aws.String(s.tableName),
		Select:    types.SelectCount,
	})
	if err != nil {
		return 0, err
	}
	return int(out.Count), nil
}

func (s *UserStore) List() ([]*model.User, error) {
	out, err := s.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName:        aws.String(s.tableName),
		FilterExpression: aws.String("is_admin = :false"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":false": &types.AttributeValueMemberBOOL{Value: false},
		},
	})
	if err != nil {
		return nil, err
	}
	users := make([]*model.User, 0, len(out.Items))
	for _, item := range out.Items {
		u, err := unmarshalUser(item)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *UserStore) Update(u *model.User) error {
	_, err := s.client.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: u.ID},
		},
		UpdateExpression: aws.String("SET first_name = :fn, last_name = :ln, #em = :e, password_hash = :ph, token_version = :tv"),
		ExpressionAttributeNames: map[string]string{
			"#em": "email",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":fn": &types.AttributeValueMemberS{Value: u.FirstName},
			":ln": &types.AttributeValueMemberS{Value: u.LastName},
			":e":  &types.AttributeValueMemberS{Value: u.Email},
			":ph": &types.AttributeValueMemberS{Value: u.PasswordHash},
			":tv": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", u.TokenVersion)},
		},
		ConditionExpression: aws.String("attribute_exists(id)"),
	})
	if err != nil {
		var cfe *types.ConditionalCheckFailedException
		if errors.As(err, &cfe) {
			return fmt.Errorf("user %q not found", u.ID)
		}
		return err
	}
	return nil
}

func (s *UserStore) Delete(id string) error {
	_, err := s.client.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		ConditionExpression: aws.String("attribute_exists(id)"),
	})
	if err != nil {
		var cfe *types.ConditionalCheckFailedException
		if errors.As(err, &cfe) {
			return fmt.Errorf("user %q not found", id)
		}
		return err
	}
	return nil
}

func (s *UserStore) GetByEmail(email string) (*model.User, error) {
	out, err := s.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName:        aws.String(s.tableName),
		FilterExpression: aws.String("#em = :e"),
		ExpressionAttributeNames: map[string]string{
			"#em": "email",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":e": &types.AttributeValueMemberS{Value: email},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}
	if len(out.Items) == 0 {
		return nil, fmt.Errorf("user with email %q not found", email)
	}
	return unmarshalUser(out.Items[0])
}

func unmarshalUser(av map[string]types.AttributeValue) (*model.User, error) {
	var item userItem
	if err := attributevalue.UnmarshalMap(av, &item); err != nil {
		return nil, err
	}
	createdAt, _ := time.Parse(time.RFC3339Nano, item.CreatedAt)
	return &model.User{
		ID:           item.ID,
		Username:     item.Username,
		PasswordHash: item.PasswordHash,
		IsAdmin:      item.IsAdmin,
		FirstName:    item.FirstName,
		LastName:     item.LastName,
		Email:        item.Email,
		TokenVersion: item.TokenVersion,
		CreatedAt:    createdAt,
	}, nil
}
