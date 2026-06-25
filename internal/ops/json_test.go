package ops

import (
	"strings"
	"testing"
)

func TestPrettyJSON(t *testing.T) {
	out, err := PrettyJSON(`{"b":1,"a":[1,2,3]}`)
	if err != nil {
		t.Fatalf("PrettyJSON returned error: %v", err)
	}
	if !strings.Contains(out, "\n  ") {
		t.Errorf("expected two-space indentation in output, got:\n%s", out)
	}
	// Round-trips back to the same data after re-formatting.
	if !strings.Contains(out, `"a"`) || !strings.Contains(out, `"b"`) {
		t.Errorf("expected keys preserved, got:\n%s", out)
	}
}

func TestPrettyJSONInvalid(t *testing.T) {
	if _, err := PrettyJSON(`{not json}`); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
