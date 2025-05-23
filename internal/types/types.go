package types

import (
	"time"

	"github.com/gorilla/websocket"
)

type IncomingMessage struct {
	Type        string `json:"type"`
	From        string `json:"from"`
	To          string `json:"to"`
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
	Token       string `json:"token"`
}

type ActionRequest struct {
	Type     string `json:"type"`
	Action   string `json:"action"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type HistoryRequest struct {
	Type     string `json:"type"`
	ChatWith string `json:"chatWith"`
	Token    string `json:"token"`
	Limit    int    `json:"limit"`
}

type MessageRow struct {
	ID          int64
	Sender      string
	Recipient   string
	ContentType string
	Content     string
	SentAt      time.Time
}

type ConnectionInfo struct {
	Conn        *websocket.Conn // the WebSocket itself
	IP          string          // conn.RemoteAddr().String()
	ConnectedAt time.Time       // when this connection was opened
}

type ChatsRequest struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

type Chat struct {
	With            string    `json:"with"`            // the *other* user
	LastMessage     string    `json:"lastMessage"`     // the snippet
	LastMessageTime time.Time `json:"lastMessageTime"` // sortable timestamp
}
