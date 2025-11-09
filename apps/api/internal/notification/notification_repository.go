package notification

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) CreateNotification(n *Notification) error {
	now := time.Now().Format(time.RFC3339)
	_, err := r.db.Exec(`
		INSERT INTO notifications (
			id, user_id, title, message, type, status, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		n.ID, n.UserID, n.Title, n.Message, n.Type, n.Status, n.Metadata, now, now)
	return err
}

func (r *NotificationRepository) GetUserNotifications(userID uuid.UUID) ([]*NotificationList, error) {
	query := `
        SELECT id, title, message, type, status, metadata, created_at
        FROM notifications
        WHERE user_id = $1
        AND (status = 'unread' OR created_at > datetime('now', '-7 days'))
        ORDER BY 
            CASE WHEN status = 'unread' THEN 0 ELSE 1 END,
            created_at DESC
        LIMIT 50`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*NotificationList
	for rows.Next() {
		n := &NotificationList{}
		err := rows.Scan(&n.ID, &n.Title, &n.Message, &n.Type, &n.Status, &n.Metadata, &n.CreatedAt)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

func (r *NotificationRepository) MarkAsRead(userID uuid.UUID, notificationID uuid.UUID) error {
	query := `
        UPDATE notifications 
        SET status = $1, updated_at = $2
        WHERE user_id = $3 AND id = $4`

	_, err := r.db.Exec(query, StatusRead, time.Now().Format(time.RFC3339), userID, notificationID)
	return err
}

func (r *NotificationRepository) DeleteNotifications(userID uuid.UUID, notificationIDs []uuid.UUID) error {
	_, err := r.db.Exec(
		"DELETE FROM notifications WHERE user_id = $1 AND id = ANY($2)",
		userID, notificationIDs)
	return err
}
