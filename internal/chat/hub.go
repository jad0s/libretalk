package chat

import (
	"time"

	"github.com/gorilla/websocket"
)

// ConnectionInfo holds a single WS connection and some metadata.
type ConnectionInfo struct {
	Conn        *websocket.Conn // the WebSocket itself
	IP          string          // conn.RemoteAddr().String()
	ConnectedAt time.Time       // when this connection was opened
}

// connections maps username -> all active connections for that user.
var connections = make(map[string][]ConnectionInfo)
