package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// MessageRow mirrors your messages table.
type MessageRow struct {
	ID          int64
	Sender      string
	Recipient   string
	ContentType string
	Content     string
	SentAt      time.Time
}

// SaveMessage writes a new message and returns its auto-increment ID.
func SaveMessage(db *sql.DB, sender, recipient, contentType, content string) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO messages (sender, recipient, content_type, content)
		VALUES (?, ?, ?, ?)`,
		sender, recipient, contentType, content,
	)
	if err != nil {
		return 0, fmt.Errorf("save message: %w", err)
	}
	return res.LastInsertId()
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
func LoadUndelivered(db *sql.DB, username string) ([]MessageRow, error) {
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

	var msgs []MessageRow
	var ids []int64
	for rows.Next() {
		var m MessageRow
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
func LoadHistory(db *sql.DB, user, withUser string, limit int) ([]MessageRow, error) {
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

	var msgs []MessageRow
	for rows.Next() {
		var m MessageRow
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
