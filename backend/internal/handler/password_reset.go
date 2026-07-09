package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/almc/cocktails/internal/auth"
	"github.com/almc/cocktails/internal/email"
	"github.com/almc/cocktails/internal/store"
)

const (
	resetExpiryMins   = 15
	resetRateLimit    = 6    // max requests per window per user
	resetRateWindow   = 3600 // seconds (1 hour)
	neutralResetReply = "If an account exists for that email, a password reset link has been sent."
)

// Timing-neutrality mitigation (feature 026, research Decision 5): the Forgot
// handler enforces a uniform minimum wall-clock duration before responding, so
// the response time does not leak whether an account exists (the real-account
// path performs an extra DynamoDB Update + SES send). Package-level so tests can
// disable it via a TestMain. Overridable in prod is unnecessary — 250ms floors
// well above the observable difference while staying imperceptible to users.
var (
	forgotMinDuration = 250 * time.Millisecond
	forgotSleep       = time.Sleep
)

// PasswordResetHandler serves the public forgot/reset endpoints.
type PasswordResetHandler struct {
	users   store.UserStore
	sender  email.Sender
	baseURL string // e.g. https://cocktails.albertomcastro.com
	now     func() time.Time
}

func NewPasswordResetHandler(us store.UserStore, sender email.Sender, baseURL string) *PasswordResetHandler {
	return &PasswordResetHandler{users: us, sender: sender, baseURL: strings.TrimRight(baseURL, "/"), now: time.Now}
}

// Forgot handles POST /api/v1/auth/forgot-password. It always returns the same
// neutral response (no account enumeration); an email is sent only for a real
// account that is under the per-user rate limit.
func (h *PasswordResetHandler) Forgot(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Email) == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "email is required")
		return
	}
	// Floor the total handler time so account-existence isn't observable via latency.
	start := h.now()
	neutral := func() {
		if forgotMinDuration > 0 {
			if elapsed := h.now().Sub(start); elapsed < forgotMinDuration {
				forgotSleep(forgotMinDuration - elapsed)
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"message": neutralResetReply})
	}

	user, err := h.users.GetByEmail(strings.TrimSpace(body.Email))
	if err != nil {
		neutral() // unknown email — same response, no email
		return
	}

	now := h.now().Unix()
	switch {
	case now-user.ResetWindowStart >= resetRateWindow:
		user.ResetWindowStart = now
		user.ResetRequestCount = 1
	case user.ResetRequestCount < resetRateLimit:
		user.ResetRequestCount++
	default:
		neutral() // over the rate limit — same response, no email
		return
	}

	token, err := auth.GenerateResetToken()
	if err != nil {
		neutral()
		return
	}
	user.ResetTokenHash = auth.HashResetToken(token)
	user.ResetTokenExpires = now + resetExpiryMins*60
	if err := h.users.Update(user); err != nil {
		neutral()
		return
	}

	link := fmt.Sprintf("%s/#/reset?uid=%s&token=%s", h.baseURL, url.QueryEscape(user.ID), url.QueryEscape(token))
	_ = h.sender.SendPasswordReset(user.Email, email.PasswordResetData{ResetURL: link, ExpiryMins: resetExpiryMins})
	neutral()
}

// Reset handles POST /api/v1/auth/reset-password. It validates the token, then
// the new password, updates it, invalidates existing sessions, and consumes the
// token.
func (h *PasswordResetHandler) Reset(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UID      string `json:"uid"`
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	invalid := func() {
		writeError(w, http.StatusBadRequest, "INVALID_RESET", "this reset link is invalid or has expired")
	}

	user, err := h.users.GetByID(body.UID)
	if err != nil {
		invalid()
		return
	}
	now := h.now().Unix()
	if user.ResetTokenExpires < now || !auth.VerifyResetToken(body.Token, user.ResetTokenHash) {
		invalid()
		return
	}
	if err := auth.ValidateComplexity(body.Password); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "WEAK_PASSWORD", err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to set password")
		return
	}
	user.PasswordHash = string(hash)
	user.TokenVersion++ // invalidate existing sessions
	user.ResetTokenHash = ""
	user.ResetTokenExpires = 0
	if err := h.users.Update(user); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to set password")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": "Your password has been reset. You can now sign in."})
}
