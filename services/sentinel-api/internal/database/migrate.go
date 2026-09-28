package database

import (
	"fmt"

	"github.com/stellar-account-inspector/sentinel-api/internal/model"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.MonitoredAccount{},
		&model.RelevantOperation{},
		&model.SentinelAlert{},
	); err != nil {
		return fmt.Errorf("migrar esquema Sentinel: %w", err)
	}
	return nil
}
