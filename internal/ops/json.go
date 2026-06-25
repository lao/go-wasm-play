// Package ops contains the pure-Go building blocks that the WebAssembly module
// exposes to JavaScript. None of the code in this package imports syscall/js,
// which keeps it portable and unit-testable on the host toolchain.
package ops

import "encoding/json"

// PrettyJSON re-indents a JSON document with two-space indentation. It returns
// an error if the input is not valid JSON.
func PrettyJSON(input string) (string, error) {
	var raw any
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		return "", err
	}
	pretty, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return "", err
	}
	return string(pretty), nil
}
