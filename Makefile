build-wasm:
	cd ./cmd/wasm && GOOS=js GOARCH=wasm go build -o  ../../assets/json.wasm

build-wasm-tiny-go:
	cd ./cmd/wasm && tinygo build -o ../../assets/tiny-go/json.wasm -target wasm main.go

run-server:
	cd ./cmd/server && go run main.go