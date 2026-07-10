package handler

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/almc/cocktails/internal/auth"
	"github.com/almc/cocktails/internal/logging"
	"github.com/almc/cocktails/internal/store"
)

type AuthHandler struct {
	users store.UserStore
}

func NewAuthHandler(us store.UserStore) *AuthHandler {
	return &AuthHandler{users: us}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log := logging.FromContext(r.Context())
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Warn("login rejected", "action", "auth.login", "outcome", "failure", "reason", "invalid body")
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	user, err := h.users.GetByUsername(body.Username)
	if err != nil {
		// Auth rejection is a recoverable anomaly (WARN); username is a login
		// identifier, not a secret. Never log the submitted password.
		log.Warn("login rejected", "action", "auth.login", "outcome", "failure",
			"username", body.Username, "reason", "unknown user")
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		log.Warn("login rejected", "action", "auth.login", "outcome", "failure",
			"user_id", user.ID, "reason", "bad password")
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
		return
	}

	token, expiresAt, err := auth.Issue(user.ID, user.Username, user.IsAdmin, user.TokenVersion)
	if err != nil {
		log.Error("login token issue failed", "action", "auth.login", "outcome", "failure",
			"user_id", user.ID, "error", err.Error())
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to issue token")
		return
	}
	log.Info("login succeeded", "action", "auth.login", "outcome", "success", "user_id", user.ID)
	writeJSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"expires_at": expiresAt,
	})
}
