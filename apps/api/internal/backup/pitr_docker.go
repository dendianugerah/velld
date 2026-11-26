package backup

import (
	"archive/tar"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type recoveryFile struct {
	name, path, content string
	mode                os.FileMode
}

func (d *dockerEngine) uploadRecovery(ctx context.Context, id, data string, files []recoveryFile) error {
	reader, writer := io.Pipe()
	defer reader.Close()
	go func() {
		writer.CloseWithError(writeRecoveryArchive(ctx, writer, data, files))
	}()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, "http://docker/containers/"+id+"/archive?path=%2F", reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-tar")
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Docker recovery archive upload: %s", resp.Status)
	}
	return nil
}

func writeRecoveryArchive(ctx context.Context, output io.Writer, data string, files []recoveryFile) error {
	tw := tar.NewWriter(output)
	writeFile := func(header *tar.Header, path string) error {
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if header.Typeflag == tar.TypeDir {
			return nil
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		_, err = io.CopyN(tw, in, header.Size)
		return err
	}
	if err := tw.WriteHeader(&tar.Header{Name: "velld-pitr/", Mode: 0755, Typeflag: tar.TypeDir}); err != nil {
		return err
	}
	err := filepath.WalkDir(data, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		relative, err := filepath.Rel(data, path)
		if err != nil {
			return err
		}
		switch relative {
		case "postgresql.conf", "postgresql.auto.conf", "pg_hba.conf", "pg_ident.conf", "recovery.signal", "standby.signal":
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported file type in PostgreSQL base backup")
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = "velld-pitr/data/" + filepath.ToSlash(relative)
		return writeFile(header, path)
	})
	if err != nil {
		return err
	}
	for _, file := range files {
		header := &tar.Header{Name: file.name, Mode: 0644}
		switch {
		case file.mode.IsDir():
			header.Mode, header.Typeflag = int64(file.mode.Perm()), tar.TypeDir
			err = tw.WriteHeader(header)
		case file.path != "":
			info, statErr := os.Lstat(file.path)
			if statErr != nil {
				return statErr
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf("non-regular PostgreSQL archive segment")
			}
			header.Size = info.Size()
			err = writeFile(header, file.path)
		default:
			header.Size = int64(len(file.content))
			if err = tw.WriteHeader(header); err == nil {
				_, err = io.WriteString(tw, file.content)
			}
		}
		if err != nil {
			return err
		}
	}
	return tw.Close()
}

func (d *dockerEngine) recoveryRunning(ctx context.Context, id string) bool {
	var details struct{ State struct{ Running bool } }
	code, err := d.do(ctx, http.MethodGet, "/containers/"+id+"/json", nil, &details)
	return err == nil && code == http.StatusOK && details.State.Running
}

func (d *dockerEngine) recoveryExec(ctx context.Context, id string, args []string) (string, error) {
	var exec struct{ Id string }
	code, err := d.do(ctx, http.MethodPost, "/containers/"+id+"/exec", map[string]any{
		"Cmd": args, "AttachStdout": true, "AttachStderr": true,
	}, &exec)
	if err != nil || code != http.StatusCreated {
		return "", fmt.Errorf("create PostgreSQL recovery query: Docker status %d: %w", code, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://docker/exec/"+exec.Id+"/start", strings.NewReader("{}"))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("start PostgreSQL recovery query: Docker status %s", resp.Status)
	}
	var stream [8]byte
	var output strings.Builder
	for {
		_, err := io.ReadFull(resp.Body, stream[:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		size := int64(binary.BigEndian.Uint32(stream[4:]))
		if size+int64(output.Len()) > 64<<10 {
			return "", fmt.Errorf("PostgreSQL recovery query returned too much output")
		}
		if _, err := io.CopyN(&output, resp.Body, size); err != nil {
			return "", err
		}
	}
	var state struct{ ExitCode int }
	code, err = d.do(ctx, http.MethodGet, "/exec/"+exec.Id+"/json", nil, &state)
	if err != nil || code != http.StatusOK || state.ExitCode != 0 {
		return "", fmt.Errorf("PostgreSQL recovery query failed: %s", output.String())
	}
	return output.String(), nil
}
