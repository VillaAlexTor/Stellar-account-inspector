package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/stellar-account-inspector/sentinel-api/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNotFound = errors.New("registro no encontrado")

type Repository struct {
	db *gorm.DB
}

type AlertPage struct {
	Alerts     []model.SentinelAlert
	NextCursor uint
}

type RetentionResult struct {
	Alerts     int64
	Operations int64
	Deliveries int64
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Ready(ctx context.Context) error {
	database, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("obtener conexión PostgreSQL: %w", err)
	}
	if err := database.PingContext(ctx); err != nil {
		return fmt.Errorf("comprobar PostgreSQL: %w", err)
	}
	return nil
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
	page, err := r.ListAlertsPage(accountID, limit, 0)
	return page.Alerts, err
}

func (r *Repository) ListAlertsPage(accountID uint, limit int, beforeID uint) (AlertPage, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var alerts []model.SentinelAlert
	query := r.db.Where("monitored_account_id = ?", accountID)
	if beforeID > 0 {
		query = query.Where("id < ?", beforeID)
	}
	if err := query.Order("id DESC").Limit(limit + 1).Find(&alerts).Error; err != nil {
		return AlertPage{}, fmt.Errorf("listar alertas: %w", err)
	}
	page := AlertPage{Alerts: alerts}
	if len(alerts) > limit {
		page.NextCursor = alerts[limit-1].ID
		page.Alerts = alerts[:limit]
	}
	return page, nil
}

func (r *Repository) Prune(ctx context.Context, alertsBefore, operationsBefore time.Time) (RetentionResult, error) {
	result := RetentionResult{}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		alertIDs := tx.Model(&model.SentinelAlert{}).Select("id").Where("created_at < ?", alertsBefore)
		deletedDeliveries := tx.Where("sentinel_alert_id IN (?)", alertIDs).Delete(&model.NotificationDelivery{})
		if deletedDeliveries.Error != nil {
			return fmt.Errorf("eliminar entregas vencidas: %w", deletedDeliveries.Error)
		}
		result.Deliveries = deletedDeliveries.RowsAffected

		deletedAlerts := tx.Where("created_at < ?", alertsBefore).Delete(&model.SentinelAlert{})
		if deletedAlerts.Error != nil {
			return fmt.Errorf("eliminar alertas vencidas: %w", deletedAlerts.Error)
		}
		result.Alerts = deletedAlerts.RowsAffected

		deletedOperations := tx.Where("created_at < ?", operationsBefore).Delete(&model.RelevantOperation{})
		if deletedOperations.Error != nil {
			return fmt.Errorf("eliminar operaciones vencidas: %w", deletedOperations.Error)
		}
		result.Operations = deletedOperations.RowsAffected
		return nil
	})
	return result, err
}

func (r *Repository) EnqueueNotifications(alertID uint, channels []string, now time.Time) error {
	deliveries := make([]model.NotificationDelivery, 0, len(channels))
	for _, channel := range channels {
		deliveries = append(deliveries, model.NotificationDelivery{
			SentinelAlertID: alertID,
			Channel:         channel,
			Status:          "pending",
			NextAttemptAt:   now.UTC(),
		})
	}
	if len(deliveries) == 0 {
		return nil
	}
	if err := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&deliveries).Error; err != nil {
		return fmt.Errorf("crear entregas de notificación: %w", err)
	}
	return nil
}

