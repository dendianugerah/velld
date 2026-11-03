package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestTrimTrailingSlash(t *testing.T) {
	cases := map[string]string{
		"http://x/":    "http://x",
		"http://x//":   "http://x",
		"http://x/api": "http://x/api",
	}
	for input, want := range cases {
		if got := trimTrailingSlash(input); got != want {
			t.Fatalf("trimTrailingSlash(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestWalPushUploadsSegment(t *testing.T) {
	var gotPath, gotAuth string
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		buf := make([]byte, 64)
		n, _ := r.Body.Read(buf)
		gotBody = buf[:n]
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	segment := filepath.Join(t.TempDir(), "000000010000000000000001")
	if err := os.WriteFile(segment, []byte("wal-bytes"), 0o600); err != nil {
		t.Fatalf("write segment: %v", err)
	}

	t.Setenv("VELD_TOKEN", "test-token")

	err := walPush([]string{"--connection", "conn-1", "--api", server.URL, segment})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/wal/conn-1/segment/000000010000000000000001" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer test-token" {
		t.Fatalf("auth = %q", gotAuth)
	}
	if string(gotBody) != "wal-bytes" {
		t.Fatalf("body = %q", string(gotBody))
	}
}

func TestWalPushFailsOnServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	segment := filepath.Join(t.TempDir(), "000000010000000000000001")
	if err := os.WriteFile(segment, []byte("wal-bytes"), 0o600); err != nil {
		t.Fatalf("write segment: %v", err)
	}

	t.Setenv("VELD_TOKEN", "test-token")

	if err := walPush([]string{"--connection", "conn-1", "--api", server.URL, segment}); err == nil {
		t.Fatal("expected error on non-2xx response")
	}
}

func TestWalPushRequiresToken(t *testing.T) {
	segment := filepath.Join(t.TempDir(), "000000010000000000000001")
	os.WriteFile(segment, []byte("x"), 0o600)

	t.Setenv("VELD_TOKEN", "")

	if err := walPush([]string{"--connection", "conn-1", "--api", "http://x", segment}); err == nil {
		t.Fatal("expected error without a token")
	}
}
