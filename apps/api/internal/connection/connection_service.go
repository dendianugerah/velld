package connection

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dendianugerah/velld/internal/d1"
	"github.com/google/uuid"
)

type ConnectionService struct {
	repo    *ConnectionRepository
	manager *ConnectionManager
}

func NewConnectionService(repo *ConnectionRepository, manager *ConnectionManager) *ConnectionService {
	return &ConnectionService{
		repo:    repo,
		manager: manager,
	}
}

func (s *ConnectionService) TestConnection(config ConnectionConfig) (bool, error) {
	if config.Type == "d1" {
		normalizeConnection(&config)
		_, err := s.checkConnection(config)
		return err == nil, err
	}
	if err := s.manager.Connect(config); err != nil {
		return false, err
	}
	defer s.manager.Disconnect(config.ID)
	return true, nil
}

func (s *ConnectionService) checkConnection(config ConnectionConfig) (int64, error) {
	if config.Type == "d1" {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return d1.NewClient().DatabaseSize(ctx, d1.Credentials{
			AccountID: config.D1AccountID, DatabaseID: config.D1DatabaseID, APIToken: config.D1APIToken,
		})
	}
	if err := s.manager.Connect(config); err != nil {
		return 0, err
	}
	defer s.manager.Disconnect(config.ID)
	size, _ := s.manager.GetDatabaseSize(config.ID)
	return size, nil
}

func normalizeConnection(config *ConnectionConfig) {
	if config.Type != "d1" {
		config.D1AccountID, config.D1DatabaseID, config.D1APIToken = "", "", ""
		return
	}
	config.D1AccountID = strings.TrimSpace(config.D1AccountID)
	config.D1DatabaseID = strings.TrimSpace(config.D1DatabaseID)
	config.D1APIToken = strings.TrimSpace(config.D1APIToken)
	config.Host, config.Port, config.SSL = d1.Endpoint, 443, true
	config.Database = config.D1DatabaseID
	config.Username, config.Password, config.MongoAuthSource = "", "", ""
	config.SSHEnabled, config.SSHPort = false, 0
	config.SSHHost, config.SSHUsername, config.SSHPassword, config.SSHPrivateKey = "", "", "", ""
}

func (s *ConnectionService) SaveConnection(config ConnectionConfig, userID uuid.UUID) (*StoredConnection, error) {
	if config.ID == "" {
		config.ID = uuid.New().String()
	}

	normalizeConnection(&config)
	dbSize, err := s.checkConnection(config)
	if err != nil {
		return nil, err
	}

	storedConn := StoredConnection{
		ID:              config.ID,
		Name:            config.Name,
		Type:            config.Type,
		Host:            config.Host,
		Port:            config.Port,
		Username:        config.Username,
		Password:        config.Password,
		DatabaseName:    config.Database,
		MongoAuthSource: config.MongoAuthSource,
		D1AccountID:     config.D1AccountID,
		D1DatabaseID:    config.D1DatabaseID,
		D1APIToken:      config.D1APIToken,
		SSL:             config.SSL,
		SSHEnabled:      config.SSHEnabled,
		SSHHost:         config.SSHHost,
		SSHPort:         config.SSHPort,
		SSHUsername:     config.SSHUsername,
		SSHPassword:     config.SSHPassword,
		SSHPrivateKey:   config.SSHPrivateKey,
		UserID:          userID,
		Status:          "connected",
		DatabaseSize:    dbSize,
	}

	if err := s.repo.Save(storedConn); err != nil {
		return nil, err
	}

	return &storedConn, nil
}

func (s *ConnectionService) ListConnections(userID uuid.UUID) ([]ConnectionListItem, error) {
	return s.repo.ListByUserID(userID)
}

func (s *ConnectionService) GetConnection(id string) (*StoredConnection, error) {
	return s.repo.GetConnection(id)
}

func (s *ConnectionService) UpdateConnection(config ConnectionConfig, userID uuid.UUID) (*StoredConnection, error) {
	// Get existing connection to preserve fields that aren't being updated
	existingConn, err := s.repo.GetConnection(config.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing connection: %w", err)
	}
	if existingConn.Type == "d1" || config.Type == "d1" {
		if existingConn.UserID != userID {
			return nil, fmt.Errorf("connection not found")
		}
		if existingConn.Type != config.Type {
			return nil, fmt.Errorf("cannot change the engine of a D1 connection")
		}
		if strings.TrimSpace(config.D1APIToken) == "" {
			config.D1APIToken = existingConn.D1APIToken
		}
	}
	normalizeConnection(&config)
	dbSize, err := s.checkConnection(config)
	if err != nil {
		return nil, err
	}

	storedConn := StoredConnection{
		ID:                   config.ID,
		Name:                 config.Name,
		Type:                 config.Type,
		Host:                 config.Host,
		Port:                 config.Port,
		Username:             config.Username,
		Password:             config.Password,
		DatabaseName:         config.Database,
		MongoAuthSource:      config.MongoAuthSource,
		D1AccountID:          config.D1AccountID,
		D1DatabaseID:         config.D1DatabaseID,
		D1APIToken:           config.D1APIToken,
		SSL:                  config.SSL,
		SSHEnabled:           config.SSHEnabled,
		SSHHost:              config.SSHHost,
		SSHPort:              config.SSHPort,
		SSHUsername:          config.SSHUsername,
		SSHPassword:          config.SSHPassword,
		SSHPrivateKey:        config.SSHPrivateKey,
		UserID:               userID,
		Status:               "connected",
		DatabaseSize:         dbSize,
		S3CleanupOnRetention: existingConn.S3CleanupOnRetention, // preserve existing value
	}

	// Update S3 cleanup setting if provided
	if config.S3CleanupOnRetention != nil {
		storedConn.S3CleanupOnRetention = *config.S3CleanupOnRetention
	}

	if err := s.repo.Update(storedConn); err != nil {
		return nil, err
	}

	return &storedConn, nil
}

// UpdateConnectionSettings updates connection settings without testing the connection
func (s *ConnectionService) UpdateConnectionSettings(id string, s3CleanupOnRetention *bool) error {
	existingConn, err := s.repo.GetConnection(id)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	if s3CleanupOnRetention != nil {
		existingConn.S3CleanupOnRetention = *s3CleanupOnRetention
	}

	return s.repo.Update(*existingConn)
}

func (s *ConnectionService) DeleteConnection(id string) error {
	return s.repo.Delete(id)
}

func (s *ConnectionService) DiscoverDatabases(id string) ([]string, error) {
	conn, err := s.repo.GetConnection(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	if conn.Type == "d1" {
		return []string{}, nil
	}
	config := ConnectionConfig{
		ID:              conn.ID,
		Type:            conn.Type,
		Host:            conn.Host,
		Port:            conn.Port,
		Username:        conn.Username,
		Password:        conn.Password,
		Database:        conn.DatabaseName,
		MongoAuthSource: conn.MongoAuthSource,
		SSL:             conn.SSL,
		SSHEnabled:      conn.SSHEnabled,
		SSHHost:         conn.SSHHost,
		SSHPort:         conn.SSHPort,
		SSHUsername:     conn.SSHUsername,
		SSHPassword:     conn.SSHPassword,
		SSHPrivateKey:   conn.SSHPrivateKey,
	}

	return s.manager.DiscoverDatabases(config)
}

func (s *ConnectionService) UpdateSelectedDatabases(id string, databases []string) error {
	return s.repo.UpdateSelectedDatabases(id, databases)
}
