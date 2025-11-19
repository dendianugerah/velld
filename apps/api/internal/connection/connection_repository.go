package connection

import (
	"database/sql"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/google/uuid"
)

type ConnectionRepository struct {
	db     *sql.DB
	crypto *common.EncryptionService
}

func NewConnectionRepository(db *sql.DB, crypto *common.EncryptionService) *ConnectionRepository {
	return &ConnectionRepository{
		db:     db,
		crypto: crypto,
	}
}

func (r *ConnectionRepository) Save(conn StoredConnection) error {
	username, err := r.crypto.Encrypt(conn.Username)
	if err != nil {
		return err
	}

	password, err := r.crypto.Encrypt(conn.Password)
	if err != nil {
		return err
	}

	sshPassword := ""
	if conn.SSHPassword != "" {
		sshPassword, err = r.crypto.Encrypt(conn.SSHPassword)
		if err != nil {
			return err
		}
	}

	sshPrivateKey := ""
	if conn.SSHPrivateKey != "" {
		sshPrivateKey, err = r.crypto.Encrypt(conn.SSHPrivateKey)
		if err != nil {
			return err
		}
	}

	sslInt := 0
	if conn.SSL {
		sslInt = 1
	}

	sshEnabledInt := 0
	if conn.SSHEnabled {
		sshEnabledInt = 1
	}

	s3CleanupInt := 1 // default to true
	if !conn.S3CleanupOnRetention {
		s3CleanupInt = 0
	}

	query := `
		INSERT INTO connections (
			id, name, type, host, port, username, password, 
			database_name, ssl, database_size, created_at, updated_at, 
			last_connected_at, user_id, status, ssh_enabled, ssh_host, 
			ssh_port, ssh_username, ssh_password, ssh_private_key, s3_cleanup_on_retention
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
		)`

	_, err = r.db.Exec(
		query,
		conn.ID,
		conn.Name,
		conn.Type,
		conn.Host,
		conn.Port,
		username,
		password,
		conn.DatabaseName,
		sslInt,
		conn.DatabaseSize,
		conn.CreatedAt,
		conn.UpdatedAt,
		conn.LastConnectedAt,
		conn.UserID,
		conn.Status,
		sshEnabledInt,
		conn.SSHHost,
		conn.SSHPort,
		conn.SSHUsername,
		sshPassword,
		sshPrivateKey,
		s3CleanupInt,
	)

	return err
}

