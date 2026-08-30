package model

import "time"

// Board is a Kanban board owned by a user.
type Board struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BoardDetail is a Board along with its ordered lists and cards, used for
// the "load full board" endpoint.
type BoardDetail struct {
	Board
	Lists []ListWithCards `json:"lists"`
}
