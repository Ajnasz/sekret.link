// +build redis,!test

package main

import "github.com/Ajnasz/sekret.link/storage"

func newStorage() storage.EntryStorage {
	return newRedisStorage(getConnectionString(redisDB, "REDIS_URL"), redisKeyPrefix)
}
