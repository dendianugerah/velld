package backup

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/common/response"
	"github.com/gorilla/mux"
)

type BackupHandler struct {
	backupService *BackupService
}

func NewBackupHandler(bs *BackupService) *BackupHandler {
	return &BackupHandler{
		backupService: bs,
	}
}

func (h *BackupHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	var req BackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	backup, err := h.backupService.CreateBackup(req.ConnectionID)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendSuccess(w, "Backup created successfully", backup)
}

func (h *BackupHandler) GetBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	backupID := vars["id"]

	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	backups, _, err := h.backupService.GetAllBackupsWithPagination(BackupListOptions{UserID: userID, BackupID: backupID, Limit: 1})
	if err != nil {
		if err == sql.ErrNoRows {
			response.SendError(w, http.StatusNotFound, "Backup not found")
			return
		}
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(backups) == 0 {
		response.SendError(w, http.StatusNotFound, "Backup not found")
		return
	}

	response.SendSuccess(w, "Backup retrieved successfully", backups[0])
}

func (h *BackupHandler) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	backupID := vars["id"]

	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.backupService.DeleteBackup(backupID, userID); err != nil {
		if err == sql.ErrNoRows {
			response.SendError(w, http.StatusNotFound, "Backup not found")
			return
		}
		if err.Error() == "unauthorized" {
			response.SendError(w, http.StatusForbidden, "Not authorized to delete this backup")
			return
		}
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendSuccess(w, "Backup deleted successfully", nil)
}

func (h *BackupHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	page := 1
	limit := 10
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	search := r.URL.Query().Get("search")
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	opts := BackupListOptions{
		UserID:       userID,
		Limit:        limit,
		Offset:       offset,
		Search:       search,
		Status:       r.URL.Query().Get("status"),
		DatabaseType: r.URL.Query().Get("engine"),
		Verification: r.URL.Query().Get("verification"),
		ConnectionID: r.URL.Query().Get("connection_id"),
	}
	dateRange := r.URL.Query().Get("date")
	if dateRange != "" && dateRange != "all" {
		days, ok := map[string]int{"24hours": 1, "7days": 7, "30days": 30}[dateRange]
		if !ok {
			response.SendError(w, http.StatusBadRequest, "Invalid date filter")
			return
		}
		from := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
		opts.DateFrom = &from
	}

	backups, total, err := h.backupService.GetAllBackupsWithPagination(opts)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendPaginatedSuccess(w, "Backups retrieved successfully", backups, page, limit, total)
}

func (h *BackupHandler) ScheduleBackup(w http.ResponseWriter, r *http.Request) {
	var req ScheduleBackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.ConnectionID == "" {
		response.SendError(w, http.StatusBadRequest, "connection_id is required")
		return
	}
	if req.CronSchedule == "" {
		response.SendError(w, http.StatusBadRequest, "cron_schedule is required")
		return
	}
	err := h.backupService.ScheduleBackup(&req)
	if err != nil {
		if err == errInvalidRetention {
			response.SendError(w, http.StatusBadRequest, err.Error())
			return
		}
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendSuccess(w, "Backup scheduled successfully", nil)
}

func (h *BackupHandler) DisableBackupSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectionID := vars["connection_id"]

	err := h.backupService.DisableBackupSchedule(connectionID)
	if err != nil {
		if err == sql.ErrNoRows {
			response.SendError(w, http.StatusNotFound, "No active schedule found")
			return
		}
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendSuccess(w, "Backup schedule disabled successfully", nil)
}

func (h *BackupHandler) UpdateBackupSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectionID := vars["connection_id"]

	var req UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.CronSchedule == "" {
		response.SendError(w, http.StatusBadRequest, "cron_schedule is required")
		return
	}
	err := h.backupService.UpdateBackupSchedule(connectionID, &req)
	if err != nil {
		if err == errInvalidRetention {
			response.SendError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err == sql.ErrNoRows {
			response.SendError(w, http.StatusNotFound, "No active schedule found")
			return
		}
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendSuccess(w, "Backup schedule updated successfully", nil)
}

func (h *BackupHandler) DownloadBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	backupID := vars["id"]

	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	backup, err := h.backupService.GetBackup(backupID)
	if err != nil {
		if err == sql.ErrNoRows {
			response.SendError(w, http.StatusNotFound, "Backup not found")
			return
		}
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Ensure backup file is available (local or download from S3)
	filePath, isTemp, err := h.backupService.ensureBackupFileAvailable(backup, userID)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Clean up temp file after download if needed
	if isTemp {
		defer func() {
			if err := os.Remove(filePath); err != nil {
				fmt.Printf("Warning: Failed to remove temp file %s: %v\n", filePath, err)
			}
		}()
	}

	file, err := os.Open(filePath)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, "Failed to open backup file")
		return
	}
	defer file.Close()

	filename := filepath.Base(backup.Path)
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "application/octet-stream")

	_, err = io.Copy(w, file)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, "Failed to send file")
		return
	}
}

func (h *BackupHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	var req RestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.BackupID == "" {
		response.SendError(w, http.StatusBadRequest, "backup_id is required")
		return
	}

	if req.ConnectionID == "" {
		response.SendError(w, http.StatusBadRequest, "connection_id is required")
		return
	}

	err := h.backupService.RestoreBackup(req.BackupID, req.ConnectionID)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendSuccess(w, "Backup restored successfully", nil)
}
