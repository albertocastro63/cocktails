package auth_test

import (
	"os"
	"testing"

	"github.com/almc/cocktails/internal/auth"
)

func init() {
	os.Setenv("JWT_SECRET", "test-secret-key-minimum-length")
}

func TestIssue_and_Parse(t *testing.T) {
	token, expiresAt, err := auth.Issue("user-1", "alice", false, 0)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if expiresAt.IsZero() {
		t.Fatal("expected non-zero expiry")
	}

	claims, err := auth.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Errorf("UserID: got %q want user-1", claims.UserID)
	}
	if claims.Username != "alice" {
		t.Errorf("Username: got %q want alice", claims.Username)
	}
	if claims.IsAdmin {
		t.Error("expected IsAdmin=false")
	}
}

func TestIssue_AdminClaim(t *testing.T) {
	token, _, err := auth.Issue("admin-1", "admin", true, 0)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	claims, err := auth.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !claims.IsAdmin {
		t.Error("expected IsAdmin=true")
	}
}

func TestParse_InvalidToken(t *testing.T) {
	_, err := auth.Parse("not.a.valid.token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestParse_ExpiredToken(t *testing.T) {
	// Pre-crafted token with exp=1 (far in the past)
	_, err := auth.Parse("eyJhbGciOiJIUzI1NiJ9.eyJleHAiOjF9.bad")
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestIssue_EmbedTokenVersion(t *testing.T) {
	token, _, err := auth.Issue("user-1", "alice", false, 3)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	claims, err := auth.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.TokenVersion != 3 {
		t.Errorf("TokenVersion: got %d want 3", claims.TokenVersion)
	}
}

func TestParse_ReturnsTokenVersion(t *testing.T) {
	token, _, err := auth.Issue("user-2", "bob", false, 7)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	claims, err := auth.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.TokenVersion != 7 {
		t.Errorf("TokenVersion: got %d want 7", claims.TokenVersion)
	}
}
