package durable

import (
	"context"
	"database/sql"
	"fmt"
	"time"
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

	// Cap total connections. Tune based on:
	//   (postgres max_connections - admin_reserved) / num_app_instances
	// Default postgres max_connections is 100; 25 is a safe conservative
	// value for a single instance leaving room for admin sessions and
	// pgBouncer/monitoring tools.
	db.SetMaxOpenConns(25)

	// Keep half the max open connections warm to avoid handshake overhead
	// on bursts, but don't hold more idle connections than we realistically
	// need between requests.
	db.SetMaxIdleConns(10)

	// Recycle connections periodically so they don't survive PostgreSQL
	// restarts, firewall idle-timeouts, or load-balancer resets.
	db.SetConnMaxLifetime(time.Hour)

	// Release connections that have been idle longer than this to free
	// PostgreSQL backend memory during quiet periods.
	db.SetConnMaxIdleTime(10 * time.Minute)

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
