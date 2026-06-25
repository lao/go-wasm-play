GOROOT    := $(shell go env GOROOT)
WASM_OUT  := assets/json.wasm
WASM_EXEC := assets/wasm_exec.js

.PHONY: all build wasm wasm-exec tinygo run test test-go test-wasm vet fmt tidy clean help

all: build ## default target: build the WebAssembly bundle

help: ## list the available targets
	@grep -hE '^[a-zA-Z._-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

build: wasm wasm-exec ## compile to Wasm and refresh wasm_exec.js

wasm: ## build cmd/wasm to assets/json.wasm (stripped to shrink the binary)
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o $(WASM_OUT) ./cmd/wasm

wasm-exec: ## copy the Go runtime's wasm_exec.js next to the binary
	cp "$(GOROOT)/lib/wasm/wasm_exec.js" $(WASM_EXEC)

tinygo: ## (experimental) smaller build via TinyGo — requires tinygo on PATH
	tinygo build -o assets/tiny-go/json.wasm -target wasm ./cmd/wasm

run: build ## build, then serve the assets at http://localhost:9090
	go run ./cmd/server

test: test-go test-wasm ## run Go unit tests and the Wasm end-to-end smoke test

test-go: ## run the host Go unit tests
	go test ./...

test-wasm: build ## load the built Wasm in Node and exercise every export
	node test/wasm_smoke.mjs

vet: ## static analysis for both the host and Wasm builds
	go vet ./...
	GOOS=js GOARCH=wasm go vet ./cmd/wasm

fmt: ## gofmt the module
	go fmt ./...

tidy: ## tidy go.mod / go.sum
	go mod tidy

clean: ## remove generated Wasm binaries
	rm -f $(WASM_OUT) assets/tiny-go/json.wasm
