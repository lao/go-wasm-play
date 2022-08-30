build.wasm:
	cd ./cmd/wasm && GOOS=js GOARCH=wasm go build -o  ../../assets/json.wasm

build.wasm.tinygo:
	cd ./cmd/wasm && tinygo build -o ../../assets/tiny-go/json.wasm -target wasm main.go

server.run:
	go run ./cmd/server/main.go