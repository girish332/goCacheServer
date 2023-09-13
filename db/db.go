package db

import (
	"context"
	"database/sql"
)

// DatabaseClient ...
type DatabaseClient interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Begin() (*sql.Tx, error)
	PingContext(ctx context.Context) error
	Close() error
}
