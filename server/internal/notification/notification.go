package notification

import (
	"time"

	"github.com/samokw/ssl_tracker/internal/types"
)

type NotificationType string

const (
	NotificationTypeEmail   NotificationType = "email"
	NotificationTypeDiscord NotificationType = "discord"
	NotificationTypeSlack   NotificationType = "slack"
)

func NewNotificationType(nType string) NotificationType {
	return NotificationType(nType)
}

func (n NotificationType) String() string {
	return string(n)
}

type Notification struct {
	NotificationID   uint             `db:"id"`
	DomainID         types.DomainID   `db:"domain_id"`
	DaysBefore       int              `db:"days_before"`
	SentAt           time.Time        `db:"sent_at"`
	NotificationType NotificationType `db:"notification_type"`
}
