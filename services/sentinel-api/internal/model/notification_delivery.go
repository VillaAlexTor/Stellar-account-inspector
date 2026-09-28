package model

import "time"

type NotificationDelivery struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	SentinelAlertID uint       `gorm:"uniqueIndex:idx_delivery_alert_channel,priority:1;index;not null" json:"sentinelAlertId"`
	Channel         string     `gorm:"size:24;uniqueIndex:idx_delivery_alert_channel,priority:2;index;not null" json:"channel"`
	Status          string     `gorm:"size:16;index;not null" json:"status"`
	Attempts        int        `gorm:"not null;default:0" json:"attempts"`
	LastError       string     `gorm:"type:text" json:"lastError,omitempty"`
	NextAttemptAt   time.Time  `gorm:"index;not null" json:"nextAttemptAt"`
	SentAt          *time.Time `json:"sentAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type NotificationJob struct {
	DeliveryID uint
	Channel    string
	Attempts   int
	Alert      SentinelAlert
	PublicKey  string
	Network    string
}
