package backup

import (
	"fmt"
	"os"
	"strings"
)

// writeMySQLOptionFile writes a mode 0600 MySQL option file containing the
// password. Passing it via --defaults-extra-file keeps the credential out of
// the process argv, where any local user could see it with ps.
func writeMySQLOptionFile(password string) (string, func(), error) {
	content := fmt.Sprintf("[client]\npassword=%s\n", escapeMySQLOptionValue(password))
	return writeCredentialFile("velld-mysql-", content)
}

// writeMongoConfigFile writes a mode 0600 YAML config with the credentials so
// mongodump does not receive --password on the command line.
func writeMongoConfigFile(username, password string) (string, func(), error) {
	var b strings.Builder
	b.WriteString("auth:\n")
	if username != "" {
		fmt.Fprintf(&b, "  username: %s\n", yamlQuote(username))
	}
	fmt.Fprintf(&b, "  password: %s\n", yamlQuote(password))
	return writeCredentialFile("velld-mongo-", b.String())
}

func writeCredentialFile(prefix, content string) (string, func(), error) {
	file, err := os.CreateTemp("", prefix+"*.cnf")
	if err != nil {
		return "", nil, fmt.Errorf("create credential file: %w", err)
	}

	name := file.Name()
	cleanup := func() {
		if err := os.Remove(name); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to remove credential file %s: %v\n", name, err)
		}
	}

	if err := file.Chmod(0o600); err != nil {
		file.Close()
		cleanup()
		return "", nil, fmt.Errorf("restrict credential file permissions: %w", err)
	}

	if _, err := file.WriteString(content); err != nil {
		file.Close()
		cleanup()
		return "", nil, fmt.Errorf("write credential file: %w", err)
	}

	if err := file.Close(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("close credential file: %w", err)
	}

	return name, cleanup, nil
}

func escapeMySQLOptionValue(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r")
	return replacer.Replace(value)
}

func yamlQuote(value string) string {
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}
