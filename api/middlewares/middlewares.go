package middlewares

import (
	"log/slog"
	"net/http"
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
		(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, x-entry-uuid, x-entry-key, x-entry-delete-key, x-entry-expire")
	}
}
