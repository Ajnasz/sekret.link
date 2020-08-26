package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
)

var storage EntryStorage
var externalURLParam string
var expireSeconds int
var sqliteDB string
var postgresDB string
var redisDB string
var redisKeyPrefix string
var webExternalURL *url.URL

func init() {
	flag.StringVar(&externalURLParam, "webExternalURL", "", "Web server external url")
	flag.StringVar(&sqliteDB, "sqliteDB", "", "Path to sqlite database file")
	flag.StringVar(&postgresDB, "postgresDB", "", "Connection string for postgresql database backend")
	flag.StringVar(&redisDB, "redisDB", "", "Path to redis database")
	flag.StringVar(&redisKeyPrefix, "redisKeyPrefix", "entries", "Prefix of keys in redis db (in case redis is used as database backend)")
	flag.IntVar(&expireSeconds, "expireSeconds", 60*60*24*7, "Expire in seconds")
}

func main() {
	flag.Parse()

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

	defer storage.Close()

	http.HandleFunc("/", handleRequest)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
