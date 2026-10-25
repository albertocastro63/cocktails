package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/almc/cocktails/internal/handler"
)

// doList issues GET /api/v1/recipes with the given raw query string and decodes
// the paginated envelope.
func doList(t *testing.T, rs *stubRecipeStore, rawQuery string) (int, map[string]any) {
	t.Helper()
	h := handler.NewRecipeHandler(rs)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes?"+rawQuery, nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)
	var resp map[string]any
	if rec.Body.Len() > 0 {
		_ = json.NewDecoder(rec.Body).Decode(&resp)
	}
	return rec.Code, resp
}

func names(resp map[string]any) []string {
	var out []string
	for _, item := range resp["data"].([]any) {
		out = append(out, item.(map[string]any)["name"].(string))
	}
	return out
}

func storeWith(n int) *stubRecipeStore {
	rs := newStubRecipeStore()
	for i := 0; i < n; i++ {
		r := sampleRecipe(fmt.Sprintf("r%02d", i), fmt.Sprintf("Recipe %02d", i), "u1")
		_ = rs.Create(r)
	}
	return rs
}

// TestRecipeList_PageAndLimit: total reflects the full set regardless of the
// page size, and each page returns at most `limit` items (FR-002, FR-005).
func TestRecipeList_PageAndLimit(t *testing.T) {
	rs := storeWith(30)

	code, resp := doList(t, rs, "page=1&limit=20")
	if code != http.StatusOK {
		t.Fatalf("got %d want 200", code)
	}
	if int(resp["total"].(float64)) != 30 {
		t.Errorf("total = %v, want 30", resp["total"])
	}
	if got := len(resp["data"].([]any)); got != 20 {
		t.Errorf("page 1 size = %d, want 20", got)
	}

	_, resp2 := doList(t, rs, "page=2&limit=20")
	if got := len(resp2["data"].([]any)); got != 10 {
		t.Errorf("page 2 size = %d, want 10", got)
	}
	if int(resp2["total"].(float64)) != 30 {
		t.Errorf("page 2 total = %v, want 30", resp2["total"])
	}
}

// TestRecipeList_OutOfRangePage returns an empty page with the correct total,
// not an error (FR-011).
func TestRecipeList_OutOfRangePage(t *testing.T) {
	rs := storeWith(5)
	code, resp := doList(t, rs, "page=99&limit=20")
	if code != http.StatusOK {
		t.Fatalf("got %d want 200", code)
	}
	if got := len(resp["data"].([]any)); got != 0 {
		t.Errorf("out-of-range page size = %d, want 0", got)
	}
	if int(resp["total"].(float64)) != 5 {
		t.Errorf("total = %v, want 5", resp["total"])
	}
}

// TestRecipeList_SortAcrossPages: name_asc orders the full set before paging, so
// the last item of page 1 precedes the first item of page 2 (FR-012).
func TestRecipeList_SortAcrossPages(t *testing.T) {
	rs := storeWith(5) // names: Recipe 00..04

	_, p1 := doList(t, rs, "page=1&limit=2&sort=name_asc")
	_, p2 := doList(t, rs, "page=2&limit=2&sort=name_asc")

	n1, n2 := names(p1), names(p2)
	if len(n1) != 2 || len(n2) != 2 {
		t.Fatalf("unexpected page sizes: %v %v", n1, n2)
	}
	if n1[0] != "Recipe 00" || n1[1] != "Recipe 01" {
		t.Errorf("page 1 = %v, want [Recipe 00, Recipe 01]", n1)
	}
	if n2[0] != "Recipe 02" {
		t.Errorf("page 2 first = %q, want Recipe 02", n2[0])
	}
	if n1[len(n1)-1] >= n2[0] {
		t.Errorf("boundary not ordered: %q should precede %q", n1[len(n1)-1], n2[0])
	}
}

func TestRecipeList_SortDesc(t *testing.T) {
	rs := storeWith(3) // Recipe 00,01,02
	_, resp := doList(t, rs, "sort=name_desc")
	got := names(resp)
	want := []string{"Recipe 02", "Recipe 01", "Recipe 00"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("desc order = %v, want %v", got, want)
			break
		}
	}
}

// TestRecipeList_UnknownSortDegrades: an unrecognized sort does not error and
// returns the full set (FR default / INV-5).
func TestRecipeList_UnknownSortDegrades(t *testing.T) {
	rs := storeWith(3)
	code, resp := doList(t, rs, "sort=bogus")
	if code != http.StatusOK {
		t.Fatalf("got %d want 200", code)
	}
	if int(resp["total"].(float64)) != 3 {
		t.Errorf("total = %v, want 3", resp["total"])
	}
}
