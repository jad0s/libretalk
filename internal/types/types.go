package types

// IncomingMessage is the payload shape for your WebSocket chat messages.
type IncomingMessage struct {
	Type        string `json:"type"`
	From        string `json:"from"`
	To          string `json:"to"`
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
	Token       string `json:"token"`
}
