package main

import (
	"log"
	"net/http"
)

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	if req.Header.Get("ORIGIN") != "" {
		(*w).Header().Set("Access-Control-Allow-Origin", req.Header.Get("ORIGIN"))
		(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	}
}

type secretHandler struct {
	config HandlerConfig
}

func NewSecretHandler(config HandlerConfig) *secretHandler {
	return &secretHandler{config}
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
		NewCreateHandler(s.config).Handle(w, r)
	} else if r.Method == http.MethodGet {
		setupResponse(&w, r)
		handleGetEntry(s.config.EntryStorage, w, r)
	} else if r.Method == http.MethodDelete {
		setupResponse(&w, r)
		handleDeleteEntry(s.config.EntryStorage, w, r)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}
}
