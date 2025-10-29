package common

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func ParseTime(timeStr string) (time.Time, error) {
	formats := []string{
		time.RFC3339,                // "2006-01-02T15:04:05Z07:00"
		"2006-01-02 15:04:05-07:00", // SQLite format with timezone
		"2006-01-02 15:04:05+07:00", // SQLite format with timezone
		"2006-01-02 15:04:05",       // SQLite format without timezone
	}

	var lastErr error
	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, fmt.Errorf("could not parse time '%s': %v", timeStr, lastErr)
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	claims, ok := ctx.Value("user").(jwt.MapClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid user claims")
	}

	userIDStr := fmt.Sprintf("%v", claims["user_id"])
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID format: %v", err)
	}

	return userID, nil
}

var CommonBinaryPaths = map[string][]string{
	"windows": {
		"C:\\Program Files\\PostgreSQL\\*\\bin",
		"C:\\Program Files\\MySQL\\*\\bin",
		"C:\\Program Files\\MariaDB*\\bin",
		"C:\\Program Files\\MongoDB\\*\\bin",
	},
	"linux": {
		"/usr/bin",
		"/usr/local/bin",
		"/opt/postgresql*/bin",
		"/opt/mysql*/bin",
	},
	"darwin": {
		"/opt/homebrew/bin",
		"/usr/local/bin",
		"/opt/homebrew/opt/postgresql@*/bin",
		"/opt/homebrew/opt/mysql@*/bin",
	},
}

// Cache for binary paths to avoid repeated filesystem searches
var (
	binaryPathCache = make(map[string]string)
	binaryPathMutex sync.RWMutex
)

func FindBinaryPath(dbType, toolName string) string {
	cacheKey := dbType + ":" + toolName
	
	// Check cache first
	binaryPathMutex.RLock()
	if cachedPath, found := binaryPathCache[cacheKey]; found {
		binaryPathMutex.RUnlock()
		return cachedPath
	}
	binaryPathMutex.RUnlock()

	execName := GetPlatformExecutableName(toolName)

	// 1. Try user-defined path if provided
	// if userPath != nil && *userPath != "" {
	// 	toolPath := filepath.Join(*userPath, execName)
	// 	if _, err := os.Stat(toolPath); err == nil {
	// 		return *userPath
	// 	}
	// }

	var foundPath string

	// 2. Search common installation paths with wildcard support
	if paths, ok := CommonBinaryPaths[runtime.GOOS]; ok {
		for _, pathPattern := range paths {
			matches, _ := filepath.Glob(pathPattern)
			for _, path := range matches {
				toolPath := filepath.Join(path, execName)
				if _, err := os.Stat(toolPath); err == nil {
					foundPath = path
					goto cacheAndReturn
				}
			}
		}
	}

	// 3. Try PATH environment as last resort
	if path, err := exec.LookPath(execName); err == nil {
		foundPath = filepath.Dir(path)
		goto cacheAndReturn
	}

cacheAndReturn:
	// Cache the result (even if empty) to avoid repeated searches
	binaryPathMutex.Lock()
	binaryPathCache[cacheKey] = foundPath
	binaryPathMutex.Unlock()

	return foundPath
}

func GetPlatformExecutableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func SanitizeConnectionName(name string) string {
	name = strings.ToLower(name)

	re := regexp.MustCompile(`[^a-z0-9]+`)
	sanitized := re.ReplaceAllString(name, "_")

	sanitized = strings.Trim(sanitized, "_")

	if len(sanitized) > 200 {
		sanitized = sanitized[:200]
		sanitized = strings.TrimRight(sanitized, "_")
	}

	if sanitized == "" {
		sanitized = "backup"
	}

	return sanitized
}
