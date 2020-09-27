// +build redis,!test

package main

func newStorage() EntryStorage {
	return newRedisStorage(getConnectionString(redisDB, "REDIS_URL"), redisKeyPrefix)
}
