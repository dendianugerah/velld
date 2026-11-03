package app

import (
	"database/sql"

	"github.com/dendianugerah/velld/internal"
	"github.com/dendianugerah/velld/internal/auth"
	"github.com/dendianugerah/velld/internal/backup"
	"github.com/dendianugerah/velld/internal/connection"
	"github.com/dendianugerah/velld/internal/middleware"
	"github.com/dendianugerah/velld/internal/notification"
	"github.com/dendianugerah/velld/internal/settings"
	"github.com/gorilla/mux"
)

type routeDeps struct {
	db                  *sql.DB
	authHandler         *auth.AuthHandler
	authMiddleware      *middleware.AuthMiddleware
	connHandler         *connection.ConnectionHandler
	backupHandler       *backup.BackupHandler
	settingsHandler     *settings.SettingsHandler
	notificationHandler *notification.NotificationHandler
}

func newRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.CORS)
	return r
}

func registerRoutes(r *mux.Router, d routeDeps) {
	r.HandleFunc("/health", internal.NewHealthHandler(d.db).CheckHealth).Methods("GET", "OPTIONS")

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/auth/register", d.authHandler.Register).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/login", d.authHandler.Login).Methods("POST", "OPTIONS")

	protected := api.PathPrefix("").Subrouter()
	protected.Use(d.authMiddleware.RequireAuth)

	registerAuthRoutes(protected, d.authHandler)
	registerConnectionRoutes(protected, d.connHandler)
	registerBackupRoutes(protected, d.backupHandler)
	registerSettingsRoutes(protected, d.settingsHandler)
	registerNotificationRoutes(protected, d.notificationHandler)
}

func registerAuthRoutes(r *mux.Router, h *auth.AuthHandler) {
	r.HandleFunc("/auth/profile", h.GetProfile).Methods("GET", "OPTIONS")
}

func registerConnectionRoutes(r *mux.Router, h *connection.ConnectionHandler) {
	r.HandleFunc("/connections/test", h.TestConnection).Methods("POST", "OPTIONS")
	r.HandleFunc("/connections/{id}/discover", h.DiscoverDatabases).Methods("GET", "OPTIONS")
	r.HandleFunc("/connections/{id}/databases", h.UpdateSelectedDatabases).Methods("PUT", "OPTIONS")
	r.HandleFunc("/connections/{id}/settings", h.UpdateConnectionSettings).Methods("POST", "OPTIONS")
	r.HandleFunc("/connections/{id}", h.GetConnection).Methods("GET", "OPTIONS")
	r.HandleFunc("/connections/{id}", h.DeleteConnection).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/connections", h.SaveConnection).Methods("POST", "OPTIONS")
	r.HandleFunc("/connections", h.ListConnections).Methods("GET", "OPTIONS")
	r.HandleFunc("/connections", h.UpdateConnection).Methods("PUT", "OPTIONS")
}

func registerBackupRoutes(r *mux.Router, h *backup.BackupHandler) {
	r.HandleFunc("/backups/stats", h.GetBackupStats).Methods("GET", "OPTIONS")
	r.HandleFunc("/backups/schedule", h.ScheduleBackup).Methods("POST", "OPTIONS")
	r.HandleFunc("/backups", h.CreateBackup).Methods("POST", "OPTIONS")
	r.HandleFunc("/backups", h.ListBackups).Methods("GET", "OPTIONS")
	r.HandleFunc("/backups/{id}", h.GetBackup).Methods("GET", "OPTIONS")
	r.HandleFunc("/backups/{id}", h.DeleteBackup).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/backups/{id}/download", h.DownloadBackup).Methods("GET", "OPTIONS")
	r.HandleFunc("/backups/restore", h.RestoreBackup).Methods("POST", "OPTIONS")
	r.HandleFunc("/backups/{id}/verify", h.VerifyBackup).Methods("POST", "OPTIONS")
	r.HandleFunc("/backups/{id}/verification", h.GetVerification).Methods("GET", "OPTIONS")
	r.HandleFunc("/backups/compare/{sourceId}/{targetId}", h.CompareBackups).Methods("GET", "OPTIONS")
	r.HandleFunc("/backups/{connection_id}/schedule/disable", h.DisableBackupSchedule).Methods("POST", "OPTIONS")
	r.HandleFunc("/backups/{connection_id}/schedule", h.UpdateBackupSchedule).Methods("PUT", "OPTIONS")
}

func registerSettingsRoutes(r *mux.Router, h *settings.SettingsHandler) {
	r.HandleFunc("/settings", h.GetSettings).Methods("GET", "OPTIONS")
	r.HandleFunc("/settings", h.UpdateSettings).Methods("PUT", "OPTIONS")
}

func registerNotificationRoutes(r *mux.Router, h *notification.NotificationHandler) {
	r.HandleFunc("/notifications", h.GetNotifications).Methods("GET", "OPTIONS")
	r.HandleFunc("/notifications/mark-read", h.MarkAsRead).Methods("POST", "OPTIONS")
}
