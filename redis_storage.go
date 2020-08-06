package main

import (
	"context"
	"fmt"
	"log"
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

func redisEntryToEntry(val map[string]string) (*Entry, error) {
	var created time.Time
	if val["created"] != "" {
		c, err := time.Parse(time.RFC3339, val["created"])
		if err != nil {
			return nil, err
		}

		created = c
	}

	var accessed time.Time
	if val["accessed"] != "" {
		a, err := time.Parse(time.RFC3339, val["accessed"])
		if err != nil {
			return nil, err
		}
		accessed = a
	}

	var expire time.Time
	if val["expire"] != "" {
		e, err := time.Parse(time.RFC3339, val["expire"])
		if err != nil {
			return nil, err
		}

		expire = e
	}

	return &Entry{
		Data:     []byte(val["data"]),
		Accessed: accessed,
		Created:  created,
		Expire:   expire,
	}, nil
}

func (r *RedisStorage) Get(UUID string) (*Entry, error) {
	ctx := context.Background()
	val, err := r.rdb.HGetAll(ctx, r.GetKey(UUID)).Result()
	if err != nil {
		return nil, err
	}

	ret, err := redisEntryToEntry(val)

	if err != nil {
		return nil, err
	}

	ret.UUID = UUID

	return ret, nil
}

func (r *RedisStorage) GetAndDelete(UUID string) (*Entry, error) {
	ctx := context.Background()
	pipe := r.rdb.TxPipeline()
	key := r.GetKey(UUID)

	val := pipe.HGetAll(ctx, key)
	pipe.HSet(ctx, key, "data", nil, "accessed", time.Now())

	_, err := pipe.Exec(ctx)

	if err != nil {
		return nil, err
	}

	ret, err := redisEntryToEntry(val.Val())

	if err != nil {
		return nil, err
	}

	ret.UUID = UUID

	return ret, nil
}

func NewRedisStorage(redisDB string, prefix string) *RedisStorage {
	connOptions, err := redis.ParseURL(redisDB)

	if err != nil {
		log.Fatal(err)
	}
	rdb := redis.NewClient(connOptions)

	return &RedisStorage{rdb, prefix}
}
