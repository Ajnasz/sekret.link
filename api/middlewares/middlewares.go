package middlewares

import (
	"log/slog"
	"net/http"
	"slices"
	"strings"
)

// safeSegments are the only path segment literals that may appear in a log
// verbatim. Everything else is redacted by default, so a segment carrying a
// secret (UUID, key, delete key) can never leak even from a route added
// later or a malformed near-miss request.
var safeSegments = []string{"key"}

// RedactPath masks every path segment except known-safe literals, keeping
// the request's shape (segment count and safe literals) visible for
// debugging.
func RedactPath(path string) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if segment == "" || slices.Contains(safeSegments, segment) {
			continue
		}
		segments[i] = ":redacted"
	}
	return strings.Join(segments, "/")
}

func SetupLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("handle", "method", r.Method, "path", RedactPath(r.URL.Path))
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
