package backup

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestChecksumWriterMatchesPlainSHA256(t *testing.T) {
	payload := []byte("database dump contents")

	var sink bytes.Buffer
	cw := NewChecksumWriter(&sink)

	if _, err := cw.Write(payload); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	sum := sha256.Sum256(payload)
	want := hex.EncodeToString(sum[:])

	if got := cw.Checksum(); got != want {
		t.Fatalf("checksum = %s, want %s", got, want)
	}
	if sink.String() != string(payload) {
		t.Fatalf("writer did not forward payload: %q", sink.String())
	}
	if cw.Size() != int64(len(payload)) {
		t.Fatalf("size = %d, want %d", cw.Size(), len(payload))
	}
}

func TestChecksumWriterAcrossMultipleWrites(t *testing.T) {
	var sink bytes.Buffer
	cw := NewChecksumWriter(&sink)

	parts := [][]byte{[]byte("a"), []byte("bc"), []byte("def")}
	for _, p := range parts {
		if _, err := cw.Write(p); err != nil {
			t.Fatalf("unexpected write error: %v", err)
		}
	}

	sum := sha256.Sum256([]byte("abcdef"))
	want := hex.EncodeToString(sum[:])

	if got := cw.Checksum(); got != want {
		t.Fatalf("checksum = %s, want %s", got, want)
	}
	if cw.Size() != 6 {
		t.Fatalf("size = %d, want 6", cw.Size())
	}
}
