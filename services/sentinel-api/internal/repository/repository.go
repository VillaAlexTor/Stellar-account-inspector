package repository

import (
	"errors"
	"fmt"

	"github.com/stellar-account-inspector/sentinel-api/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNotFound = errors.New("registro no encontrado")

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetOrCreateAccount(publicKey, network string) (model.MonitoredAccount, error) {
	account := model.MonitoredAccount{PublicKey: publicKey, Network: network}
	result := r.db.Where("public_key = ? AND network = ?", publicKey, network).FirstOrCreate(&account)
	if result.Error != nil {
		return model.MonitoredAccount{}, fmt.Errorf("obtener cuenta monitoreada: %w", result.Error)
	}
	return account, nil
}

func (r *Repository) FindAccount(publicKey, network string) (model.MonitoredAccount, error) {
	var account model.MonitoredAccount
	result := r.db.Where("public_key = ? AND network = ?", publicKey, network).First(&account)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return model.MonitoredAccount{}, ErrNotFound
	}
	if result.Error != nil {
		return model.MonitoredAccount{}, fmt.Errorf("buscar cuenta monitoreada: %w", result.Error)
	}
	return account, nil
}

func (r *Repository) ListAccounts() ([]model.MonitoredAccount, error) {
	var accounts []model.MonitoredAccount
	if err := r.db.Order("id ASC").Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("listar cuentas monitoreadas: %w", err)
	}
	return accounts, nil
}

func (r *Repository) ListAlerts(accountID uint, limit int) ([]model.SentinelAlert, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var alerts []model.SentinelAlert
	if err := r.db.Where("monitored_account_id = ?", accountID).
		Order("created_at DESC, id DESC").Limit(limit).Find(&alerts).Error; err != nil {
		return nil, fmt.Errorf("listar alertas: %w", err)
	}
	return alerts, nil
}

func (r *Repository) SaveCheckpoint(accountID uint, cursor, stateJSON string) error {
	if err := r.db.Model(&model.MonitoredAccount{}).Where("id = ?", accountID).
		Updates(map[string]any{"last_cursor": cursor, "state_json": stateJSON}).Error; err != nil {
		return fmt.Errorf("guardar checkpoint: %w", err)
	}
	return nil
}

func (r *Repository) RecordEvaluation(
	accountID uint,
	operation model.RelevantOperation,
	alerts []model.SentinelAlert,
	cursor string,
	stateJSON string,
) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if len(alerts) > 0 {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&operation).Error; err != nil {
				return fmt.Errorf("persistir operación relevante: %w", err)
			}
			for index := range alerts {
				if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&alerts[index]).Error; err != nil {
					return fmt.Errorf("persistir alerta: %w", err)
				}
			}
		}

		if err := tx.Model(&model.MonitoredAccount{}).Where("id = ?", accountID).
			Updates(map[string]any{"last_cursor": cursor, "state_json": stateJSON}).Error; err != nil {
			return fmt.Errorf("actualizar cursor: %w", err)
		}
		return nil
	})
}
