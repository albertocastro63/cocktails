package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/almc/cocktails/internal/auth"
	"github.com/almc/cocktails/internal/email"
	"github.com/almc/cocktails/internal/handler"
	"github.com/almc/cocktails/internal/logging"
)

// T018 — secrets must never appear in logs at any level (SC-004, FR-009).
func TestLog_NoSecretsInAuthAndReset(t *testing.T) {
	// Login: the submitted password must not appear anywhere in the output.
	buf, install := bufLogger()
	const pw = "SuperSecret-123!"
	us := newStubUserStore(userWithPassword("u1", "alice", pw, false))
	h := install(http.HandlerFunc(handler.NewAuthHandler(us).Login))
	h.ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"username":"alice","password":"`+pw+`"}`)))
	if strings.Contains(buf.String(), pw) {
		t.Errorf("submitted password leaked into logs: %s", buf.String())
	}

	// Reset: neither the reset token nor the new password may appear.
	buf2, install2 := bufLogger()
	const token = "TOK-abc-123"
	const newPw = "BrandNewPass-9!x"
	u := resetUser()
	u.ResetTokenHash = auth.HashResetToken(token)
	u.ResetTokenExpires = time.Now().Add(time.Hour).Unix()
	rh := handler.NewPasswordResetHandler(newStubUserStore(u), &email.StubSender{}, "https://x")
	h2 := install2(http.HandlerFunc(rh.Reset))
	body := `{"uid":"u1","token":"` + token + `","password":"` + newPw + `"}`
	h2.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(body)))

	out := buf2.String()
	if strings.Contains(out, token) {
		t.Errorf("reset token leaked into logs: %s", out)
	}
	if strings.Contains(out, newPw) {
		t.Errorf("new password leaked into logs: %s", out)
	}
	findEntry(t, buf2, "password.reset", "success", "INFO")
}

// T019 — all lines emitted while handling one request share the same rid (SC-005).
func TestLog_RequestCorrelationGroupsLines(t *testing.T) {
	buf, install := bufLogger()
	multi := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := logging.FromContext(r.Context())
		l.Info("first", "action", "demo.one")
		l.Info("second", "action", "demo.two")
	})
	h := install(handler.RequestLogger(multi))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/thing", nil))

	var rids []string
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		var e map[string]any
		if json.Unmarshal([]byte(line), &e) != nil {
			continue
		}
		rid, _ := e["rid"].(string)
		if rid == "" {
			t.Fatalf("entry missing rid: %s", line)
		}
		rids = append(rids, rid)
	}
	if len(rids) < 2 {
		t.Fatalf("expected >=2 correlated lines, got %d", len(rids))
	}
	for _, rid := range rids[1:] {
		if rid != rids[0] {
			t.Errorf("rid mismatch across one request: %v", rids)
		}
	}
}
