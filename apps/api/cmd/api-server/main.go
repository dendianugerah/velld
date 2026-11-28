package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dendianugerah/velld/internal"
	"github.com/dendianugerah/velld/internal/auth"
	"github.com/dendianugerah/velld/internal/backup"
	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/connection"
	"github.com/dendianugerah/velld/internal/database"
	"github.com/dendianugerah/velld/internal/middleware"
	"github.com/dendianugerah/velld/internal/settings"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	secrets := common.GetSecrets()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = filepath.Join("data", "velld.db")
	}

	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	db, err := database.Init(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	connManager := connection.NewConnectionManager()

	authRepo := auth.NewAuthRepository(db)
	authService := auth.NewAuthService(authRepo, secrets.JWTSecret)

	if secrets.AdminUsernameCredential != "" && secrets.AdminPasswordCredential != "" {
		_, err := authService.CreateNewUserByEnvData(secrets.AdminUsernameCredential, secrets.AdminPasswordCredential)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Admin user created")
		}
	}

	cryptoService, err := common.NewEncryptionService(secrets.EncryptionKey)
	if err != nil {
		log.Fatal(err)
	}

	connRepo := connection.NewConnectionRepository(db, cryptoService)
	connService := connection.NewConnectionService(connRepo, connManager)

	authHandler := auth.NewAuthHandler(authService)

	authMiddleware := middleware.NewAuthMiddleware(secrets.JWTSecret)

	r := mux.NewRouter()
	r.Use(middleware.CORS)

	healthHandler := internal.NewHealthHandler(db)
	r.HandleFunc("/health", healthHandler.CheckHealth).Methods("GET", "OPTIONS")

	// Public routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST", "OPTIONS")

	// Protected routes
	protected := api.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)
	protected.HandleFunc("/auth/profile", authHandler.GetProfile).Methods("GET", "OPTIONS")

	backupRepo := backup.NewBackupRepository(db)
	settingsRepo := settings.NewSettingsRepository(db)
	settingsService := settings.NewSettingsService(settingsRepo, cryptoService)

	backupService := backup.NewBackupService(
		connRepo,
		"./backups",
		backupRepo,
		settingsService,
		cryptoService,
	)
	if err := backupService.InitializePG18Rehearsals(); err != nil {
		log.Fatalf("Failed to initialize PG 18 rehearsals: %v", err)
	}
	go backupService.ResumeRestoreChecks()
	go backupService.CleanExpiredBackupBranches()

	// Create connHandler after backupService is available
	connHandler := connection.NewConnectionHandler(connService, backupService)
	var adminID uuid.UUID
	if admin, err := authRepo.GetUserByUsername(secrets.AdminUsernameCredential); err == nil {
		adminID = admin.ID
	}
	dockerHandler := connection.NewDockerDiscoveryHandler(connRepo, adminID, func(id string) error {
		return backupService.ScheduleBackup(&backup.ScheduleBackupRequest{
			ConnectionID: id, CronSchedule: "0 0 0 * * *", RetentionDays: 30,
		})
	})

	protected.HandleFunc("/connections/test", connHandler.TestConnection).Methods("POST", "OPTIONS")
	protected.HandleFunc("/connections/docker", dockerHandler.List).Methods("GET", "OPTIONS")
	protected.HandleFunc("/connections/docker/{id}/test", dockerHandler.Test).Methods("POST", "OPTIONS")
	protected.HandleFunc("/connections/docker/{id}", dockerHandler.Import).Methods("POST", "OPTIONS")
	protected.HandleFunc("/connections/{id}/discover", connHandler.DiscoverDatabases).Methods("GET", "OPTIONS")
	protected.HandleFunc("/connections/{id}/databases", connHandler.UpdateSelectedDatabases).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/connections/{id}/settings", connHandler.UpdateConnectionSettings).Methods("POST", "OPTIONS")
	protected.HandleFunc("/connections/{id}", connHandler.GetConnection).Methods("GET", "OPTIONS")
	protected.HandleFunc("/connections/{id}", connHandler.DeleteConnection).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/connections", connHandler.SaveConnection).Methods("POST", "OPTIONS")
	protected.HandleFunc("/connections", connHandler.ListConnections).Methods("GET", "OPTIONS")
	protected.HandleFunc("/connections", connHandler.UpdateConnection).Methods("PUT", "OPTIONS")

	backupHandler := backup.NewBackupHandler(backupService)

	protected.HandleFunc("/connections/{id}/pitr", backupHandler.PointInTimeRecovery).Methods("GET", "POST", "DELETE", "OPTIONS")
	protected.HandleFunc("/backups/schedule", backupHandler.ScheduleBackup).Methods("POST", "OPTIONS")
	protected.HandleFunc("/backups", backupHandler.CreateBackup).Methods("POST", "OPTIONS")
	protected.HandleFunc("/backups", backupHandler.ListBackups).Methods("GET", "OPTIONS")
	protected.HandleFunc("/backups/{id}", backupHandler.GetBackup).Methods("GET", "OPTIONS")
	protected.HandleFunc("/backups/{id}", backupHandler.DeleteBackup).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/backups/{id}/download", backupHandler.DownloadBackup).Methods("GET", "OPTIONS")
	protected.HandleFunc("/backups/{id}/branch", backupHandler.CreateBackupBranch).Methods("POST", "OPTIONS")
	protected.HandleFunc("/backups/{id}/pg18-rehearsal", backupHandler.PG18Rehearsal).Methods("GET", "POST", "OPTIONS")
	protected.HandleFunc("/backups/{id}/schema-diff", backupHandler.SchemaDiff).Methods("POST", "OPTIONS")
	protected.HandleFunc("/connections/{id}/upgrade", backupHandler.GetUpgradeInfo).Methods("GET", "OPTIONS")
	protected.HandleFunc("/connections/{id}/upgrade", backupHandler.UpgradeConnection).Methods("POST", "OPTIONS")
	protected.HandleFunc("/backups/restore", backupHandler.RestoreBackup).Methods("POST", "OPTIONS")
	protected.HandleFunc("/backups/{connection_id}/schedule/disable", backupHandler.DisableBackupSchedule).Methods("POST", "OPTIONS")
	protected.HandleFunc("/backups/{connection_id}/schedule", backupHandler.UpdateBackupSchedule).Methods("PUT", "OPTIONS")

	settingsHandler := settings.NewSettingsHandler(settingsService)

	protected.HandleFunc("/settings", settingsHandler.GetSettings).Methods("GET", "OPTIONS")
	protected.HandleFunc("/settings", settingsHandler.UpdateSettings).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/settings/storage/test", settingsHandler.TestStorage).Methods("POST", "OPTIONS")

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
