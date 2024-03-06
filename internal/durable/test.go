package durable

import (
	"context"
	"database/sql"
)

func TestConnection(ctx context.Context) (*sql.DB, error) {
	config := ConnectionInfo{
		Host:     "localhost",
		Port:     5432,
		Username: "postgres",
		Password: "password",
		Database: "sekret_link_test",
	}
	return OpenDatabaseClient(ctx, config)
}
