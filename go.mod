module github.com/jad0s/libretalk

go 1.24.1

require (
	github.com/go-sql-driver/mysql v1.9.2
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	golang.org/x/crypto v0.37.0
	golang.org/x/term v0.31.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
)

replace (
	github.com/jad0s/libretalk/internal/auth => ./internal/auth
	github.com/jad0s/libretalk/internal/chat => ./internal/chat
	github.com/jad0s/libretalk/internal/db => ./internal/db
	github.com/jad0s/libretalk/internal/types => ./internal/types
)
