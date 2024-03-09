package durable

import (
	"context"
	"database/sql"

	"github.com/Ajnasz/sekret.link/internal/durable"
)

func TestConnection(ctx context.Context) (*sql.DB, error) {
	config := durable.ConnectionInfo{
		Host:     "localhost",
		Port:     5432,
		Username: "postgres",
		Password: "password",
		Database: "sekret_link_test",
		SslMode:  "disable",
	}
	return durable.OpenDatabaseClient(ctx, config.String())
}
