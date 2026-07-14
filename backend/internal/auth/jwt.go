package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultTokenDuration = 24 * time.Hour

// tokenDuration is how long an issued JWT stays valid. It defaults to 24h but
// can be overridden with JWT_TOKEN_DURATION (any Go duration string, e.g. "30s",
// "5m") — handy for exercising the session-expiry / auto-logout flow without a
// 24h wait. An unset or unparseable value falls back to the default.
func tokenDuration() time.Duration {
	if v := os.Getenv("JWT_TOKEN_DURATION"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return defaultTokenDuration
}

type Claims struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	IsAdmin      bool   `json:"is_admin"`
	TokenVersion int    `json:"token_version"`
	jwt.RegisteredClaims
}

func secret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		panic("JWT_SECRET environment variable is required")
	}
	return []byte(s)
}

func Issue(userID, username string, isAdmin bool, tokenVersion int) (string, time.Time, error) {
	exp := time.Now().Add(tokenDuration())
	claims := Claims{
		UserID:       userID,
		Username:     username,
		IsAdmin:      isAdmin,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret())
	return signed, exp, err
}

func Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret(), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
