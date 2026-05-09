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

