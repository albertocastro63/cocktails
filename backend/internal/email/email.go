// Package email builds and sends transactional emails. The Sender interface is
// the seam that lets handlers be tested with a stub instead of live SES.
package email

import "fmt"

// PasswordResetData is the input for a password-reset email.
type PasswordResetData struct {
	ResetURL   string // full https link to the reset page
	ExpiryMins int    // link validity in minutes (15)
}

// Message is a rendered email (multipart: HTML + plain text).
type Message struct {
	Subject string
	HTML    string
	Text    string
}

// Sender delivers a reset email to a recipient address.
type Sender interface {
	SendPasswordReset(to string, data PasswordResetData) error
}

// BuildResetEmail renders the branded reset email. Pure and testable — no
// credential is ever included; the only actionable content is the reset link.
func BuildResetEmail(data PasswordResetData) Message {
	subject := "Reset your Cocktail Recipes password"

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en"><body style="margin:0;background:#fafaf9;font-family:ui-sans-serif,system-ui,-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;">
  <div style="max-width:520px;margin:0 auto;">
    <div style="background:#1c1917;color:#fff;padding:20px 24px;font-weight:700;font-size:18px;">Cocktail Recipes</div>
    <div style="padding:24px;color:#292524;">
      <h1 style="font-size:20px;margin:0 0 12px;">Reset your password</h1>
      <p style="margin:0 0 16px;color:#57534e;">We received a request to reset your password. Click the button below to choose a new one. This link expires in %d minutes.</p>
      <p style="margin:0 0 24px;">
        <a href="%s" style="display:inline-block;background:#b45309;color:#fff;text-decoration:none;font-weight:600;padding:10px 20px;border-radius:8px;">Reset password</a>
      </p>
      <p style="margin:0 0 8px;color:#57534e;font-size:13px;">Or paste this link into your browser:</p>
      <p style="margin:0 0 24px;word-break:break-all;font-size:13px;"><a href="%s" style="color:#b45309;">%s</a></p>
      <p style="margin:0;color:#78716c;font-size:13px;">Didn't request this? You can safely ignore this email — your password won't change.</p>
    </div>
  </div>
</body></html>`, data.ExpiryMins, data.ResetURL, data.ResetURL, data.ResetURL)

	text := fmt.Sprintf("Reset your Cocktail Recipes password\n\n"+
		"We received a request to reset your password. Open this link to choose a new one "+
		"(it expires in %d minutes):\n\n%s\n\n"+
		"Didn't request this? You can safely ignore this email — your password won't change.\n",
		data.ExpiryMins, data.ResetURL)

	return Message{Subject: subject, HTML: html, Text: text}
}
