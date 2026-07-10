package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
)

// ctxKey is the private context key under which a request-scoped logger is stored.
type ctxKey struct{}

// defaultLogger is returned by FromContext when no request-scoped logger is set.
// It is stored atomically so SetDefault at startup is safe against concurrent
// reads. It starts at error-only until configured.
var defaultLogger atomic.Pointer[slog.Logger]

func init() {
	defaultLogger.Store(New(slog.LevelError))
}

// New builds a JSON logger writing to stdout at the given level. One JSON object
// per line — the shape CloudWatch Logs Insights can filter by field.
func New(level slog.Leveler) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

// SetDefault installs the process-wide default logger (called once at startup
// after the level is resolved from the environment).
func SetDefault(l *slog.Logger) {
	if l != nil {
		defaultLogger.Store(l)
	}
}

// Default returns the process-wide default logger (never nil).
func Default() *slog.Logger {
	return defaultLogger.Load()
}

// IntoContext returns a child context carrying the request-scoped logger.
func IntoContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext returns the request-scoped logger, or the process default when
// none is set. It never returns nil, so call sites can log unconditionally.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return Default()
}

// ParseLevel maps a level name (case-insensitive) to a slog.Level. ok is false
// for empty or unrecognized input, in which case the returned level is the
// safe error-only fallback.
func ParseLevel(s string) (slog.Level, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug, true
	case "info":
		return slog.LevelInfo, true
	case "warn", "warning":
		return slog.LevelWarn, true
	case "error":
		return slog.LevelError, true
	default:
		return slog.LevelError, false
	}
}

// LevelFromEnv resolves the active level from LOG_LEVEL. On a missing or invalid
// value it falls back to error-only and emits one ERROR line recording the
// fallback (emitted at error so it is visible even under the fallback level).
func LevelFromEnv() slog.Level {
	raw := os.Getenv("LOG_LEVEL")
	lvl, ok := ParseLevel(raw)
	if !ok {
		New(slog.LevelError).Error("LOG_LEVEL fallback applied",
			"requested", raw, "using", "error")
	}
	return lvl
}
