// Command server serves the WebAssembly demo assets over HTTP with on-the-fly
// gzip compression. Configure it with the PORT and ASSETS_DIR environment
// variables or the matching command-line flags.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"go-wasm-test/internal/web"
)

func main() {
	addr := flag.String("addr", envOr("ADDR", ":9090"), "address to listen on")
	dir := flag.String("dir", envOr("ASSETS_DIR", "./assets"), "directory of static assets to serve")
	flag.Parse()

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(*dir)))

	handler := web.Gzip(mux)

	log.Printf("serving %s at http://localhost%s", *dir, *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
