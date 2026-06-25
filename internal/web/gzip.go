// Package web provides small HTTP helpers for the static asset server.
package web

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// gzipResponseWriter streams the handler's output through a gzip.Writer.
type gzipResponseWriter struct {
	http.ResponseWriter
	gz          *gzip.Writer
	wroteHeader bool
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	// The compressed length is unknown up front, and Content-Length set by the
	// underlying handler (e.g. http.FileServer) would be wrong, so drop it.
	w.Header().Del("Content-Length")
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.gz.Write(b)
}

// Flush forwards an explicit flush through the gzip writer to the client.
func (w *gzipResponseWriter) Flush() {
	_ = w.gz.Flush()
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Gzip wraps next so responses are gzip-compressed for clients that advertise
// support via the Accept-Encoding header. It is a dependency-free replacement
// for the previously used (and now archived) NYTimes/gziphandler package.
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept-Encoding")
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, gz: gz}, r)
	})
}

// compile-time assertion that the wrapper still satisfies io.Writer.
var _ io.Writer = (*gzipResponseWriter)(nil)
