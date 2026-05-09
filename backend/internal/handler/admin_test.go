package handler_test

import (
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
