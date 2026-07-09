package model

import "time"

type Ingredient struct {
	Name          string `json:"name"`
	Quantity      string `json:"quantity"`
	Unit          string `json:"unit,omitempty"`
	IsBaseSpirit  bool   `json:"is_base_spirit,omitempty"`
}

type Recipe struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Ingredients []Ingredient      `json:"ingredients"`
	Steps       []string          `json:"steps"`
	Properties  map[string]string `json:"properties,omitempty"`
	Notes       string            `json:"notes"`
	Garnishes   []string          `json:"garnishes,omitempty"`
	CreatorID   string            `json:"creator_id"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type Favorite struct {
	UserID    string    `json:"user_id"`
	RecipeID  string    `json:"recipe_id"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	IsAdmin      bool      `json:"is_admin"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	TokenVersion int       `json:"token_version"`
	CreatedAt    time.Time `json:"created_at"`

	// Password-recovery state (never serialized to clients).
	ResetTokenHash    string `json:"-"` // SHA-256 hex of the active reset token
	ResetTokenExpires int64  `json:"-"` // unix seconds; 0 when no active token
	ResetWindowStart  int64  `json:"-"` // unix seconds; start of the rate-limit window
	ResetRequestCount int    `json:"-"` // requests made in the current window
}
