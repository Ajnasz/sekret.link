package durable

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/Ajnasz/sekret.link/config"
)

type ConnectionInfo struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

// getPSQLTestConn returns connection string for tests
func getPSQLTestConn() string {
	password := os.Getenv("POSTGRES_PASSWORD")

	if password == "" {
		password = "sekret_link_test"
	}
	return config.GetConnectionString(fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", "postgres", "password", "localhost", 5432, password))
}

func OpenDatabaseClient(ctx context.Context, c ConnectionInfo) (*sql.DB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", c.Username, c.Password, c.Host, c.Port, c.Database)
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
