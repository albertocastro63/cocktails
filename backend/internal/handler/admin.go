package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/almc/cocktails/internal/logging"
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

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := h.users.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}
	if user.IsAdmin {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "cannot manage admin accounts")
		return
	}
	if err := h.users.Delete(id); err != nil {
		logging.FromContext(r.Context()).Error("admin delete user failed", "action", "admin.user.delete",
			"outcome", "failure", "user_id", actorID(r), "target_id", id, "error", err.Error())
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete user")
		return
	}
	logging.FromContext(r.Context()).Info("admin deleted user", "action", "admin.user.delete",
		"outcome", "success", "user_id", actorID(r), "target_id", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := h.users.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}
	if user.IsAdmin {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "cannot manage admin accounts")
		return
	}
	logging.FromContext(r.Context()).Debug("admin fetched user", "action", "admin.user.get",
		"outcome", "success", "user_id", actorID(r), "target_id", id)
	writeJSON(w, http.StatusOK, user)
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := h.users.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}
	if user.IsAdmin {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "cannot manage admin accounts")
		return
	}

	var body struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if body.Email != "" && body.Email != user.Email {
		if existing, err := h.users.GetByEmail(body.Email); err == nil && existing.ID != id {
			writeError(w, http.StatusConflict, "EMAIL_CONFLICT", "email already in use")
			return
		}
	}

	user.FirstName = body.FirstName
	user.LastName = body.LastName
	user.Email = body.Email

	if body.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to hash password")
			return
		}
		user.PasswordHash = string(hash)
		user.TokenVersion++
	}

	if err := h.users.Update(user); err != nil {
		logging.FromContext(r.Context()).Error("admin update user failed", "action", "admin.user.update",
			"outcome", "failure", "user_id", actorID(r), "target_id", id, "error", err.Error())
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update user")
		return
	}
	logging.FromContext(r.Context()).Info("admin updated user", "action", "admin.user.update",
		"outcome", "success", "user_id", actorID(r), "target_id", id)
	writeJSON(w, http.StatusOK, user)
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List()
	if err != nil {
		logging.FromContext(r.Context()).Error("admin list users failed", "action", "admin.user.list",
			"outcome", "failure", "user_id", actorID(r), "error", err.Error())
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list users")
		return
	}
	logging.FromContext(r.Context()).Debug("admin listed users", "action", "admin.user.list",
		"outcome", "success", "user_id", actorID(r), "count", len(users))
	writeJSON(w, http.StatusOK, users)
}

func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	if strings.TrimSpace(body.Username) == "" || body.Password == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "username and password are required")
		return
	}

	if body.Email != "" {
		if _, err := h.users.GetByEmail(body.Email); err == nil {
			writeError(w, http.StatusConflict, "EMAIL_CONFLICT", "email already in use")
			return
		}
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
		FirstName:    body.FirstName,
		LastName:     body.LastName,
		Email:        body.Email,
		CreatedAt:    time.Now().UTC(),
	}
	if err := h.users.Create(user); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			logging.FromContext(r.Context()).Warn("admin create user conflict", "action", "admin.user.create",
				"outcome", "failure", "user_id", actorID(r), "reason", "duplicate_username")
			writeError(w, http.StatusConflict, "CONFLICT", "username already exists")
			return
		}
		logging.FromContext(r.Context()).Error("admin create user failed", "action", "admin.user.create",
			"outcome", "failure", "user_id", actorID(r), "error", err.Error())
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create user")
		return
	}
	logging.FromContext(r.Context()).Info("admin created user", "action", "admin.user.create",
		"outcome", "success", "user_id", actorID(r), "target_id", user.ID)
	writeJSON(w, http.StatusCreated, user)
}
