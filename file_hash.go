// Package checksum computes checksums, like MD5 or SHA256, for large files
package checksum

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
)

const bufferSize = 65536

// MD5sum returns MD5 checksum of filename
func MD5sum(filename string) (string, error) {
	return Sum(md5.New(), filename)
}

// SHA256sum returns SHA256 checksum of filename
func SHA256sum(filename string) (string, error) {
	return Sum(sha256.New(), filename)
}

// Sum calculates the hash based on a provided hash provider
func Sum(hashAlgorithm hash.Hash, filename string) (string, error) {
	if info, err := os.Stat(filename); err != nil || info.IsDir() {
		return "", err
	}

	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	buf := make([]byte, bufferSize)
	reader := bufio.NewReader(file)
here:
	for {
		switch n, err := reader.Read(buf); err {
		case nil:
			hashAlgorithm.Write(buf[:n])
		case io.EOF:
			break here
		default:
			return "", err
		}
	}
	return fmt.Sprintf("%x", hashAlgorithm.Sum(nil)), nil
}
