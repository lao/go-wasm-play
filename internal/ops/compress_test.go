package ops

import (
	"bytes"
	"testing"
)

func TestCompressRoundTrip(t *testing.T) {
	// Repetitive data so every algorithm compresses well.
	original := bytes.Repeat([]byte("the quick brown fox "), 1024)
	for _, algorithm := range CompressionAlgorithms() {
		compressed, err := Compress(algorithm, original)
		if err != nil {
			t.Fatalf("Compress(%q) error: %v", algorithm, err)
		}
		if len(compressed) >= len(original) {
			t.Errorf("Compress(%q) did not shrink data: %d >= %d", algorithm, len(compressed), len(original))
		}
		restored, err := Decompress(algorithm, compressed)
		if err != nil {
			t.Fatalf("Decompress(%q) error: %v", algorithm, err)
		}
		if !bytes.Equal(restored, original) {
			t.Errorf("round trip for %q did not restore the original data", algorithm)
		}
	}
}

func TestCompressUnsupported(t *testing.T) {
	if _, err := Compress("brotli", []byte("x")); err == nil {
		t.Fatal("expected error for unsupported algorithm, got nil")
	}
}

func TestCompressStats(t *testing.T) {
	data := bytes.Repeat([]byte("compress me "), 512)
	res, err := CompressStats("gzip", data)
	if err != nil {
		t.Fatalf("CompressStats error: %v", err)
	}
	if res.Algorithm != "gzip" {
		t.Errorf("Algorithm = %q, want gzip", res.Algorithm)
	}
	if res.OriginalSize != len(data) {
		t.Errorf("OriginalSize = %d, want %d", res.OriginalSize, len(data))
	}
	if res.Ratio <= 0 || res.Ratio >= 1 {
		t.Errorf("Ratio = %f, want between 0 and 1 for compressible data", res.Ratio)
	}
	if delta := res.Ratio + res.Saved; delta < 0.999 || delta > 1.001 {
		t.Errorf("Ratio + Saved = %f, want ~1", delta)
	}
	if res.Duration < 0 {
		t.Errorf("Duration = %v, want non-negative", res.Duration)
	}
}

func TestBenchmarkCompressionSorted(t *testing.T) {
	data := bytes.Repeat([]byte("benchmark all algorithms "), 800)
	results, err := BenchmarkCompression(data)
	if err != nil {
		t.Fatalf("BenchmarkCompression error: %v", err)
	}
	if len(results) != len(CompressionAlgorithms()) {
		t.Fatalf("got %d results, want %d", len(results), len(CompressionAlgorithms()))
	}
	for i := 1; i < len(results); i++ {
		if results[i-1].CompressedSize > results[i].CompressedSize {
			t.Errorf("results not sorted by size: %d before %d", results[i-1].CompressedSize, results[i].CompressedSize)
		}
	}
}

// TestCompressBigFile checks the pipeline handles a multi-megabyte payload.
func TestCompressBigFile(t *testing.T) {
	big := bytes.Repeat([]byte("0123456789abcdef"), 1<<20) // ~16 MiB
	results, err := BenchmarkCompression(big)
	if err != nil {
		t.Fatalf("BenchmarkCompression on big input error: %v", err)
	}
	for _, res := range results {
		if res.CompressedSize == 0 {
			t.Errorf("%s produced an empty output", res.Algorithm)
		}
	}
}

func BenchmarkCompressGzip(b *testing.B) {
	data := bytes.Repeat([]byte("benchmark gzip throughput "), 1<<14)
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Compress("gzip", data); err != nil {
			b.Fatal(err)
		}
	}
}
