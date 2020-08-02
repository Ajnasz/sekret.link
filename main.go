package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		handleCreateEntry(w, r)
	} else if r.Method == "GET" {
		handleGetEntry(w, r)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}
}

var storage EntryStorage
var externalURLParam string
var sqliteDB string
var redisDB string
var redisKeyPrefix string
var webExternalURL *url.URL

func init() {
	flag.StringVar(&externalURLParam, "webExternalURL", "", "Web server external url")
	flag.StringVar(&sqliteDB, "sqliteDB", "", "Path to sqlite database file")
	flag.StringVar(&redisDB, "redisDB", "", "Path to redis database")
	flag.StringVar(&redisKeyPrefix, "redisKeyPrefix", "entries", "Prefix of keys in redis db (in case redis is used as database backend)")
}

func main() {
	flag.Parse()

	extURL, err := url.Parse(externalURLParam)

	if err != nil {
		log.Fatal(err)
	}

	webExternalURL = extURL
	if sqliteDB != "" {
		storage = NewSQLiteStorage(sqliteDB)
		log.Println("Using SQLite database")
	} else if redisDB != "" {
		storage = NewRedisStorage(redisDB, redisKeyPrefix)
		log.Println("Using Redis database")
	} else {
		log.Fatal("No database backend selected")
	}

	http.HandleFunc("/", handleRequest)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
