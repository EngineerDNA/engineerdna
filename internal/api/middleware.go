package api

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("%s %s completed in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// CORSMiddleware adds CORS headers - restricted to localhost only
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Only allow localhost origins (for local desktop app security)
		if origin == "" || isLocalhostOrigin(origin) {
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isLocalhostOrigin checks if an origin is from localhost
func isLocalhostOrigin(origin string) bool {
	// Allow localhost and 127.0.0.1 with any port
	return strings.HasPrefix(origin, "http://localhost:") ||
		strings.HasPrefix(origin, "http://127.0.0.1:") ||
		strings.HasPrefix(origin, "http://[::1]:") ||
		origin == "http://localhost" ||
		origin == "http://127.0.0.1" ||
		origin == "http://[::1]"
}

// MaxBytesMiddleware limits request body size to prevent DoS
func MaxBytesMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// CompressionMiddleware adds gzip compression for responses >1KB
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Wrap response writer with gzip
		gz := gzip.NewWriter(w)
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		gzw := &gzipResponseWriter{ResponseWriter: w, Writer: gz}
		next.ServeHTTP(gzw, r)
	})
}

// gzipResponseWriter wraps http.ResponseWriter with gzip compression
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	// Always compress when Content-Encoding header is set
	return w.Writer.Write(b)
}
