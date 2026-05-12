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

func TestAdminCreateUser_Success(t *testing.T) {
	us := newStubUserStore()
	h := handler.NewAdminHandler(us)

	body := `{"username":"bob","password":"pass1234"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.CreateUser(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got %d want 201", rec.Code)
	}
}

func TestAdminCreateUser_DuplicateUsername(t *testing.T) {
	us := newStubUserStore(userWithPassword("u1", "bob", "pass", false))
	h := handler.NewAdminHandler(us)

	body := `{"username":"bob","password":"other"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.CreateUser(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("got %d want 409", rec.Code)
	}
}

func TestAdminCreateUser_MissingFields(t *testing.T) {
	us := newStubUserStore()
	h := handler.NewAdminHandler(us)

	body := `{"username":"bob"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.CreateUser(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d want 400", rec.Code)
	}
}

func TestAdminCreateUser_StoreError(t *testing.T) {
	us := newStubUserStore()
	us.createErr = store.ErrDuplicate // not ErrDuplicate path, so use generic error
	us.createErr = errGeneric

	h := handler.NewAdminHandler(us)
	body := `{"username":"bob","password":"pass1234"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.CreateUser(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d want 500", rec.Code)
	}
}

func TestAdminCreateUser_StoresProfileFields(t *testing.T) {
	us := newStubUserStore()
	h := handler.NewAdminHandler(us)

	body := `{"username":"alice","password":"pass1234","first_name":"Alice","last_name":"Smith","email":"alice@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.CreateUser(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got %d want 201", rec.Code)
	}
	var user map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &user); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if user["first_name"] != "Alice" || user["last_name"] != "Smith" || user["email"] != "alice@example.com" {
		t.Errorf("profile fields not stored: %v", user)
	}
}

func TestAdminCreateUser_EmailConflict(t *testing.T) {
	existing := userWithPassword("u1", "existing", "pass", false)
	existing.Email = "taken@example.com"
	us := newStubUserStore(existing)
	h := handler.NewAdminHandler(us)

	body := `{"username":"newuser","password":"pass1234","email":"taken@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.CreateUser(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("got %d want 409", rec.Code)
	}
	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp) //nolint:errcheck
	if errObj, ok := resp["error"].(map[string]any); ok {
		if errObj["code"] != "EMAIL_CONFLICT" {
			t.Errorf("expected EMAIL_CONFLICT code, got %v", errObj["code"])
		}
	}
}

func TestAdminCreateUser_NoEmailSucceedsWithDuplicate(t *testing.T) {
	existing := userWithPassword("u1", "existing", "pass", false)
	us := newStubUserStore(existing)
	h := handler.NewAdminHandler(us)

	body := `{"username":"newuser","password":"pass1234"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.CreateUser(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got %d want 201, body: %s", rec.Code, rec.Body.String())
	}
}

func TestGetUser_ReturnsUser(t *testing.T) {
	u := userWithPassword("u1", "alice", "pass", false)
	us := newStubUserStore(u)
	h := handler.NewAdminHandler(us)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/u1", nil)
	req.SetPathValue("id", "u1")
	rec := httptest.NewRecorder()
	h.GetUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d want 200", rec.Code)
	}
	var got map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got) //nolint:errcheck
	if got["id"] != "u1" {
		t.Errorf("wrong id: %v", got["id"])
	}
}

func TestGetUser_NotFound(t *testing.T) {
	us := newStubUserStore()
	h := handler.NewAdminHandler(us)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()
	h.GetUser(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got %d want 404", rec.Code)
	}
}

func TestGetUser_ForbiddenForAdmin(t *testing.T) {
	admin := userWithPassword("a1", "admin", "pass", true)
	us := newStubUserStore(admin)
	h := handler.NewAdminHandler(us)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/a1", nil)
	req.SetPathValue("id", "a1")
	rec := httptest.NewRecorder()
	h.GetUser(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("got %d want 403", rec.Code)
	}
}

func TestUpdateUser_UpdatesFields(t *testing.T) {
	u := userWithPassword("u1", "alice", "pass", false)
	us := newStubUserStore(u)
	h := handler.NewAdminHandler(us)

	body := `{"first_name":"Alice","last_name":"Smith","email":"alice@example.com"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/u1", strings.NewReader(body))
	req.SetPathValue("id", "u1")
	rec := httptest.NewRecorder()
	h.UpdateUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d want 200, body: %s", rec.Code, rec.Body.String())
	}
	var got map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got) //nolint:errcheck
	if got["first_name"] != "Alice" || got["email"] != "alice@example.com" {
		t.Errorf("fields not updated: %v", got)
	}
}

