package dynamo_test

import (
	"testing"
	"time"

	"github.com/almc/cocktails/internal/model"
	dynstore "github.com/almc/cocktails/internal/store/dynamo"
	"github.com/google/uuid"
)

func makeDynamoRecipe(name string, ingredients ...string) *model.Recipe {
	var ings []model.Ingredient
	for _, n := range ingredients {
		ings = append(ings, model.Ingredient{Name: n, Quantity: "30", Unit: "ml"})
	}
	return &model.Recipe{
		ID:          uuid.NewString(),
		Name:        name,
		Ingredients: ings,
		Steps:       []string{"mix"},
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
}

func TestDynamo_SearchByIngredients_TwoIngredients(t *testing.T) {
	client := newTestClient(t)
	table := testRecipesTable + "-ing-" + uuid.NewString()[:8]
	createTable(t, client, table, false)
	rs := dynstore.NewRecipeStore(client, table)

	if err := rs.Create(makeDynamoRecipe("Gin Fizz", "Gin", "Lemon Juice", "Sugar")); err != nil {
		t.Fatal(err)
	}
	if err := rs.Create(makeDynamoRecipe("Daiquiri", "Rum", "Lime Juice", "Sugar")); err != nil {
		t.Fatal(err)
	}
	if err := rs.Create(makeDynamoRecipe("Bee's Knees", "Gin", "Lemon Juice", "Honey")); err != nil {
		t.Fatal(err)
	}

	got, total, err := rs.SearchByIngredients([]string{"gin", "lemon juice"}, 1, 20)
	if err != nil {
		t.Fatalf("SearchByIngredients: %v", err)
	}
	if total != 2 {
		t.Errorf("total: got %d want 2", total)
	}
	if len(got) != 2 {
		t.Errorf("len: got %d want 2", len(got))
	}
	for _, r := range got {
		if r.Name == "Daiquiri" {
			t.Error("Daiquiri should not match (no gin)")
		}
	}
}

func TestDynamo_SearchByIngredients_NoMatch(t *testing.T) {
	client := newTestClient(t)
	table := testRecipesTable + "-ing-nomatch-" + uuid.NewString()[:8]
	createTable(t, client, table, false)
	rs := dynstore.NewRecipeStore(client, table)

	if err := rs.Create(makeDynamoRecipe("Daiquiri", "Rum", "Lime Juice")); err != nil {
		t.Fatal(err)
	}

	got, total, err := rs.SearchByIngredients([]string{"gin", "rum"}, 1, 20)
	if err != nil {
		t.Fatalf("SearchByIngredients: %v", err)
	}
	if total != 0 || len(got) != 0 {
		t.Errorf("expected no results, got %d", len(got))
	}
}

func TestDynamo_SearchByIngredients_CaseInsensitive(t *testing.T) {
	client := newTestClient(t)
	table := testRecipesTable + "-ing-case-" + uuid.NewString()[:8]
	createTable(t, client, table, false)
	rs := dynstore.NewRecipeStore(client, table)

	if err := rs.Create(makeDynamoRecipe("Gin Sour", "Gin", "Lemon Juice")); err != nil {
		t.Fatal(err)
	}

	got, _, err := rs.SearchByIngredients([]string{"GIN", "LEMON JUICE"}, 1, 20)
	if err != nil {
		t.Fatalf("SearchByIngredients: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 result for uppercase tokens, got %d", len(got))
	}
}
