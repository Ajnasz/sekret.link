package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func handleCreateEntry(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	UUID := newUUIDString()

	newURL, err := getUUIDUrl(webExternalURL, UUID)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	err = storage.Create(UUID, body)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("x-entry-uuid", UUID)
	fmt.Fprintf(w, "%s", newURL.String())
}
