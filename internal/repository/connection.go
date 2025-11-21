package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type DatabaseConnection interface {
	DB() *sql.DB
}

type PostgresConnection struct {
	db *sql.DB
}

func (p *PostgresConnection) DB() *sql.DB {
	return p.db
}

func NewPostgresConnection(dsn string) (*PostgresConnection, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &PostgresConnection{
		db: db,
	}, nil
}
