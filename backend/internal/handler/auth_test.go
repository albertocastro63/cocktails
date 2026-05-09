package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/almc/cocktails/internal/handler"
	"github.com/almc/cocktails/internal/model"
)

func userWithPassword(id, username, password string, isAdmin bool) *model.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return &model.User{ID: id, Username: username, PasswordHash: string(hash), IsAdmin: isAdmin}
}

func TestAuthLogin_ValidCredentials(t *testing.T) {
	us := newStubUserStore(userWithPassword("u1", "alice", "s3cr3t", false))
	h := handler.NewAuthHandler(us)

	body := `{"username":"alice","password":"s3cr3t"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rec.Code)
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("expected non-empty token in response")
	}
}

func TestAuthLogin_WrongPassword(t *testing.T) {
	us := newStubUserStore(userWithPassword("u1", "alice", "s3cr3t", false))
	h := handler.NewAuthHandler(us)

	body := `{"username":"alice","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got %d want 401", rec.Code)
	}
}

func TestAuthLogin_UnknownUser(t *testing.T) {
	us := newStubUserStore()
	h := handler.NewAuthHandler(us)

	body := `{"username":"nobody","password":"any"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got %d want 401", rec.Code)
	}
}

func TestAuthLogin_InvalidBody(t *testing.T) {
	us := newStubUserStore()
	h := handler.NewAuthHandler(us)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader("not-json"))
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d want 400", rec.Code)
	}
}
