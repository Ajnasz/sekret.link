package main

import (
	"net/http"
)

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", req.Header.Get("ORIGIN"))
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if (*r).Method == "OPTIONS" {
		setupResponse(&w, r)
		return
	}
	if r.Method == "POST" {
		setupResponse(&w, r)
		handleCreateEntry(w, r)
	} else if r.Method == "GET" {
		setupResponse(&w, r)
		handleGetEntry(w, r)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}
}