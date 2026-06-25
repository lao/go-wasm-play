package ops

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
	"sort"
)

// checksumFactories maps an algorithm name to a constructor for its hash.Hash.
var checksumFactories = map[string]func() hash.Hash{
	"md5":    md5.New,
	"sha1":   sha1.New,
	"sha256": sha256.New,
	"crc32":  func() hash.Hash { return crc32.NewIEEE() },
}

// ChecksumAlgorithms returns the supported checksum algorithm names, sorted.
func ChecksumAlgorithms() []string {
	names := make([]string, 0, len(checksumFactories))
	for name := range checksumFactories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Checksum returns the hex-encoded digest of data using the named algorithm.
// Supported algorithms are md5, sha1, sha256 and crc32.
func Checksum(algorithm string, data []byte) (string, error) {
	factory, ok := checksumFactories[algorithm]
	if !ok {
		return "", fmt.Errorf("unsupported checksum algorithm %q", algorithm)
	}
	h := factory()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// MD5 hashes a string and returns its hex digest. It is kept as a small
// convenience wrapper for the original tutorial demo.
func MD5(input string) string {
	sum, _ := Checksum("md5", []byte(input))
	return sum
}
