package backup

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

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

	sourceContent, err := readBackupFile(sourceBackup.Path)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to read source backup: %v", err))
		return
	}

	targetContent, err := readBackupFile(targetBackup.Path)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to read target backup: %v", err))
		return
	}

	diff := generateDiff(sourceContent, targetContent)

	response.SendSuccess(w, "Backup comparison completed", diff)
}

// readBackupFile reads a backup file and returns its content
func readBackupFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var content strings.Builder
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

	// Simple line-by-line comparison
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
				// Try to find matching line in target
				found := false
				for k := j + 1; k < min(j+5, len(targetLines)); k++ {
					if sourceLines[i] == targetLines[k] {
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
