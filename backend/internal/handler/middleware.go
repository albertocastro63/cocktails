package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/google/uuid"

	"github.com/almc/cocktails/internal/auth"
	"github.com/almc/cocktails/internal/logging"
	"github.com/almc/cocktails/internal/store"
)

// RequestLogger installs a request-scoped structured logger into the request
// context, bound to a correlation id (rid) and the request line (method+path),
// so every entry a handler emits while serving the request can be grouped.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := logging.FromContext(r.Context()).With(
			"rid", requestID(r.Context()),
			"req", r.Method+" "+r.URL.Path,
		)
		next.ServeHTTP(w, r.WithContext(logging.IntoContext(r.Context(), l)))
	})
}

// requestID reuses the platform-provided request id (API Gateway v2, then the
// Lambda request id), falling back to a generated id off-Lambda / in tests.
func requestID(ctx context.Context) string {
	if pc, ok := core.GetAPIGatewayV2ContextFromContext(ctx); ok && pc.RequestID != "" {
		return pc.RequestID
	}
	if lc, ok := lambdacontext.FromContext(ctx); ok && lc.AwsRequestID != "" {
		return lc.AwsRequestID
	}
	return uuid.NewString()
}

// Recover turns an unhandled panic into one ERROR log entry (with request
// correlation) and a generic 500, so an uncaught failure is captured and never
// leaks a stack trace to the client.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logging.FromContext(r.Context()).Error("panic recovered",
					"action", "panic", "outcome", "failure", "error", fmt.Sprint(rec))
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type contextKey string

const claimsKey contextKey = "claims"

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			return
		}
		claims, err := auth.Parse(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuthWithStore extends RequireAuth with a DB lookup to validate token_version
// and detect deleted users, enabling immediate session invalidation.
func RequireAuthWithStore(us store.UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
				return
			}
			claims, err := auth.Parse(token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
				return
			}
			user, err := us.GetByID(claims.UserID)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not found")
				return
			}
			if user.TokenVersion != claims.TokenVersion {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "token has been invalidated")
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := ClaimsFromContext(r.Context())
		if claims == nil || !claims.IsAdmin {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ClaimsFromContext(ctx context.Context) *auth.Claims {
	c, _ := ctx.Value(claimsKey).(*auth.Claims)
	return c
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
