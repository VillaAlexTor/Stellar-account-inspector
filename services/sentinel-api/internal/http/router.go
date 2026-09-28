package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/stellar-account-inspector/sentinel-api/internal/http/middleware"
	"github.com/stellar-account-inspector/sentinel-api/internal/model"
	"github.com/stellar-account-inspector/sentinel-api/internal/observability"
	"github.com/stellar-account-inspector/sentinel-api/internal/repository"
	"github.com/stellar-account-inspector/sentinel-api/internal/sentinel"
)

var stellarPublicKeyPattern = regexp.MustCompile(`^G[A-Z2-7]{55}$`)

type API struct {
	repository *repository.Repository
	manager    *sentinel.Manager
	logger     *slog.Logger
	limiter    *middleware.RateLimiter
	metrics    *observability.Metrics
}

type SecurityOptions struct {
	Auth      middleware.AuthOptions
	RateLimit middleware.RateLimitOptions
}

type monitorRequest struct {
	PublicKey string `json:"publicKey"`
	Network   string `json:"network"`
}

func NewRouter(
	repository *repository.Repository,
	manager *sentinel.Manager,
	logger *slog.Logger,
	allowedOrigins []string,
	metrics *observability.Metrics,
	security SecurityOptions,
) http.Handler {
	if metrics == nil {
		metrics = observability.NewMetrics()
	}
	security.Auth.Metrics = metrics
	security.RateLimit.Metrics = metrics
	authenticator := middleware.NewAuthenticator(security.Auth)
	limiter := middleware.NewRateLimiter(security.RateLimit)
	api := &API{repository: repository, manager: manager, logger: logger, limiter: limiter, metrics: metrics}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", api.health)
	mux.HandleFunc("GET /readyz", api.ready)
	mux.Handle("GET /metrics", metrics.Handler())
	mux.HandleFunc("GET /api/v1/auth/session", authenticator.Session)
	mux.HandleFunc("POST /api/v1/auth/session", authenticator.Session)
	mux.HandleFunc("DELETE /api/v1/auth/session", authenticator.Session)
	mux.HandleFunc("POST /api/v1/monitored-accounts", api.monitorAccount)
	mux.HandleFunc("GET /api/v1/monitored-accounts/{publicKey}", api.getAccount)
	mux.HandleFunc("GET /api/v1/monitored-accounts/{publicKey}/alerts", api.listAlerts)
	mux.HandleFunc("GET /api/v1/monitored-accounts/{publicKey}/events", api.streamEvents)
	secured := limiter.Handler(authenticator.Require(mux))
	withCORS := middleware.CORS(allowedOrigins, secured)
	return middleware.RequestID(middleware.ObserveHTTP(logger, metrics, withCORS))
}

func (api *API) ready(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), 2*time.Second)
	defer cancel()
	if err := api.repository.Ready(ctx); err != nil {
		api.metrics.SetDatabaseReady(false)
		api.logger.ErrorContext(request.Context(), "readiness_failed",
			"request_id", middleware.RequestIDFromContext(request.Context()),
			"error", err,
		)
		writeError(writer, http.StatusServiceUnavailable, "PostgreSQL no está disponible.")
		return
	}
	api.metrics.SetDatabaseReady(true)
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ready"})
}

func (api *API) health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "stellar-sentinel-api",
		"time":    time.Now().UTC(),
	})
}