func (r *Repository) ClaimNotificationJobs(now time.Time, limit int) ([]model.NotificationJob, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	jobs := make([]model.NotificationJob, 0, limit)
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.NotificationDelivery{}).
			Where("status = ? AND updated_at < ?", "sending", now.Add(-5*time.Minute)).
			Updates(map[string]any{"status": "pending", "next_attempt_at": now}).Error; err != nil {
			return fmt.Errorf("recuperar entregas bloqueadas: %w", err)
		}

		var deliveries []model.NotificationDelivery
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ? AND next_attempt_at <= ?", "pending", now).
			Order("next_attempt_at ASC, id ASC").Limit(limit).Find(&deliveries).Error; err != nil {
			return fmt.Errorf("reclamar entregas pendientes: %w", err)
		}
		if len(deliveries) == 0 {
			return nil
		}
		ids := make([]uint, 0, len(deliveries))
		for _, delivery := range deliveries {
			ids = append(ids, delivery.ID)
		}
		if err := tx.Model(&model.NotificationDelivery{}).Where("id IN ?", ids).
			Updates(map[string]any{"status": "sending", "attempts": gorm.Expr("attempts + 1")}).Error; err != nil {
			return fmt.Errorf("marcar entregas en proceso: %w", err)
		}

		type row struct {
			DeliveryID     uint
			Channel        string
			Attempts       int
			AlertID        uint
			AccountID      uint
			RuleID         string
			Severity       string
			Message        string
			OperationID    string
			AlertCreatedAt time.Time
			PublicKey      string
			Network        string
		}
		var rows []row
		if err := tx.Table("notification_deliveries AS nd").
			Select("nd.id AS delivery_id, nd.channel, nd.attempts, sa.id AS alert_id, sa.monitored_account_id AS account_id, sa.rule_id, sa.severity, sa.message, sa.operation_id, sa.created_at AS alert_created_at, ma.public_key, ma.network").
			Joins("JOIN sentinel_alerts AS sa ON sa.id = nd.sentinel_alert_id").
			Joins("JOIN monitored_accounts AS ma ON ma.id = sa.monitored_account_id").
			Where("nd.id IN ?", ids).Scan(&rows).Error; err != nil {
			return fmt.Errorf("cargar entregas reclamadas: %w", err)
		}
		for _, item := range rows {
			jobs = append(jobs, model.NotificationJob{
				DeliveryID: item.DeliveryID,
				Channel:    item.Channel,
				Attempts:   item.Attempts,
				PublicKey:  item.PublicKey,
				Network:    item.Network,
				Alert:      model.SentinelAlert{ID: item.AlertID, MonitoredAccountID: item.AccountID, RuleID: item.RuleID, Severity: item.Severity, Message: item.Message, OperationID: item.OperationID, CreatedAt: item.AlertCreatedAt},
			})
		}
		return nil
	})
	return jobs, err
}

func (r *Repository) MarkNotificationSent(deliveryID uint, now time.Time) error {
	if err := r.db.Model(&model.NotificationDelivery{}).Where("id = ?", deliveryID).
		Updates(map[string]any{"status": "sent", "sent_at": now.UTC(), "last_error": ""}).Error; err != nil {
		return fmt.Errorf("marcar notificación enviada: %w", err)
	}
	return nil
}

func (r *Repository) MarkNotificationFailed(deliveryID uint, terminal bool, nextAttempt time.Time, message string) error {
	status := "pending"
	if terminal {
		status = "failed"
	}
	if len(message) > 2000 {
		message = message[:2000]
	}
	if err := r.db.Model(&model.NotificationDelivery{}).Where("id = ?", deliveryID).
		Updates(map[string]any{"status": status, "next_attempt_at": nextAttempt.UTC(), "last_error": message}).Error; err != nil {
		return fmt.Errorf("registrar fallo de notificación: %w", err)
	}
	return nil
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
	notificationChannels []string,
	cursor string,
	stateJSON string,
) ([]model.SentinelAlert, error) {
	createdAlerts := make([]model.SentinelAlert, 0, len(alerts))
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if len(alerts) > 0 {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&operation).Error; err != nil {
				return fmt.Errorf("persistir operación relevante: %w", err)
			}
			for index := range alerts {
				created := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&alerts[index])
				if created.Error != nil {
					return fmt.Errorf("persistir alerta: %w", created.Error)
				}
				if created.RowsAffected == 0 {
					continue
				}
				createdAlerts = append(createdAlerts, alerts[index])
				for _, channel := range notificationChannels {
					delivery := model.NotificationDelivery{
						SentinelAlertID: alerts[index].ID,
						Channel:         channel,
						Status:          "pending",
						NextAttemptAt:   time.Now().UTC(),
					}
					if err := tx.Create(&delivery).Error; err != nil {
						return fmt.Errorf("crear entrega de notificación: %w", err)
					}
				}
			}
		}

		if err := tx.Model(&model.MonitoredAccount{}).Where("id = ?", accountID).
			Updates(map[string]any{"last_cursor": cursor, "state_json": stateJSON}).Error; err != nil {
			return fmt.Errorf("actualizar cursor: %w", err)
		}
		return nil
	})
	return createdAlerts, err
}
