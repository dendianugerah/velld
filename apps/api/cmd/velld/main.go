package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// velld is the helper binary that runs on a PostgreSQL host. Its wal-push
// subcommand is what postgresql.conf's archive_command invokes to ship a
// finished WAL segment to a Velld server.
func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "wal-push":
		if err := walPush(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "velld wal-push: %v\n", err)
			os.Exit(1)
		}
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "velld: unknown command %q\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `velld - helper for PostgreSQL WAL archiving

Usage:
  velld wal-push --connection <id> --api <url> <segment-path>

Environment:
  VELD_TOKEN   bearer token used to authenticate against the Velld API

Exit status is non-zero on failure so postgresql.conf archive_command retries.
`)
}

type walPushOptions struct {
	connection string
	api        string
	segment    string
}

func walPush(args []string) error {
	fs := flag.NewFlagSet("wal-push", flag.ContinueOnError)
	var opts walPushOptions
	fs.StringVar(&opts.connection, "connection", "", "connection id")
	fs.StringVar(&opts.api, "api", os.Getenv("VELD_API"), "Velld API base URL")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return fmt.Errorf("expected exactly one WAL segment path")
	}
	opts.segment = fs.Arg(0)

	if opts.connection == "" {
		return fmt.Errorf("--connection is required")
	}
	if opts.api == "" {
		return fmt.Errorf("--api (or VELD_API) is required")
	}

	token := os.Getenv("VELD_TOKEN")
	if token == "" {
		return fmt.Errorf("VELD_TOKEN is not set")
	}

	info, err := os.Stat(opts.segment)
	if err != nil {
		return fmt.Errorf("stat segment: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("segment %s is empty", opts.segment)
	}

	content, err := os.ReadFile(opts.segment)
	if err != nil {
		return fmt.Errorf("read segment: %w", err)
	}

	segment := filepath.Base(opts.segment)
	url := fmt.Sprintf("%s/api/wal/%s/segment/%s", trimTrailingSlash(opts.api), opts.connection, segment)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload segment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	return nil
}

func trimTrailingSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
