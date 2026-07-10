// Package logging provides the backend's structured, level-controlled logging.
//
// Log entries are emitted as JSON (one object per line) to stdout, where AWS
// Lambda forwards them to CloudWatch for field-based querying in Logs Insights.
// The active minimum severity is selected by the LOG_LEVEL environment variable
// (see LevelFromEnv); production defaults to warn and previews to debug, and the
// value is overridable from the AWS console without a redeploy.
//
// Handlers obtain a request-scoped logger via FromContext (installed by the
// handler.RequestLogger middleware, which binds the request-correlation id and
// method+path). The canonical action names and their severities are defined in
// specs/027-backend-logging/contracts/action-catalog.md.
//
// Secrets (passwords, session tokens, password-reset tokens) must never be
// passed into a log call at any level.
package logging
