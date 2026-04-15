package api

import (
	"fmt"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     200, // Default to 200
		}

		next.ServeHTTP(rw, r) // Call next handler

		duration := time.Since(start)

		fmt.Printf("%s %s → %d → %v\n", r.Method, r.URL.Path, rw.statusCode, duration)
	})
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
