package durable

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/Ajnasz/sekret.link/internal/config"
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

// getPSQLTestConn returns connection string for tests
func getPSQLTestConn() string {
	password := os.Getenv("POSTGRES_PASSWORD")

	if password == "" {
		password = "sekret_link_test"
	}
	return config.GetConnectionString(fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", "postgres", "password", "localhost", 5432, password))
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
