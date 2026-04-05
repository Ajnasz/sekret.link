package middlewares

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
)

func SetupLogging(withPath bool, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if withPath {
			slog.Info("handle", "method", r.Method, "path", r.URL.Path)
		} else {
			slog.Info("handle", "method", r.Method)
		}
		h.ServeHTTP(w, r)
	})
}

func SetupHeaders(origins []string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(origins)(w, r)
		h.ServeHTTP(w, r)
	})
}

func setCORSHeaders(origins []string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("ORIGIN: %s\n", req.Header.Get("ORIGIN"))
		if req.Header.Get("ORIGIN") != "" {
			if slices.Contains(origins, req.Header.Get("ORIGIN")) {
				(w).Header().Set("Access-Control-Allow-Origin", req.Header.Get("ORIGIN"))
				(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
				(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, x-entry-uuid, x-entry-key, x-entry-delete-key, x-entry-expire")
			} else {
				slog.Warn("origin not allowed", "origin", req.Header.Get("ORIGIN"))
			}

		}
	}
}
