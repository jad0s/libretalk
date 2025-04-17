package chat

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"libretalk/internal/auth"

	"log"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type IncomingMessage struct {
	Type        string `json:"type"`
	From        string `json:"from"`
	To          string `json:"to"`
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
}

type ActionRequest struct {
	Type     string `json:"type"`
	Action   string `json:"action"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Handler upgrades to WS, handles register/login, then echoes messages.
func Handler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		http.Error(w, "upgrade failed", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	for {
		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		// Inspect "type" field
		var peek map[string]interface{}
		if err := json.Unmarshal(rawMsg, &peek); err != nil {
			log.Println("invalid JSON:", err)
			continue
		}
		t, _ := peek["type"].(string)

		switch t {
		case "action":
			var req ActionRequest
			if err := json.Unmarshal(rawMsg, &req); err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "msg": "bad action"})
				continue
			}
			switch req.Action {
			case "register":
				if err := auth.Register(db, req.Username, req.Password); err != nil {
					conn.WriteJSON(map[string]string{"type": "error", "msg": err.Error()})
				} else {
					conn.WriteJSON(map[string]string{"type": "register", "status": "ok"})
				}
			case "login":
				if err := auth.Login(db, req.Username, req.Password); err != nil {
					conn.WriteJSON(map[string]string{"type": "error", "msg": err.Error()})
				} else {
					token, err := auth.GenerateToken(req.Username)
					if err != nil {
						conn.WriteJSON(map[string]string{"type": "error", "msg": "token error"})
					} else {
						conn.WriteJSON(map[string]string{"type": "login", "status": "ok", "token": token})
					}
				}
			default:
				conn.WriteJSON(map[string]string{"type": "error", "msg": "unknown action"})
			}

		case "message":
			var im IncomingMessage
			if err := json.Unmarshal(rawMsg, &im); err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "msg": "bad message"})
				continue
			}
			// TODO: verify sender with JWT, route to recipient, persist, etc.
			conn.WriteJSON(im)

		default:
			conn.WriteJSON(map[string]string{"type": "error", "msg": "unknown type"})
		}
	}
}
