package auth

import (
	"errors"
	"unicode"
)

// Password length bounds. The upper bound matches bcrypt's effective input
// limit (72 bytes) so a long password is never silently truncated.
const (
	MinPasswordBytes = 12
	MaxPasswordBytes = 72
)

// ValidateComplexity enforces the reset password policy: 12–72 bytes with at
// least one upper-case letter, lower-case letter, digit, and symbol. A "symbol"
// is any non-alphanumeric printable ASCII character. Returns a message naming
// the unmet requirement.
func ValidateComplexity(pw string) error {
	if len(pw) < MinPasswordBytes {
		return errors.New("password must be at least 12 characters")
	}
	if len(pw) > MaxPasswordBytes {
		return errors.New("password must be at most 72 characters")
	}
	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, r := range pw {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		case r >= 33 && r <= 126 && !unicode.IsLetter(r) && !unicode.IsDigit(r):
			hasSymbol = true
		}
	}
	switch {
	case !hasUpper:
		return errors.New("password must include an upper-case letter")
	case !hasLower:
		return errors.New("password must include a lower-case letter")
	case !hasDigit:
		return errors.New("password must include a number")
	case !hasSymbol:
		return errors.New("password must include a symbol")
	}
	return nil
}
