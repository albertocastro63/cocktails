package auth

import "testing"

func TestGenerateResetToken_UniqueAndURLSafe(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		tok, err := GenerateResetToken()
		if err != nil {
			t.Fatalf("GenerateResetToken: %v", err)
		}
		if len(tok) < 40 {
			t.Errorf("token too short: %q", tok)
		}
		for _, r := range tok {
			ok := (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
			if !ok {
				t.Errorf("token has non-URL-safe char %q", r)
			}
		}
		if seen[tok] {
			t.Errorf("duplicate token generated")
		}
		seen[tok] = true
	}
}

func TestHashAndVerifyResetToken(t *testing.T) {
	tok, _ := GenerateResetToken()
	hash := HashResetToken(tok)

	if hash == tok {
		t.Error("hash must not equal the raw token")
	}
	if len(hash) != 64 {
		t.Errorf("expected 64-hex sha256, got len %d", len(hash))
	}
	if !VerifyResetToken(tok, hash) {
		t.Error("matching token should verify")
	}
	if VerifyResetToken("some-other-token", hash) {
		t.Error("non-matching token should not verify")
	}
	if VerifyResetToken(tok, "") {
		t.Error("empty stored hash should never verify")
	}
}