func (api *API) monitorAccount(writer http.ResponseWriter, request *http.Request) {
	var input monitorRequest
	decoder := json.NewDecoder(http.MaxBytesReader(writer, request.Body, 32*1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(writer, http.StatusBadRequest, "El cuerpo JSON no es válido.")
		return
	}
	publicKey, network, ok := validateIdentity(input.PublicKey, input.Network)
	if !ok {
		writeError(writer, http.StatusBadRequest, "Usa una clave pública Stellar válida y network testnet o mainnet.")
		return
	}
	if !api.limiter.AllowAccount(writer, publicKey) {
		return
	}

	account, err := api.repository.GetOrCreateAccount(publicKey, network)
	if err != nil {
		api.internalError(writer, request, err)
		return
	}
	session := api.manager.Start(account)
	writeJSON(writer, http.StatusAccepted, map[string]any{
		"account": account,
		"status":  session.CurrentStatus(),
	})
}

func (api *API) getAccount(writer http.ResponseWriter, request *http.Request) {
	account, ok := api.accountFromRequest(writer, request)
	if !ok {
		return
	}
	session := api.manager.Start(account)
	writeJSON(writer, http.StatusOK, map[string]any{
		"account": account,
		"status":  session.CurrentStatus(),
	})
}

func (api *API) listAlerts(writer http.ResponseWriter, request *http.Request) {
	account, ok := api.accountFromRequest(writer, request)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(request.URL.Query().Get("limit"))
	if request.URL.Query().Get("limit") != "" && (limit <= 0 || limit > 200) {
		writeError(writer, http.StatusBadRequest, "limit debe ser un entero entre 1 y 200.")
		return
	}
	var cursor uint64
	if rawCursor := request.URL.Query().Get("cursor"); rawCursor != "" {
		parsed, err := strconv.ParseUint(rawCursor, 10, 64)
		if err != nil || parsed == 0 {
			writeError(writer, http.StatusBadRequest, "cursor no es válido.")
			return
		}
		cursor = parsed
	}
	page, err := api.repository.ListAlertsPage(account.ID, limit, uint(cursor))
	if err != nil {
		api.internalError(writer, request, err)
		return
	}
	response := map[string]any{"alerts": page.Alerts}
	if page.NextCursor > 0 {
		response["nextCursor"] = strconv.FormatUint(uint64(page.NextCursor), 10)
	}
	writeJSON(writer, http.StatusOK, response)
}

func (api *API) streamEvents(writer http.ResponseWriter, request *http.Request) {
	account, ok := api.accountFromRequest(writer, request)
	if !ok {
		return
	}
	flusher, ok := writer.(http.Flusher)
	if !ok {
		writeError(writer, http.StatusInternalServerError, "El servidor no soporta streaming.")
		return
	}

	session := api.manager.Start(account)
	events, unsubscribe := session.Subscribe()
	defer unsubscribe()
	disconnectMetric := api.metrics.SSEConnected()
	defer disconnectMetric()

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache, no-transform")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("X-Accel-Buffering", "no")
	writer.WriteHeader(http.StatusOK)
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-request.Context().Done():
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprint(writer, ": heartbeat\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case event, open := <-events:
			if !open {
				return
			}
			payload, err := json.Marshal(event.Data)
			if err != nil {
				api.logger.ErrorContext(request.Context(), "sse_serialization_failed",
					"request_id", middleware.RequestIDFromContext(request.Context()),
					"error", err,
				)
				continue
			}
			if _, err := fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", event.Type, payload); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (api *API) accountFromRequest(writer http.ResponseWriter, request *http.Request) (account model.MonitoredAccount, ok bool) {
	publicKey, network, valid := validateIdentity(request.PathValue("publicKey"), request.URL.Query().Get("network"))
	if !valid {
		writeError(writer, http.StatusBadRequest, "Cuenta o red inválida.")
		return model.MonitoredAccount{}, false
	}
	stored, err := api.repository.FindAccount(publicKey, network)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(writer, http.StatusNotFound, "La cuenta aún no está registrada en Sentinel.")
		return model.MonitoredAccount{}, false
	}
	if err != nil {
		api.internalError(writer, request, err)
		return model.MonitoredAccount{}, false
	}
	return stored, true
}

func validateIdentity(publicKey, network string) (string, string, bool) {
	key := strings.ToUpper(strings.TrimSpace(publicKey))
	selectedNetwork := strings.ToLower(strings.TrimSpace(network))
	if selectedNetwork == "" {
		selectedNetwork = "testnet"
	}
	return key, selectedNetwork, stellarPublicKeyPattern.MatchString(key) && (selectedNetwork == "testnet" || selectedNetwork == "mainnet")
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func writeError(writer http.ResponseWriter, status int, message string) {
	writeJSON(writer, status, map[string]string{"error": message})
}

func (api *API) internalError(writer http.ResponseWriter, request *http.Request, err error) {
	api.logger.ErrorContext(request.Context(), "request_failed",
		"request_id", middleware.RequestIDFromContext(request.Context()),
		"error", err,
	)
	writeError(writer, http.StatusInternalServerError, "Sentinel no pudo completar la operación.")
}
