package main

import (
	"context"
	"fmt"
	"log"
)

type CleanableStorage interface {
	EntryStorage
	Clean()
}

type PostgresCleanableStorage struct {
	*PostgresqlStorage
}

func (s PostgresCleanableStorage) Clean() {
	_, err := s.db.Exec("TRUNCATE entries;")

	if err != nil {
		log.Fatal(err)
	}
}

type RedisCleanableStorage struct {
	*RedisStorage
}

func (s RedisCleanableStorage) Clean() {
	ctx := context.Background()
	keys, err := s.rdb.Keys(ctx, fmt.Sprintf("%s:*", s.Prefix)).Result()

	if err != nil {
		panic(err)
	}

	if len(keys) > 0 {
		err = s.rdb.Del(ctx, keys...).Err()

		if err != nil {
			panic(err)
		}
	}
}

type SQLiteCleanableStorage struct {
	*SQLiteStorage
}

func (s SQLiteCleanableStorage) Clean() {
	_, err := s.db.Exec("DELETE FROM entries;")

	if err != nil {
		log.Fatal(err)
	}
}

type MemoryCleanbleStorage struct {
	*MemoryStorage
}

func (s MemoryCleanbleStorage) Clean() {
	s.entries.RLock()
	defer s.entries.RUnlock()

	s.entries.m = make(map[string]*memoryEntry)
}

type CleanableSecretStorage struct {
	*SecretStorage
	internalStorage CleanableStorage
}

func (s CleanableSecretStorage) Clean() {
	s.internalStorage.Clean()
}
