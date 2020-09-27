package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func cleanRedisStorage(storage *redisStorage) {
	ctx := context.Background()
	keys, err := storage.rdb.Keys(ctx, fmt.Sprintf("%s:*", storage.Prefix)).Result()

	if err != nil {
		panic(err)
	}

	if len(keys) > 0 {
		err = storage.rdb.Del(ctx, keys...).Err()

		if err != nil {
			panic(err)
		}
	}

}

func getRedisTestConn() string {
	return getConnectionString("redis://localhost:6379/0", "REDIS_URL")
}

func TestRedisStorage(t *testing.T) {
	connStr := getRedisTestConn()

	connOptions, err := redis.ParseURL(connStr)

	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(connOptions)

	storage := newRedisStorage(connStr, "entries_test")

	t.Run("Create", func(t *testing.T) {
		cleanRedisStorage(storage)
		testCase := "foo"
		UUID := newUUIDString()
		err := storage.Create(UUID, []byte(testCase), time.Second*10)

		if err != nil {
			t.Fatal(err)
		}

		ctx := context.Background()
		res, err := rdb.HGet(ctx, storage.GetKey(UUID), "data").Result()

		if err != nil {
			t.Fatal(err)
		}

		if res != testCase {
			t.Errorf("expected %q, got %q", testCase, res)
		}
	})

	t.Run("Get", func(t *testing.T) {
		cleanRedisStorage(storage)
		testCase := "foo"
		UUID := newUUIDString()
		err := storage.Create(UUID, []byte(testCase), time.Second*10)

		if err != nil {
			t.Fatal(err)
		}

		res, err := storage.Get(UUID)

		if err != nil {
			t.Fatal(err)
		}

		if string(res.Data) != testCase {
			t.Errorf("expected %q, got %q", testCase, res)
		}

		ctx := context.Background()
		data, err := rdb.HGet(ctx, storage.GetKey(UUID), "data").Result()

		if data != testCase {
			t.Errorf("Data is not ok, expected %q, got %q", testCase, data)
		}
	})

	t.Run("GetAndDelete", func(t *testing.T) {
		cleanRedisStorage(storage)
		testCase := "foo"
		UUID := newUUIDString()
		err := storage.Create(UUID, []byte(testCase), time.Second*10)

		if err != nil {
			t.Fatal(err)
		}

		res, err := storage.GetAndDelete(UUID)

		if err != nil {
			t.Fatal(err)
		}

		if string(res.Data) != testCase {
			t.Errorf("expected %q, got %q", testCase, res)
		}

		ctx := context.Background()
		data, err := rdb.HGet(ctx, storage.GetKey(UUID), "data").Result()

		if len(data) != 0 {
			t.Errorf("Data is not ok, expected to be empty, got %q", data)
		}

	})

	cleanRedisStorage(storage)
}
