package main

import (
	"context"
	"flag"
	"log/slog"

	_ "github.com/lib/pq"

	"github.com/Ajnasz/sekret.link/internal/config"
	"github.com/Ajnasz/sekret.link/internal/durable"
	"github.com/Ajnasz/sekret.link/internal/models/migrate"
)

func prepareDatabase(ctx context.Context) error {
	var postgresDB string
	flag.StringVar(&postgresDB, "postgresDB", "", "Connection string for postgresql database backend")
	flag.Parse()

	db, err := durable.OpenDatabaseClient(context.Background(), config.GetConnectionString(postgresDB))
	if err != nil {
		return err
	}

	defer db.Close()
	if err := migrate.PrepareDatabase(ctx, db); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := prepareDatabase(context.Background()); err != nil {
		slog.Error("Failed to prepare database", "error", err)
	}
}
