package auth

import (
	"database/sql"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// Register creates a new user record with a bcrypt-hashed password.
func Register(db *sql.DB, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	query := "INSERT INTO users (username, password_hash) VALUES (?, ?)"
	if _, err := db.Exec(query, username, string(hash)); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

// Login verifies a username/password against the DB.
func Login(db *sql.DB, username, password string) (string, error) {
	var hash string
	query := "SELECT password_hash FROM users WHERE username = ?"
	if err := db.QueryRow(query, username).Scan(&hash); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user not found")
		}
		return "", fmt.Errorf("query user: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid password")
	}
	log.Printf("user %q logged in", username)

	return GenerateToken(username)
}
