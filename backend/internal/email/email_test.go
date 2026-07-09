package email

import (
	"strings"
	"testing"
)

func TestBuildResetEmail_ContentContract(t *testing.T) {
	url := "https://cocktails.albertomcastro.com/#/reset?uid=u1&token=abc123"
	msg := BuildResetEmail(PasswordResetData{ResetURL: url, ExpiryMins: 15})

	if !strings.Contains(strings.ToLower(msg.Subject), "reset") {
		t.Errorf("subject should mention reset: %q", msg.Subject)
	}
	for _, part := range []string{msg.HTML, msg.Text} {
		if strings.Count(part, url) < 1 {
			t.Error("email must contain the reset link (E2)")
		}
		if !strings.Contains(part, "15 minutes") {
			t.Error("email must state the 15-minute expiry (E3)")
		}
		if strings.Contains(strings.ToLower(part), "password:") || strings.Contains(part, "token=") && !strings.Contains(part, url) {
			t.Error("email must not leak a credential (E4)")
		}
	}
	if !strings.Contains(msg.HTML, "Cocktail Recipes") {
		t.Error("HTML must carry the brand (E6)")
	}
}

func TestStubSender_Records(t *testing.T) {
	s := &StubSender{}
	_ = s.SendPasswordReset("a@b.com", PasswordResetData{ResetURL: "x", ExpiryMins: 15})
	if s.Count() != 1 || s.Sent[0].To != "a@b.com" {
		t.Errorf("stub did not record send: %+v", s.Sent)
	}
}
