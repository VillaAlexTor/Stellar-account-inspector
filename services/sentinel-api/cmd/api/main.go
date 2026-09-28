package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stellar-account-inspector/sentinel-api/internal/config"
	"github.com/stellar-account-inspector/sentinel-api/internal/database"
	"github.com/stellar-account-inspector/sentinel-api/internal/horizon"
	httpapi "github.com/stellar-account-inspector/sentinel-api/internal/http"
	"github.com/stellar-account-inspector/sentinel-api/internal/http/middleware"
	"github.com/stellar-account-inspector/sentinel-api/internal/notifications"
	"github.com/stellar-account-inspector/sentinel-api/internal/observability"
	"github.com/stellar-account-inspector/sentinel-api/internal/repository"
	"github.com/stellar-account-inspector/sentinel-api/internal/retention"
	"github.com/stellar-account-inspector/sentinel-api/internal/sentinel"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	configuration, err := config.Load()
	if err != nil {
		logger.Error("configuración inválida", "error", err)
		os.Exit(1)
	}
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: configuration.LogLevel})).With(
		"service", "stellar-sentinel-api",
		"version", configuration.Version,
		"environment", configuration.Environment,
	)

	db, err := database.Open(configuration.DatabaseURL)
	if err != nil {
		logger.Error("PostgreSQL no disponible", "error", err)
		os.Exit(1)
	}
	if err := database.Migrate(db); err != nil {
		logger.Error("migración fallida", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repo := repository.New(db)
	metrics := observability.NewMetrics()
	metrics.SetDatabaseReady(true)
	senders := make(map[string]notifications.Sender)
	if configuration.WebhookURL != "" {
		senders["webhook"] = notifications.NewWebhookSender(configuration.WebhookURL, configuration.WebhookSecret, configuration.NotificationTimeout)
	}
	if configuration.TelegramBotToken != "" {
		senders["telegram"] = notifications.NewTelegramSender(configuration.TelegramAPIURL, configuration.TelegramBotToken, configuration.TelegramChatID, configuration.NotificationTimeout)
	}
	if configuration.SMTPHost != "" {
		senders["email"] = notifications.NewSMTPSender(
			configuration.SMTPHost, configuration.SMTPPort, configuration.SMTPUsername, configuration.SMTPPassword,
			configuration.SMTPFrom, configuration.SMTPTo, configuration.SMTPTLS, configuration.NotificationTimeout,
		)
	}
	dispatcher := notifications.NewDispatcher(repo, senders, configuration.NotificationPoll, configuration.NotificationAttempts, logger, metrics)
	go dispatcher.Run(ctx)
	retentionWorker := retention.NewWorker(repo, configuration.AlertRetention, configuration.OperationRetention, configuration.RetentionInterval, logger)
	go retentionWorker.Run(ctx)
	horizonClient := horizon.NewClient(
		configuration.HorizonTestnetURL,
		configuration.HTTPTimeout,
	)
	manager := sentinel.NewManager(ctx, repo, horizonClient, configuration.ReconnectMaxBackoff, logger, metrics, dispatcher)

	accounts, err := repo.ListAccounts()
	if err != nil {
		logger.Error("no se pudieron restaurar las sesiones", "error", err)
		os.Exit(1)
	}
	for _, account := range accounts {
		manager.Start(account)
	}

	router := httpapi.NewRouter(repo, manager, logger, configuration.AllowedOrigins, metrics, httpapi.SecurityOptions{
		Auth: middleware.AuthOptions{
			Tokens:        configuration.AuthTokens,
			SessionSecret: configuration.SessionSecret,
			SessionTTL:    configuration.SessionTTL,
			SecureCookie:  configuration.SecureCookie,
		},
		RateLimit: middleware.RateLimitOptions{
			Window:            configuration.RateLimitWindow,
			IPLimit:           configuration.RateLimitIP,
			AccountLimit:      configuration.RateLimitAccount,
			TrustedProxyCIDRs: configuration.TrustedProxyCIDRs,
		},
	})

	server := &http.Server{
		Addr:              configuration.Address,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       75 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		logger.Info("Sentinel API escuchando", "address", configuration.Address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("servidor HTTP detenido", "error", err)
			cancel()
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	select {
	case <-stop:
	case <-ctx.Done():
	}
	cancel()

	shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("apagado incompleto", "error", err)
	}
}