func TestUpdateUser_PasswordIncreasesTokenVersion(t *testing.T) {
	u := userWithPassword("u1", "alice", "oldpass", false)
	us := newStubUserStore(u)
	h := handler.NewAdminHandler(us)

	body := `{"password":"newpassword"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/u1", strings.NewReader(body))
	req.SetPathValue("id", "u1")
	rec := httptest.NewRecorder()
	h.UpdateUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d want 200", rec.Code)
	}
	var got map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got) //nolint:errcheck
	if v, _ := got["token_version"].(float64); v != 1 {
		t.Errorf("token_version: got %v want 1", got["token_version"])
	}
}

func TestUpdateUser_BlankPasswordPreservesTokenVersion(t *testing.T) {
	u := userWithPassword("u1", "alice", "pass", false)
	us := newStubUserStore(u)
	h := handler.NewAdminHandler(us)

	body := `{"first_name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/u1", strings.NewReader(body))
	req.SetPathValue("id", "u1")
	rec := httptest.NewRecorder()
	h.UpdateUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d want 200", rec.Code)
	}
	var got map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got) //nolint:errcheck
	if v, _ := got["token_version"].(float64); v != 0 {
		t.Errorf("token_version should stay 0, got %v", got["token_version"])
	}
}

func TestUpdateUser_EmailConflict(t *testing.T) {
	u := userWithPassword("u1", "alice", "pass", false)
	other := userWithPassword("u2", "bob", "pass", false)
	other.Email = "taken@example.com"
	us := newStubUserStore(u, other)
	h := handler.NewAdminHandler(us)

	body := `{"email":"taken@example.com"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/u1", strings.NewReader(body))
	req.SetPathValue("id", "u1")
	rec := httptest.NewRecorder()
	h.UpdateUser(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("got %d want 409", rec.Code)
	}
}

func TestUpdateUser_ForbiddenForAdmin(t *testing.T) {
	admin := userWithPassword("a1", "admin", "pass", true)
	us := newStubUserStore(admin)
	h := handler.NewAdminHandler(us)

	body := `{"first_name":"Changed"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/a1", strings.NewReader(body))
	req.SetPathValue("id", "a1")
	rec := httptest.NewRecorder()
	h.UpdateUser(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("got %d want 403", rec.Code)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	us := newStubUserStore()
	h := handler.NewAdminHandler(us)

	body := `{"first_name":"Ghost"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/nonexistent", strings.NewReader(body))
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()
	h.UpdateUser(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got %d want 404", rec.Code)
	}
}

func TestDeleteUser_Succeeds(t *testing.T) {
	u := userWithPassword("u1", "alice", "pass", false)
	us := newStubUserStore(u)
	h := handler.NewAdminHandler(us)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/users/u1", nil)
	req.SetPathValue("id", "u1")
	rec := httptest.NewRecorder()
	h.DeleteUser(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("got %d want 204", rec.Code)
	}
	// Verify user is removed
	if _, err := us.GetByID("u1"); err == nil {
		t.Error("expected user to be deleted")
	}
}

func TestDeleteUser_ForbiddenForAdmin(t *testing.T) {
	admin := userWithPassword("a1", "admin", "pass", true)
	us := newStubUserStore(admin)
	h := handler.NewAdminHandler(us)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/users/a1", nil)
	req.SetPathValue("id", "a1")
	rec := httptest.NewRecorder()
	h.DeleteUser(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("got %d want 403", rec.Code)
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	us := newStubUserStore()
	h := handler.NewAdminHandler(us)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/users/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()
	h.DeleteUser(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got %d want 404", rec.Code)
	}
}

func TestListUsers_ReturnsNonAdminUsers(t *testing.T) {
	admin := userWithPassword("a1", "admin", "pass", true)
	user := userWithPassword("u1", "alice", "pass", false)
	us := newStubUserStore(admin, user)
	h := handler.NewAdminHandler(us)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	rec := httptest.NewRecorder()
	h.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d want 200", rec.Code)
	}
	var users []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &users); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 non-admin user, got %d", len(users))
	}
}

func TestListUsers_ReturnsEmptyArray(t *testing.T) {
	us := newStubUserStore()
	h := handler.NewAdminHandler(us)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	rec := httptest.NewRecorder()
	h.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d want 200", rec.Code)
	}
	body := strings.TrimSpace(rec.Body.String())
	if body != "[]" {
		t.Errorf("expected empty array, got %q", body)
	}
}
