package connection

import (
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
	err := s.manager.Connect(config)
	if err != nil {
		return false, err
	}
	defer s.manager.Disconnect(config.ID)
	return true, nil
}

func (s *ConnectionService) SaveConnection(config ConnectionConfig, userID uuid.UUID) (*StoredConnection, error) {
	if config.ID == "" {
		config.ID = uuid.New().String()
	}

	if err := s.manager.Connect(config); err != nil {
		return nil, err
	}
	defer s.manager.Disconnect(config.ID)

	dbSize, err := s.manager.GetDatabaseSize(config.ID)
	if err != nil {
		dbSize = 0 // Set to 0 if we can't get the size
	}

	storedConn := StoredConnection{
		ID:            config.ID,
		Name:          config.Name,
		Type:          config.Type,
		Host:          config.Host,
		Port:          config.Port,
		Username:      config.Username,
		Password:      config.Password,
		DatabaseName:  config.Database,
		SSL:           config.SSL,
		SSHEnabled:    config.SSHEnabled,
		SSHHost:       config.SSHHost,
		SSHPort:       config.SSHPort,
		SSHUsername:   config.SSHUsername,
		SSHPassword:   config.SSHPassword,
		SSHPrivateKey: config.SSHPrivateKey,
		UserID:        userID,
		Status:        "connected",
		DatabaseSize:  dbSize,
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
	if err := s.manager.Connect(config); err != nil {
		return nil, err
	}
	defer s.manager.Disconnect(config.ID)

	dbSize, err := s.manager.GetDatabaseSize(config.ID)
	if err != nil {
		dbSize = 0 // Set to 0 if we can't get the size
	}

	storedConn := StoredConnection{
		ID:            config.ID,
		Name:          config.Name,
		Type:          config.Type,
		Host:          config.Host,
		Port:          config.Port,
		Username:      config.Username,
		Password:      config.Password,
		DatabaseName:  config.Database,
		SSL:           config.SSL,
		SSHEnabled:    config.SSHEnabled,
		SSHHost:       config.SSHHost,
		SSHPort:       config.SSHPort,
		SSHUsername:   config.SSHUsername,
		SSHPassword:   config.SSHPassword,
		SSHPrivateKey: config.SSHPrivateKey,
		UserID:        userID,
		Status:        "connected",
		DatabaseSize:  dbSize,
	}

	if err := s.repo.Update(storedConn); err != nil {
		return nil, err
	}

	return &storedConn, nil
}

func (s *ConnectionService) UpdateConnectionStatus(id string, status string) error {
	return s.repo.UpdateStatus(id, status)
}

func (s *ConnectionService) DeleteConnection(id string) error {
	return s.repo.Delete(id)
}
