package notification

import (
	"time"
	"github.com/samokw/ssl_tracker/internal/types"
)

type NotificationID uint

type DaysBefore time.Time

type SentAt time.Time

type Notification struct {
	NotificationID NotificationID
	DomainID       types.DomainID
	DaysBefore     DaysBefore
	SentAt         SentAt
}
