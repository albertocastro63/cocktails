package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/almc/cocktails/internal/logging"
)

func TestParseLevel(t *testing.T) {
	cases := []struct {
		in   string
		want slog.Level
		ok   bool
	}{
		{"debug", slog.LevelDebug, true},
		{"DEBUG", slog.LevelDebug, true},
		{"info", slog.LevelInfo, true},
		{" Info ", slog.LevelInfo, true},
		{"warn", slog.LevelWarn, true},
		{"warning", slog.LevelWarn, true},
		{"error", slog.LevelError, true},
		{"", slog.LevelError, false},
		{"bogus", slog.LevelError, false},
	}
	for _, c := range cases {
		got, ok := logging.ParseLevel(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("ParseLevel(%q) = (%v,%v), want (%v,%v)", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestNewEmitsJSONWithCoreFields(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	l.Info("hello", "action", "recipe.create", "outcome", "success")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("output is not JSON: %v (%q)", err, buf.String())
	}
	for _, f := range []string{"time", "level", "msg", "action", "outcome"} {
		if _, ok := entry[f]; !ok {
			t.Errorf("missing field %q in %v", f, entry)
		}
	}
	if entry["level"] != "INFO" {
		t.Errorf("level = %v, want INFO", entry["level"])
	}
}

func TestFromContextReturnsDefaultWhenUnset(t *testing.T) {
	if got := logging.FromContext(context.Background()); got == nil {
		t.Fatal("FromContext returned nil; must fall back to the default logger")
	}
}

func TestIntoAndFromContextRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, nil))
	ctx := logging.IntoContext(context.Background(), l)
	if got := logging.FromContext(ctx); got != l {
		t.Fatal("FromContext did not return the logger stored by IntoContext")
	}
}

func TestLevelFromEnv(t *testing.T) {
	t.Setenv("LOG_LEVEL", "warn")
	if got := logging.LevelFromEnv(); got != slog.LevelWarn {
		t.Errorf("LevelFromEnv(warn) = %v, want WARN", got)
	}
	t.Setenv("LOG_LEVEL", "bogus") // invalid -> error-only fallback
	if got := logging.LevelFromEnv(); got != slog.LevelError {
		t.Errorf("LevelFromEnv(bogus) = %v, want ERROR fallback", got)
	}
	t.Setenv("LOG_LEVEL", "") // missing -> error-only fallback
	if got := logging.LevelFromEnv(); got != slog.LevelError {
		t.Errorf("LevelFromEnv(empty) = %v, want ERROR fallback", got)
	}
}

func TestSetDefaultRoundTrip(t *testing.T) {
	l := logging.New(slog.LevelInfo)
	logging.SetDefault(l)
	if logging.Default() != l {
		t.Error("Default did not return the logger set by SetDefault")
	}
	logging.SetDefault(nil) // nil is ignored; default unchanged
	if logging.Default() != l {
		t.Error("SetDefault(nil) must not replace the default logger")
	}
}

func TestSuppressionByLevel(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	l.Debug("read", "action", "recipe.get") // below warn -> suppressed
	l.Error("boom", "action", "recipe.get") // at/above warn -> emitted
	out := buf.String()
	if strings.Contains(out, "recipe.get") && strings.Contains(out, "DEBUG") {
		t.Errorf("DEBUG entry should be suppressed at warn level: %q", out)
	}
	if !strings.Contains(out, "ERROR") {
		t.Errorf("ERROR entry should be emitted at warn level: %q", out)
	}
}
