package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/almc/cocktails/internal/auth"
	"github.com/almc/cocktails/internal/email"
	"github.com/almc/cocktails/internal/handler"
	"github.com/almc/cocktails/internal/model"
)

func resetUser() *model.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPassword1!"), bcrypt.MinCost)
	return &model.User{ID: "u1", Username: "alberto", Email: "a@b.com", PasswordHash: string(hash), TokenVersion: 1}
}

func postJSON(h http.HandlerFunc, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

// ---- Forgot (US1) ----

func TestForgot_KnownEmail_SendsAndStoresToken(t *testing.T) {
	us := newStubUserStore(resetUser())
	sender := &email.StubSender{}
	h := handler.NewPasswordResetHandler(us, sender, "https://cocktails.albertomcastro.com")

	rec := postJSON(h.Forgot, `{"email":"a@b.com"}`)
	if rec.Code != 200 {
		t.Fatalf("F1 status: got %d", rec.Code)
	}
	if sender.Count() != 1 {
		t.Errorf("F1 expected 1 email, got %d", sender.Count())
	}
	u, _ := us.GetByEmail("a@b.com")
	if u.ResetTokenHash == "" || u.ResetTokenExpires == 0 {
		t.Errorf("F1 token not stored: %+v", u)
	}
	if !strings.Contains(sender.Sent[0].Data.ResetURL, "uid=u1&token=") {
		t.Errorf("F1 link malformed: %s", sender.Sent[0].Data.ResetURL)
	}
}

func TestForgot_UnknownEmail_NeutralNoSend(t *testing.T) {
	us := newStubUserStore(resetUser())
	sender := &email.StubSender{}
	h := handler.NewPasswordResetHandler(us, sender, "https://x")
	known := postJSON(h.Forgot, `{"email":"a@b.com"}`).Body.String()
	unknown := postJSON(h.Forgot, `{"email":"nope@x.com"}`).Body.String()
	if sender.Count() != 1 {
		t.Errorf("F2 expected only the known email to send, got %d", sender.Count())
	}
	if known != unknown {
		t.Errorf("F5 responses differ: %q vs %q", known, unknown)
	}
}

func TestForgot_RateLimit_SixPerHour(t *testing.T) {
	us := newStubUserStore(resetUser())
	sender := &email.StubSender{}
	h := handler.NewPasswordResetHandler(us, sender, "https://x")
	fixed := time.Unix(1_000_000, 0)
	handler.SetResetClock(h, func() time.Time { return fixed })

	for i := 0; i < 8; i++ {
		if rec := postJSON(h.Forgot, `{"email":"a@b.com"}`); rec.Code != 200 {
			t.Fatalf("status %d on request %d", rec.Code, i)
		}
	}
	if sender.Count() != 6 {
		t.Errorf("F3 expected 6 emails within the hour, got %d", sender.Count())
	}
}

func TestForgot_BadBody(t *testing.T) {
	h := handler.NewPasswordResetHandler(newStubUserStore(), &email.StubSender{}, "https://x")
	if postJSON(h.Forgot, `{}`).Code != 400 {
		t.Error("F4 missing email should be 400")
	}
}

// ---- Reset (US2 + US3) ----

func withActiveToken(u *model.User, token string, expires int64) {
	u.ResetTokenHash = auth.HashResetToken(token)
	u.ResetTokenExpires = expires
}

func TestReset_Valid_UpdatesPasswordAndInvalidatesSessions(t *testing.T) {
	u := resetUser()
	now := time.Unix(1_000_000, 0)
	withActiveToken(u, "goodtoken", now.Unix()+900)
	us := newStubUserStore(u)
	h := handler.NewPasswordResetHandler(us, &email.StubSender{}, "https://x")
	handler.SetResetClock(h, func() time.Time { return now })

	rec := postJSON(h.Reset, `{"uid":"u1","token":"goodtoken","password":"NewStrong1!aa"}`)
	if rec.Code != 200 {
		t.Fatalf("R1 status: got %d body %s", rec.Code, rec.Body.String())
	}
	got, _ := us.GetByID("u1")
	if bcrypt.CompareHashAndPassword([]byte(got.PasswordHash), []byte("NewStrong1!aa")) != nil {
		t.Error("R1 new password should verify")
	}
	if bcrypt.CompareHashAndPassword([]byte(got.PasswordHash), []byte("OldPassword1!")) == nil {
		t.Error("R1 old password should no longer verify")
	}
	if got.TokenVersion != 2 {
		t.Errorf("R1/R5 TokenVersion should bump to 2, got %d", got.TokenVersion)
	}
	if got.ResetTokenHash != "" {
		t.Error("R1 token should be consumed (cleared)")
	}
}

func TestReset_WeakPassword_422(t *testing.T) {
	u := resetUser()
	now := time.Unix(1_000_000, 0)
	withActiveToken(u, "goodtoken", now.Unix()+900)
	h := handler.NewPasswordResetHandler(newStubUserStore(u), &email.StubSender{}, "https://x")
	handler.SetResetClock(h, func() time.Time { return now })
	if rec := postJSON(h.Reset, `{"uid":"u1","token":"goodtoken","password":"weak"}`); rec.Code != 422 {
		t.Errorf("R3 weak password should be 422, got %d", rec.Code)
	}
}

func TestReset_InvalidCases_GenericError(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	cases := []struct {
		name  string
		setup func(u *model.User)
		body  string
	}{
		{"expired", func(u *model.User) { withActiveToken(u, "goodtoken", now.Unix()-1) }, `{"uid":"u1","token":"goodtoken","password":"NewStrong1!aa"}`},
		{"used/cleared", func(u *model.User) {}, `{"uid":"u1","token":"goodtoken","password":"NewStrong1!aa"}`},
		{"tampered", func(u *model.User) { withActiveToken(u, "goodtoken", now.Unix()+900) }, `{"uid":"u1","token":"WRONG","password":"NewStrong1!aa"}`},
		{"unknown uid", func(u *model.User) { withActiveToken(u, "goodtoken", now.Unix()+900) }, `{"uid":"nope","token":"goodtoken","password":"NewStrong1!aa"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			u := resetUser()
			c.setup(u)
			h := handler.NewPasswordResetHandler(newStubUserStore(u), &email.StubSender{}, "https://x")
			handler.SetResetClock(h, func() time.Time { return now })
			rec := postJSON(h.Reset, c.body)
			if rec.Code != 400 || !strings.Contains(rec.Body.String(), "INVALID_RESET") {
				t.Errorf("R2 %s: expected generic 400 INVALID_RESET, got %d %s", c.name, rec.Code, rec.Body.String())
			}
		})
	}
}
