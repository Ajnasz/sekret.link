package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	rdb    *redis.Client
	Prefix string
}

func (r *RedisStorage) GetKey(UUID string) string {
	return fmt.Sprintf("%s:%s", r.Prefix, UUID)
}

func (r *RedisStorage) Create(UUID string, entry []byte) error {
	ctx := context.Background()
	err := r.rdb.HSet(ctx, r.GetKey(UUID), "data", entry, "created", time.Now()).Err()

	return err
}

func (r *RedisStorage) Get(UUID string) ([]byte, error) {
	ctx := context.Background()
	val, err := r.rdb.HGet(ctx, r.GetKey(UUID), "data").Result()

	return []byte(val), err
}

func (r *RedisStorage) GetAndDelete(UUID string) ([]byte, error) {
	ctx := context.Background()
	pipe := r.rdb.TxPipeline()
	key := r.GetKey(UUID)

	val := pipe.HGet(ctx, key, "data")
	pipe.HSet(ctx, key, "data", nil, "accessed", time.Now())

	_, err := pipe.Exec(ctx)

	if err != nil {
		return nil, err
	}

	return []byte(val.Val()), nil
}
