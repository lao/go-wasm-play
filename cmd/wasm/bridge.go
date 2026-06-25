//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
)

// handler is the signature every asynchronous binding implements. Returning an
// error rejects the JavaScript Promise; returning a value resolves it.
type handler func(args []js.Value) (any, error)

// asyncFunc adapts a handler into a JavaScript function that returns a Promise.
// The work runs in a goroutine so the JS event loop is never blocked while Go
// executes, which keeps the page responsive even for large inputs.
func asyncFunc(h handler) js.Func {
	return js.FuncOf(func(_ js.Value, args []js.Value) any {
		// Copy the arguments: the js.Value slice is only valid for the
		// duration of this call, but the goroutine below outlives it.
		captured := make([]js.Value, len(args))
		copy(captured, args)

		var executor js.Func
		executor = js.FuncOf(func(_ js.Value, pargs []js.Value) any {
			resolve, reject := pargs[0], pargs[1]
			go func() {
				// Release the executor once the goroutine finishes; it has
				// already been invoked synchronously by the Promise
				// constructor and is never called again.
				defer executor.Release()
				defer func() {
					if r := recover(); r != nil {
						reject.Invoke(jsError(fmt.Sprintf("panic: %v", r)))
					}
				}()
				result, err := h(captured)
				if err != nil {
					reject.Invoke(jsError(err.Error()))
					return
				}
				resolve.Invoke(result)
			}()
			return nil
		})

		return js.Global().Get("Promise").New(executor)
	})
}

// syncString wraps a string-to-string transform, returning the error text on
// failure to preserve the original demo's behaviour.
func syncString(fn func(string) (string, error)) js.Func {
	return js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) != 1 {
			return "expected exactly one argument"
		}
		out, err := fn(args[0].String())
		if err != nil {
			return err.Error()
		}
		return out
	})
}

func requireArgs(args []js.Value, n int) error {
	if len(args) != n {
		return fmt.Errorf("expected %d argument(s), got %d", n, len(args))
	}
	return nil
}

// toBytes copies a JavaScript Uint8Array into a Go byte slice using a single
// bulk memcpy, which scales to multi-megabyte files.
func toBytes(v js.Value) []byte {
	buf := make([]byte, v.Get("length").Int())
	js.CopyBytesToGo(buf, v)
	return buf
}

// toUint8Array copies a Go byte slice into a fresh JavaScript Uint8Array.
func toUint8Array(b []byte) js.Value {
	dst := js.Global().Get("Uint8Array").New(len(b))
	js.CopyBytesToJS(dst, b)
	return dst
}

func toAnySlice(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

func jsError(msg string) js.Value {
	return js.Global().Get("Error").New(msg)
}
