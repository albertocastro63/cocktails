package sqlite_test

import (
	"testing"
	"time"

	"github.com/almc/cocktails/internal/model"
	sqstore "github.com/almc/cocktails/internal/store/sqlite"
	"github.com/google/uuid"
)

func newTestStores(t *testing.T) (*sqstore.RecipeStore, *sqstore.UserStore) {
	t.Helper()
	rs, us, err := sqstore.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return rs, us
}

func seedUser(t *testing.T, us *sqstore.UserStore) string {
	t.Helper()
	id := uuid.NewString()
	err := us.Create(&model.User{
		ID:           id,
		Username:     "testuser-" + id[:8],
		PasswordHash: "$2a$12$placeholder",
		CreatedAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

func sampleRecipe(creatorID string) *model.Recipe {
	return &model.Recipe{
		ID:   uuid.NewString(),
		Name: "Margarita",
		Ingredients: []model.Ingredient{
			{Name: "tequila", Quantity: "50", Unit: "ml"},
			{Name: "lime juice", Quantity: "25", Unit: "ml"},
		},
		Steps:      []string{"Combine ingredients", "Shake with ice", "Strain into glass"},
		Properties: map[string]string{"base_spirit": "Tequila", "style": "Sour"},
		CreatorID:  creatorID,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
}

func TestCreate_and_GetByID(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	r := sampleRecipe(uid)
	if err := rs.Create(r); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := rs.GetByID(r.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != r.Name {
		t.Errorf("name: got %q want %q", got.Name, r.Name)
	}
	if got.Properties["base_spirit"] != "Tequila" {
		t.Errorf("property base_spirit: got %q", got.Properties["base_spirit"])
	}
}

func TestGetByID_NotFound(t *testing.T) {
	rs, _ := newTestStores(t)
	_, err := rs.GetByID("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestList(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	for i := 0; i < 3; i++ {
		r := sampleRecipe(uid)
		r.ID = uuid.NewString()
		r.Name = "Recipe " + string(rune('A'+i))
		if err := rs.Create(r); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}
	recipes, total, err := rs.List(1, 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 3 {
		t.Errorf("total: got %d want 3", total)
	}
	if len(recipes) != 3 {
		t.Errorf("len: got %d want 3", len(recipes))
	}
}

func TestSearch_ByIngredient(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	if err := rs.Create(sampleRecipe(uid)); err != nil {
		t.Fatalf("Create: %v", err)
	}
	results, total, err := rs.Search("lime juice", 1, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if total == 0 || len(results) == 0 {
		t.Error("expected results for 'lime juice'")
	}
}

func TestSearch_ByPropertyValue(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	if err := rs.Create(sampleRecipe(uid)); err != nil {
		t.Fatalf("Create: %v", err)
	}
	results, _, err := rs.Search("Tequila", 1, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for property value 'Tequila'")
	}
}

func TestSearch_ByStep(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	if err := rs.Create(sampleRecipe(uid)); err != nil {
		t.Fatalf("Create: %v", err)
	}
	results, _, err := rs.Search("Shake", 1, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for step text 'Shake'")
	}
}

func TestUpdate(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	r := sampleRecipe(uid)
	if err := rs.Create(r); err != nil {
		t.Fatalf("Create: %v", err)
	}
	r.Name = "Updated Margarita"
	r.Properties["occasion"] = "Brunch"
	if err := rs.Update(r); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err := rs.GetByID(r.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Updated Margarita" {
		t.Errorf("name: got %q", got.Name)
	}
	if got.Properties["occasion"] != "Brunch" {
		t.Errorf("new property not persisted")
	}
}

func TestDelete(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	r := sampleRecipe(uid)
	if err := rs.Create(r); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := rs.Delete(r.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := rs.GetByID(r.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestExistsByName(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	r := sampleRecipe(uid)
	if err := rs.Create(r); err != nil {
		t.Fatalf("Create: %v", err)
	}
	exists, err := rs.ExistsByName(r.Name)
	if err != nil {
		t.Fatalf("ExistsByName: %v", err)
	}
	if !exists {
		t.Error("expected true for existing name")
	}
	exists, err = rs.ExistsByName("Nonexistent")
	if err != nil {
		t.Fatalf("ExistsByName: %v", err)
	}
	if exists {
		t.Error("expected false for nonexistent name")
	}
}

func TestRandom_EmptyDB(t *testing.T) {
	rs, _ := newTestStores(t)
	r, err := rs.Random()
	if err != nil {
		t.Fatalf("Random on empty DB: %v", err)
	}
	if r != nil {
		t.Error("expected nil from empty DB")
	}
}

func TestListAll(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)

	all, err := rs.ListAll()
	if err != nil {
		t.Fatalf("ListAll on empty DB: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected 0 recipes, got %d", len(all))
	}

	for i := 0; i < 3; i++ {
		r := sampleRecipe(uid)
		r.ID = uuid.NewString()
		r.Name = "Recipe " + string(rune('A'+i))
		if err := rs.Create(r); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	all, err = rs.ListAll()
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 recipes, got %d", len(all))
	}
}

func TestImportBatch(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)

	existing := sampleRecipe(uid)
	existing.Name = "Existing"
	if err := rs.Create(existing); err != nil {
		t.Fatalf("seed Create: %v", err)
	}

	toImport := []*model.Recipe{
		{ID: uuid.NewString(), Name: "New Cocktail", CreatorID: uid},
		{ID: uuid.NewString(), Name: "Existing", CreatorID: uid},
	}
	created, skipped, err := rs.ImportBatch(toImport, uid)
	if err != nil {
		t.Fatalf("ImportBatch: %v", err)
	}
	if created != 1 {
		t.Errorf("created: got %d want 1", created)
	}
	if skipped != 1 {
		t.Errorf("skipped: got %d want 1", skipped)
	}

	all, _ := rs.ListAll()
	if len(all) != 2 {
		t.Errorf("total recipes: got %d want 2", len(all))
	}
}

func TestImportBatch_EmptyInput(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	created, skipped, err := rs.ImportBatch([]*model.Recipe{}, uid)
	if err != nil {
		t.Fatalf("ImportBatch empty: %v", err)
	}
	if created != 0 || skipped != 0 {
		t.Errorf("expected 0/0 got %d/%d", created, skipped)
	}
}

func TestListByCreator_ReturnsOwnedRecipes(t *testing.T) {
	rs, us := newTestStores(t)
	uid1 := seedUser(t, us)
	uid2 := seedUser(t, us)

	r1 := sampleRecipe(uid1)
	r1.Name = "Owner Recipe 1"
	r2 := sampleRecipe(uid1)
	r2.Name = "Owner Recipe 2"
	r3 := sampleRecipe(uid2)
	r3.Name = "Other User Recipe"
	for _, r := range []*model.Recipe{r1, r2, r3} {
		if err := rs.Create(r); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	recipes, total, err := rs.ListByCreator(uid1, 1, 20)
	if err != nil {
		t.Fatalf("ListByCreator: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(recipes) != 2 {
		t.Errorf("expected 2 recipes, got %d", len(recipes))
	}
	for _, r := range recipes {
		if r.CreatorID != uid1 {
			t.Errorf("unexpected creator_id %q", r.CreatorID)
		}
	}
}

func TestListByCreator_ExcludesLegacyRecipes(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)

	legacy := sampleRecipe("")
	legacy.Name = "Legacy Recipe"
	owned := sampleRecipe(uid)
	owned.Name = "Owned Recipe"
	for _, r := range []*model.Recipe{legacy, owned} {
		if err := rs.Create(r); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	recipes, total, err := rs.ListByCreator(uid, 1, 20)
	if err != nil {
		t.Fatalf("ListByCreator: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(recipes) != 1 || recipes[0].Name != "Owned Recipe" {
		t.Errorf("expected only owned recipe, got %v", recipes)
	}
}

func TestListByCreator_EmptyForUnknownUser(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	r := sampleRecipe(uid)
	if err := rs.Create(r); err != nil {
		t.Fatalf("create: %v", err)
	}

	recipes, total, err := rs.ListByCreator("no-such-user", 1, 20)
	if err != nil {
		t.Fatalf("ListByCreator: %v", err)
	}
	if total != 0 || len(recipes) != 0 {
		t.Errorf("expected empty, got total=%d len=%d", total, len(recipes))
	}
}

func TestListByCreator_Pagination(t *testing.T) {
	rs, us := newTestStores(t)
	uid := seedUser(t, us)
	for i := 0; i < 5; i++ {
		r := sampleRecipe(uid)
		r.Name = uuid.NewString()
		if err := rs.Create(r); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	page1, total, err := rs.ListByCreator(uid, 1, 3)
	if err != nil {
		t.Fatalf("ListByCreator page1: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(page1) != 3 {
		t.Errorf("expected 3 on page1, got %d", len(page1))
	}

	page2, _, err := rs.ListByCreator(uid, 2, 3)
	if err != nil {
		t.Fatalf("ListByCreator page2: %v", err)
	}
	if len(page2) != 2 {
		t.Errorf("expected 2 on page2, got %d", len(page2))
	}
}
