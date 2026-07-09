package handler

import "time"

// SetResetClock lets external tests control the handler's clock (for expiry and
// rate-limit windows).
func SetResetClock(h *PasswordResetHandler, f func() time.Time) { h.now = f }
