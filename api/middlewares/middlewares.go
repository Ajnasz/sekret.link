package middlewares

import (
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
		setSecurityHeaders(w)
		setCORSHeaders(origins)(w, r)
		h.ServeHTTP(w, r)
	})
}

// setSecurityHeaders sets headers that should apply to every response:
// secrets and their URLs must never be cached or leaked via Referer, and
// responses must never be MIME-sniffed into something else.
func setSecurityHeaders(w http.ResponseWriter) {
	h := w.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Cache-Control", "no-store")
	h.Set("Referrer-Policy", "no-referrer")
}

func setCORSHeaders(origins []string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
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
