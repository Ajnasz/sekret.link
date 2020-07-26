package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"

	"github.com/google/uuid"
)

func createHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create entry")
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Unknown error", http.StatusInternalServerError)
	}

	uuidString := storage.Create(body)

	fmt.Fprintf(w, "%s", uuidString)
}

func getUUIDFromPath(urlPath string) (string, error) {
	_, uuidFromPath := path.Split(urlPath)
	UUID, err := uuid.Parse(uuidFromPath)

	if err != nil {
		return "", err
	}
	return UUID.String(), nil
}

func getHandler(w http.ResponseWriter, r *http.Request) {

	UUID, err := getUUIDFromPath(r.URL.Path)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte{})
		return
	}

	entry := storage.Get(UUID)

	w.WriteHeader(http.StatusOK)
	w.Write(entry)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		createHandler(w, r)
	} else if r.Method == "GET" {
		getHandler(w, r)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}
}

var storage EntryStorage

func main() {
	storage = NewMemoryStorage()

	http.HandleFunc("/", handleRequest)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
