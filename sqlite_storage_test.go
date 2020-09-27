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
			storage := newSQLiteStorage(dbname)

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

func TestSQLiteStorageCreateGetAndDelete(t *testing.T) {
	dbname := "./test.sqlite"
	testCase := "foo"

	clearDatabase(dbname)
	storage := newSQLiteStorage(dbname)

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

	row := storage.db.QueryRow("SELECT data, accessed, created FROM entries WHERE uuid=?", UUID)
	err = row.Scan(&data, &accessed, &created)
	if err != sql.ErrNoRows {
		t.Fatal("Expected a sql.ErrNoRows but got", err)
	}
}
