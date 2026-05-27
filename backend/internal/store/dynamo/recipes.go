package dynamo

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/almc/cocktails/internal/model"
)

type RecipeStore struct {
	client    *dynamodb.Client
	tableName string
}

func NewRecipeStore(client *dynamodb.Client, tableName string) *RecipeStore {
	return &RecipeStore{client: client, tableName: tableName}
}

type recipeItem struct {
	ID          string            `dynamodbav:"id"`
	Name        string            `dynamodbav:"name"`
	Ingredients []ingItem         `dynamodbav:"ingredients"`
	Steps       []string          `dynamodbav:"steps"`
	Properties  map[string]string `dynamodbav:"properties"`
	Notes       string            `dynamodbav:"notes"`
	Garnishes   []string          `dynamodbav:"garnishes"`
	CreatorID   string            `dynamodbav:"creator_id"`
	CreatedAt   string            `dynamodbav:"created_at"`
	UpdatedAt   string            `dynamodbav:"updated_at"`
}

type ingItem struct {
	Name         string `dynamodbav:"name"`
	Quantity     string `dynamodbav:"quantity"`
	Unit         string `dynamodbav:"unit"`
	BaseSpirit   bool   `dynamodbav:"is_base_spirit,omitempty"`
}

func (s *RecipeStore) Create(r *model.Recipe) error {
	item := toItem(r)
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

func (s *RecipeStore) GetByID(id string) (*model.Recipe, error) {
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
		return nil, errors.New("recipe not found")
	}
	return unmarshalRecipe(out.Item)
}

func (s *RecipeStore) List(page, limit int) ([]*model.Recipe, int, error) {
	out, err := s.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		return nil, 0, err
	}
	all, err := scanToRecipes(out.Items)
	if err != nil {
		return nil, 0, err
	}
	total := len(all)
	start := (page - 1) * limit
	if start >= total {
		return []*model.Recipe{}, total, nil
	}
	end := start + limit
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func (s *RecipeStore) Search(query string, page, limit int) ([]*model.Recipe, int, error) {
	q := strings.ToLower(query)
	out, err := s.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		return nil, 0, err
	}
	all, err := scanToRecipes(out.Items)
	if err != nil {
		return nil, 0, err
	}
	var matches []*model.Recipe
	for _, r := range all {
		if matchesQuery(r, q) {
			matches = append(matches, r)
		}
	}
	total := len(matches)
	start := (page - 1) * limit
	if start >= total {
		return []*model.Recipe{}, total, nil
	}
	end := start + limit
	if end > total {
		end = total
	}
	return matches[start:end], total, nil
}

func (s *RecipeStore) Random() (*model.Recipe, error) {
	out, err := s.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		return nil, err
	}
	if len(out.Items) == 0 {
		return nil, nil
	}
	item := out.Items[rand.Intn(len(out.Items))]
	return unmarshalRecipe(item)
}

func (s *RecipeStore) Update(r *model.Recipe) error {
	r.UpdatedAt = time.Now().UTC()
	return s.Create(r)
}

func (s *RecipeStore) Delete(id string) error {
	_, err := s.client.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}

func (s *RecipeStore) ExistsByName(name string) (bool, error) {
	n := strings.ToLower(name)
	out, err := s.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		return false, err
	}
	for _, av := range out.Items {
		var item recipeItem
		if err := attributevalue.UnmarshalMap(av, &item); err != nil {
			continue
		}
		if strings.ToLower(item.Name) == n {
			return true, nil
		}
	}
	return false, nil
}

func (s *RecipeStore) ListByCreator(creatorID string, page, limit int) ([]*model.Recipe, int, error) {
	out, err := s.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName:        aws.String(s.tableName),
		FilterExpression: aws.String("creator_id = :cid AND creator_id <> :empty"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":cid":   &types.AttributeValueMemberS{Value: creatorID},
			":empty": &types.AttributeValueMemberS{Value: ""},
		},
	})
	if err != nil {
		return nil, 0, err
	}
	all, err := scanToRecipes(out.Items)
	if err != nil {
		return nil, 0, err
	}
	total := len(all)
	start := (page - 1) * limit
	if start >= total {
		return []*model.Recipe{}, total, nil
	}
	end := start + limit
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func (s *RecipeStore) ListAll() ([]*model.Recipe, error) {
	out, err := s.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		return nil, err
	}
	return scanToRecipes(out.Items)
}

func (s *RecipeStore) ImportBatch(recipes []*model.Recipe, _ string) (int, int, error) {
	created, skipped := 0, 0
	var createdIDs []string
	for _, r := range recipes {
		exists, err := s.ExistsByName(r.Name)
		if err != nil {
			// Best-effort compensation: delete items created so far
			for _, id := range createdIDs {
				_ = s.Delete(id)
			}
			return 0, 0, err
		}
		if exists {
			skipped++
			continue
		}
		if err := s.Create(r); err != nil {
			for _, id := range createdIDs {
				_ = s.Delete(id)
			}
			return 0, 0, err
		}
		createdIDs = append(createdIDs, r.ID)
		created++
	}
	return created, skipped, nil
}

