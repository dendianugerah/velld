package internal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HealthHandler struct {
	startTime time.Time
	db        *sql.DB
}

type ComponentHealth struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type HealthDetails struct {
	Database ComponentHealth `json:"database"`
}

type HealthResponse struct {
	Status  string        `json:"status"`
	Version string        `json:"version"`
	Uptime  string        `json:"uptime"`
	Details HealthDetails `json:"details"`
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		db:        db,
	}
}

func (h *HealthHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(h.startTime)
	days := int(uptime.Hours() / 24)
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60

	now := time.Now()
	dbHealth := ComponentHealth{
		Status:    "up",
		Timestamp: now,
	}

	// Check database connection
	if err := h.db.Ping(); err != nil {
		dbHealth.Status = "down"
	}

	response := HealthResponse{
		Version: "v0.1.0",
		Status:  "up",
		Uptime:  fmt.Sprintf("%dd %dh %dm", days, hours, minutes),
		Details: HealthDetails{
			Database: dbHealth,
		},
	}

	// If any component is down, mark overall status as down
	if dbHealth.Status == "down" {
		response.Status = "down"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
