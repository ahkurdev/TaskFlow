package ws

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Same-origin isn't guaranteed to match in dev (Vite dev server proxies
	// to a different port), so origin checks are relaxed here. Tighten this
	// if the API is ever exposed directly to untrusted origins.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Upgrade promotes an HTTP connection to a WebSocket, subscribes it to
// boardID on hub, and blocks running its read pump until the connection
// closes (the write pump runs in its own goroutine).
func Upgrade(hub *Hub, w http.ResponseWriter, r *http.Request, boardID int64) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	client := &Client{send: make(chan []byte, 16)}
	hub.register(boardID, client)

	go client.writePump(conn)
	client.readPump(conn, hub, boardID)

	return nil
}

// readPump only exists to detect client disconnects and respond to pongs;
// this server doesn't expect any meaningful messages from clients.
func (c *Client) readPump(conn *websocket.Conn, hub *Hub, boardID int64) {
	defer func() {
		hub.unregister(boardID, c)
		close(c.send)
		conn.Close()
	}()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (c *Client) writePump(conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
