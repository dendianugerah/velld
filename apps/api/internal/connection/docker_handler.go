package connection

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/common/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type DockerDiscoveryHandler struct {
	discovery *dockerDiscovery
	repo      *ConnectionRepository
	adminID   uuid.UUID
	schedule  func(connectionID string) error
	importMu  sync.Mutex
}

func NewDockerDiscoveryHandler(repo *ConnectionRepository, adminID uuid.UUID, schedule func(string) error) *DockerDiscoveryHandler {
	var hostname string
	if _, err := os.Stat("/.dockerenv"); err == nil {
		hostname, _ = os.Hostname()
	}
	return &DockerDiscoveryHandler{
		discovery: newDockerDiscovery("/var/run/docker.sock", hostname),
		repo:      repo,
		adminID:   adminID,
		schedule:  schedule,
	}
}

func (h *DockerDiscoveryHandler) authorize(w http.ResponseWriter, r *http.Request) bool {
	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil || h.adminID == uuid.Nil || h.adminID != userID {
		response.SendError(w, http.StatusForbidden, "Docker discovery is available to the configured Velld admin only.")
		return false
	}
	return true
}

func (h *DockerDiscoveryHandler) List(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r) {
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	candidates, hint := h.discovery.scan(r.Context())
	added, err := h.repo.connectionEndpoints(h.adminID)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, "Could not check existing connections.")
		return
	}
	available := make([]dockerCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if !candidate.alreadyAdded(added) {
			available = append(available, candidate)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Containers []dockerCandidate `json:"containers"`
		Hint       string            `json:"hint,omitempty"`
	}{available, hint})
}

func (h *DockerDiscoveryHandler) Import(w http.ResponseWriter, r *http.Request) {
	h.handleCandidate(w, r, true)
}

func (h *DockerDiscoveryHandler) Test(w http.ResponseWriter, r *http.Request) {
	h.handleCandidate(w, r, false)
}

func (h *DockerDiscoveryHandler) handleCandidate(w http.ResponseWriter, r *http.Request, save bool) {
	if !h.authorize(w, r) {
		return
	}

	var req struct {
		Database string `json:"database"`
		Name     string `json:"name"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, "Invalid Docker import request.")
		return
	}
	id := mux.Vars(r)["id"]
	if save {
		h.importMu.Lock()
		defer h.importMu.Unlock()
	}
	candidates, _ := h.discovery.scan(r.Context())
	for _, candidate := range candidates {
		if candidate.ID != id {
			continue
		}
		added, err := h.repo.connectionEndpoints(h.adminID)
		if err != nil {
			response.SendError(w, http.StatusInternalServerError, "Could not check existing connections.")
			return
		}
		if candidate.alreadyAdded(added) {
			response.SendError(w, http.StatusConflict, "This Docker database is already connected.")
			return
		}
		if database := strings.TrimSpace(req.Database); database != "" {
			candidate.config.Database = database
		}
		if name := strings.TrimSpace(req.Name); name != "" {
			candidate.config.Name = name
		}
		if candidate.config.Database == "" {
			response.SendError(w, http.StatusBadRequest, "Enter the database name to back up.")
			return
		}
		if strings.ContainsAny(candidate.config.Database, "/\\\x00") {
			response.SendError(w, http.StatusBadRequest, "The database name cannot contain path separators or a null character.")
			return
		}
		service := NewConnectionService(h.repo, NewConnectionManager())
		if !save {
			if connected, err := service.TestConnection(candidate.config); err != nil || !connected {
				response.SendError(w, http.StatusBadRequest, "Could not connect with the detected settings. Check the database name and credentials.")
				return
			}
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]bool{"isConnected": true})
			return
		}
		conn, err := service.SaveConnection(candidate.config, h.adminID)
		if err != nil {
			response.SendError(w, http.StatusBadRequest, "Could not connect with the detected settings. Check the database name and credentials, or add it manually.")
			return
		}
		if err := h.schedule(conn.ID); err != nil {
			if rollbackErr := service.DeleteConnection(conn.ID); rollbackErr != nil {
				log.Printf("Docker import schedule rollback failed for %s: %v", conn.ID, errors.Join(err, rollbackErr))
				response.SendError(w, http.StatusInternalServerError, "The connection was added but the backup schedule failed. Set it up from Connections.")
				return
			}
			response.SendError(w, http.StatusInternalServerError, "Could not set up the backup schedule. The connection was not added.")
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": conn.ID})
		return
	}
	response.SendError(w, http.StatusNotFound, "Docker database unavailable. Refresh discovery and try again.")
}
