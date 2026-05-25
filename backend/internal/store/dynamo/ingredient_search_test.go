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

func makeDynamoBaseSpiritRecipe(name, baseSpirit string, otherIngredients ...string) *model.Recipe {
	ings := []model.Ingredient{{Name: baseSpirit, IsBaseSpirit: true, Quantity: "50", Unit: "ml"}}
	for _, n := range otherIngredients {
		ings = append(ings, model.Ingredient{Name: n, Quantity: "20", Unit: "ml"})
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

func TestDynamo_SearchByBaseSpirit(t *testing.T) {
	client := newTestClient(t)
	table := testRecipesTable + "-bs-" + uuid.NewString()[:8]
	createTable(t, client, table, false)
	rs := dynstore.NewRecipeStore(client, table)

	recipeA := makeDynamoBaseSpiritRecipe("Gimlet", "gin", "lime juice")
	recipeB := makeDynamoBaseSpiritRecipe("Daiquiri", "rum", "ginger beer")
	recipeC := makeDynamoRecipe("Lemonade", "lemon juice", "sugar")

	for _, r := range []*model.Recipe{recipeA, recipeB, recipeC} {
		if err := rs.Create(r); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	t.Run("returns only gin-based recipe", func(t *testing.T) {
		got, total, err := rs.SearchByBaseSpirit("gin", 1, 20)
		if err != nil {
			t.Fatalf("SearchByBaseSpirit: %v", err)
		}
		if total != 1 || len(got) != 1 || got[0].ID != recipeA.ID {
			t.Errorf("expected recipeA only, got total=%d", total)
		}
	})

	t.Run("case-insensitive", func(t *testing.T) {
		got, total, err := rs.SearchByBaseSpirit("GIN", 1, 20)
		if err != nil {
			t.Fatalf("SearchByBaseSpirit: %v", err)
		}
		if total != 1 || got[0].ID != recipeA.ID {
			t.Errorf("expected recipeA for GIN, got total=%d", total)
		}
	})

	t.Run("no match returns empty", func(t *testing.T) {
		got, total, err := rs.SearchByBaseSpirit("whis", 1, 20)
		if err != nil {
			t.Fatalf("SearchByBaseSpirit: %v", err)
		}
		if total != 0 || len(got) != 0 {
			t.Errorf("expected empty, got total=%d", total)
		}
	})

	t.Run("empty filter falls through to List", func(t *testing.T) {
		_, total, err := rs.SearchByBaseSpirit("", 1, 20)
		if err != nil {
			t.Fatalf("SearchByBaseSpirit empty: %v", err)
		}
		if total != 3 {
			t.Errorf("expected 3, got %d", total)
		}
	})
}

func TestDynamo_SearchByBaseSpiritAndIngredients(t *testing.T) {
	client := newTestClient(t)
	table := testRecipesTable + "-bsi-" + uuid.NewString()[:8]
	createTable(t, client, table, false)
	rs := dynstore.NewRecipeStore(client, table)

	recipeA := makeDynamoBaseSpiritRecipe("Gimlet", "gin", "lime juice")
	recipeB := makeDynamoBaseSpiritRecipe("Dark and Stormy", "rum", "ginger beer")

	for _, r := range []*model.Recipe{recipeA, recipeB} {
		if err := rs.Create(r); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	t.Run("gin base and lime ingredient returns recipeA", func(t *testing.T) {
		got, total, err := rs.SearchByBaseSpiritAndIngredients("gin", []string{"lime"}, 1, 20)
		if err != nil {
			t.Fatalf("SearchByBaseSpiritAndIngredients: %v", err)
		}
		if total != 1 || got[0].ID != recipeA.ID {
			t.Errorf("expected recipeA, got total=%d", total)
		}
	})

	t.Run("gin base but wrong ingredient returns empty", func(t *testing.T) {
		got, total, err := rs.SearchByBaseSpiritAndIngredients("gin", []string{"rum"}, 1, 20)
		if err != nil {
			t.Fatalf("SearchByBaseSpiritAndIngredients: %v", err)
		}
		if total != 0 || len(got) != 0 {
			t.Errorf("expected empty, got total=%d", total)
		}
	})

	t.Run("rum base and ginger ingredient returns recipeB", func(t *testing.T) {
		got, total, err := rs.SearchByBaseSpiritAndIngredients("rum", []string{"ginger"}, 1, 20)
		if err != nil {
			t.Fatalf("SearchByBaseSpiritAndIngredients: %v", err)
		}
		if total != 1 || got[0].ID != recipeB.ID {
			t.Errorf("expected recipeB, got total=%d", total)
		}
	})

	t.Run("rum base but no lime ingredient returns empty", func(t *testing.T) {
		got, total, err := rs.SearchByBaseSpiritAndIngredients("rum", []string{"lime"}, 1, 20)
		if err != nil {
			t.Fatalf("SearchByBaseSpiritAndIngredients: %v", err)
		}
		if total != 0 || len(got) != 0 {
			t.Errorf("expected empty, got total=%d", total)
		}
	})
}
