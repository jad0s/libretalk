package auth

import (
	"fmt"
	"time"

	"encoding/base64"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret, _ = base64.StdEncoding.DecodeString("12SNtOQ9LS3nSAHAZl0EQaciRDCsLDCFCEzeIJvx5ss=") // TODO load from .env

// GenerateToken returns a JWT signed with HS256, expiring in 24h.
func GenerateToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken validates the token string and returns the username.
func ParseToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	user, ok := claims["username"].(string)
	if !ok {
		return "", fmt.Errorf("username claim missing")
	}
	return user, nil
}
