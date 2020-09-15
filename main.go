package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

var (
	storage          EntryStorage
	externalURLParam string
	expireSeconds    int
	maxExpireSeconds int
	sqliteDB         string
	postgresDB       string
	redisDB          string
	redisKeyPrefix   string
	webExternalURL   *url.URL
	maxDataSize      int64
	version          string
	queryVersion     bool
)

func init() {
	flag.StringVar(&externalURLParam, "webExternalURL", "", "Web server external url")
	flag.StringVar(&sqliteDB, "sqliteDB", "", "Path to sqlite database file")
	flag.StringVar(&postgresDB, "postgresDB", "", "Connection string for postgresql database backend")
	flag.StringVar(&redisDB, "redisDB", "", "Path to redis database")
	flag.StringVar(&redisKeyPrefix, "redisKeyPrefix", "entries", "Prefix of keys in redis db (in case redis is used as database backend)")
	flag.IntVar(&expireSeconds, "expireSeconds", 60*60*24*7, "Default expiration time in seconds")
	flag.IntVar(&maxExpireSeconds, "maxExpireSeconds", 60*60*24*30, "Max expiration time in seconds")
	flag.Int64Var(&maxDataSize, "maxDataSize", 1024, "Max data size")
	flag.BoolVar(&queryVersion, "version", false, "Get version information")
}

func main() {
	flag.Parse()

	if queryVersion {
		fmt.Println(version)
		return
	}

	if maxExpireSeconds < expireSeconds {
		log.Fatal("`expireSeconds` must be less or equal then `maxExpireSeconds`")
	}

	extURL, err := url.Parse(externalURLParam)

	if err != nil {
		log.Fatal(err)
	}

	webExternalURL = extURL

	if postgresDB != "" {
		storage = NewPostgresqlStorage(postgresDB)
	} else if sqliteDB != "" {
		storage = NewSQLiteStorage(sqliteDB)
		log.Println("Using SQLite database")
	} else if redisDB != "" {
		storage = NewRedisStorage(redisDB, redisKeyPrefix)
		log.Println("Using Redis database")
	} else {
		log.Fatal("No database backend selected")
	}

	stopChan := make(chan interface{})
	go func() {
		for {
			select {
			case <-time.After(time.Second):
				storage.DeleteExpired()
			case <-stopChan:
				return
			}
		}
	}()

	defer storage.Close()
	defer func() { stopChan <- struct{}{} }()

	http.HandleFunc("/", handleRequest)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
