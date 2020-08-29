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

func (r *RedisStorage) Close() error {
	return r.rdb.Close()
}

func (r *RedisStorage) GetKey(UUID string) string {
	return fmt.Sprintf("%s:%s", r.Prefix, UUID)
}

func (r *RedisStorage) Create(UUID string, entry []byte, expire time.Duration) error {
	ctx := context.Background()
	now := time.Now()
	err := r.rdb.HSet(ctx, r.GetKey(UUID), "data", entry, "created", now, "expire", now.Add(expire)).Err()

	return err
}

func redisEntryToMeta(val map[string]string) (*EntryMeta, error) {
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

	return &EntryMeta{
		Accessed: accessed,
		Created:  created,
		Expire:   expire,
	}, nil

}

func redisEntryToEntry(val map[string]string) (*Entry, error) {
	meta, err := redisEntryToMeta(val)
	if err != nil {
		return nil, err
	}
	return &Entry{
		EntryMeta: *meta,
		Data:      []byte(val["data"]),
	}, nil
}

func (r *RedisStorage) Get(UUID string) (*Entry, error) {
	ctx := context.Background()
	val, err := r.rdb.HGetAll(ctx, r.GetKey(UUID)).Result()
	if err != nil {
		return nil, err
	}

	if len(val) == 0 {
		return nil, entryNotFound
	}

	ret, err := redisEntryToEntry(val)

	if err != nil {
		return nil, err
	}

	if ret.IsExpired() {
		return nil, &entryExpiredError{}
	}

	ret.UUID = UUID

	return ret, nil
}

func (r *RedisStorage) GetMeta(UUID string) (*EntryMeta, error) {
	ctx := context.Background()
	exists := r.rdb.Exists(ctx, r.GetKey(UUID))
	if exists.Val() == 0 {
		return nil, entryNotFound
	}

	metaKeys := []string{"created", "accessed", "expire"}

	val := map[string]string{}
	for _, metaKey := range metaKeys {
		keyVal, err := r.rdb.HGet(ctx, r.GetKey(UUID), metaKey).Result()
		if err != nil && err != redis.Nil {
			log.Println(err)
			return nil, err
		}
		val[metaKey] = keyVal
	}

	ret, err := redisEntryToMeta(val)

	if err != nil {
		return nil, err
	}

	if ret.IsExpired() {
		return nil, &entryExpiredError{}
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

	value := val.Val()

	if len(value) == 0 {
		return nil, entryNotFound
	}

	ret, err := redisEntryToEntry(value)

	if err != nil {
		return nil, err
	}

	if ret.IsExpired() {
		return nil, &entryExpiredError{}
	}

	ret.UUID = UUID

	return ret, nil
}

func (r *RedisStorage) Delete(UUID string) error {
	ctx := context.Background()
	err := r.rdb.Del(ctx, r.GetKey(UUID)).Err()
	return err
}

func NewRedisStorage(redisDB string, prefix string) *RedisStorage {
	connOptions, err := redis.ParseURL(redisDB)

	if err != nil {
		log.Fatal(err)
	}
	rdb := redis.NewClient(connOptions)

	return &RedisStorage{rdb, prefix}
}
