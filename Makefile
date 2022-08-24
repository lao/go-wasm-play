build-wasm:
	cd ./cmd/wasm && GOOS=js GOARCH=wasm go build -o  ../../assets/json.wasm

run-server:
	cd ./cmd/server && go run main.go