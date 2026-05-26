package dynamo

import (
	"testing"

	"github.com/almc/cocktails/internal/model"
)

func makeRecipeWithIngredients(ings []model.Ingredient) *model.Recipe {
	return &model.Recipe{Ingredients: ings}
}

func TestMatchesBaseSpirit(t *testing.T) {
	gin := model.Ingredient{Name: "Gin", IsBaseSpirit: true}
	lime := model.Ingredient{Name: "Lime Juice", IsBaseSpirit: false}
	rum := model.Ingredient{Name: "Rum", IsBaseSpirit: true}

	cases := []struct {
		name     string
		recipe   *model.Recipe
		q        string
		expected bool
	}{
		{"exact match base spirit", makeRecipeWithIngredients([]model.Ingredient{gin, lime}), "gin", true},
		{"partial match base spirit", makeRecipeWithIngredients([]model.Ingredient{rum}), "ru", true},
		{"ingredient name compared case-insensitively", makeRecipeWithIngredients([]model.Ingredient{{Name: "GIN", IsBaseSpirit: true}}), "gin", true},
		{"non-base spirit ingredient not matched", makeRecipeWithIngredients([]model.Ingredient{lime}), "lime", false},
		{"no match", makeRecipeWithIngredients([]model.Ingredient{gin}), "rum", false},
		{"empty recipe", makeRecipeWithIngredients(nil), "gin", false},
		{"mixed: base spirit wins", makeRecipeWithIngredients([]model.Ingredient{lime, gin}), "gin", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := matchesBaseSpirit(tc.recipe, tc.q)
			if got != tc.expected {
				t.Errorf("matchesBaseSpirit(%q) = %v, want %v", tc.q, got, tc.expected)
			}
		})
	}
}

func TestMatchesAllIngredients(t *testing.T) {
	ings := []model.Ingredient{
		{Name: "Gin", IsBaseSpirit: true},
		{Name: "Lime Juice"},
		{Name: "Sugar Syrup"},
	}
	recipe := makeRecipeWithIngredients(ings)

	cases := []struct {
		name        string
		ingredients []string
		expected    bool
	}{
		{"single match", []string{"gin"}, true},
		{"multiple match", []string{"gin", "lime"}, true},
		{"all match", []string{"gin", "lime juice", "sugar"}, true},
		{"one missing", []string{"gin", "bitters"}, false},
		{"case insensitive", []string{"GIN", "LIME JUICE"}, true},
		{"empty tokens always matches", []string{}, true},
		{"partial name match", []string{"sug"}, true},
		{"no match at all", []string{"vodka"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := matchesAllIngredients(recipe, tc.ingredients)
			if got != tc.expected {
				t.Errorf("matchesAllIngredients(%v) = %v, want %v", tc.ingredients, got, tc.expected)
			}
		})
	}
}
