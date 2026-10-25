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

// provisionRecipes is a white-box helper (package dynamo) so the scan loop can
// be exercised directly. It mirrors the external test harness.
func provisionRecipes(t *testing.T) *RecipeStore {
	t.Helper()
	if os.Getenv("DYNAMODB_ENDPOINT") == "" {
		t.Fatal("DYNAMODB_ENDPOINT is required for DynamoDB store tests (see specs/029-local-dynamodb-emulator/quickstart.md).")
	}
	client, err := NewClient(context.Background())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	names := TableNames{
		Recipes:   "test-scan-recipes-" + uuid.NewString()[:8],
		Users:     "test-scan-users-" + uuid.NewString()[:8],
		Favorites: "test-scan-favorites-" + uuid.NewString()[:8],
	}
	if err := EnsureSchema(context.Background(), client, names); err != nil {
		t.Fatalf("EnsureSchema: %v", err)
	}
	t.Cleanup(func() {
		for _, n := range []string{names.Recipes, names.Users, names.Favorites} {
			_, _ = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{TableName: aws.String(n)})
		}
	})
	return NewRecipeStore(client, names.Recipes)
}

// TestScanAll_PagesThroughAllBatches forces the Scan to return one item per call
// (Limit=1) so LastEvaluatedKey is set on every page but the last. scanAll MUST
// follow the key and return every item, not just the first batch. This is the
// regression guard for the "only the first Scan page counted" bug (FR-009).
func TestScanAll_PagesThroughAllBatches(t *testing.T) {
	rs := provisionRecipes(t)
	const n = 7
	for i := 0; i < n; i++ {
		r := &model.Recipe{ID: uuid.NewString(), Name: "Recipe " + uuid.NewString()[:4]}
		if err := rs.Create(r); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	items, err := scanAll(context.Background(), rs.client, &dynamodb.ScanInput{
		TableName: aws.String(rs.tableName),
		Limit:     aws.Int32(1), // force a new page per item
	})
	if err != nil {
		t.Fatalf("scanAll: %v", err)
	}
	if len(items) != n {
		t.Errorf("scanAll returned %d items across batches, want %d", len(items), n)
	}
}

// TestList_TotalCountsEntireTable asserts List reports the true total even when
// the underlying table spans more than one Scan batch (verified independently of
// the returned page size).
func TestList_TotalCountsEntireTable(t *testing.T) {
	rs := provisionRecipes(t)
	const n = 25
	for i := 0; i < n; i++ {
		r := &model.Recipe{ID: uuid.NewString(), Name: "Recipe " + uuid.NewString()[:6]}
		if err := rs.Create(r); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	page1, total, err := rs.List(1, 20, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != n {
		t.Errorf("total = %d, want %d", total, n)
	}
	if len(page1) != 20 {
		t.Errorf("page 1 size = %d, want 20", len(page1))
	}
	page2, _, err := rs.List(2, 20, "")
	if err != nil {
		t.Fatalf("List page 2: %v", err)
	}
	if len(page2) != n-20 {
		t.Errorf("page 2 size = %d, want %d", len(page2), n-20)
	}
}
