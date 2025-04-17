package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// Connect opens and pings the database.
func Connect(dsn string) (*sql.DB, error) {
	dbc, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := dbc.Ping(); err != nil {
		dbc.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return dbc, nil
}
