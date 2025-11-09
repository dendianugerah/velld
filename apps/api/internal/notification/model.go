package notification

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	BackupFailed    NotificationType = "backup_failed"
	BackupCompleted NotificationType = "backup_completed"
)

type NotificationStatus string

const (
	StatusUnread NotificationStatus = "unread"
	StatusRead   NotificationStatus = "read"
)

type Notification struct {
	ID        uuid.UUID          `json:"id"`
	UserID    uuid.UUID          `json:"user_id"`
	Title     string             `json:"title"`
	Message   string             `json:"message"`
	Type      NotificationType   `json:"type"`
	Status    NotificationStatus `json:"status"`
	Metadata  json.RawMessage    `json:"metadata,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type NotificationList struct {
	ID        uuid.UUID          `json:"id"`
	Title     string             `json:"title"`
	Message   string             `json:"message"`
	Type      NotificationType   `json:"type"`
	Status    NotificationStatus `json:"status"`
	Metadata  json.RawMessage    `json:"metadata,omitempty"`
	CreatedAt string             `json:"created_at"`
}

type NotificationListOptions struct {
	UserID  uuid.UUID
	Status  *NotificationStatus
	Type    *NotificationType
	Limit   int
	Offset  int
}
