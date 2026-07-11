package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/almc/cocktails/internal/handler"
	"github.com/almc/cocktails/internal/store"
)

func relMux(rs store.RecipeStore) *http.ServeMux {
	h := handler.NewRecipeHandler(rs)
	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/recipes", handler.RequireAuth(http.HandlerFunc(h.Create)))
	mux.Handle("PUT /api/v1/recipes/{id}", handler.RequireAuth(http.HandlerFunc(h.Update)))
	mux.HandleFunc("GET /api/v1/recipes/{id}", h.GetByID)
	return mux
}

func putRelated(t *testing.T, mux *http.ServeMux, id, token, body string) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/"+id, strings.NewReader(body))
	req.SetPathValue("id", id)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec.Code
}

func TestUpdate_RelatedIDs_Symmetric(t *testing.T) {
	a := sampleRecipe("a", "Negroni", "u1")
	b := sampleRecipe("b", "Left Hand", "u1")
	rs := newStubRecipeStore(a, b)
	mux := relMux(rs)

	if code := putRelated(t, mux, "a", validToken(t, "u1", "alice", false), `{"related_ids":["b"]}`); code != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200", code)
	}
	if !contains(a.RelatedIDs, "b") {
		t.Errorf("A.related = %v, want to contain b", a.RelatedIDs)
	}
	if !contains(b.RelatedIDs, "a") {
		t.Errorf("B.related = %v, want to contain a (reverse recorded)", b.RelatedIDs)
	}
}

func TestUpdate_RelatedIDs_AbsentLeavesUnchanged(t *testing.T) {
	a := sampleRecipe("a", "Negroni", "u1")
	b := sampleRecipe("b", "Left Hand", "u1")
	a.RelatedIDs = []string{"b"}
	b.RelatedIDs = []string{"a"}
	rs := newStubRecipeStore(a, b)
	mux := relMux(rs)

	// No related_ids in the body → relations must be preserved.
	if code := putRelated(t, mux, "a", validToken(t, "u1", "alice", false), `{"name":"Negroni Sbagliato"}`); code != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200", code)
	}
	if !contains(a.RelatedIDs, "b") {
		t.Errorf("A.related = %v, want unchanged [b]", a.RelatedIDs)
	}
}

func TestUpdate_RelatedIDs_NormalizesSelfDupNonexistent(t *testing.T) {
	a := sampleRecipe("a", "Negroni", "u1")
	b := sampleRecipe("b", "Left Hand", "u1")
	rs := newStubRecipeStore(a, b)
	mux := relMux(rs)

	putRelated(t, mux, "a", validToken(t, "u1", "alice", false), `{"related_ids":["b","b","a","ghost"]}`)
	if len(a.RelatedIDs) != 1 || a.RelatedIDs[0] != "b" {
		t.Errorf("A.related = %v, want exactly [b] (dedupe, drop self + non-existent)", a.RelatedIDs)
	}
}

// FR-016: a non-admin editor of their own recipe A may relate to a recipe B
// owned by a different user; the reverse is written to B with no ownership check.
func TestUpdate_RelatedIDs_CrossOwnershipAllowed(t *testing.T) {
	a := sampleRecipe("a", "Negroni", "u1")   // editor owns A
	b := sampleRecipe("b", "Left Hand", "u2") // B owned by someone else
	rs := newStubRecipeStore(a, b)
	mux := relMux(rs)

	code := putRelated(t, mux, "a", validToken(t, "u1", "alice", false), `{"related_ids":["b"]}`)
	if code != http.StatusOK {
		t.Fatalf("cross-owner relate status = %d, want 200", code)
	}
	if !contains(b.RelatedIDs, "a") {
		t.Errorf("B (owned by u2).related = %v, want to contain a — no ownership check on B", b.RelatedIDs)
	}
}

// FR-003: relations are not transitive. A–B and B–C must not make A related to C.
func TestUpdate_RelatedIDs_NonTransitive(t *testing.T) {
	a := sampleRecipe("a", "A", "u1")
	b := sampleRecipe("b", "B", "u1")
	c := sampleRecipe("c", "C", "u1")
	rs := newStubRecipeStore(a, b, c)
	mux := relMux(rs)
	tok := validToken(t, "u1", "alice", false)

	putRelated(t, mux, "a", tok, `{"related_ids":["b"]}`)
	putRelated(t, mux, "b", tok, `{"related_ids":["a","c"]}`)

	if contains(a.RelatedIDs, "c") {
		t.Errorf("A.related = %v, must not contain C (non-transitive)", a.RelatedIDs)
	}
}

func TestGetByID_EnrichesRelatedSortedAlpha(t *testing.T) {
	a := sampleRecipe("a", "Negroni", "u1")
	lh := sampleRecipe("lh", "Left Hand", "u1")
	rh := sampleRecipe("rh", "Right Hand", "u1")
	a.RelatedIDs = []string{"rh", "lh"} // deliberately unsorted
	h := handler.NewRecipeHandler(newStubRecipeStore(a, lh, rh))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/a", nil)
	req.SetPathValue("id", "a")
	rec := httptest.NewRecorder()
	h.GetByID(rec, req)

	var resp struct {
		Related []struct{ ID, Name string } `json:"related"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("bad JSON: %v", err)
	}
	if len(resp.Related) != 2 {
		t.Fatalf("related len = %d, want 2: %s", len(resp.Related), rec.Body)
	}
	if resp.Related[0].Name != "Left Hand" || resp.Related[1].Name != "Right Hand" {
		t.Errorf("related not alphabetical: %+v", resp.Related)
	}
}

func TestList_DoesNotEnrichRelated(t *testing.T) {
	a := sampleRecipe("a", "Negroni", "u1")
	a.RelatedIDs = []string{"b"}
	h := handler.NewRecipeHandler(newStubRecipeStore(a, sampleRecipe("b", "Left Hand", "u1")))

	rec := httptest.NewRecorder()
	h.List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/recipes", nil))
	var resp struct {
		Data []map[string]any `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	for _, item := range resp.Data {
		if _, ok := item["related"]; ok {
			t.Errorf("list response must not include enriched `related`: %v", item)
		}
	}
}

func TestNames_ReturnsMinimalIDName(t *testing.T) {
	rs := newStubRecipeStore(sampleRecipe("a", "Negroni", "u1"), sampleRecipe("b", "Left Hand", "u1"))
	h := handler.NewRecipeHandler(rs)
	rec := httptest.NewRecorder()
	h.Names(rec, httptest.NewRequest(http.MethodGet, "/api/v1/recipes/names", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var out []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("bad JSON: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("got %d names, want 2", len(out))
	}
	for _, e := range out {
		if e["id"] == nil || e["name"] == nil {
			t.Errorf("entry missing id/name: %v", e)
		}
		if len(e) != 2 { // id + name only, no ingredients/steps/etc.
			t.Errorf("entry has extra fields (should be minimal): %v", e)
		}
	}
}

var _ = json.Marshal
