package durable

import (
	"context"
	"database/sql"
	"fmt"
)

type ConnectionInfo struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SslMode  string
}

func (c ConnectionInfo) String() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.Username, c.Password, c.Host, c.Port, c.Database, c.SslMode)
}

func OpenDatabaseClient(ctx context.Context, connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
