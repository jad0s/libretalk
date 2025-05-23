package store

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jad0s/libretalk/internal/types"
)

// SaveMessage writes a new message to db and returns its auto-increment ID.
func SaveMessage(db *sql.DB, sender, recipient, contentType, content string) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO messages (sender, recipient, content_type, content)
		VALUES (?, ?, ?, ?)`,
		sender, recipient, contentType, content,
	)
	if err != nil {
		return 0, fmt.Errorf("save message: %w", err)
	}
	msgID, _ := res.LastInsertId()
	user1, user2 := sortTwoUsers(sender, recipient)
	if _, err := db.Exec(`
		INSERT INTO conversations (user1, user2, last_message, updated_at)
        VALUES (?, ?, ?, NOW())
        ON DUPLICATE KEY UPDATE
          last_message = VALUES(last_message),
          updated_at   = VALUES(updated_at)
    `, user1, user2, content); err != nil {
		// Log the error but don’t fail the whole send
		log.Println("upsert conversation:", err)
	}
	return msgID, nil
}

// MarkDelivered flips the delivered flag and stamps delivered_at.
func MarkDelivered(db *sql.DB, msgID int64) error {
	_, err := db.Exec(
		`UPDATE messages
		 SET delivered = TRUE, delivered_at = ?
		 WHERE id = ?`,
		time.Now(), msgID,
	)
	return err
}

// LoadUndelivered fetches all undelivered messages for a user,
// in ascending sent_at order, and marks them delivered.
func LoadUndelivered(db *sql.DB, username string) ([]types.MessageRow, error) {
	rows, err := db.Query(`
		SELECT id, sender, recipient, content_type, content, sent_at
		  FROM messages
		 WHERE recipient = ? AND delivered = FALSE
	     ORDER BY sent_at`,
		username,
	)
	if err != nil {
		return nil, fmt.Errorf("load undelivered: %w", err)
	}
	defer rows.Close()

	var msgs []types.MessageRow
	var ids []int64
	for rows.Next() {
		var m types.MessageRow
		if err := rows.Scan(
			&m.ID, &m.Sender, &m.Recipient,
			&m.ContentType, &m.Content, &m.SentAt,
		); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msgs = append(msgs, m)
		ids = append(ids, m.ID)
	}

	if len(ids) > 0 {
		// Build "IN (?,?,...)" placeholder list
		ph := strings.Repeat("?,", len(ids))
		ph = ph[:len(ph)-1]
		query := fmt.Sprintf(
			"UPDATE messages SET delivered = TRUE, delivered_at = ? WHERE id IN (%s)",
			ph,
		)
		args := make([]interface{}, len(ids)+1)
		args[0] = time.Now()
		for i, v := range ids {
			args[i+1] = v
		}
		if _, err := db.Exec(query, args...); err != nil {
			return nil, fmt.Errorf("mark delivered: %w", err)
		}
	}

	return msgs, nil
}

// LoadHistory fetches the most recent `limit` messages exchanged
// between `user` and `withUser`, in chronological order (oldest first).
func LoadHistory(db *sql.DB, user, withUser string, limit int) ([]types.MessageRow, error) {
	// Query newest first, limited
	rows, err := db.Query(`
        SELECT id, sender, recipient, content_type, content, sent_at
          FROM messages
         WHERE (sender = ? AND recipient = ?)
            OR (sender = ? AND recipient = ?)
         ORDER BY sent_at DESC
         LIMIT ?`,
		user, withUser,
		withUser, user,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("load history: %w", err)
	}
	defer rows.Close()

	var msgs []types.MessageRow
	for rows.Next() {
		var m types.MessageRow
		var rawTime []byte // placeholder for the DATETIME column
		if err := rows.Scan(
			&m.ID,
			&m.Sender,
			&m.Recipient,
			&m.ContentType,
			&m.Content,
			&rawTime,
		); err != nil {
			return nil, fmt.Errorf("scan history row: %w", err)
		}

		// MySQL DATETIME comes back as "YYYY-MM-DD HH:MM:SS"
		t, err := time.Parse("2006-01-02 15:04:05", string(rawTime))
		if err != nil {
			return nil, fmt.Errorf("parse sent_at: %w", err)
		}
		m.SentAt = t

		msgs = append(msgs, m)
	}

	// Reverse so oldest are first
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	return msgs, nil
}

func LoadChats(db *sql.DB, me string) ([]types.Chat, error) {
	const q = `
	  SELECT
		CASE
		  WHEN user1 = ? THEN user2
		  ELSE user1
		END AS peer,
		last_message,
		updated_at
	  FROM conversations
	  WHERE user1 = ? OR user2 = ?
	  ORDER BY updated_at DESC
	`
	rows, err := db.Query(q, me, me, me)
	if err != nil {
		return nil, fmt.Errorf("LoadChats query: %w", err)
	}
	defer rows.Close()

	var chats []types.Chat
	for rows.Next() {
		var c types.Chat
		if err := rows.Scan(&c.With, &c.LastMessage, &c.LastMessageTime); err != nil {
			return nil, fmt.Errorf("LoadChats scan: %w", err)
		}
		chats = append(chats, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("LoadChats rows error: %w", err)
	}
	return chats, nil
}

func sortTwoUsers(a, b string) (string, string) {
	if a < b {
		return a, b
	}
	return b, a
}
