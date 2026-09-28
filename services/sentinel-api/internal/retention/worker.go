package retention

import (
	"context"
	"log/slog"
	"time"

	"github.com/stellar-account-inspector/sentinel-api/internal/repository"
)

type Worker struct {
	repository         *repository.Repository
	alertRetention     time.Duration
	operationRetention time.Duration
	interval           time.Duration
	logger             *slog.Logger
	now                func() time.Time
}

func NewWorker(repo *repository.Repository, alertRetention, operationRetention, interval time.Duration, logger *slog.Logger) *Worker {
	return &Worker{
		repository:         repo,
		alertRetention:     alertRetention,
		operationRetention: operationRetention,
		interval:           interval,
		logger:             logger,
		now:                time.Now,
	}
}

func (worker *Worker) Run(ctx context.Context) {
	worker.prune(ctx)
	ticker := time.NewTicker(worker.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			worker.prune(ctx)
		}
	}
}

func (worker *Worker) prune(ctx context.Context) {
	now := worker.now().UTC()
	result, err := worker.repository.Prune(ctx, now.Add(-worker.alertRetention), now.Add(-worker.operationRetention))
	if err != nil {
		worker.logger.ErrorContext(ctx, "retention_failed", "error", err)
		return
	}
	worker.logger.InfoContext(ctx, "retention_completed",
		"alerts_deleted", result.Alerts,
		"operations_deleted", result.Operations,
		"deliveries_deleted", result.Deliveries,
	)
}
