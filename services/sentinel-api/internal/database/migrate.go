package database

import (
	"fmt"

	"github.com/stellar-account-inspector/sentinel-api/internal/model"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	migrator := db.Migrator()
	legacyIndexes := []struct {
		model any
		name  string
	}{
		{model: &model.SentinelAlert{}, name: "idx_alert_operation_rule"},
		{model: &model.RelevantOperation{}, name: "idx_relevant_operations_operation_id"},
	}
	for _, index := range legacyIndexes {
		if migrator.HasTable(index.model) && migrator.HasIndex(index.model, index.name) {
			if err := migrator.DropIndex(index.model, index.name); err != nil {
				return fmt.Errorf("retirar índice legado %s: %w", index.name, err)
			}
		}
	}
	if err := db.AutoMigrate(
		&model.MonitoredAccount{},
		&model.RelevantOperation{},
		&model.SentinelAlert{},
		&model.NotificationDelivery{},
	); err != nil {
		return fmt.Errorf("migrar esquema Sentinel: %w", err)
	}
	return nil
}
