package handler

import "time"

// SetResetClock lets external tests control the handler's clock (for expiry and
// rate-limit windows).
func SetResetClock(h *PasswordResetHandler, f func() time.Time) { h.now = f }

// SetForgotFloor overrides the Forgot timing-neutrality minimum duration (tests
// set it to 0 to avoid sleeping).
func SetForgotFloor(d time.Duration) { forgotMinDuration = d }
