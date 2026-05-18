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
	CreatorID   string            `json:"creator_id"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
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
}
