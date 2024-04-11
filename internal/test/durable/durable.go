package durable

import (
	"context"
	"database/sql"

	"github.com/Ajnasz/sekret.link/internal/config"
	"github.com/Ajnasz/sekret.link/internal/durable"
)

func TestConnection(ctx context.Context) (*sql.DB, error) {
	conf := durable.ConnectionInfo{
		Host:     "localhost",
		Port:     5432,
		Username: "postgres",
		Password: "password",
		Database: "sekret_link_test",
		SslMode:  "disable",
	}
	db, err := durable.OpenDatabaseClient(ctx, config.GetConnectionString(conf.String()))

	if err != nil {
		return nil, err
	}

	return db, nil
}
