# Email Contract: Password Reset Email

**Branch**: `026-password-recovery` | **Date**: 2026-07-09

## Sender / transport

- Sent via Amazon SES from `no-reply@cocktails.albertomcastro.com` (site domain, DKIM-signed).
- Multipart: HTML + plain-text alternative.

## Content

| # | Requirement | Maps to |
|---|-------------|---------|
| E1 | Subject clearly identifies a password reset for "Cocktail Recipes" | FR-004 |
| E2 | Contains exactly one reset link: `https://cocktails.albertomcastro.com/#/reset?uid=<id>&token=<token>` | FR-004 |
| E3 | States the link expires in 15 minutes | FR-005 |
| E4 | Contains NO password or other credential | FR-004 |
| E5 | Includes a "didn't request this? ignore this email" note | security UX |
| E6 | HTML matches the site design: dark stone header, "Cocktail Recipes" brand, amber call-to-action button, stone/amber palette | FR-013 |
| E7 | Plain-text alternative includes the same link and expiry note | accessibility/deliverability |

## Branding tokens (reuse the site palette)

- Header background: stone-900 (`#1c1917`); brand text light.
- Primary button: amber (`#b45309` / `#d97706`) with dark text, matching the site CTA.
- Body text: stone-800 on white; muted notes in stone-500.

## Notes

- The button and a fallback text link both point to the same reset URL (some clients strip buttons).
- No tracking pixels or external resources beyond what the design needs; keep it simple and deliverable.
