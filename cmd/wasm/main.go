//go:build js && wasm

// Command wasm is the WebAssembly entry point. It binds the pure-Go helpers in
// internal/ops to JavaScript globals. Heavy or fallible operations are exposed
// as Promise-returning functions so they integrate naturally with async/await
// in the browser and never silently block on bad input.
package main

import (
	"fmt"
	"syscall/js"

	"go-wasm-test/internal/ops"
)

func main() {
	fmt.Println("Go WebAssembly module initialised")

	// Synchronous string helpers kept for backwards compatibility with the
	// original tutorial demo.
	js.Global().Set("formatJSON", syncString(func(s string) (string, error) {
		return ops.PrettyJSON(s)
	}))
	js.Global().Set("md5Hash", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) != 1 {
			return "expected exactly one argument"
		}
		return ops.MD5(args[0].String())
	}))

	// Asynchronous, byte-oriented helpers. Each returns a Promise.
	js.Global().Set("checksum", asyncFunc(checksum))
	js.Global().Set("compress", asyncFunc(compress))
	js.Global().Set("benchmarkCompression", asyncFunc(benchmarkCompression))
	js.Global().Set("imageInfo", asyncFunc(imageInfo))
	js.Global().Set("resizeImage", asyncFunc(resizeImage))
	js.Global().Set("grayscaleImage", asyncFunc(grayscaleImage))

	// Advertise which algorithms the module understands so the UI can build
	// its menus without hard-coding them.
	js.Global().Set("wasmCapabilities", js.ValueOf(map[string]any{
		"checksum": toAnySlice(ops.ChecksumAlgorithms()),
		"compress": toAnySlice(ops.CompressionAlgorithms()),
	}))

	// Block forever so the exported functions stay callable.
	select {}
}

// checksum(algorithm string, data Uint8Array) -> string
func checksum(args []js.Value) (any, error) {
	if err := requireArgs(args, 2); err != nil {
		return nil, err
	}
	return ops.Checksum(args[0].String(), toBytes(args[1]))
}

// compress(algorithm string, data Uint8Array) -> CompressionResult
func compress(args []js.Value) (any, error) {
	if err := requireArgs(args, 2); err != nil {
		return nil, err
	}
	res, err := ops.CompressStats(args[0].String(), toBytes(args[1]))
	if err != nil {
		return nil, err
	}
	return compressionResultJS(res), nil
}

// benchmarkCompression(data Uint8Array) -> []CompressionResult
func benchmarkCompression(args []js.Value) (any, error) {
	if err := requireArgs(args, 1); err != nil {
		return nil, err
	}
	results, err := ops.BenchmarkCompression(toBytes(args[0]))
	if err != nil {
		return nil, err
	}
	out := make([]any, len(results))
	for i, res := range results {
		out[i] = compressionResultJS(res)
	}
	return out, nil
}

// imageInfo(data Uint8Array) -> {format, width, height}
func imageInfo(args []js.Value) (any, error) {
	if err := requireArgs(args, 1); err != nil {
		return nil, err
	}
	info, err := ops.Info(toBytes(args[0]))
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"format": info.Format,
		"width":  info.Width,
		"height": info.Height,
	}, nil
}

// resizeImage(data Uint8Array, width int, height int, format string) -> Uint8Array
func resizeImage(args []js.Value) (any, error) {
	if err := requireArgs(args, 4); err != nil {
		return nil, err
	}
	out, err := ops.Resize(toBytes(args[0]), args[1].Int(), args[2].Int(), args[3].String())
	if err != nil {
		return nil, err
	}
	return toUint8Array(out), nil
}

// grayscaleImage(data Uint8Array, format string) -> Uint8Array
func grayscaleImage(args []js.Value) (any, error) {
	if err := requireArgs(args, 2); err != nil {
		return nil, err
	}
	out, err := ops.Grayscale(toBytes(args[0]), args[1].String())
	if err != nil {
		return nil, err
	}
	return toUint8Array(out), nil
}

func compressionResultJS(res ops.CompressionResult) map[string]any {
	return map[string]any{
		"algorithm":      res.Algorithm,
		"originalSize":   res.OriginalSize,
		"compressedSize": res.CompressedSize,
		"ratio":          res.Ratio,
		"saved":          res.Saved,
		"durationMs":     float64(res.Duration.Microseconds()) / 1000.0,
	}
}
