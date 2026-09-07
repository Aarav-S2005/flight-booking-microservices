package database

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	NotificationID uuid.UUID `db:"notification_id" json:"notification_id"`
	RecipientEmail string    `db:"recipient_email" json:"recipient_email"`
	Subject        string    `db:"subject" json:"subject"`
	Body           string    `db:"body" json:"body"`
	SentAt         time.Time `db:"sent_at" json:"sent_at"`
	Success        bool      `db:"success" json:"success"`
}
