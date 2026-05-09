package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/almc/cocktails/internal/model"
	"github.com/almc/cocktails/internal/store"
	"github.com/google/uuid"
)

type AdminHandler struct {
	users store.UserStore
}

func NewAdminHandler(us store.UserStore) *AdminHandler {
	return &AdminHandler{users: us}
}

func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	if strings.TrimSpace(body.Username) == "" || body.Password == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "username and password are required")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to hash password")
		return
	}

	user := &model.User{
		ID:           uuid.NewString(),
		Username:     body.Username,
		PasswordHash: string(hash),
		IsAdmin:      false,
		CreatedAt:    time.Now().UTC(),
	}
	if err := h.users.Create(user); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			writeError(w, http.StatusConflict, "CONFLICT", "username already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create user")
		return
	}
	writeJSON(w, http.StatusCreated, user)
}
