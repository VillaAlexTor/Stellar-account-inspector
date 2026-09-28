package model

import "time"

type MonitoredAccount struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PublicKey  string    `gorm:"size:56;uniqueIndex:idx_account_network,priority:1;not null" json:"publicKey"`
	Network    string    `gorm:"size:16;uniqueIndex:idx_account_network,priority:2;not null" json:"network"`
	LastCursor string    `gorm:"size:128" json:"lastCursor"`
	StateJSON  string    `gorm:"type:text" json:"-"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
