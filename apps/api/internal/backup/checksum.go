package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
)

const defaultChecksumAlgorithm = "sha256"

// ChecksumWriter hashes every byte written through it while forwarding to the
// underlying writer. Wrapping the destination (instead of hashing the file
// afterwards) means the digest is produced in a single pass, which is what the
// streaming backup path relies on.
type ChecksumWriter struct {
	w    io.Writer
	hash hash.Hash
	n    int64
}

func NewChecksumWriter(w io.Writer) *ChecksumWriter {
	return &ChecksumWriter{w: w, hash: sha256.New()}
}

func (c *ChecksumWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	if n > 0 {
		c.hash.Write(p[:n])
		c.n += int64(n)
	}
	return n, err
}

// Checksum returns the hex-encoded digest of everything written so far.
func (c *ChecksumWriter) Checksum() string {
	return hex.EncodeToString(c.hash.Sum(nil))
}

// Size returns the number of bytes written.
func (c *ChecksumWriter) Size() int64 {
	return c.n
}
