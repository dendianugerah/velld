package backup

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

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

	backup, err := h.backupService.GetBackup(backupID)
	if err != nil {
		if err == sql.ErrNoRows {
			response.SendError(w, http.StatusNotFound, "Backup not found")
			return
		}
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendSuccess(w, "Backup retrieved successfully", backup)
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
	offset := (page - 1) * limit

	opts := BackupListOptions{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
		Search: search,
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
	if req.RetentionDays <= 0 {
		response.SendError(w, http.StatusBadRequest, "retention_days must be greater than 0")
		return
	}

	err := h.backupService.ScheduleBackup(&req)
	if err != nil {
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
	if req.RetentionDays <= 0 {
		response.SendError(w, http.StatusBadRequest, "retention_days must be greater than 0")
		return
	}

	err := h.backupService.UpdateBackupSchedule(connectionID, &req)
	if err != nil {
		if err == sql.ErrNoRows {
			response.SendError(w, http.StatusNotFound, "No active schedule found")
			return
		}
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendSuccess(w, "Backup schedule updated successfully", nil)
}

func (h *BackupHandler) GetBackupStats(w http.ResponseWriter, r *http.Request) {
	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	stats, err := h.backupService.GetBackupStats(userID)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendSuccess(w, "Backup statistics retrieved successfully", stats)
}

func (h *BackupHandler) DownloadBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	backupID := vars["id"]

	backup, err := h.backupService.GetBackup(backupID)
	if err != nil {
		if err == sql.ErrNoRows {
			response.SendError(w, http.StatusNotFound, "Backup not found")
			return
		}
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	file, err := os.Open(backup.Path)
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
