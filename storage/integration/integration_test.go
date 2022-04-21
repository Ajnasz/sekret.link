package integration

import (
	"context"
	"testing"
	"time"

	"github.com/Ajnasz/sekret.link/encrypter/dummy"
	"github.com/Ajnasz/sekret.link/entries"
	"github.com/Ajnasz/sekret.link/storage"
	"github.com/Ajnasz/sekret.link/storage/postgresql"
	"github.com/Ajnasz/sekret.link/storage/secret"
	"github.com/Ajnasz/sekret.link/testhelper"
	"github.com/Ajnasz/sekret.link/uuid"
)

func TestStorages(t *testing.T) {
	connection := postgresql.NewStorage(testhelper.GetPSQLTestConn())
	t.Cleanup(func() {
		connection.Close()
	})
	psqlStorage := postgresql.NewPostgresCleanableStorage(connection)

	storages := map[string]storage.Cleanable{
		"Postgres": psqlStorage,
		"Secret": secret.NewCleanableSecretStorage(
			secret.NewSecretStorage(
				psqlStorage,
				dummy.NewEncrypter(),
			),
			psqlStorage,
		),
	}

	ctx := context.TODO()

	for name, storage := range storages {
		t.Run(name, func(t *testing.T) {
			t.Run("GetMeta", func(t *testing.T) {
				UUID := uuid.NewUUIDString()
				err := storage.Create(ctx, UUID, []byte("foo"), time.Second*-10, 1)

				if err != nil {
					t.Fatal(err)
				}

				data, err := storage.GetMeta(ctx, UUID)

				if data != nil {
					t.Errorf("Expected expired data to be nil")
				}

				if err != entries.ErrEntryExpired {
					t.Errorf("Expected expire error but got %v", err)
				}
			})
			t.Run("GetAndDelete", func(t *testing.T) {
				UUID := uuid.NewUUIDString()
				err := storage.Create(ctx, UUID, []byte("foo"), time.Second*-10, 1)

				if err != nil {
					t.Fatal(err)
				}

				data, err := storage.GetAndDelete(ctx, UUID)

				if data != nil {
					t.Errorf("Expected expired data to be nil")
				}

				if err != entries.ErrEntryExpired {
					t.Errorf("Expected expire error but got %v", err)
				}
			})
			t.Run("Delete", func(t *testing.T) {
				UUID := uuid.NewUUIDString()
				err := storage.Create(ctx, UUID, []byte("foo"), time.Second*-10, 1)

				if err != nil {
					t.Fatal(err)
				}

				err = storage.Delete(ctx, UUID)

				if err != nil {
					t.Error(err)
				}

				retMeta, err := storage.GetMeta(ctx, UUID)

				if err != entries.ErrEntryNotFound {
					t.Errorf("Storage GetMeta should return an entry not found error, but returned %v", err)
				}

				if retMeta != nil {
					t.Errorf("Expected to not be able to retreive deleted item, but got: %v", retMeta)
				}

				ret, err := storage.GetAndDelete(ctx, UUID)

				if err != entries.ErrEntryNotFound {
					t.Errorf("Storage GetAndDelete should return an entry not found error, but returned %v", err)
				}

				if ret != nil {
					t.Errorf("Expected to not be able to retreive deleted item, but got: %v", ret)
				}
			})

			t.Run("DeleteExpired", func(t *testing.T) {
				items := []struct {
					UUID         string
					Expire       time.Duration
					Value        []byte
					ShouldExpire bool
				}{
					{
						UUID:         uuid.NewUUIDString(),
						Expire:       time.Second * 10,
						Value:        []byte("FOO"),
						ShouldExpire: false,
					},
					{
						UUID:         uuid.NewUUIDString(),
						Expire:       time.Second * -10,
						Value:        []byte("BAR"),
						ShouldExpire: true,
					},
				}

				for _, item := range items {
					err := storage.Create(ctx, item.UUID, item.Value, item.Expire, 1)

					if err != nil {
						t.Fatal(err)
					}

					err = storage.DeleteExpired(ctx)

					if err != nil {
						t.Error(err)
					}

					ret, err := storage.GetMeta(ctx, item.UUID)

					if item.ShouldExpire {
						if err != entries.ErrEntryNotFound {
							t.Errorf("Expected entry to return a not found error, but got %s", err)
						}

						if ret != nil {
							t.Errorf("Expected to retrun nil for expired and deleted item")
						}
					} else {
						if err != nil {
							t.Error(err)
						}

						if ret == nil {
							t.Error("Returned a nil data")
						}
					}
				}
			})
		})
	}
}
