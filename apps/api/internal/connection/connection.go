package connection

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/common/response"
	"github.com/gorilla/mux"
)

type BackupService interface {
	CleanupS3BackupsForConnection(connectionID string) error
	RenameS3FolderForConnection(connectionID string, oldName string, newName string) error
}

type ConnectionHandler struct {
	service       *ConnectionService
	backupService BackupService
}

func NewConnectionHandler(service *ConnectionService, backupService BackupService) *ConnectionHandler {
	return &ConnectionHandler{
		service:       service,
		backupService: backupService,
	}
}

func (h *ConnectionHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	var config ConnectionConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isConnected, err := h.service.TestConnection(config)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"isConnected": false,
			"error":       err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"isConnected": isConnected,
		"lastSync":    "",
	})
}

func (h *ConnectionHandler) SaveConnection(w http.ResponseWriter, r *http.Request) {
	var config ConnectionConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	storedConn, err := h.service.SaveConnection(config, userID)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(storedConn)
}

func (h *ConnectionHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	connections, err := h.service.ListConnections(userID)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connections)
}

func (h *ConnectionHandler) GetConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		response.SendError(w, http.StatusBadRequest, "connection id is required")
		return
	}

	connection, err := h.service.GetConnection(id)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(connection)
}

func (h *ConnectionHandler) UpdateConnection(w http.ResponseWriter, r *http.Request) {
	var config ConnectionConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get existing connection to check if name changed
	existingConn, err := h.service.GetConnection(config.ID)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, "Failed to get existing connection")
		return
	}

	storedConn, err := h.service.UpdateConnection(config, userID)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Rename S3 folder if connection name changed
	if existingConn.Name != config.Name && h.backupService != nil {
		if err := h.backupService.RenameS3FolderForConnection(config.ID, existingConn.Name, config.Name); err != nil {
			fmt.Printf("Warning: Failed to rename S3 folder for connection %s: %v\n", config.ID, err)
			// Continue even if S3 rename fails
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(storedConn)
}

func (h *ConnectionHandler) UpdateConnectionSettings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		response.SendError(w, http.StatusBadRequest, "connection id is required")
		return
	}

	var req struct {
		S3CleanupOnRetention *bool `json:"s3_cleanup_on_retention"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.UpdateConnectionSettings(id, req.S3CleanupOnRetention); err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.SendSuccess(w, "Connection settings updated successfully", nil)
}

func (h *ConnectionHandler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		response.SendError(w, http.StatusBadRequest, "connection id is required")
		return
	}

	// Check if cleanup_s3 query parameter is provided
	cleanupS3 := r.URL.Query().Get("cleanup_s3") == "true"

	// Cleanup S3 backups if requested
	if cleanupS3 && h.backupService != nil {
		if err := h.backupService.CleanupS3BackupsForConnection(id); err != nil {
			fmt.Printf("Warning: Failed to cleanup S3 backups for connection %s: %v\n", id, err)
			// Continue with connection deletion even if S3 cleanup fails
		}
	}

	if err := h.service.DeleteConnection(id); err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ConnectionHandler) DiscoverDatabases(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		response.SendError(w, http.StatusBadRequest, "connection id is required")
		return
	}

	databases, err := h.service.DiscoverDatabases(id)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"databases": databases,
	})
}

func (h *ConnectionHandler) UpdateSelectedDatabases(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		response.SendError(w, http.StatusBadRequest, "connection id is required")
		return
	}

	var req struct {
		Databases []string `json:"databases"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.UpdateSelectedDatabases(id, req.Databases); err != nil {
		response.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Selected databases updated successfully",
	})
}
