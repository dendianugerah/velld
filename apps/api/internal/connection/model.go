package connection

import (
	"time"

	"github.com/google/uuid"
)

type StoredConnection struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Type            string     `json:"type"`
	Host            string     `json:"host"`
	Port            int        `json:"port"`
	Username        string     `json:"username"`
	Password        string     `json:"password"`
	DatabaseName    string     `json:"database_name"`
	SSL             bool       `json:"ssl"`
	SSHEnabled      bool       `json:"ssh_enabled"`
	SSHHost         string     `json:"ssh_host"`
	SSHPort         int        `json:"ssh_port"`
	SSHUsername     string     `json:"ssh_username"`
	SSHPassword     string     `json:"ssh_password"`
	SSHPrivateKey   string     `json:"ssh_private_key"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
	LastConnectedAt *time.Time `json:"last_connected_at"`
	UserID          uuid.UUID  `json:"user_id"`
	Status          string     `json:"status"`
	DatabaseSize    int64      `json:"database_size"`
}

type ConnectionConfig struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	Database      string `json:"database"`
	SSL           bool   `json:"ssl"`
	SSHEnabled    bool   `json:"ssh_enabled"`
	SSHHost       string `json:"ssh_host"`
	SSHPort       int    `json:"ssh_port"`
	SSHUsername   string `json:"ssh_username"`
	SSHPassword   string `json:"ssh_password"`
	SSHPrivateKey string `json:"ssh_private_key"`
}

type ConnectionStats struct {
	TotalConnections int     `json:"total_connections"`
	TotalSize        int64   `json:"total_size"`
	AverageSize      float64 `json:"average_size"`
	ActiveCount      int     `json:"active_count"`
	InactiveCount    int     `json:"inactive_count"`
	SSLCount         int64   `json:"ssl_count"`
	SSLPercentage    float64 `json:"ssl_percentage"`
}

type ConnectionListItem struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Host           string  `json:"host"`
	Status         string  `json:"status"`
	DatabaseSize   int64   `json:"database_size"`
	LastBackupTime *string `json:"last_backup_time"`
	BackupEnabled  bool    `json:"backup_enabled"`
	CronSchedule   *string `json:"cron_schedule"`
	RetentionDays  *int    `json:"retention_days"`
}
