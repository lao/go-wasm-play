# Playing with Go and WebAssembly

repo for this [tutorial](https://golangbot.com/webassembly-using-go/), just to play around with Go building to Wasm

## Scripts

```bash
make build-wasm 
make run-server
```

## Tiny-go problems

> **syscall/js.finalizeRef not implemented**
> follow: https://github.com/tinygo-org/tinygo/issues/1140

# TODO:
- [x] basic go package built to webassembly and called from js
- [x] reduce size of bin generated by go (maybe follow [this](https://dev.bitolog.com/minimizing-go-webassembly-binary-size/))

- [ ] load files from html input tag and checksum with go
- [ ] compress files with different algorithms from go and benchmark
- [ ] resize and other image related actions and benchmark
- [ ] test with big files

