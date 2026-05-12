package sqlite_test

import (
	"testing"
	"time"

	"github.com/almc/cocktails/internal/model"
	"github.com/google/uuid"
)

func sampleUser() *model.User {
	return &model.User{
		ID:           uuid.NewString(),
		Username:     "alice-" + uuid.NewString()[:8],
		PasswordHash: "$2a$12$placeholder",
		IsAdmin:      false,
		CreatedAt:    time.Now().UTC(),
	}
}

func TestUserCreate_and_GetByUsername(t *testing.T) {
	_, us := newTestStores(t)
	u := sampleUser()
	if err := us.Create(u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := us.GetByUsername(u.Username)
	if err != nil {
		t.Fatalf("GetByUsername: %v", err)
	}
	if got.ID != u.ID {
		t.Errorf("id mismatch: got %q want %q", got.ID, u.ID)
	}
}

func TestUserGetByID(t *testing.T) {
	_, us := newTestStores(t)
	u := sampleUser()
	if err := us.Create(u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := us.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Username != u.Username {
		t.Errorf("username mismatch")
	}
}

func TestUserList_NonAdminOnly(t *testing.T) {
	_, us := newTestStores(t)
	admin := sampleUser()
	admin.IsAdmin = true
	if err := us.Create(admin); err != nil {
		t.Fatalf("Create admin: %v", err)
	}
	u := sampleUser()
	if err := us.Create(u); err != nil {
		t.Fatalf("Create user: %v", err)
	}
	users, err := us.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 non-admin user, got %d", len(users))
	}
	if users[0].ID != u.ID {
		t.Errorf("expected user %q, got %q", u.ID, users[0].ID)
	}
}

func TestUserList_Empty(t *testing.T) {
	_, us := newTestStores(t)
	users, err := us.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected empty slice, got %d", len(users))
	}
}

func TestUserUpdate_PersistsFields(t *testing.T) {
	_, us := newTestStores(t)
	u := sampleUser()
	u.FirstName = "Alice"
	u.LastName = "Smith"
	u.Email = "alice@example.com"
	if err := us.Create(u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	u.FirstName = "Alicia"
	u.LastName = "Jones"
	u.Email = "alicia@example.com"
	u.TokenVersion = 1
	if err := us.Update(u); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err := us.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.FirstName != "Alicia" || got.LastName != "Jones" || got.Email != "alicia@example.com" {
		t.Errorf("fields not updated: %+v", got)
	}
	if got.TokenVersion != 1 {
		t.Errorf("token_version: got %d want 1", got.TokenVersion)
	}
}

func TestUserDelete_RemovesUser(t *testing.T) {
	_, us := newTestStores(t)
	u := sampleUser()
	if err := us.Create(u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := us.Delete(u.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := us.GetByID(u.ID)
	if err == nil {
		t.Fatal("expected error getting deleted user, got nil")
	}
}

func TestUserDelete_UnknownID(t *testing.T) {
	_, us := newTestStores(t)
	if err := us.Delete("nonexistent"); err == nil {
		t.Fatal("expected error deleting nonexistent user")
	}
}

func TestUserGetByEmail(t *testing.T) {
	_, us := newTestStores(t)
	u := sampleUser()
	u.Email = "test@example.com"
	if err := us.Create(u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := us.GetByEmail("test@example.com")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if got.ID != u.ID {
		t.Errorf("id mismatch: got %q want %q", got.ID, u.ID)
	}
}

func TestUserGetByEmail_NotFound(t *testing.T) {
	_, us := newTestStores(t)
	_, err := us.GetByEmail("nobody@example.com")
	if err == nil {
		t.Fatal("expected error for unknown email")
	}
}

func TestUserCreate_StoresProfileFields(t *testing.T) {
	_, us := newTestStores(t)
	u := sampleUser()
	u.FirstName = "Bob"
	u.LastName = "Marley"
	u.Email = "bob@example.com"
	if err := us.Create(u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := us.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.FirstName != "Bob" || got.LastName != "Marley" || got.Email != "bob@example.com" {
		t.Errorf("profile fields not stored: %+v", got)
	}
}

func TestUserCount(t *testing.T) {
	_, us := newTestStores(t)
	n, err := us.Count()
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
	if err := us.Create(sampleUser()); err != nil {
		t.Fatalf("Create: %v", err)
	}
	n, err = us.Count()
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

