package connection

import (
	"time"

	"github.com/dendianugerah/velld/internal/d1"
	"github.com/google/uuid"
)

type StoredConnection struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	Type                 string     `json:"type"`
	Host                 string     `json:"host"`
	Port                 int        `json:"port"`
	Username             string     `json:"username"`
	Password             string     `json:"password"`
	DatabaseName         string     `json:"database_name"`
	MongoAuthSource      string     `json:"mongo_auth_source,omitempty"`
	D1AccountID          string     `json:"d1_account_id,omitempty"`
	D1DatabaseID         string     `json:"d1_database_id,omitempty"`
	D1APIToken           string     `json:"-"`
	SelectedDatabases    []string   `json:"selected_databases"`
	SSL                  bool       `json:"ssl"`
	SSHEnabled           bool       `json:"ssh_enabled"`
	SSHHost              string     `json:"ssh_host"`
	SSHPort              int        `json:"ssh_port"`
	SSHUsername          string     `json:"ssh_username"`
	SSHPassword          string     `json:"ssh_password"`
	SSHPrivateKey        string     `json:"ssh_private_key"`
	S3CleanupOnRetention bool       `json:"s3_cleanup_on_retention"`
	CreatedAt            string     `json:"created_at"`
	UpdatedAt            string     `json:"updated_at"`
	LastConnectedAt      *time.Time `json:"last_connected_at"`
	UserID               uuid.UUID  `json:"user_id"`
	Status               string     `json:"status"`
	DatabaseSize         int64      `json:"database_size"`
}

type ConnectionConfig struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Type                 string `json:"type"`
	Host                 string `json:"host"`
	Port                 int    `json:"port"`
	Username             string `json:"username"`
	Password             string `json:"password"`
	Database             string `json:"database"`
	MongoAuthSource      string `json:"mongo_auth_source,omitempty"`
	D1AccountID          string `json:"d1_account_id,omitempty"`
	D1DatabaseID         string `json:"d1_database_id,omitempty"`
	D1APIToken           string `json:"d1_api_token,omitempty"`
	SSL                  bool   `json:"ssl"`
	SSHEnabled           bool   `json:"ssh_enabled"`
	SSHHost              string `json:"ssh_host"`
	SSHPort              int    `json:"ssh_port"`
	SSHUsername          string `json:"ssh_username"`
	SSHPassword          string `json:"ssh_password"`
	SSHPrivateKey        string `json:"ssh_private_key"`
	S3CleanupOnRetention *bool  `json:"s3_cleanup_on_retention,omitempty"`
}

func (conn StoredConnection) D1Credentials() d1.Credentials {
	return d1.Credentials{AccountID: conn.D1AccountID, DatabaseID: conn.D1DatabaseID, APIToken: conn.D1APIToken}
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
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	Type                 string  `json:"type"`
	Host                 string  `json:"host"`
	Status               string  `json:"status"`
	DatabaseSize         int64   `json:"database_size"`
	LastBackupTime       *string `json:"last_backup_time"`
	BackupEnabled        bool    `json:"backup_enabled"`
	CronSchedule         *string `json:"cron_schedule"`
	RetentionDays        *int    `json:"retention_days"`
	RetentionCount       *int    `json:"retention_count"`
	S3CleanupOnRetention bool    `json:"s3_cleanup_on_retention"`
}
