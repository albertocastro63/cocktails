package sqlite_test

import (
	"testing"

	sqstore "github.com/almc/cocktails/internal/store/sqlite"
)

func newTestAllStores(t *testing.T) (*sqstore.RecipeStore, *sqstore.UserStore, *sqstore.FavoriteStore) {
	t.Helper()
	rs, us, fs, err := sqstore.OpenAll(":memory:")
	if err != nil {
		t.Fatalf("open stores: %v", err)
	}
	return rs, us, fs
}

// (1) Add inserts a row without error
func TestFavoriteStore_Add(t *testing.T) {
	_, us, fs := newTestAllStores(t)
	userID := seedUser(t, us)
	if err := fs.Add(userID, "r1"); err != nil {
		t.Fatalf("Add: %v", err)
	}
}

// (2) IsFavorite returns true after Add
func TestFavoriteStore_IsFavorite_AfterAdd(t *testing.T) {
	_, us, fs := newTestAllStores(t)
	userID := seedUser(t, us)
	if err := fs.Add(userID, "r1"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	ok, err := fs.IsFavorite(userID, "r1")
	if err != nil {
		t.Fatalf("IsFavorite: %v", err)
	}
	if !ok {
		t.Error("expected IsFavorite=true after Add")
	}
}

// (3) Remove deletes the row without error
func TestFavoriteStore_Remove(t *testing.T) {
	_, us, fs := newTestAllStores(t)
	userID := seedUser(t, us)
	if err := fs.Add(userID, "r1"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := fs.Remove(userID, "r1"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
}

// (4) IsFavorite returns false after Remove
func TestFavoriteStore_IsFavorite_AfterRemove(t *testing.T) {
	_, us, fs := newTestAllStores(t)
	userID := seedUser(t, us)
	if err := fs.Add(userID, "r1"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := fs.Remove(userID, "r1"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	ok, err := fs.IsFavorite(userID, "r1")
	if err != nil {
		t.Fatalf("IsFavorite: %v", err)
	}
	if ok {
		t.Error("expected IsFavorite=false after Remove")
	}
}

// (5) ListByUser returns all favorites for a user
func TestFavoriteStore_ListByUser(t *testing.T) {
	_, us, fs := newTestAllStores(t)
	userID := seedUser(t, us)
	for _, id := range []string{"r1", "r2", "r3"} {
		if err := fs.Add(userID, id); err != nil {
			t.Fatalf("Add(%s): %v", id, err)
		}
	}
	favs, err := fs.ListByUser(userID)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(favs) != 3 {
		t.Errorf("expected 3 favorites, got %d", len(favs))
	}
}

// (6) Add is idempotent — calling it twice does not return an error
func TestFavoriteStore_Add_Idempotent(t *testing.T) {
	_, us, fs := newTestAllStores(t)
	userID := seedUser(t, us)
	if err := fs.Add(userID, "r1"); err != nil {
		t.Fatalf("first Add: %v", err)
	}
	if err := fs.Add(userID, "r1"); err != nil {
		t.Fatalf("second Add (idempotent) failed: %v", err)
	}
}

// (7) CountByRecipe returns 0 (stub)
func TestFavoriteStore_CountByRecipe_Stub(t *testing.T) {
	_, _, fs := newTestAllStores(t)
	count, err := fs.CountByRecipe("any-recipe")
	if err != nil {
		t.Fatalf("CountByRecipe: %v", err)
	}
	if count != 0 {
		t.Errorf("expected stub to return 0, got %d", count)
	}
}
