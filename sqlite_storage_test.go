package main

import (
	"database/sql"
	"os"
	"testing"
	"time"
)

func clearDatabase(dbname string) {
	os.Remove(dbname)
}

func TestSQLiteStorageCreateGet(t *testing.T) {
	dbname := "./test.sqlite"

	testCases := []string{
		"foo",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			clearDatabase(dbname)
			storage := NewSQLiteStorage(dbname)

			UUID := newUUIDString()
			err := storage.Create(UUID, []byte("foo"))

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

func TestSQLiteStorageCreateGetAndDelete(t *testing.T) {
	dbname := "./test.sqlite"

	testCases := []string{
		"foo",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			clearDatabase(dbname)
			storage := NewSQLiteStorage(dbname)

			UUID := newUUIDString()
			err := storage.Create(UUID, []byte("foo"))

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

			row := storage.db.QueryRow("SELECT data, accessed, created FROM entries WHERE uuid=?", UUID)
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
