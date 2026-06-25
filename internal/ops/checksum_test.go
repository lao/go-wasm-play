package ops

import (
	"bytes"
	"testing"
)

func TestChecksumKnownVectors(t *testing.T) {
	// Reference digests for the input "abc".
	cases := map[string]string{
		"md5":    "900150983cd24fb0d6963f7d28e17f72",
		"sha1":   "a9993e364706816aba3e25717850c26c9cd0d89d",
		"sha256": "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		"crc32":  "352441c2",
	}
	for algorithm, want := range cases {
		got, err := Checksum(algorithm, []byte("abc"))
		if err != nil {
			t.Fatalf("Checksum(%q) error: %v", algorithm, err)
		}
		if got != want {
			t.Errorf("Checksum(%q) = %s, want %s", algorithm, got, want)
		}
	}
}

func TestChecksumUnsupported(t *testing.T) {
	if _, err := Checksum("sha512", []byte("x")); err == nil {
		t.Fatal("expected error for unsupported algorithm, got nil")
	}
}

func TestChecksumAlgorithmsAllResolve(t *testing.T) {
	for _, algorithm := range ChecksumAlgorithms() {
		if _, err := Checksum(algorithm, []byte("data")); err != nil {
			t.Errorf("advertised algorithm %q failed: %v", algorithm, err)
		}
	}
}

func TestMD5Wrapper(t *testing.T) {
	if got := MD5("abc"); got != "900150983cd24fb0d6963f7d28e17f72" {
		t.Errorf("MD5(abc) = %s", got)
	}
}

// TestChecksumBigFile exercises the hashing path with a multi-megabyte buffer,
// covering the "test with big files" goal from the README.
func TestChecksumBigFile(t *testing.T) {
	big := bytes.Repeat([]byte("go-wasm-play"), 1<<20) // ~12 MiB
	got, err := Checksum("sha256", big)
	if err != nil {
		t.Fatalf("Checksum on big input error: %v", err)
	}
	if len(got) != 64 {
		t.Errorf("sha256 digest length = %d, want 64", len(got))
	}
}

func BenchmarkChecksumSHA256(b *testing.B) {
	data := bytes.Repeat([]byte("benchmark"), 1<<16) // ~576 KiB
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Checksum("sha256", data); err != nil {
			b.Fatal(err)
		}
	}
}
