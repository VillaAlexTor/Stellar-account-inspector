package sentinel

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/stellar-account-inspector/sentinel-api/internal/horizon"
	"github.com/stellar-account-inspector/sentinel-api/internal/model"
	"github.com/stellar-account-inspector/sentinel-api/internal/observability"
	"github.com/stellar-account-inspector/sentinel-api/internal/repository"
)

type Manager struct {
	ctx        context.Context
	repository *repository.Repository
	horizon    *horizon.Client
	maxBackoff time.Duration
	logger     *slog.Logger
	metrics    *observability.Metrics
	notifier   AlertNotifier
	mu         sync.Mutex
	sessions   map[string]*Session
}

type AlertNotifier interface {
	Channels() []string
	Wake()
}

func NewManager(
	ctx context.Context,
	repository *repository.Repository,
	horizonClient *horizon.Client,
	maxBackoff time.Duration,
	logger *slog.Logger,
	metrics *observability.Metrics,
	notifier AlertNotifier,
) *Manager {
	return &Manager{
		ctx:        ctx,
		repository: repository,
		horizon:    horizonClient,
		maxBackoff: maxBackoff,
		logger:     logger,
		metrics:    metrics,
		notifier:   notifier,
		sessions:   make(map[string]*Session),
	}
}

func (manager *Manager) Start(account model.MonitoredAccount) *Session {
	key := sessionKey(account.PublicKey, account.Network)
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if session, exists := manager.sessions[key]; exists {
		return session
	}
	session := newSession()
	manager.sessions[key] = session
	if manager.metrics != nil {
		manager.metrics.SetMonitoredSessions(len(manager.sessions))
	}
	go manager.run(account, session)
	return session
}

func (manager *Manager) Session(publicKey, network string) (*Session, bool) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	session, exists := manager.sessions[sessionKey(publicKey, network)]
	return session, exists
}

func (manager *Manager) run(account model.MonitoredAccount, session *Session) {
	state, cursor, err := manager.initialState(account)
	backoff := time.Second
	for err != nil {
		if manager.metrics != nil {
			manager.metrics.HorizonEvent("initialization_error")
		}
		session.SetStatus("down", err.Error())
		manager.logger.Error("no se pudo inicializar Sentinel", "account", account.PublicKey, "network", account.Network, "error", err)
		if !wait(manager.ctx, backoff) {
			return
		}
		backoff = nextBackoff(backoff, manager.maxBackoff)
		state, cursor, err = manager.initialState(account)
	}

	backoff = time.Second
	for {
		streamErr := manager.horizon.StreamOperations(
			manager.ctx,
			account.PublicKey,
			account.Network,
			cursor,
			func() {
				if manager.metrics != nil {
					manager.metrics.HorizonEvent("connected")
				}
				session.SetStatus("connected", "Stream activo con Horizon")
				backoff = time.Second
			},
			func(operation horizon.Operation) error {
				if manager.metrics != nil {
					manager.metrics.HorizonEvent("operation")
				}
				alerts := Evaluate(operation, state)
				nextState := state.Clone()
				nextState.Apply(operation)
				stateJSON, marshalErr := nextState.JSON()
				if marshalErr != nil {
					return fmt.Errorf("serializar estado Sentinel: %w", marshalErr)
				}

				createdAt := parseOperationTime(operation.CreatedAt)
				persistedAlerts := make([]model.SentinelAlert, 0, len(alerts))
				for _, alert := range alerts {
					persistedAlerts = append(persistedAlerts, model.SentinelAlert{
						MonitoredAccountID: account.ID,
						RuleID:             alert.RuleID,
						Severity:           alert.Severity,
						Message:            alert.Message,
						OperationID:        operation.ID,
						CreatedAt:          createdAt,
					})
				}
				payload := operation.Raw
				if len(payload) == 0 {
					payload, _ = json.Marshal(operation)
				}
				relevantOperation := model.RelevantOperation{
					MonitoredAccountID: account.ID,
					OperationID:        operation.ID,
					OperationType:      operation.Type,
					Payload:            string(payload),
					CreatedAt:          createdAt,
				}
				notificationChannels := []string(nil)
				if manager.notifier != nil {
					notificationChannels = manager.notifier.Channels()
				}
				createdAlerts, recordErr := manager.repository.RecordEvaluation(
					account.ID,
					relevantOperation,
					persistedAlerts,
					notificationChannels,
					operation.PagingToken,
					stateJSON,
				)
				if recordErr != nil {
					return recordErr
				}
				if manager.notifier != nil && len(createdAlerts) > 0 {
					manager.notifier.Wake()
				}
				state = nextState
				cursor = operation.PagingToken
				for _, alert := range createdAlerts {
					if manager.metrics != nil {
						manager.metrics.Alert(alert.RuleID, alert.Severity)
					}
					session.PublishAlert(alert)
				}
				return nil
			},
		)

		if manager.ctx.Err() != nil {
			return
		}
		if streamErr != nil {
			if manager.metrics != nil {
				manager.metrics.HorizonEvent("stream_error")
			}
			manager.logger.Warn("stream Horizon interrumpido", "account", account.PublicKey, "network", account.Network, "error", streamErr)
		}
		session.SetStatus("reconnecting", "Reconectando con Horizon")
		if !wait(manager.ctx, backoff) {
			return
		}
		backoff = nextBackoff(backoff, manager.maxBackoff)
	}
}

func (manager *Manager) initialState(account model.MonitoredAccount) (AccountState, string, error) {
	if account.StateJSON != "" && account.LastCursor != "" {
		state, err := StateFromJSON(account.StateJSON)
		if err == nil {
			return state, account.LastCursor, nil
		}
		manager.logger.Warn("checkpoint inválido; se reconstruirá desde Horizon", "account", account.PublicKey, "error", err)
	}

	snapshot, err := manager.horizon.Account(manager.ctx, account.PublicKey, account.Network)
	if err != nil {
		return AccountState{}, "", err
	}
	state := StateFromSnapshot(snapshot)
	stateJSON, err := state.JSON()
	if err != nil {
		return AccountState{}, "", err
	}
	if err := manager.repository.SaveCheckpoint(account.ID, "now", stateJSON); err != nil {
		return AccountState{}, "", err
	}
	return state, "now", nil
}

func sessionKey(publicKey, network string) string {
	return network + ":" + publicKey
}

func nextBackoff(current, maximum time.Duration) time.Duration {
	next := current * 2
	if next > maximum {
		return maximum
	}
	return next
}

func wait(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func parseOperationTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Now().UTC()
	}
	return parsed.UTC()
}
