package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/almc/cocktails/internal/email"
	"github.com/almc/cocktails/internal/handler"
	"github.com/almc/cocktails/internal/model"
)

// findEntry scans the buffered JSON log lines for one whose action/outcome match,
// and returns it. level="" skips the level check.
func findEntry(t *testing.T, buf *bytes.Buffer, action, outcome, level string) map[string]any {
	t.Helper()
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var e map[string]any
		if json.Unmarshal([]byte(line), &e) != nil {
			continue
		}
		if e["action"] != action {
			continue
		}
		if outcome != "" && e["outcome"] != outcome {
			continue
		}
		if level != "" && e["level"] != level {
			t.Fatalf("action %q logged at level %v, want %s", action, e["level"], level)
		}
		return e
	}
	t.Fatalf("no log entry for action=%q outcome=%q in:\n%s", action, outcome, buf.String())
	return nil
}

func TestLog_AuthLogin_SuccessAndFailure(t *testing.T) {
	// Success (INFO)
	buf, install := bufLogger()
	pw := "correct horse battery"
	us := newStubUserStore(userWithPassword("u1", "alice", pw, false))
	h := install(http.HandlerFunc(handler.NewAuthHandler(us).Login))
	body := `{"username":"alice","password":"` + pw + `"}`
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(body)))
	e := findEntry(t, buf, "auth.login", "success", "INFO")
	if e["user_id"] != "u1" {
		t.Errorf("user_id = %v, want u1", e["user_id"])
	}

	// Failure — bad password (WARN), and no secret leakage
	buf2, install2 := bufLogger()
	h2 := install2(http.HandlerFunc(handler.NewAuthHandler(us).Login))
	bad := `{"username":"alice","password":"wrong-secret-xyz"}`
	h2.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(bad)))
	findEntry(t, buf2, "auth.login", "failure", "WARN")
	if strings.Contains(buf2.String(), "wrong-secret-xyz") {
		t.Errorf("submitted password leaked into logs: %s", buf2.String())
	}
}

func TestLog_RecipeCreate_Info(t *testing.T) {
	buf, install := bufLogger()
	rs := newStubRecipeStore()
	h := install(handler.RequireAuth(http.HandlerFunc(handler.NewRecipeHandler(rs).Create)))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/recipes", strings.NewReader(`{"name":"Mojito","ingredients":[],"steps":[]}`))
	req.Header.Set("Authorization", "Bearer "+validToken(t, "u1", "alice", false))
	h.ServeHTTP(httptest.NewRecorder(), req)
	findEntry(t, buf, "recipe.create", "success", "INFO")
}

func TestLog_FavoriteAdd_Info(t *testing.T) {
	buf, install := bufLogger()
	rs := newStubRecipeStore(sampleRecipe("r1", "Mojito", "other-user"))
	h := install(handler.RequireAuth(http.HandlerFunc(handler.NewFavoriteHandler(&stubFavoriteStore{}, rs).Add)))
	req := httptest.NewRequest(http.MethodPut, "/api/v1/recipes/r1/favorite", nil)
	req.SetPathValue("id", "r1")
	req.Header.Set("Authorization", "Bearer "+validToken(t, "u1", "alice", false))
	h.ServeHTTP(httptest.NewRecorder(), req)
	findEntry(t, buf, "favorite.add", "success", "INFO")
}

func TestLog_PasswordResetRequest_Info(t *testing.T) {
	buf, install := bufLogger()
	us := newStubUserStore(resetUser())
	rh := handler.NewPasswordResetHandler(us, &email.StubSender{}, "https://x")
	h := install(http.HandlerFunc(rh.Forgot))
	body := `{"email":"` + resetUser().Email + `"}`
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(body)))
	findEntry(t, buf, "password.reset_request", "success", "INFO")
}

var _ = model.User{}
