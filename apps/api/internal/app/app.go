package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dendianugerah/velld/internal/auth"
	"github.com/dendianugerah/velld/internal/backup"
	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/connection"
	"github.com/dendianugerah/velld/internal/database"
	"github.com/dendianugerah/velld/internal/middleware"
	"github.com/dendianugerah/velld/internal/notification"
	"github.com/dendianugerah/velld/internal/settings"
	"github.com/gorilla/mux"
)

// App holds the constructed dependencies and the HTTP handler. It is returned
// by Build so main can run the server and tests can exercise the wiring.
type App struct {
	Handler   *mux.Router
	db        *sql.DB
	backups   *backup.BackupService
	verifier  *backup.VerificationWorker
	closeOnce bool
}

// Build wires the whole application from environment-provided secrets and
// configuration. It does not start any server; the caller owns the lifecycle.
func Build() (*App, error) {
	secrets := common.GetSecrets()

	dbPath := envOrDefault("DB_PATH", filepath.Join("data", "velld.db"))
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := database.Init(dbPath)
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	app := &App{db: db}
	app.Handler, app.backups, app.verifier, err = app.buildRouter(secrets)
	if err != nil {
		db.Close()
		return nil, err
	}

	return app, nil
}

func (a *App) buildRouter(secrets *common.Secrets) (*mux.Router, *backup.BackupService, *backup.VerificationWorker, error) {
	cryptoService, err := common.NewEncryptionService(secrets.EncryptionKey)
	if err != nil {
		return nil, nil, nil, err
	}

	connManager := connection.NewConnectionManager()
	connRepo := connection.NewConnectionRepository(a.db, cryptoService)
	connService := connection.NewConnectionService(connRepo, connManager)

	authRepo := auth.NewAuthRepository(a.db)
	authService := auth.NewAuthService(authRepo, secrets.JWTSecret)
	authHandler := auth.NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(secrets.JWTSecret)

	if err := seedAdminUser(secrets, authService); err != nil {
		return nil, nil, nil, err
	}

	r := newRouter()

	backupRepo := backup.NewBackupRepository(a.db)
	settingsRepo := settings.NewSettingsRepository(a.db)
	notificationRepo := notification.NewNotificationRepository(a.db)
	settingsService := settings.NewSettingsService(settingsRepo, cryptoService)

	backupDir := envOrDefault("BACKUP_DIR", "./backups")
	backupService := backup.NewBackupService(
		connRepo,
		backupDir,
		backupRepo,
		settingsService,
		notificationRepo,
		cryptoService,
	)

	verifier := backup.NewVerificationWorker(backupService, backup.DefaultVerificationWorkers)

	registerRoutes(r, routeDeps{
		db:              a.db,
		authHandler:     authHandler,
		authMiddleware:  authMiddleware,
		connHandler:     connection.NewConnectionHandler(connService, backupService),
		backupHandler:   backup.NewBackupHandler(backupService).WithVerificationWorker(verifier),
		settingsHandler: settings.NewSettingsHandler(settingsService),
		notificationHandler: notification.NewNotificationHandler(
			notification.NewNotificationService(notificationRepo),
		),
	})

	return r, backupService, verifier, nil
}

// seedAdminUser creates the bootstrap admin when registration is disabled.
func seedAdminUser(secrets *common.Secrets, authService *auth.AuthService) error {
	if secrets.IsAllowSignup {
		return nil
	}

	if secrets.AdminUsernameCredential == "" || secrets.AdminPasswordCredential == "" {
		return fmt.Errorf("admin credentials are required when registration is disabled")
	}

	if _, err := authService.CreateNewUserByEnvData(secrets.AdminUsernameCredential, secrets.AdminPasswordCredential); err != nil {
		log.Println(err)
	} else {
		log.Println("Admin user created")
	}
	return nil
}

// Close releases resources held by the app. It is safe to call once.
func (a *App) Close(ctx context.Context) error {
	if a.closeOnce {
		return nil
	}
	a.closeOnce = true

	if a.verifier != nil {
		if err := a.verifier.Stop(ctx); err != nil {
			log.Printf("verification worker did not stop cleanly: %v", err)
		}
	}

	if a.backups != nil {
		if err := a.backups.Stop(ctx); err != nil {
			log.Printf("backup scheduler did not stop cleanly: %v", err)
		}
	}

	return a.db.Close()
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Run builds the application, serves HTTP until ctx is cancelled, then shuts
// down gracefully: the HTTP server stops accepting and drains in-flight
// requests, and the backup scheduler waits for running jobs.
func Run(ctx context.Context, addr string) error {
	app, err := Build()
	if err != nil {
		return err
	}
	defer app.Close(context.Background())

	server := &http.Server{
		Addr:              addr,
		Handler:           app.Handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Server starting on %s", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
	}

	log.Println("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}

	return nil
}
