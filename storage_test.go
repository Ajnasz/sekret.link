package main

import (
	"testing"
	"time"
)

func TestStorages(t *testing.T) {
	storages := map[string]CleanableStorage{
		"Postgres": PostgresCleanableStorage{NewPostgresqlStorage(getConn())},
		"Redis":    RedisCleanableStorage{NewRedisStorage("redis://localhost:6379/0", "entries_test")},
		"SQLite":   SQLiteCleanableStorage{NewSQLiteStorage("./test.sqlite")},
		"Memory":   MemoryCleanbleStorage{NewMemoryStorage()},
		"Secret": CleanableSecretStorage{
			&SecretStorage{NewMemoryStorage(),
				NewDummyEncrypter(),
			},
			MemoryCleanbleStorage{NewMemoryStorage()},
		},
	}

	for name, storage := range storages {
		t.Run(name, func(t *testing.T) {
			t.Run("Get", func(t *testing.T) {
				storage.Clean()
				UUID := newUUIDString()
				err := storage.Create(UUID, []byte("foo"), time.Second*-10)

				if err != nil {
					t.Fatal(err)
				}

				data, err := storage.Get(UUID)

				if data != nil {
					t.Errorf("Expected expired data to be nil")
				}

				if err != ErrEntryExpired {
					t.Errorf("Expected expire error but got %v", err)
				}
			})
			t.Run("GetMeta", func(t *testing.T) {
				storage.Clean()
				UUID := newUUIDString()
				err := storage.Create(UUID, []byte("foo"), time.Second*-10)

				if err != nil {
					t.Fatal(err)
				}

				data, err := storage.GetMeta(UUID)

				if data != nil {
					t.Errorf("Expected expired data to be nil")
				}

				if err != ErrEntryExpired {
					t.Errorf("Expected expire error but got %v", err)
				}
			})
			t.Run("GetAndDelete", func(t *testing.T) {
				storage.Clean()
				UUID := newUUIDString()
				err := storage.Create(UUID, []byte("foo"), time.Second*-10)

				if err != nil {
					t.Fatal(err)
				}

				data, err := storage.GetAndDelete(UUID)

				if data != nil {
					t.Errorf("Expected expired data to be nil")
				}

				if err != ErrEntryExpired {
					t.Errorf("Expected expire error but got %v", err)
				}
			})
			t.Run("Delete", func(t *testing.T) {
				storage.Clean()
				UUID := newUUIDString()
				err := storage.Create(UUID, []byte("foo"), time.Second*-10)

				if err != nil {
					t.Fatal(err)
				}

				err = storage.Delete(UUID)

				if err != nil {
					t.Error(err)
				}

				ret, err := storage.Get(UUID)

				if err != ErrEntryNotFound {
					t.Errorf("Storage Get should return an entry not found error, but returned %v", err)
				}

				if ret != nil {
					t.Errorf("Expected to not be able to retreive deleted item, but got: %v", ret)
				}

				retMeta, err := storage.GetMeta(UUID)

				if err != ErrEntryNotFound {
					t.Errorf("Storage GetMeta should return an entry not found error, but returned %v", err)
				}

				if retMeta != nil {
					t.Errorf("Expected to not be able to retreive deleted item, but got: %v", ret)
				}

				ret, err = storage.GetAndDelete(UUID)

				if err != ErrEntryNotFound {
					t.Errorf("Storage GetAndDelete should return an entry not found error, but returned %v", err)
				}

				if retMeta != nil {
					t.Errorf("Expected to not be able to retreive deleted item, but got: %v", ret)
				}
			})

			t.Run("DeleteExpired", func(t *testing.T) {
				storage.Clean()
				items := []struct {
					UUID         string
					Expire       time.Duration
					Value        []byte
					ShouldExpire bool
				}{
					{
						UUID:         newUUIDString(),
						Expire:       time.Second * 10,
						Value:        []byte("FOO"),
						ShouldExpire: false,
					},
					{
						UUID:         newUUIDString(),
						Expire:       time.Second * -10,
						Value:        []byte("BAR"),
						ShouldExpire: true,
					},
				}

				for _, item := range items {
					err := storage.Create(item.UUID, item.Value, item.Expire)

					if err != nil {
						t.Fatal(err)
					}

					err = storage.DeleteExpired()

					if err != nil {
						t.Error(err)
					}

					ret, err := storage.Get(item.UUID)

					if item.ShouldExpire {
						if err != ErrEntryNotFound {
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
						} else if string(ret.Data) != string(item.Value) {
							t.Errorf("Expected data to be %q but got %q", string(item.Value), string(ret.Data))
						}
					}
				}
			})
			storage.Clean()
		})
	}
}
