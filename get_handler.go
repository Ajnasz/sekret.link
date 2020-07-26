package main

import (
	"log"
	"net/http"
)

func handleGetEntry(w http.ResponseWriter, r *http.Request) {
	UUID, err := getUUIDFromPath(r.URL.Path)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	entry, err := storage.GetAndDelete(UUID)

	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(entry)
}
