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

				if _, ok := err.(*entryExpiredError); !ok {
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

				if _, ok := err.(*entryExpiredError); !ok {
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

				if _, ok := err.(*entryExpiredError); !ok {
					t.Errorf("Expected expire error but got %v", err)
				}
			})
			storage.Clean()
		})
	}
}
