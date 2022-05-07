// Package middleware provides response compressing functionality.
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// Type gzipWriter redefines http.ResponseWriter changing its Writer method to use gzip.
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write method redefines default http.ResponseWriter Write method.
func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// GzipHandle serves as a middleware handler implementing gzip compressing.
func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")

		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
