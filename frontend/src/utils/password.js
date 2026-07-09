// Shared password-complexity rules for live UX feedback. Mirrors the backend
// authoritative check (auth.ValidateComplexity): 12–72 bytes with at least one
// upper, lower, digit, and symbol (any non-alphanumeric printable ASCII).

export const PASSWORD_RULES = [
  { id: 'length', label: 'At least 12 characters', test: (p) => byteLength(p) >= 12 && byteLength(p) <= 72 },
  { id: 'upper', label: 'An upper-case letter', test: (p) => /[A-Z]/.test(p) },
  { id: 'lower', label: 'A lower-case letter', test: (p) => /[a-z]/.test(p) },
  { id: 'digit', label: 'A number', test: (p) => /[0-9]/.test(p) },
  { id: 'symbol', label: 'A symbol', test: (p) => /[!-/:-@[-`{-~]/.test(p) },
];

function byteLength(s) {
  return new TextEncoder().encode(s).length;
}

// Returns the list of unmet rule ids (empty when the password is valid).
export function passwordProblems(password) {
  return PASSWORD_RULES.filter((r) => !r.test(password)).map((r) => r.id);
}

export function isPasswordValid(password) {
  return passwordProblems(password).length === 0;
}
