package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-redis/redis/v8"
)

func cleanRedisStorage(storage *RedisStorage) {
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

func TestRedisStorage(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	storage := RedisStorage{rdb, "entries_test"}

	testCases := []string{
		"foo",
	}

	t.Run("Create", func(t *testing.T) {
		cleanRedisStorage(&storage)
		for _, testCase := range testCases {
			UUID := newUUIDString()
			err := storage.Create(UUID, []byte(testCase))

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
		}
	})

	t.Run("Get", func(t *testing.T) {
		cleanRedisStorage(&storage)
		for _, testCase := range testCases {
			UUID := newUUIDString()
			err := storage.Create(UUID, []byte(testCase))

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

		}
	})

	t.Run("GetAndDelete", func(t *testing.T) {
		cleanRedisStorage(&storage)
		for _, testCase := range testCases {
			UUID := newUUIDString()
			err := storage.Create(UUID, []byte(testCase))

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

			if data != "" {
				t.Errorf("Data is not ok, expected %q, got %q", "", data)
			}

		}
	})

	cleanRedisStorage(&storage)
}