func (r *ConnectionRepository) GetConnection(id string) (*StoredConnection, error) {
	var conn StoredConnection
	var encryptedUsername, encryptedPassword string
	var encryptedSSHPassword, encryptedSSHPrivateKey sql.NullString
	var selectedDatabasesStr sql.NullString
	var sslInt, sshEnabledInt, s3CleanupInt int

	query := `SELECT 
		id, name, type, host, port, username, password, database_name, ssl, 
		database_size, created_at, updated_at, last_connected_at, user_id, status,
		ssh_enabled, ssh_host, ssh_port, ssh_username, ssh_password, ssh_private_key,
		COALESCE(selected_databases, '') as selected_databases,
		COALESCE(s3_cleanup_on_retention, 1) as s3_cleanup_on_retention
	FROM connections WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&conn.ID,
		&conn.Name,
		&conn.Type,
		&conn.Host,
		&conn.Port,
		&encryptedUsername,
		&encryptedPassword,
		&conn.DatabaseName,
		&sslInt,
		&conn.DatabaseSize,
		&conn.CreatedAt,
		&conn.UpdatedAt,
		&conn.LastConnectedAt,
		&conn.UserID,
		&conn.Status,
		&sshEnabledInt,
		&conn.SSHHost,
		&conn.SSHPort,
		&conn.SSHUsername,
		&encryptedSSHPassword,
		&encryptedSSHPrivateKey,
		&selectedDatabasesStr,
		&s3CleanupInt,
	)
	if err != nil {
		return nil, err
	}

	conn.SSL = sslInt != 0
	conn.SSHEnabled = sshEnabledInt != 0
	conn.S3CleanupOnRetention = s3CleanupInt != 0

	// Parse selected_databases from comma-separated string
	if selectedDatabasesStr.Valid && selectedDatabasesStr.String != "" {
		conn.SelectedDatabases = []string{}
		for i := 0; i < len(selectedDatabasesStr.String); {
			end := i
			for end < len(selectedDatabasesStr.String) && selectedDatabasesStr.String[end] != ',' {
				end++
			}
			if end > i {
				conn.SelectedDatabases = append(conn.SelectedDatabases, selectedDatabasesStr.String[i:end])
			}
			i = end + 1
		}
	}

	conn.Username, err = r.crypto.Decrypt(encryptedUsername)
	if err != nil {
		return nil, err
	}

	conn.Password, err = r.crypto.Decrypt(encryptedPassword)
	if err != nil {
		return nil, err
	}

	if encryptedSSHPassword.Valid && encryptedSSHPassword.String != "" {
		conn.SSHPassword, err = r.crypto.Decrypt(encryptedSSHPassword.String)
		if err != nil {
			return nil, err
		}
	}

	if encryptedSSHPrivateKey.Valid && encryptedSSHPrivateKey.String != "" {
		conn.SSHPrivateKey, err = r.crypto.Decrypt(encryptedSSHPrivateKey.String)
		if err != nil {
			return nil, err
		}
	}

	return &conn, nil
}

func (r *ConnectionRepository) Update(conn StoredConnection) error {
	username, err := r.crypto.Encrypt(conn.Username)
	if err != nil {
		return err
	}

	password, err := r.crypto.Encrypt(conn.Password)
	if err != nil {
		return err
	}

	sshPassword := ""
	if conn.SSHPassword != "" {
		sshPassword, err = r.crypto.Encrypt(conn.SSHPassword)
		if err != nil {
			return err
		}
	}

	sshPrivateKey := ""
	if conn.SSHPrivateKey != "" {
		sshPrivateKey, err = r.crypto.Encrypt(conn.SSHPrivateKey)
		if err != nil {
			return err
		}
	}

	sslInt := 0
	if conn.SSL {
		sslInt = 1
	}

	sshEnabledInt := 0
	if conn.SSHEnabled {
		sshEnabledInt = 1
	}

	s3CleanupInt := 0
	if conn.S3CleanupOnRetention {
		s3CleanupInt = 1
	}

	query := `
		UPDATE connections SET 
			name = $1, type = $2, host = $3, port = $4, 
			username = $5, password = $6, database_name = $7, 
			ssl = $8, ssh_enabled = $9, ssh_host = $10, ssh_port = $11,
			ssh_username = $12, ssh_password = $13, ssh_private_key = $14,
			database_size = $15, s3_cleanup_on_retention = $16, updated_at = CURRENT_TIMESTAMP
		WHERE id = $17`

	_, err = r.db.Exec(
		query,
		conn.Name,
		conn.Type,
		conn.Host,
		conn.Port,
		username,
		password,
		conn.DatabaseName,
		sslInt,
		sshEnabledInt,
		conn.SSHHost,
		conn.SSHPort,
		conn.SSHUsername,
		sshPassword,
		sshPrivateKey,
		conn.DatabaseSize,
		s3CleanupInt,
		conn.ID,
	)

	return err
}

func (r *ConnectionRepository) ListByUserID(userID uuid.UUID) ([]ConnectionListItem, error) {
	query := `
		SELECT 
			c.id,
			c.name,
			c.type,
			c.host,
			c.status,
			c.database_size,
			b.completed_time as last_backup_time,
			COALESCE(bs.enabled, false) as backup_enabled,
			bs.cron_schedule,
			bs.retention_days,
			COALESCE(c.s3_cleanup_on_retention, 1) as s3_cleanup_on_retention
		FROM connections c
		LEFT JOIN backup_schedules bs ON c.id = bs.connection_id AND bs.enabled = true
		LEFT JOIN backups b ON c.id = b.connection_id
			AND b.completed_time = (
				SELECT MAX(completed_time)
				FROM backups
				WHERE connection_id = c.id
			)
		WHERE c.user_id = $1
		GROUP BY c.id, c.name, c.type, c.host, c.status, c.database_size, b.completed_time, bs.enabled, bs.cron_schedule, bs.retention_days, c.s3_cleanup_on_retention
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []ConnectionListItem
	for rows.Next() {
		var conn ConnectionListItem
		var lastBackupTime sql.NullString
		var cronSchedule sql.NullString
		var retentionDays sql.NullInt64
		var s3CleanupInt int

		err := rows.Scan(
			&conn.ID,
			&conn.Name,
			&conn.Type,
			&conn.Host,
			&conn.Status,
			&conn.DatabaseSize,
			&lastBackupTime,
			&conn.BackupEnabled,
			&cronSchedule,
			&retentionDays,
			&s3CleanupInt,
		)
		if err != nil {
			return nil, err
		}

		if lastBackupTime.Valid {
			conn.LastBackupTime = &lastBackupTime.String
		}
		if cronSchedule.Valid {
			conn.CronSchedule = &cronSchedule.String
		}
		if retentionDays.Valid {
			days := int(retentionDays.Int64)
			conn.RetentionDays = &days
		}
		conn.S3CleanupOnRetention = s3CleanupInt != 0

		connections = append(connections, conn)
	}
	return connections, nil
}

func (r *ConnectionRepository) Delete(id string) error {
	query := `DELETE FROM connections WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *ConnectionRepository) UpdateSelectedDatabases(id string, databases []string) error {
	// Convert []string to comma-separated string for storage
	var dbString string
	if len(databases) > 0 {
		for i, db := range databases {
			if i > 0 {
				dbString += ","
			}
			dbString += db
		}
	}

	query := `UPDATE connections SET selected_databases = $1, updated_at = datetime('now') WHERE id = $2`
	_, err := r.db.Exec(query, dbString, id)
	return err
}
