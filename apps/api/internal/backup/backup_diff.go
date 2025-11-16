package backup

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/common/response"
	"github.com/gorilla/mux"
)

type DiffChange struct {
	Type       string `json:"type"`        // "added", "removed", "modified", "unchanged"
	Content    string `json:"content"`     // The actual line content
	LineNumber int    `json:"line_number"` // Line number in the file
	OldLine    int    `json:"old_line,omitempty"`
	NewLine    int    `json:"new_line,omitempty"`
}

type DiffResponse struct {
	Added     int          `json:"added"`
	Removed   int          `json:"removed"`
	Modified  int          `json:"modified"`
	Unchanged int          `json:"unchanged"`
	Changes   []DiffChange `json:"changes"`
}

// CompareBackups handles the comparison of two backup files
func (h *BackupHandler) CompareBackups(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourceID := vars["sourceId"]
	targetID := vars["targetId"]

	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	sourceBackup, err := h.backupService.GetBackup(sourceID)
	if err != nil {
		response.SendError(w, http.StatusNotFound, "Source backup not found")
		return
	}

	targetBackup, err := h.backupService.GetBackup(targetID)
	if err != nil {
		response.SendError(w, http.StatusNotFound, "Target backup not found")
		return
	}

	// Ensure both backup files are available (local or download from S3)
	sourceFilePath, sourceIsTemp, err := h.backupService.ensureBackupFileAvailable(sourceBackup, userID)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to access source backup: %v", err))
		return
	}

	targetFilePath, targetIsTemp, err := h.backupService.ensureBackupFileAvailable(targetBackup, userID)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to access target backup: %v", err))
		return
	}

	// Clean up temp files after comparison
	defer func() {
		if sourceIsTemp {
			if err := os.Remove(sourceFilePath); err != nil {
				fmt.Printf("Warning: Failed to remove temp source file %s: %v\n", sourceFilePath, err)
			}
		}
		if targetIsTemp {
			if err := os.Remove(targetFilePath); err != nil {
				fmt.Printf("Warning: Failed to remove temp target file %s: %v\n", targetFilePath, err)
			}
		}
	}()

	sourceContent, err := readBackupFile(sourceFilePath)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to read source backup: %v", err))
		return
	}

	targetContent, err := readBackupFile(targetFilePath)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to read target backup: %v", err))
		return
	}

	diff := generateDiff(sourceContent, targetContent)

	response.SendSuccess(w, "Backup comparison completed", diff)
}

// readBackupFile reads a backup file and returns its content
// For very large files, this limits the content to prevent memory issues
func readBackupFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Check file size first
	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	// Limit file size to 10MB to prevent memory issues
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if fileInfo.Size() > maxFileSize {
		return "", fmt.Errorf("file too large for comparison (%d bytes). Maximum size is %d bytes", fileInfo.Size(), maxFileSize)
	}

	var content strings.Builder
	// Pre-allocate capacity based on file size for better performance
	content.Grow(int(fileInfo.Size()))
	
	scanner := bufio.NewScanner(file)

	// Read file with a larger buffer for big dump files
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return content.String(), nil
}

// generateDiff creates a structured diff between two texts
func generateDiff(source, target string) DiffResponse {
	sourceLines := strings.Split(source, "\n")
	targetLines := strings.Split(target, "\n")

	var changes []DiffChange
	added := 0
	removed := 0
	unchanged := 0
	lineNumber := 0

	// Build a map of target lines for faster lookups
	targetLineMap := make(map[string][]int)
	for i, line := range targetLines {
		targetLineMap[line] = append(targetLineMap[line], i)
	}

	// Simple line-by-line comparison with optimized matching
	i, j := 0, 0
	for i < len(sourceLines) || j < len(targetLines) {
		lineNumber++

		if i >= len(sourceLines) {
			// Only target lines remaining (additions)
			added++
			changes = append(changes, DiffChange{
				Type:       "added",
				Content:    "+ " + targetLines[j],
				LineNumber: lineNumber,
				NewLine:    j,
			})
			j++
		} else if j >= len(targetLines) {
			// Only source lines remaining (deletions)
			removed++
			changes = append(changes, DiffChange{
				Type:       "removed",
				Content:    "- " + sourceLines[i],
				LineNumber: lineNumber,
				OldLine:    i,
			})
			i++
		} else if sourceLines[i] == targetLines[j] {
			// Lines are the same
			if shouldShowLine(changes, lineNumber) {
				unchanged++
				changes = append(changes, DiffChange{
					Type:       "unchanged",
					Content:    "  " + sourceLines[i],
					LineNumber: lineNumber,
					OldLine:    i,
					NewLine:    j,
				})
			} else {
				unchanged++
			}
			i++
			j++
		} else {
			// Lines are different - check if it's a modification or add/remove
			// Simple heuristic: if next lines match, it's a modification
			if i+1 < len(sourceLines) && j+1 < len(targetLines) && sourceLines[i+1] == targetLines[j+1] {
				// Modified line
				removed++
				changes = append(changes, DiffChange{
					Type:       "removed",
					Content:    "- " + sourceLines[i],
					LineNumber: lineNumber,
					OldLine:    i,
				})
				lineNumber++
				added++
				changes = append(changes, DiffChange{
					Type:       "added",
					Content:    "+ " + targetLines[j],
					LineNumber: lineNumber,
					NewLine:    j,
				})
				i++
				j++
			} else {
				// Use map lookup instead of nested loop for better performance
				found := false
				lookAhead := 5
				if indices, exists := targetLineMap[sourceLines[i]]; exists {
					// Check if any of the matching indices are within our lookahead window
					for _, k := range indices {
						if k > j && k < j+lookAhead {
							// Line was removed from source, added to target
							for l := j; l < k; l++ {
								lineNumber++
								added++
								changes = append(changes, DiffChange{
									Type:       "added",
									Content:    "+ " + targetLines[l],
									LineNumber: lineNumber,
									NewLine:    l,
								})
							}
							j = k
							found = true
							break
						}
					}
				}

				if !found {
					// Just mark as removed
					removed++
					changes = append(changes, DiffChange{
						Type:       "removed",
						Content:    "- " + sourceLines[i],
						LineNumber: lineNumber,
						OldLine:    i,
					})
					i++
				}
			}
		}

		// Limit to prevent memory issues
		if len(changes) > 1000 {
			break
		}
	}

	return DiffResponse{
		Added:     added,
		Removed:   removed,
		Modified:  0,
		Unchanged: unchanged,
		Changes:   changes,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// shouldShowLine determines if an unchanged line should be shown
func shouldShowLine(changes []DiffChange, lineNumber int) bool {
	contextLines := 3

	if lineNumber <= contextLines {
		return true
	}

	recentChanges := 0
	for i := len(changes) - 1; i >= 0 && i >= len(changes)-10; i-- {
		if changes[i].Type != "unchanged" {
			recentChanges++
		}
	}

	return recentChanges > 0
}
