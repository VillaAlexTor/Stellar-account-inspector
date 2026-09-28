package model

import "time"

type RelevantOperation struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	MonitoredAccountID uint      `gorm:"index;uniqueIndex:idx_operation_account_operation,priority:1;not null" json:"monitoredAccountId"`
	OperationID        string    `gorm:"size:128;uniqueIndex:idx_operation_account_operation,priority:2;not null" json:"operationId"`
	OperationType      string    `gorm:"size:64;index;not null" json:"operationType"`
	Payload            string    `gorm:"type:jsonb;not null" json:"payload"`
	CreatedAt          time.Time `gorm:"index" json:"createdAt"`
}
