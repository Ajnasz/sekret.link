package middlewares

import (
	"fmt"
	"log"
	"net/http"
)

func SetupLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(fmt.Sprintf("%s: %s", r.Method, r.URL.Path))
		h.ServeHTTP(w, r)
	})
}

func SetupHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w, r)
		h.ServeHTTP(w, r)
	})
}

func setCORSHeaders(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("ORIGIN") != "" {
		(w).Header().Set("Access-Control-Allow-Origin", req.Header.Get("ORIGIN"))
		(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
		(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	}
}