func (s *RecipeStore) SearchByIngredients(ingredients []string, page, limit int) ([]*model.Recipe, int, error) {
	if len(ingredients) == 0 {
		return s.List(page, limit)
	}
	out, err := s.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		return nil, 0, err
	}
	all, err := scanToRecipes(out.Items)
	if err != nil {
		return nil, 0, err
	}
	var matches []*model.Recipe
	for _, r := range all {
		if matchesAllIngredients(r, ingredients) {
			matches = append(matches, r)
		}
	}
	total := len(matches)
	start := (page - 1) * limit
	if start >= total {
		return []*model.Recipe{}, total, nil
	}
	end := start + limit
	if end > total {
		end = total
	}
	return matches[start:end], total, nil
}

func matchesBaseSpirit(r *model.Recipe, q string) bool {
	for _, ing := range r.Ingredients {
		if ing.IsBaseSpirit && strings.Contains(strings.ToLower(ing.Name), q) {
			return true
		}
	}
	return false
}

func (s *RecipeStore) SearchByBaseSpirit(baseSpirit string, page, limit int) ([]*model.Recipe, int, error) {
	if strings.TrimSpace(baseSpirit) == "" {
		return s.List(page, limit)
	}
	q := strings.ToLower(baseSpirit)
	out, err := s.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		return nil, 0, err
	}
	all, err := scanToRecipes(out.Items)
	if err != nil {
		return nil, 0, err
	}
	var matches []*model.Recipe
	for _, r := range all {
		if matchesBaseSpirit(r, q) {
			matches = append(matches, r)
		}
	}
	total := len(matches)
	start := (page - 1) * limit
	if start >= total {
		return []*model.Recipe{}, total, nil
	}
	end := start + limit
	if end > total {
		end = total
	}
	return matches[start:end], total, nil
}

func (s *RecipeStore) SearchByBaseSpiritAndIngredients(baseSpirit string, ingredients []string, page, limit int) ([]*model.Recipe, int, error) {
	q := strings.ToLower(baseSpirit)
	out, err := s.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		return nil, 0, err
	}
	all, err := scanToRecipes(out.Items)
	if err != nil {
		return nil, 0, err
	}
	var matches []*model.Recipe
	for _, r := range all {
		if matchesBaseSpirit(r, q) && matchesAllIngredients(r, ingredients) {
			matches = append(matches, r)
		}
	}
	total := len(matches)
	start := (page - 1) * limit
	if start >= total {
		return []*model.Recipe{}, total, nil
	}
	end := start + limit
	if end > total {
		end = total
	}
	return matches[start:end], total, nil
}

func matchesAllIngredients(r *model.Recipe, ingredients []string) bool {
	for _, token := range ingredients {
		t := strings.ToLower(token)
		found := false
		for _, ing := range r.Ingredients {
			if strings.Contains(strings.ToLower(ing.Name), t) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func toItem(r *model.Recipe) recipeItem {
	ings := make([]ingItem, len(r.Ingredients))
	for i, ing := range r.Ingredients {
		ings[i] = ingItem{Name: ing.Name, Quantity: ing.Quantity, Unit: ing.Unit, BaseSpirit: ing.IsBaseSpirit}
	}
	return recipeItem{
		ID:          r.ID,
		Name:        r.Name,
		Ingredients: ings,
		Steps:       r.Steps,
		Properties:  r.Properties,
		Notes:       r.Notes,
		Garnishes:   r.Garnishes,
		CreatorID:   r.CreatorID,
		CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:   r.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func unmarshalRecipe(av map[string]types.AttributeValue) (*model.Recipe, error) {
	var item recipeItem
	if err := attributevalue.UnmarshalMap(av, &item); err != nil {
		return nil, err
	}
	ings := make([]model.Ingredient, len(item.Ingredients))
	for i, ing := range item.Ingredients {
		ings[i] = model.Ingredient{Name: ing.Name, Quantity: ing.Quantity, Unit: ing.Unit, IsBaseSpirit: ing.BaseSpirit}
	}
	createdAt, _ := time.Parse(time.RFC3339Nano, item.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339Nano, item.UpdatedAt)
	return &model.Recipe{
		ID:          item.ID,
		Name:        item.Name,
		Ingredients: ings,
		Steps:       item.Steps,
		Properties:  item.Properties,
		Notes:       item.Notes,
		Garnishes:   item.Garnishes,
		CreatorID:   item.CreatorID,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

func scanToRecipes(items []map[string]types.AttributeValue) ([]*model.Recipe, error) {
	recipes := make([]*model.Recipe, 0, len(items))
	for _, av := range items {
		r, err := unmarshalRecipe(av)
		if err != nil {
			return nil, err
		}
		recipes = append(recipes, r)
	}
	return recipes, nil
}

func matchesQuery(r *model.Recipe, q string) bool {
	if strings.Contains(strings.ToLower(r.Name), q) {
		return true
	}
	for _, ing := range r.Ingredients {
		if strings.Contains(strings.ToLower(ing.Name), q) {
			return true
		}
	}
	for _, step := range r.Steps {
		if strings.Contains(strings.ToLower(step), q) {
			return true
		}
	}
	for _, v := range r.Properties {
		if strings.Contains(strings.ToLower(v), q) {
			return true
		}
	}
	return false
}
