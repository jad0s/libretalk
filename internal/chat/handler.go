package chat

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"libretalk/internal/auth"
	"libretalk/internal/chat/store"
	"libretalk/internal/types"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Handler is the WebSocket entrypoint for chat.
func Handler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		http.Error(w, "upgrade failed", http.StatusBadRequest)
		return
	}

	// 1) Cleanup on disconnect
	defer func() {
		conn.Close()
		for user, list := range connections {
			for i, ci := range list {
				if ci.Conn == conn {
					connections[user] = append(list[:i], list[i+1:]...)
					break
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteJSON(map[string]string{"type": "ping"}); err != nil {
				log.Println("json-ping failed, closing:", err)
				conn.Close()
				return
			}
		}
	}()

	// 3) Start ping loop in a separate goroutine

	// 4) Main read loop
	for {
		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		// Peek at "type"
		var peek map[string]interface{}
		if err := json.Unmarshal(rawMsg, &peek); err != nil {
			conn.WriteJSON(map[string]string{"type": "error", "msg": "invalid JSON"})
			continue
		}
		t, _ := peek["type"].(string)

		switch t {

		// ─── AUTH ACTIONS ─────────────────────────────────────────────────────────
		case "ping":
			// client is checking server; echo back
			conn.WriteJSON(map[string]string{"type": "pong"})
			continue
		case "pong":
			// client responded to our ping; could update lastSeen here
			continue
		}

		switch t {

		case "action":
			var req types.ActionRequest
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
				// 1) authenticate + get token
				token, err := auth.Login(db, req.Username, req.Password)
				if err != nil {
					conn.WriteJSON(map[string]string{"type": "error", "msg": err.Error()})
					continue
				}
				// 2) send token back
				conn.WriteJSON(map[string]string{
					"type":   "login",
					"status": "ok",
					"token":  token,
				})
				// 3) register this connection
				ci := types.ConnectionInfo{
					Conn:        conn,
					IP:          conn.RemoteAddr().String(),
					ConnectedAt: time.Now(),
				}
				connections[req.Username] = append(connections[req.Username], ci)
				// 4) replay undelivered messages
				undelivered, err := store.LoadUndelivered(db, req.Username)
				if err != nil {
					log.Println("LoadUndelivered error:", err)
				}
				for _, row := range undelivered {
					conn.WriteJSON(types.IncomingMessage{
						Type:        "message",
						From:        row.Sender,
						To:          row.Recipient,
						ContentType: row.ContentType,
						Content:     row.Content,
					})
				}

			default:
				conn.WriteJSON(map[string]string{"type": "error", "msg": "unknown action"})
			}

		// ─── CHAT MESSAGES ────────────────────────────────────────────────────────
		case "message":
			var im types.IncomingMessage
			if err := json.Unmarshal(rawMsg, &im); err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "msg": "bad message"})
				continue
			}
			// verify JWT
			user, err := auth.ParseToken(im.Token)
			if err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "msg": "invalid token"})
				continue
			}
			if im.From != user {
				conn.WriteJSON(map[string]string{"type": "error", "msg": "sender mismatch"})
				continue
			}
			im.From = user

			// persist
			msgID, err := store.SaveMessage(db, im.From, im.To, im.ContentType, im.Content)
			if err != nil {
				log.Println("SaveMessage error:", err)
			}

			// deliver to all online devices
			for _, ci := range connections[im.To] {
				ci.Conn.WriteJSON(im)
			}

			// mark delivered
			if err := store.MarkDelivered(db, msgID); err != nil {
				log.Println("MarkDelivered error:", err)
			}

		// ─── HISTORY REQUEST ───────────────────────────────────────────────────────
		case "history":
			var req types.HistoryRequest
			if err := json.Unmarshal(rawMsg, &req); err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "msg": "bad history request"})
				continue
			}
			// authenticate
			user, err := auth.ParseToken(req.Token)
			if err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "msg": "invalid token"})
				continue
			}
			// load last N messages
			rows, err := store.LoadHistory(db, user, req.ChatWith, req.Limit)
			if err != nil {
				log.Println("LoadHistory error:", err)
				continue
			}
			// send each back
			for _, row := range rows {
				conn.WriteJSON(types.IncomingMessage{
					Type:        "message",
					From:        row.Sender,
					To:          row.Recipient,
					ContentType: row.ContentType,
					Content:     row.Content,
				})
			}

		// ─── UNKNOWN TYPE ─────────────────────────────────────────────────────────
		default:
			conn.WriteJSON(map[string]string{"type": "error", "msg": "unknown type"})
		}
	}
}
