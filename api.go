package main

import (
	"log"
	"net/http"

	"github.com/Ajnasz/sekret.link/storage"
)

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	if req.Header.Get("ORIGIN") != "" {
		(*w).Header().Set("Access-Control-Allow-Origin", req.Header.Get("ORIGIN"))
		(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	}
}

type secretHandler struct {
	storage storage.VerifyStorage
}

func NewSecretHandler(storage storage.VerifyStorage) *secretHandler {
	return &secretHandler{storage}
}

func (s secretHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if (*r).Method == http.MethodOptions {
		setupResponse(&w, r)
		return
	}

	if r.Method == http.MethodPost {
		setupResponse(&w, r)
		if r.URL.Path != "/" && r.URL.Path != "" {
			http.Error(w, "Not found", http.StatusNotFound)
			log.Println("Not found", r.URL.Path)
			return
		}
		handleCreateEntry(s.storage, w, r)
	} else if r.Method == http.MethodGet {
		setupResponse(&w, r)
		handleGetEntry(s.storage, w, r)
	} else if r.Method == http.MethodDelete {
		setupResponse(&w, r)
		handleDeleteEntry(s.storage, w, r)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}
}
