package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/almc/cocktails/internal/handler"
	"github.com/almc/cocktails/internal/model"
)

func BenchmarkRecipeSearch(b *testing.B) {
	recipes := make([]*model.Recipe, 20)
	for i := range recipes {
		recipes[i] = sampleRecipe(
			"r"+string(rune('0'+i)),
			"Recipe "+string(rune('A'+i)),
			"u1",
		)
	}
	rs := newStubRecipeStore(recipes...)
	h := handler.NewRecipeHandler(rs)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes?q=recipe", nil)
			rec := httptest.NewRecorder()
			h.List(rec, req)
			if rec.Code != http.StatusOK {
				b.Errorf("unexpected status %d", rec.Code)
			}
		}
	})
}

func BenchmarkRecipeList(b *testing.B) {
	recipes := make([]*model.Recipe, 50)
	for i := range recipes {
		recipes[i] = sampleRecipe(
			"r"+string(rune('0'+i%10))+string(rune('0'+i/10)),
			"Recipe "+string(rune('A'+i%26)),
			"u1",
		)
	}
	rs := newStubRecipeStore(recipes...)
	h := handler.NewRecipeHandler(rs)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes", nil)
			rec := httptest.NewRecorder()
			h.List(rec, req)
		}
	})
}
