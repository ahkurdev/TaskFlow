package ws

import (
	"encoding/json"
	"sync"
)

// Event is the lightweight real-time message pushed to board subscribers.
// Clients are expected to react by refetching the board rather than
// applying the payload directly, keeping the protocol simple and robust.
type Event struct {
	Type    string `json:"type"`
	BoardID int64  `json:"board_id"`
}

// Client represents a single WebSocket connection subscribed to a board.
type Client struct {
	send chan []byte
}

// Hub tracks which clients are subscribed to which boards and fans out
// broadcast events to them. Safe for concurrent use.
type Hub struct {
	mu      sync.RWMutex
	clients map[int64]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[int64]map[*Client]struct{})}
}

func (h *Hub) register(boardID int64, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[boardID] == nil {
		h.clients[boardID] = make(map[*Client]struct{})
	}
	h.clients[boardID][c] = struct{}{}
}

func (h *Hub) unregister(boardID int64, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients[boardID], c)
	if len(h.clients[boardID]) == 0 {
		delete(h.clients, boardID)
	}
}

// Broadcast notifies every client subscribed to boardID that something
// changed. It never blocks: slow/dead consumers have messages dropped
// rather than stalling the caller (which runs on the request path).
func (h *Hub) Broadcast(boardID int64, eventType string) {
	msg, err := json.Marshal(Event{Type: eventType, BoardID: boardID})
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients[boardID] {
		select {
		case c.send <- msg:
		default:
		}
	}
}
