package connection

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConnectionManager struct {
	connections map[string]interface{}
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]interface{}),
	}
}

func (cm *ConnectionManager) Connect(config ConnectionConfig) error {
	if config.SSHEnabled {
		return cm.connectWithSSH(config)
	}

	switch config.Type {
	case "mysql":
		return cm.connectMySQL(config)
	case "postgresql":
		return cm.connectPostgres(config)
	case "mongodb":
		return cm.connectMongoDB(config)
	case "redis":
		return cm.connectRedis(config)
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

func (cm *ConnectionManager) connectWithSSH(config ConnectionConfig) error {
	tunnel, err := NewSSHTunnel(
		config.SSHHost,
		config.SSHPort,
		config.SSHUsername,
		config.SSHPassword,
		config.SSHPrivateKey,
		config.Host,
		config.Port,
	)
	if err != nil {
		return fmt.Errorf("failed to create SSH tunnel: %w", err)
	}

	if err := tunnel.Start(); err != nil {
		return fmt.Errorf("failed to start SSH tunnel: %w", err)
	}

	tunnelConfig := config
	tunnelConfig.Host = "127.0.0.1"
	tunnelConfig.Port = tunnel.GetLocalPort()

	var connErr error
	switch config.Type {
	case "mysql":
		connErr = cm.connectMySQL(tunnelConfig)
	case "postgresql":
		connErr = cm.connectPostgres(tunnelConfig)
	case "mongodb":
		connErr = cm.connectMongoDB(tunnelConfig)
	case "redis":
		connErr = cm.connectRedis(tunnelConfig)
	default:
		tunnel.Stop()
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	if connErr != nil {
		tunnel.Stop()
		return connErr
	}

	// Store tunnel reference (we'll need to close it later)
	// For now, we'll let it clean up when the connection is closed
	return nil
}

func (cm *ConnectionManager) connectMySQL(config ConnectionConfig) error {
	sslMode := "false"
	if config.SSL {
		sslMode = "true"
	}

	// Use default database if not specified
	database := config.Database
	if database == "" {
		database = "information_schema"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=%s",
		config.Username, config.Password, config.Host, config.Port, database, sslMode)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	cm.connections[config.ID] = db
	return nil
}

func (cm *ConnectionManager) connectPostgres(config ConnectionConfig) error {
	sslMode := "disable"
	if config.SSL {
		sslMode = "require"
	}

	// Use default database if not specified
	database := config.Database
	if database == "" {
		database = "postgres"
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, database, sslMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	cm.connections[config.ID] = db
	return nil
}

func (cm *ConnectionManager) connectMongoDB(config ConnectionConfig) error {
	ctx := context.Background()

	// Use default database if not specified
	database := config.Database
	if database == "" {
		database = "admin"
	}

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
		config.Username, config.Password, config.Host, config.Port, database)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	if err = client.Ping(ctx, nil); err != nil {
		return err
	}

	cm.connections[config.ID] = client
	return nil
}

func (cm *ConnectionManager) connectRedis(config ConnectionConfig) error {
	ctx := context.Background()

	opts := &redis.Options{
		Addr: fmt.Sprintf("%s:%d", config.Host, config.Port),
	}

	if config.SSL {
		opts.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if config.Password != "" {
		opts.Password = config.Password
	}

	if config.Database != "" {
		var db int
		_, err := fmt.Sscanf(config.Database, "%d", &db)
		if err == nil && db >= 0 && db <= 15 {
			opts.DB = db
		}
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	cm.connections[config.ID] = client
	return nil
}

func (cm *ConnectionManager) Disconnect(id string) error {
	conn, exists := cm.connections[id]
	if !exists {
		return fmt.Errorf("connection not found: %s", id)
	}

	switch c := conn.(type) {
	case *sql.DB:
		return c.Close()
	case *mongo.Client:
		return c.Disconnect(context.Background())
	case *redis.Client:
		return c.Close()
	default:
		return fmt.Errorf("unknown connection type for id: %s", id)
	}
}

func (cm *ConnectionManager) GetDatabaseSize(id string) (int64, error) {
	conn, exists := cm.connections[id]
	if !exists {
		return 0, fmt.Errorf("connection not found: %s", id)
	}

	switch c := conn.(type) {
	case *sql.DB:
		return cm.getSQLDatabaseSize(c)
	case *mongo.Client:
		return cm.getMongoDBSize(c)
	case *redis.Client:
		return cm.getRedisSize(c)
	default:
		return 0, fmt.Errorf("unknown connection type for id: %s", id)
	}
}

func (cm *ConnectionManager) getSQLDatabaseSize(db *sql.DB) (int64, error) {
	var query string

	switch db.Driver().(type) {
	case *pq.Driver:
		query = "SELECT pg_database_size(current_database())"
	case *mysql.MySQLDriver:
		query = `SELECT SUM(data_length + index_length) 
				 FROM information_schema.tables 
				 WHERE table_schema = DATABASE()`
	case *sqlite3.SQLiteDriver:
		query = "SELECT page_count * page_size as size FROM pragma_page_count, pragma_page_size"
	default:
		return 0, fmt.Errorf("unsupported database type for size calculation")
	}

	var size int64
	err := db.QueryRow(query).Scan(&size)
	return size, err
}

func (cm *ConnectionManager) getMongoDBSize(client *mongo.Client) (int64, error) {
	ctx := context.Background()
	result := client.Database("admin").RunCommand(ctx, bson.D{
		{Key: "dbStats", Value: 1},
		{Key: "scale", Value: 1},
	})

	var stats bson.M
	if err := result.Decode(&stats); err != nil {
		return 0, err
	}

	return int64(stats["dataSize"].(float64)), nil
}

func (cm *ConnectionManager) getRedisSize(client *redis.Client) (int64, error) {
	ctx := context.Background()

	info, err := client.Info(ctx, "memory").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get Redis memory info: %w", err)
	}

	var usedMemory int64
	lines := []byte(info)
	start := 0
	for i := 0; i < len(lines); i++ {
		if lines[i] == '\n' {
			line := string(lines[start:i])
			start = i + 1

			if len(line) > 12 && line[:12] == "used_memory:" {
				fmt.Sscanf(line[12:], "%d", &usedMemory)
				return usedMemory, nil
			}
		}
	}

	return 0, nil
}

func (cm *ConnectionManager) DiscoverDatabases(config ConnectionConfig) ([]string, error) {
	tempConfig := config
	tempConfig.ID = "temp_discovery_" + config.ID

	if err := cm.Connect(tempConfig); err != nil {
		return nil, fmt.Errorf("failed to connect for discovery: %w", err)
	}
	defer cm.Disconnect(tempConfig.ID)

	conn, exists := cm.connections[tempConfig.ID]
	if !exists {
		return nil, fmt.Errorf("connection not found after connecting")
	}

	var databases []string
	var err error

	switch config.Type {
	case "postgresql":
		databases, err = cm.discoverPostgresDatabases(conn.(*sql.DB))
	case "mysql", "mariadb":
		databases, err = cm.discoverMySQLDatabases(conn.(*sql.DB))
	case "mongodb":
		databases, err = cm.discoverMongoDBDatabases(conn.(*mongo.Client))
	case "redis":
		// Redis doesn't have multiple databases in the traditional sense
		// Return the 16 default database numbers
		databases = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"}
	default:
		return nil, fmt.Errorf("unsupported database type for discovery: %s", config.Type)
	}

	return databases, err
}

func (cm *ConnectionManager) discoverPostgresDatabases(db *sql.DB) ([]string, error) {
	query := `
		SELECT datname 
		FROM pg_database 
		WHERE datistemplate = false 
		AND datname NOT IN ('postgres')
		ORDER BY datname
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, err
		}
		databases = append(databases, dbName)
	}

	return databases, rows.Err()
}

func (cm *ConnectionManager) discoverMySQLDatabases(db *sql.DB) ([]string, error) {
	query := `
		SELECT SCHEMA_NAME 
		FROM information_schema.SCHEMATA 
		WHERE SCHEMA_NAME NOT IN ('information_schema', 'mysql', 'performance_schema', 'sys')
		ORDER BY SCHEMA_NAME
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, err
		}
		databases = append(databases, dbName)
	}

	return databases, rows.Err()
}

func (cm *ConnectionManager) discoverMongoDBDatabases(client *mongo.Client) ([]string, error) {
	ctx := context.Background()

	databases, err := client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	// Filter out system databases
	var filtered []string
	for _, db := range databases {
		if db != "admin" && db != "local" && db != "config" {
			filtered = append(filtered, db)
		}
	}

	return filtered, nil
}
