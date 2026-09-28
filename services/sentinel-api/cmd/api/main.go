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
	"github.com/stellar-account-inspector/sentinel-api/internal/repository"
	"github.com/stellar-account-inspector/sentinel-api/internal/sentinel"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	configuration, err := config.Load()
	if err != nil {
		logger.Error("configuración inválida", "error", err)
		os.Exit(1)
	}

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
	horizonClient := horizon.NewClient(
		configuration.HorizonTestnetURL,
		configuration.HorizonMainnetURL,
		configuration.HTTPTimeout,
	)
	manager := sentinel.NewManager(ctx, repo, horizonClient, configuration.ReconnectMaxBackoff, logger)

	accounts, err := repo.ListAccounts()
	if err != nil {
		logger.Error("no se pudieron restaurar las sesiones", "error", err)
		os.Exit(1)
	}
	for _, account := range accounts {
		manager.Start(account)
	}

	server := &http.Server{
		Addr:              configuration.Address,
		Handler:           httpapi.NewRouter(repo, manager, logger, configuration.AllowedOrigins),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       75 * time.Second,
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
