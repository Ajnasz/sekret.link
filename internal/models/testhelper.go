package models

import (
	"fmt"
	"os"

	"github.com/Ajnasz/sekret.link/config"
)

// getPSQLTestConn returns connection string for tests
func getPSQLTestConn() string {
	password := os.Getenv("POSTGRES_PASSWORD")

	if password == "" {
		password = "sekret_link_test"
	}
	return config.GetConnectionString(fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", "postgres", "password", "localhost", 5432, password))
}
