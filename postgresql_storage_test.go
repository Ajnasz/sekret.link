package main

import (
	"context"
	"database/sql"
	"fmt"
	// "os"
	"testing"
	"time"
)

func getConn() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", "secret_link_test", "Km61HJgJbBjNA0FdABpjDmQxEz008PHAQMA8TLpUbnlaKN7U8G1bQGHk0wsm", "localhost", 54320, "secret_link_test")
}

func clearPSQLDatabase(dbname string) {
	psqlConn := getConn()
	storage := NewPostgresqlStorage(psqlConn)
	defer storage.Close()
	ctx := context.Background()
	_, err := storage.db.ExecContext(ctx, "TRUNCATE entries;")

	if err != nil {
		panic(err)
	}
}

func TestPostgresqlStorageCreateGet(t *testing.T) {
	psqlConn := getConn()
	testCases := []string{
		"foo",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			clearPSQLDatabase(psqlConn)

			storage := NewPostgresqlStorage(psqlConn)
			defer storage.Close()

			UUID := newUUIDString()
			err := storage.Create(UUID, []byte("foo"), time.Second*10)

			if err != nil {
				t.Fatal(err)
			}
			res, err := storage.Get(UUID)
			if err != nil {
				t.Fatal(err)
			}

			actual := string(res.Data)
			if actual != testCase {
				t.Errorf("expected: %s, actual: %s", testCase, actual)
			}
		})
	}
}

func TestPostgresqlStorageCreateGetAndDelete(t *testing.T) {
	psqlConn := getConn()

	testCases := []string{
		"foo",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			clearPSQLDatabase(psqlConn)

			storage := NewPostgresqlStorage(psqlConn)
			defer storage.Close()

			UUID := newUUIDString()
			err := storage.Create(UUID, []byte("foo"), time.Second*10)

			if err != nil {
				t.Fatal(err)
			}
			res, err := storage.GetAndDelete(UUID)
			if err != nil {
				t.Fatal(err)
			}

			actual := string(res.Data)
			if actual != testCase {
				t.Errorf("expected: %s, actual: %s", testCase, actual)
			}

			var data []byte
			var accessed time.Time
			var created time.Time

			row := storage.db.QueryRow("SELECT data, accessed, created FROM entries WHERE uuid=$1", UUID)
			// row := storage.db.QueryRow("SELECT accessed, created FROM entries WHERE uuid=?", "d799509c-21bd-47a3-a552-495fda8595b4")
			err = row.Scan(&data, &accessed, &created)
			if err != nil && err != sql.ErrNoRows {
				t.Fatal(err)
			}

			if data != nil {
				t.Errorf("Data returned, when it should be null %q", data)
			}

			now := time.Now()

			if now.Sub(accessed) > time.Second {
				t.Errorf("Accessed time is too large %q", accessed)
			}

			if !accessed.After(created) {
				t.Errorf("Accessed (%q) expected to be after created (%q)", accessed, created)
			}
		})
	}
}