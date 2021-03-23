package main

import (
	"fmt"
	"log"
	"net/http"
)

func setupLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(fmt.Sprintf("%s: %s", r.Method, r.URL.Path))
		h.ServeHTTP(w, r)
	})
}

func setupHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w, r)
		h.ServeHTTP(w, r)
	})
}

func setCORSHeaders(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("ORIGIN") != "" {
		(w).Header().Set("Access-Control-Allow-Origin", req.Header.Get("ORIGIN"))
		(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	}
}

type secretHandler struct{}

func (s secretHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if (*r).Method == http.MethodOptions {
		return
	}

	if r.Method == http.MethodPost {
		if r.URL.Path != "/" && r.URL.Path != "" {
			http.Error(w, "Not found", http.StatusNotFound)
			log.Println("Not found", r.URL.Path)
			return
		}
		handleCreateEntry(w, r)
	} else if r.Method == http.MethodGet {
		handleGetEntry(w, r)
	} else if r.Method == http.MethodDelete {
		handleDeleteEntry(w, r)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}
}
