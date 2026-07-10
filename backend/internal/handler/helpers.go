package handler

import (
	"encoding/json"
	"net/http"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorBody{Error: errorDetail{Code: code, Message: message}})
}

// actorID returns the authenticated user's id from the request context, or ""
// when unauthenticated. Used to attribute log entries to the acting user.
func actorID(r *http.Request) string {
	if c := ClaimsFromContext(r.Context()); c != nil {
		return c.UserID
	}
	return ""
}
