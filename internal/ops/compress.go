package ops

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"sort"
	"time"
)

// CompressionAlgorithms lists the supported algorithms in a stable order.
var compressionAlgorithms = []string{"gzip", "zlib", "flate"}

// CompressionAlgorithms returns the supported compression algorithm names.
func CompressionAlgorithms() []string {
	out := make([]string, len(compressionAlgorithms))
	copy(out, compressionAlgorithms)
	return out
}

// CompressionResult summarises a single compression run. It is returned to the
// browser to drive the benchmark table.
type CompressionResult struct {
	Algorithm      string        `json:"algorithm"`
	OriginalSize   int           `json:"originalSize"`
	CompressedSize int           `json:"compressedSize"`
	Ratio          float64       `json:"ratio"`    // CompressedSize / OriginalSize (lower is better)
	Saved          float64       `json:"saved"`    // fraction of bytes removed, 1 - Ratio
	Duration       time.Duration `json:"duration"` // nanoseconds spent compressing
}

func newWriter(algorithm string, buf *bytes.Buffer) (io.WriteCloser, error) {
	switch algorithm {
	case "gzip":
		return gzip.NewWriter(buf), nil
	case "zlib":
		return zlib.NewWriter(buf), nil
	case "flate":
		return flate.NewWriter(buf, flate.DefaultCompression)
	default:
		return nil, fmt.Errorf("unsupported compression algorithm %q", algorithm)
	}
}

// Compress compresses data with the named algorithm and returns the bytes.
func Compress(algorithm string, data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := newWriter(algorithm, &buf)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	// Close flushes any buffered data and writes the trailer; it must run
	// before the buffer is read.
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decompress reverses Compress for the named algorithm. It is primarily used to
// prove round-trip correctness in tests.
func Decompress(algorithm string, data []byte) ([]byte, error) {
	var (
		r   io.ReadCloser
		err error
	)
	switch algorithm {
	case "gzip":
		r, err = gzip.NewReader(bytes.NewReader(data))
	case "zlib":
		r, err = zlib.NewReader(bytes.NewReader(data))
	case "flate":
		r = flate.NewReader(bytes.NewReader(data))
	default:
		return nil, fmt.Errorf("unsupported compression algorithm %q", algorithm)
	}
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

// CompressStats compresses data and reports the size and timing metrics.
func CompressStats(algorithm string, data []byte) (CompressionResult, error) {
	start := time.Now()
	compressed, err := Compress(algorithm, data)
	if err != nil {
		return CompressionResult{}, err
	}
	elapsed := time.Since(start)

	res := CompressionResult{
		Algorithm:      algorithm,
		OriginalSize:   len(data),
		CompressedSize: len(compressed),
		Duration:       elapsed,
	}
	if len(data) > 0 {
		res.Ratio = float64(len(compressed)) / float64(len(data))
		res.Saved = 1 - res.Ratio
	}
	return res, nil
}

// BenchmarkCompression runs every supported algorithm over data and returns the
// results sorted by compressed size (smallest first).
func BenchmarkCompression(data []byte) ([]CompressionResult, error) {
	results := make([]CompressionResult, 0, len(compressionAlgorithms))
	for _, algorithm := range compressionAlgorithms {
		res, err := CompressStats(algorithm, data)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].CompressedSize < results[j].CompressedSize
	})
	return results, nil
}
