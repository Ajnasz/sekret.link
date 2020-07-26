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
var webExternalURL *url.URL

func init() {
	flag.StringVar(&externalURLParam, "webExternalURL", "", "Web server external url")
}
func main() {
	flag.Parse()

	extURL, err := url.Parse(externalURLParam)

	if err != nil {
		log.Fatal(err)
	}

	webExternalURL = extURL
	storage = NewMemoryStorage()

	http.HandleFunc("/", handleRequest)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
