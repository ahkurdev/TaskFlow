package model

import "time"

// Card is a single task item within a List.
type Card struct {
	ID          int64     `json:"id"`
	ListID      int64     `json:"list_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
