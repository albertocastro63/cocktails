package auth

import "testing"

func TestValidateComplexity(t *testing.T) {
	cases := []struct {
		name string
		pw   string
		ok   bool
	}{
		{"valid", "Abcdefgh1j!k", true},
		{"valid boundary symbol underscore", "Abcdefgh1j_k", true},
		{"valid boundary symbol tilde", "Abcdefgh1j~k", true},
		{"too short", "Abc1!", false},
		{"exactly 12 ok", "Abcdefghij1!", true},
		{"too long over 72 bytes", "Aa1!" + string(make([]byte, 0)) + repeat("a", 70), false},
		{"missing upper", "abcdefgh1j!k", false},
		{"missing lower", "ABCDEFGH1J!K", false},
		{"missing digit", "Abcdefghij!k", false},
		{"missing symbol", "Abcdefghij1k", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateComplexity(c.pw)
			if c.ok && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !c.ok && err == nil {
				t.Errorf("expected error for %q, got nil", c.pw)
			}
		})
	}
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
