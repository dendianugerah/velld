package backup

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/connection"
)

var pgDumpVersion = regexp.MustCompile(`^pg_dump \(PostgreSQL\) ([0-9]+)(?:\.([0-9]+))?`)

func checkPostgresVersion(clientVersion, serverVersion string) error {
	parts := pgDumpVersion.FindStringSubmatch(clientVersion)
	if parts == nil {
		return fmt.Errorf("unrecognized pg_dump version: %q", clientVersion)
	}
	clientMajor, err := strconv.Atoi(parts[1])
	if err != nil || clientMajor == 0 {
		return fmt.Errorf("invalid pg_dump version: %q", clientVersion)
	}
	client := clientMajor * 100
	if clientMajor < 10 {
		minor, err := strconv.Atoi(parts[2])
		if err != nil || minor >= 100 {
			return fmt.Errorf("invalid pg_dump version: %q", clientVersion)
		}
		client += minor
	}
	server, err := strconv.Atoi(serverVersion)
	if err != nil || server < 10000 {
		return fmt.Errorf("invalid PostgreSQL server_version_num: %q", serverVersion)
	}
	if client < server/100 {
		return fmt.Errorf("PostgreSQL version mismatch: %s is older than server (server_version_num=%s); install a pg_dump client of the same or newer major version", clientVersion, serverVersion)
	}
	return nil
}

func (s *BackupService) verifyPostgresVersion(conn *connection.StoredConnection, dump *exec.Cmd) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientVersion, err := exec.CommandContext(ctx, dump.Path, append([]string{"--version"}, dump.Args[1:]...)...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("read pg_dump version: %w: %s", err, strings.TrimSpace(string(clientVersion)))
	}
	psqlDir := s.findDatabaseRestorePath("postgresql")
	if psqlDir == "" {
		return fmt.Errorf("psql not found; install PostgreSQL client tools to check server compatibility")
	}
	psql := filepath.Join(psqlDir, common.GetPlatformExecutableName(restoreTools["postgresql"]))
	probe := postgresCmd(psql, conn, "-X", "-Atq", "-c", "SHOW server_version_num")
	cmd := exec.CommandContext(ctx, probe.Path, probe.Args[1:]...)
	cmd.Env = probe.Env
	serverVersion, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("read PostgreSQL server version: %w: %s", err, strings.TrimSpace(string(serverVersion)))
	}
	return checkPostgresVersion(strings.TrimSpace(string(clientVersion)), strings.TrimSpace(string(serverVersion)))
}

func writePostgresBackup(cmd *exec.Cmd, path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	compressed := gzip.NewWriter(file)
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = compressed, &stderr
	runErr := cmd.Run()
	gzipErr := compressed.Close()
	fileErr := file.Close()
	if runErr != nil || gzipErr != nil || fileErr != nil {
		_ = os.Remove(path)
	}
	if runErr != nil {
		return fmt.Errorf("%w: %s", runErr, strings.TrimSpace(stderr.String()))
	}
	if gzipErr != nil {
		return gzipErr
	}
	return fileErr
}

func openPostgresBackup(path string) (io.Reader, func(), error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	input := bufio.NewReader(file)
	magic, _ := input.Peek(2)
	if bytes.Equal(magic, []byte{0x1f, 0x8b}) {
		compressed, err := gzip.NewReader(input)
		if err != nil {
			_ = file.Close()
			return nil, nil, err
		}
		return compressed, func() {
			_ = compressed.Close()
			_ = file.Close()
		}, nil
	}
	return input, func() { _ = file.Close() }, nil
}

func postgresRestoreInput(input io.Reader) (io.Reader, error) {
	reader := bufio.NewReader(input)
	var header bytes.Buffer
	for {
		line, err := reader.ReadSlice('\n')
		if err != nil && err != io.EOF && err != bufio.ErrBufferFull {
			return nil, err
		}
		statement := strings.TrimRight(string(line), "\r\n")
		if statement == "SET transaction_timeout = 0;" {
			header.WriteString("SELECT pg_catalog.set_config(name, '0', false) FROM pg_catalog.pg_settings WHERE name = 'transaction_timeout';")
			header.Write(line[len(statement):])
			break
		}
		header.Write(line)
		if err != nil || (statement != "" && !strings.HasPrefix(statement, "--") &&
			!strings.HasPrefix(statement, "\\restrict ") && !strings.HasPrefix(statement, "SET ")) {
			break
		}
	}
	return io.MultiReader(&header, reader), nil
}
