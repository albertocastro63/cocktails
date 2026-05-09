package handler

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/almc/cocktails/internal/auth"
	"github.com/almc/cocktails/internal/store"
)

type AuthHandler struct {
	users store.UserStore
}

func NewAuthHandler(us store.UserStore) *AuthHandler {
	return &AuthHandler{users: us}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	user, err := h.users.GetByUsername(body.Username)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
		return
	}

	token, expiresAt, err := auth.Issue(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to issue token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"expires_at": expiresAt,
	})
}
