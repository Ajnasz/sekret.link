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
var encryptionKey string
var databasePath string
var webExternalURL *url.URL

func init() {
	flag.StringVar(&externalURLParam, "webExternalURL", "", "Web server external url")
	flag.StringVar(&encryptionKey, "encryptionKey", "", "Encryption key to encrypt/decrypt data")
	flag.StringVar(&databasePath, "databasePath", "", "Path to sqlite database file")
}
func main() {
	flag.Parse()

	if len(encryptionKey) != 32 {
		log.Fatal("Encryption key length must be 32 bytes")
	}

	extURL, err := url.Parse(externalURLParam)

	if err != nil {
		log.Fatal(err)
	}

	webExternalURL = extURL
	sqlStorage := NewSQLiteStorage(databasePath)
	key := []byte(encryptionKey)
	storage = &SecretStorage{sqlStorage, &AESEncrypter{key}}

	http.HandleFunc("/", handleRequest)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
