package model

import "time"

type RelevantOperation struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	MonitoredAccountID uint      `gorm:"index;not null" json:"monitoredAccountId"`
	OperationID        string    `gorm:"size:128;uniqueIndex;not null" json:"operationId"`
	OperationType      string    `gorm:"size:64;index;not null" json:"operationType"`
	Payload            string    `gorm:"type:jsonb;not null" json:"payload"`
	CreatedAt          time.Time `gorm:"index" json:"createdAt"`
}
