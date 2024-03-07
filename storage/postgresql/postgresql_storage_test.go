package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/Ajnasz/sekret.link/testhelper"
	"github.com/Ajnasz/sekret.link/uuid"
)

func Test_PostgresqlStorageWrite(t *testing.T) {
	psqlConn := testhelper.GetPSQLTestConn()
	storage := NewStorage(psqlConn)
	t.Cleanup(func() {
		storage.Close()
	})

	testCases := []struct {
		Name      string
		Secret    string
		Reads     int
		Remaining int
	}{
		{
			Name:      "Simple get",
			Secret:    "foo",
			Reads:     1,
			Remaining: 0,
		},
		{
			Name:      "Exist get",
			Secret:    "bar",
			Reads:     2,
			Remaining: 1,
		},
		{
			Name:      "Exist get 2",
			Secret:    "bar",
			Reads:     3,
			Remaining: 2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			UUID := uuid.NewUUIDString()
			ctx := context.Background()
			meta, err := storage.Write(ctx, UUID, []byte(testCase.Secret), time.Second*10, testCase.Reads)

			if err != nil {
				t.Fatal(err)
			}

			res, confirm, err := storage.ReadConfirm(ctx, UUID)
			if err != nil {
				t.Fatal(err)
			}
			confirm <- true
			select {
			case <-confirm:
			}

			if res.EntryMeta != *meta {
				t.Errorf("expected meta to be the same %+v, %+v", res.EntryMeta, *meta)
			}

			actual := string(res.Data)
			if actual != testCase.Secret {
				t.Errorf("%s expected: %s, actual: %s", UUID, testCase.Secret, actual)
			}

			var remainingReads int

			row := storage.db.QueryRow("SELECT remaining_reads FROM entries WHERE uuid=$1", UUID)
			err = row.Scan(&remainingReads)
			if err != nil {
				t.Fatalf("%s: %v", UUID, err)
			}

			if remainingReads != testCase.Remaining {
				t.Errorf("expected remaining to be %d, but got %d", testCase.Remaining, remainingReads)
			}
		})
	}
}

func TestPostgresqlStorageVerifyDelete(t *testing.T) {
	psqlConn := testhelper.GetPSQLTestConn()
	storage := NewStorage(psqlConn)
	t.Cleanup(func() {
		storage.Close()
	})
	testCases := []struct {
		UUID        string
		Key         string
		DeleteKey   string
		Expected    bool
		ExpectedErr error
	}{
		{
			UUID:        uuid.NewUUIDString(),
			DeleteKey:   "",
			Expected:    true,
			ExpectedErr: nil,
		},
		{
			UUID:        uuid.NewUUIDString(),
			DeleteKey:   uuid.NewUUIDString(),
			Expected:    false,
			ExpectedErr: nil,
		},
	}
	for _, testCase := range testCases {

		ctx := context.Background()
		_, err := storage.Write(ctx, testCase.UUID, []byte("foo"), time.Second*10, 1)
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

		actual, err := storage.VerifyDelete(ctx, testCase.UUID, deleteKey)

		if err != nil {
			t.Errorf("Expected error to be %+v, but got %+v", testCase.ExpectedErr, err)
		}

		if actual != testCase.Expected {
			t.Errorf("Expected %+v to be %+v", actual, testCase.Expected)
		}
	}
}
