package connection

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
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
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=%s",
		config.Username, config.Password, config.Host, config.Port, config.Database, sslMode)

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
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, sslMode)

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
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
		config.Username, config.Password, config.Host, config.Port, config.Database)

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
