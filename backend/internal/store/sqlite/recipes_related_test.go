package sqlite_test

import (
	"sort"
	"testing"
	"time"

	"github.com/almc/cocktails/internal/model"
	sqstore "github.com/almc/cocktails/internal/store/sqlite"
	"github.com/google/uuid"
)

func relRecipe(t *testing.T, rs *sqstore.RecipeStore, name string) string {
	t.Helper()
	id := uuid.NewString()
	if err := rs.Create(&model.Recipe{
		ID: id, Name: name,
		Ingredients: []model.Ingredient{}, Steps: []string{},
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create %s: %v", name, err)
	}
	return id
}

func relatedOf(t *testing.T, rs *sqstore.RecipeStore, id string) []string {
	t.Helper()
	r, err := rs.GetByID(id)
	if err != nil {
		t.Fatalf("get %s: %v", id, err)
	}
	got := append([]string(nil), r.RelatedIDs...)
	sort.Strings(got)
	return got
}

func TestSetRelated_Symmetric(t *testing.T) {
	rs, _ := newTestStores(t)
	a := relRecipe(t, rs, "Negroni")
	b := relRecipe(t, rs, "Left Hand")

	if err := rs.SetRelated(a, []string{b}); err != nil {
		t.Fatalf("SetRelated: %v", err)
	}
	if got := relatedOf(t, rs, a); len(got) != 1 || got[0] != b {
		t.Errorf("A.related = %v, want [%s]", got, b)
	}
	if got := relatedOf(t, rs, b); len(got) != 1 || got[0] != a {
		t.Errorf("B.related = %v, want [%s] (reverse must be recorded)", got, a)
	}
}

func TestSetRelated_DedupeAndNoSelf(t *testing.T) {
	rs, _ := newTestStores(t)
	a := relRecipe(t, rs, "Negroni")
	b := relRecipe(t, rs, "Left Hand")

	if err := rs.SetRelated(a, []string{b, b, a}); err != nil {
		t.Fatalf("SetRelated: %v", err)
	}
	if got := relatedOf(t, rs, a); len(got) != 1 || got[0] != b {
		t.Errorf("A.related = %v, want exactly [%s] (dedupe + drop self)", got, b)
	}
}

func TestSetRelated_DropsNonExistent(t *testing.T) {
	rs, _ := newTestStores(t)
	a := relRecipe(t, rs, "Negroni")
	b := relRecipe(t, rs, "Left Hand")

	if err := rs.SetRelated(a, []string{b, "ghost-id"}); err != nil {
		t.Fatalf("SetRelated: %v", err)
	}
	if got := relatedOf(t, rs, a); len(got) != 1 || got[0] != b {
		t.Errorf("A.related = %v, want [%s] (non-existent dropped)", got, b)
	}
}

func TestSetRelated_RemovalIsSymmetric(t *testing.T) {
	rs, _ := newTestStores(t)
	a := relRecipe(t, rs, "Negroni")
	b := relRecipe(t, rs, "Left Hand")

	if err := rs.SetRelated(a, []string{b}); err != nil {
		t.Fatalf("SetRelated add: %v", err)
	}
	if err := rs.SetRelated(a, []string{}); err != nil {
		t.Fatalf("SetRelated clear: %v", err)
	}
	if got := relatedOf(t, rs, a); len(got) != 0 {
		t.Errorf("A.related = %v, want empty after clear", got)
	}
	if got := relatedOf(t, rs, b); len(got) != 0 {
		t.Errorf("B.related = %v, want empty after A cleared (reverse removed)", got)
	}
}

func TestDelete_RemovesFromCounterparts(t *testing.T) {
	rs, _ := newTestStores(t)
	a := relRecipe(t, rs, "Negroni")
	b := relRecipe(t, rs, "Left Hand")
	if err := rs.SetRelated(a, []string{b}); err != nil {
		t.Fatalf("SetRelated: %v", err)
	}
	if err := rs.Delete(a); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	for _, id := range relatedOf(t, rs, b) {
		if id == a {
			t.Errorf("B still lists deleted A: %v", relatedOf(t, rs, b))
		}
	}
	if got := relatedOf(t, rs, b); len(got) != 0 {
		t.Errorf("B.related = %v, want empty after A deleted", got)
	}
}

func TestSetRelated_NonTransitive(t *testing.T) {
	rs, _ := newTestStores(t)
	a := relRecipe(t, rs, "A")
	b := relRecipe(t, rs, "B")
	c := relRecipe(t, rs, "C")

	if err := rs.SetRelated(a, []string{b}); err != nil {
		t.Fatalf("A-B: %v", err)
	}
	if err := rs.SetRelated(b, []string{a, c}); err != nil {
		t.Fatalf("B-C: %v", err)
	}
	// A relates only to B; C must not have leaked in.
	for _, id := range relatedOf(t, rs, a) {
		if id == c {
			t.Errorf("A must not be related to C (non-transitive)")
		}
	}
}
