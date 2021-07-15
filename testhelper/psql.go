package testhelper

import (
	"fmt"

	"github.com/Ajnasz/sekret.link/config"
)

func GetPSQLTestConn() string {
	return config.GetConnectionString(fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", "secret_link_test", "Km61HJgJbBjNA0FdABpjDmQxEz008PHAQMA8TLpUbnlaKN7U8G1bQGHk0wsm", "localhost", 5432, "secret_link_test"))
}
