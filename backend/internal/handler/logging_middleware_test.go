package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/almc/cocktails/internal/handler"
	"github.com/almc/cocktails/internal/logging"
)

// bufLogger installs a buffer-backed logger into the request context so tests
// can inspect what a handler emitted.
func bufLogger() (*bytes.Buffer, func(next http.Handler) http.Handler) {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(logging.IntoContext(r.Context(), l)))
		})
	}
	return &buf, mw
}

func TestRequestLoggerBindsRidAndReq(t *testing.T) {
	buf, install := bufLogger()

	var seen *slog.Logger
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = logging.FromContext(r.Context())
		seen.Info("did thing", "action", "recipe.get")
	})

	// install (buffer logger) -> RequestLogger (adds rid/req) -> final
	h := install(handler.RequestLogger(final))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recipes/42", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry); err != nil {
		t.Fatalf("not JSON: %v (%q)", err, buf.String())
	}
	if entry["rid"] == nil || entry["rid"] == "" {
		t.Errorf("expected non-empty rid, got %v", entry["rid"])
	}
	if entry["req"] != "GET /api/v1/recipes/42" {
		t.Errorf("req = %v, want %q", entry["req"], "GET /api/v1/recipes/42")
	}
}

func TestRecoverConvertsPanicToErrorAnd500(t *testing.T) {
	buf, install := bufLogger()

	boom := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("kaboom")
	})
	h := install(handler.Recover(boom))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/x", nil))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
	if strings.Contains(rr.Body.String(), "kaboom") {
		t.Errorf("panic value leaked to client body: %q", rr.Body.String())
	}
	out := buf.String()
	if !strings.Contains(out, "ERROR") || !strings.Contains(out, "kaboom") {
		t.Errorf("expected an ERROR log entry recording the panic, got %q", out)
	}
}

var _ = context.Background
