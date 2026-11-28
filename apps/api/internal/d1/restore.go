package d1

/*
#include <stdlib.h>
int sqlite3_complete(const char *);
*/
import "C"

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/mattn/go-sqlite3"
)

type RestoreSummary struct {
	Tables int
	Rows   int64
}

func init() {
	sql.Register("velld-d1-restore", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			if _, err := conn.Exec("PRAGMA trusted_schema=OFF", nil); err != nil {
				return err
			}
			conn.RegisterAuthorizer(func(op int, arg1, arg2, _ string) int {
				switch op {
				case sqlite3.SQLITE_ATTACH, sqlite3.SQLITE_DETACH:
					return sqlite3.SQLITE_DENY
				case sqlite3.SQLITE_PRAGMA:
					switch strings.ToLower(arg1) {
					case "defer_foreign_keys", "foreign_keys":
					case "integrity_check":
						if arg2 != "" {
							return sqlite3.SQLITE_DENY
						}
					default:
						return sqlite3.SQLITE_DENY
					}
				case sqlite3.SQLITE_FUNCTION:
					if strings.EqualFold(arg2, "load_extension") {
						return sqlite3.SQLITE_DENY
					}
				}
				return sqlite3.SQLITE_OK
			})
			return nil
		},
	})
}

func CheckRestore(ctx context.Context, path string) (RestoreSummary, error) {
	var summary RestoreSummary
	file, err := os.Open(path)
	if err != nil {
		return summary, err
	}
	defer file.Close()

	dir, err := os.MkdirTemp("", "velld-d1-restore-*")
	if err != nil {
		return summary, err
	}
	defer os.RemoveAll(dir)
	db, err := sql.Open("velld-d1-restore", filepath.Join(dir, "restored.db")+"?_journal_mode=OFF&_synchronous=OFF")
	if err != nil {
		return summary, err
	}
	defer db.Close()
	conn, err := db.Conn(ctx)
	if err != nil {
		return summary, err
	}
	defer conn.Close()

	if err := importSQL(ctx, conn, file); err != nil {
		return summary, fmt.Errorf("D1 SQL import failed: %w", err)
	}
	if err := conn.Raw(func(driver any) error {
		if !driver.(*sqlite3.SQLiteConn).AutoCommit() {
			return errors.New("D1 SQL import left an open transaction")
		}
		return nil
	}); err != nil {
		return summary, err
	}
	var integrity string
	if err := conn.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrity); err != nil {
		return summary, err
	}
	if integrity != "ok" {
		return summary, fmt.Errorf("SQLite integrity check failed: %s", integrity)
	}

	tables, err := conn.QueryContext(ctx, `SELECT name FROM sqlite_schema WHERE type = 'table' AND name NOT GLOB 'sqlite_*'`)
	if err != nil {
		return summary, err
	}
	defer tables.Close()
	for tables.Next() {
		var table string
		if err := tables.Scan(&table); err != nil {
			return summary, err
		}
		var count int64
		if err := conn.QueryRowContext(ctx, `SELECT count(*) FROM "`+strings.ReplaceAll(table, `"`, `""`)+`"`).Scan(&count); err != nil {
			return summary, err
		}
		summary.Tables++
		summary.Rows += count
	}
	return summary, tables.Err()
}

func importSQL(ctx context.Context, conn *sql.Conn, file io.Reader) error {
	reader := bufio.NewReader(file)
	var pending bytes.Buffer
	executed := false
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if bytes.IndexByte(line, 0) != -1 {
			return errors.New("SQL contains a NUL byte")
		}
		pending.Write(line)
		if len(line) != 0 && bytes.ContainsRune(line, ';') && completeSQL(pending.String()) {
			if _, err := conn.ExecContext(ctx, pending.String()); err != nil {
				return err
			}
			pending.Reset()
			executed = true
		}
		if err == io.EOF {
			break
		}
	}
	if !executed {
		return errors.New("SQL export contains no complete statements")
	}
	if pending.Len() == 0 {
		return nil
	}
	_, err := conn.ExecContext(ctx, pending.String())
	return err
}

func completeSQL(query string) bool {
	text := C.CString(query)
	defer C.free(unsafe.Pointer(text))
	return C.sqlite3_complete(text) != 0
}
