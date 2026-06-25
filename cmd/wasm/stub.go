//go:build !(js && wasm)

// This stub lets `go build`, `go vet` and `go test ./...` succeed on the host
// toolchain. The real implementation in main.go imports syscall/js and only
// compiles for GOOS=js GOARCH=wasm.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "this command must be built for WebAssembly: GOOS=js GOARCH=wasm go build -o assets/json.wasm ./cmd/wasm")
	os.Exit(1)
}
