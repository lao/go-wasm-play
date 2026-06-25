package web

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func echoHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "999") // deliberately wrong, should be dropped
		_, _ = io.WriteString(w, body)
	})
}

func TestGzipCompressesWhenAccepted(t *testing.T) {
	body := strings.Repeat("compress me ", 100)
	srv := Gzip(echoHandler(body))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	res := rec.Result()
	if got := res.Header.Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", got)
	}
	if got := res.Header.Get("Content-Length"); got != "" {
		t.Errorf("Content-Length = %q, want it dropped", got)
	}
	if got := res.Header.Get("Vary"); !strings.Contains(got, "Accept-Encoding") {
		t.Errorf("Vary = %q, want it to include Accept-Encoding", got)
	}

	gr, err := gzip.NewReader(res.Body)
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	got, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("read gzip body: %v", err)
	}
	if string(got) != body {
		t.Errorf("decompressed body mismatch")
	}
	if rec.Body.Len() >= len(body) {
		t.Errorf("body was not compressed: %d >= %d", rec.Body.Len(), len(body))
	}
}

func TestGzipPassthroughWhenNotAccepted(t *testing.T) {
	body := "plain response"
	srv := Gzip(echoHandler(body))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	res := rec.Result()
	if got := res.Header.Get("Content-Encoding"); got != "" {
		t.Errorf("Content-Encoding = %q, want empty", got)
	}
	raw, _ := io.ReadAll(res.Body)
	if !bytes.Equal(raw, []byte(body)) {
		t.Errorf("body = %q, want %q", raw, body)
	}
}
