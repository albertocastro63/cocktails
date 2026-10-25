package dynamo

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"

	"github.com/almc/cocktails/internal/model"
)

// seedNamed provisions a store and inserts recipes with the given names.
func seedNamed(t *testing.T, names ...string) *RecipeStore {
	t.Helper()
	if os.Getenv("DYNAMODB_ENDPOINT") == "" {
		t.Fatal("DYNAMODB_ENDPOINT is required for DynamoDB store tests (see specs/029-local-dynamodb-emulator/quickstart.md).")
	}
	client, err := NewClient(context.Background())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	tn := TableNames{
		Recipes:   "test-sort-recipes-" + uuid.NewString()[:8],
		Users:     "test-sort-users-" + uuid.NewString()[:8],
		Favorites: "test-sort-favorites-" + uuid.NewString()[:8],
	}
	if err := EnsureSchema(context.Background(), client, tn); err != nil {
		t.Fatalf("EnsureSchema: %v", err)
	}
	t.Cleanup(func() {
		for _, n := range []string{tn.Recipes, tn.Users, tn.Favorites} {
			_, _ = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{TableName: aws.String(n)})
		}
	})
	rs := NewRecipeStore(client, tn.Recipes)
	for _, name := range names {
		if err := rs.Create(&model.Recipe{ID: uuid.NewString(), Name: name}); err != nil {
			t.Fatalf("Create %q: %v", name, err)
		}
	}
	return rs
}

// TestList_SortNameAscAcrossPages asserts the sort is applied over the full set
// before paging: the last name on page 1 must sort before the first name on
// page 2 (FR-012). With 5 recipes and limit 2, name_asc order is Alpha, Bravo,
// Charlie, Delta, Echo.
func TestList_SortNameAscAcrossPages(t *testing.T) {
	// Insert in a deliberately unsorted order.
	rs := seedNamed(t, "Delta", "Alpha", "Echo", "Bravo", "Charlie")

	page1, total, err := rs.List(1, 2, "name_asc")
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}
	if total != 5 {
		t.Fatalf("total = %d, want 5", total)
	}
	if page1[0].Name != "Alpha" || page1[1].Name != "Bravo" {
		t.Errorf("page 1 = [%s %s], want [Alpha Bravo]", page1[0].Name, page1[1].Name)
	}
	page2, _, err := rs.List(2, 2, "name_asc")
	if err != nil {
		t.Fatalf("List page 2: %v", err)
	}
	if page2[0].Name != "Charlie" || page2[1].Name != "Delta" {
		t.Errorf("page 2 = [%s %s], want [Charlie Delta]", page2[0].Name, page2[1].Name)
	}
	// Boundary invariant: last of page 1 sorts before first of page 2.
	if page1[len(page1)-1].Name >= page2[0].Name {
		t.Errorf("page boundary not ordered: %q should precede %q", page1[len(page1)-1].Name, page2[0].Name)
	}
}

func TestList_SortNameDesc(t *testing.T) {
	rs := seedNamed(t, "Delta", "Alpha", "Echo", "Bravo", "Charlie")
	page1, _, err := rs.List(1, 2, "name_desc")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if page1[0].Name != "Echo" || page1[1].Name != "Delta" {
		t.Errorf("page 1 = [%s %s], want [Echo Delta]", page1[0].Name, page1[1].Name)
	}
}

// TestList_NoSortIsStable asserts the default (no sort) returns a consistent
// order across repeated calls for a stable table (no duplicates/gaps).
func TestList_NoSortIsStable(t *testing.T) {
	rs := seedNamed(t, "Delta", "Alpha", "Echo", "Bravo", "Charlie")
	first, _, err := rs.List(1, 10, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	second, _, err := rs.List(1, 10, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(first) != len(second) {
		t.Fatalf("length differs: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i].ID != second[i].ID {
			t.Errorf("order not stable at %d: %q vs %q", i, first[i].ID, second[i].ID)
		}
	}
}
