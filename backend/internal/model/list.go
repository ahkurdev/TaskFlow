package model

import "time"

// List is a column within a Board (e.g. "To Do", "In Progress", "Done").
type List struct {
	ID        int64     `json:"id"`
	BoardID   int64     `json:"board_id"`
	Name      string    `json:"name"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListWithCards is a List along with its ordered cards.
type ListWithCards struct {
	List
	Cards []Card `json:"cards"`
}
