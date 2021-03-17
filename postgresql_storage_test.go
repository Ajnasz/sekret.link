package main

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func getPSQLTestConn() string {
	return getConnectionString(fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", "secret_link_test", "Km61HJgJbBjNA0FdABpjDmQxEz008PHAQMA8TLpUbnlaKN7U8G1bQGHk0wsm", "localhost", 5432, "secret_link_test"), "POSTGRES_URL")
}

func clearPSQLDatabase(dbname string) {
	psqlConn := getPSQLTestConn()
	storage := newPostgresqlStorage(psqlConn)
	defer storage.Close()
	ctx := context.Background()
	_, err := storage.db.ExecContext(ctx, "TRUNCATE entries;")

	if err != nil {
		panic(err)
	}
}

func TestPostgresqlStorageCreateGet(t *testing.T) {
	psqlConn := getPSQLTestConn()
	testCases := []string{
		"foo",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			clearPSQLDatabase(psqlConn)

			storage := newPostgresqlStorage(psqlConn)
			defer storage.Close()

			UUID := newUUIDString()
			err := storage.Create(UUID, []byte("foo"), time.Second*10, 1)

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
	psqlConn := getPSQLTestConn()

	testCases := []struct {
		Name         string
		Secret       string
		Reads        int
		Remaining    int
		ExistanceErr error
	}{
		{
			Name:         "Simple get",
			Secret:       "foo",
			Reads:        1,
			Remaining:    0,
			ExistanceErr: sql.ErrNoRows,
		},
		{
			Name:         "Exist get",
			Secret:       "bar",
			Reads:        2,
			Remaining:    1,
			ExistanceErr: nil,
		},
		{
			Name:         "Exist get 2",
			Secret:       "bar",
			Reads:        3,
			Remaining:    2,
			ExistanceErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			clearPSQLDatabase(psqlConn)

			storage := newPostgresqlStorage(psqlConn)
			defer storage.Close()

			UUID := newUUIDString()
			err := storage.Create(UUID, []byte(testCase.Secret), time.Second*10, testCase.Reads)

			if err != nil {
				t.Fatal(err)
			}
			res, err := storage.GetAndDelete(UUID)
			if err != nil {
				t.Fatal(err)
			}

			actual := string(res.Data)
			if actual != testCase.Secret {
				t.Errorf("expected: %s, actual: %s", testCase.Secret, actual)
			}

			var data []byte
			var remainingReads int

			row := storage.db.QueryRow("SELECT data, remaining_reads FROM entries WHERE uuid=$1", UUID)
			err = row.Scan(&data, &remainingReads)
			if err != testCase.ExistanceErr {
				t.Fatal(err)
			}

			if remainingReads != testCase.Remaining {
				t.Errorf("expected remaining to be %d, but got %d", testCase.Remaining, remainingReads)
			}
		})
	}
}

func TestPostgresqlStorageVerifyDelete(t *testing.T) {
	psqlConn := getPSQLTestConn()
	testCases := []struct {
		UUID        string
		Key         string
		DeleteKey   string
		Expected    bool
		ExpectedErr error
	}{
		{
			UUID:        "5e8e0330-1f9c-45eb-99ef-d49eeb4952f5",
			DeleteKey:   "",
			Expected:    true,
			ExpectedErr: nil,
		},
		{
			UUID:        "9d436e32-1914-4d9f-a9b4-51f8e58b271a",
			DeleteKey:   "2e573753-93f6-4b39-918f-b3448b050d01",
			Expected:    false,
			ExpectedErr: nil,
		},
	}
	for _, testCase := range testCases {
		clearPSQLDatabase(psqlConn)

		storage := newPostgresqlStorage(psqlConn)
		defer storage.Close()

		err := storage.Create(testCase.UUID, []byte("foo"), time.Second*10, 1)
		if err != testCase.ExpectedErr {
			t.Error(err)
		}

		var deleteKey string

		if testCase.DeleteKey == "" {
			row := storage.db.QueryRow("SELECT delete_key FROM entries WHERE uuid=$1", testCase.UUID)
			row.Scan(&deleteKey)
		} else {
			deleteKey = testCase.DeleteKey
		}

		actual, err := storage.VerifyDelete(testCase.UUID, deleteKey)

		if err != nil {
			t.Errorf("Expected error to be %+v, but got %+v", testCase.ExpectedErr, err)
		}

		if actual != testCase.Expected {
			t.Errorf("Expected %+v to be %+v", actual, testCase.Expected)
		}
	}
}
