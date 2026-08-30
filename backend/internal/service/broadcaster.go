package service

// Broadcaster pushes a lightweight real-time notification to clients
// subscribed to a board (see internal/ws.Hub). Implementations must not
// block the caller, since Broadcast is invoked on the request path.
type Broadcaster interface {
	Broadcast(boardID int64, eventType string)
}

// noopBroadcaster is used when no real broadcaster is wired (e.g. in
// tests), keeping services fully functional without WebSocket support.
type noopBroadcaster struct{}

func (noopBroadcaster) Broadcast(int64, string) {}

func withDefaultBroadcaster(b Broadcaster) Broadcaster {
	if b == nil {
		return noopBroadcaster{}
	}
	return b
}

const EventBoardUpdated = "board_updated"
