package model

import "time"

type SentinelAlert struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	MonitoredAccountID uint      `gorm:"index;uniqueIndex:idx_alert_account_operation_rule,priority:1;not null" json:"monitoredAccountId"`
	RuleID             string    `gorm:"size:64;uniqueIndex:idx_alert_account_operation_rule,priority:3;not null" json:"ruleId"`
	Severity           string    `gorm:"size:16;index;not null" json:"severity"`
	Message            string    `gorm:"type:text;not null" json:"message"`
	OperationID        string    `gorm:"size:128;uniqueIndex:idx_alert_account_operation_rule,priority:2;not null" json:"operationId"`
	CreatedAt          time.Time `gorm:"index" json:"createdAt"`
}
