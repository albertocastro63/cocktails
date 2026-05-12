package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenDuration = 24 * time.Hour

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
	exp := time.Now().Add(tokenDuration)
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
