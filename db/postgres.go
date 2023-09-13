package db

import (
	"database/sql"
	"log"
)

// PostgreSQL ...
type PostgreSQL struct {
	DB *sql.DB
}

// NewPostgreSQL constructor for PostgreSQL struct.
func NewPostgreSQL(driverName string, connStr string) (*PostgreSQL, error) {
	DB, err := sql.Open(driverName, connStr)
	if err != nil {
		log.Println("could not connect to database")
		return nil, err
	}
	return &PostgreSQL{DB: DB}, nil
}
