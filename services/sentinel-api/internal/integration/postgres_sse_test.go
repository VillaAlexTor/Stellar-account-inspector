//go:build integration

package integration_test

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stellar-account-inspector/sentinel-api/internal/database"
	"github.com/stellar-account-inspector/sentinel-api/internal/horizon"
	httpapi "github.com/stellar-account-inspector/sentinel-api/internal/http"
	"github.com/stellar-account-inspector/sentinel-api/internal/http/middleware"
	"github.com/stellar-account-inspector/sentinel-api/internal/model"
	"github.com/stellar-account-inspector/sentinel-api/internal/notifications"
	"github.com/stellar-account-inspector/sentinel-api/internal/observability"
	"github.com/stellar-account-inspector/sentinel-api/internal/repository"
	"github.com/stellar-account-inspector/sentinel-api/internal/sentinel"
	"gorm.io/gorm"
)

const defaultIntegrationDatabaseURL = "postgres://stellar:stellar@localhost:5433/stellar_inspector?sslmode=disable"

func TestPostgresAndBrowserSSE(t *testing.T) {
	databaseURL := os.Getenv("SENTINEL_INTEGRATION_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultIntegrationDatabaseURL
	}

	db, err := database.Open(databaseURL)
	if err != nil {
		t.Fatalf("open real PostgreSQL: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate real PostgreSQL: %v", err)
	}

	publicKey := integrationPublicKey(t.Name(), time.Now().UTC())
	operationID := fmt.Sprintf("integration-%d", time.Now().UnixNano())
	streamReady := make(chan struct{})
	releaseOperation := make(chan struct{})
	var readyOnce sync.Once

	horizonServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.URL.Path == "/accounts/"+publicKey:
			writer.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"account_id": publicKey,
				"thresholds": map[string]int{
					"low_threshold":  1,
					"med_threshold":  1,
					"high_threshold": 1,
				},
				"signers":  []map[string]any{{"key": publicKey, "weight": 1, "type": "ed25519_public_key"}},
				"balances": []map[string]string{{"asset_type": "native"}},
			})
		case request.URL.Path == "/accounts/"+publicKey+"/operations":
			flusher, ok := writer.(http.Flusher)
			if !ok {
				http.Error(writer, "streaming unsupported", http.StatusInternalServerError)
				return
			}
			writer.Header().Set("Content-Type", "text/event-stream")
			writer.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(writer, "data: hello\n\n")
			flusher.Flush()
			readyOnce.Do(func() { close(streamReady) })

			select {
			case <-request.Context().Done():
				return
			case <-releaseOperation:
			}

			operation := map[string]any{
				"id":                operationID,
				"paging_token":      operationID,
				"type":              "set_options",
				"source_account":    publicKey,
				"created_at":        time.Now().UTC().Format(time.RFC3339),
				"master_key_weight": 0,
			}
			payload, _ := json.Marshal(operation)
			_, _ = fmt.Fprintf(writer, "data: %s\n\n", payload)
			flusher.Flush()
			<-request.Context().Done()
		default:
			http.NotFound(writer, request)
		}
	}))

	ctx, cancel := context.WithCancel(context.Background())
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	metrics := observability.NewMetrics()
	repo := repository.New(db)
	horizonClient := horizon.NewClient(horizonServer.URL, 3*time.Second)
	delivered := make(chan model.NotificationJob, 1)
	dispatcher := notifications.NewDispatcher(repo, map[string]notifications.Sender{
		"webhook": integrationSender{delivered: delivered},
	}, 10*time.Millisecond, 3, logger, metrics)
	go dispatcher.Run(ctx)
	manager := sentinel.NewManager(ctx, repo, horizonClient, 100*time.Millisecond, logger, metrics, dispatcher)
	apiServer := httptest.NewServer(httpapi.NewRouter(
		repo,
		manager,
		logger,
		[]string{"http://localhost:3000"},
		metrics,
		httpapi.SecurityOptions{
			Auth: middleware.AuthOptions{
				Tokens:        []string{"integration-token"},
				SessionSecret: "0123456789abcdef0123456789abcdef",
				SessionTTL:    time.Hour,
			},
			RateLimit: middleware.RateLimitOptions{
				Window:       time.Minute,
				IPLimit:      1000,
				AccountLimit: 1000,
			},
		},
	))

	var accountID uint
	defer func() {
		cancel()
		apiServer.Close()
		horizonServer.Close()
		if accountID != 0 {
			alertIDs := db.Model(&model.SentinelAlert{}).Select("id").Where("monitored_account_id = ?", accountID)
			_ = db.Where("sentinel_alert_id IN (?)", alertIDs).Delete(&model.NotificationDelivery{}).Error
			_ = db.Where("monitored_account_id = ?", accountID).Delete(&model.SentinelAlert{}).Error
			_ = db.Where("monitored_account_id = ?", accountID).Delete(&model.RelevantOperation{}).Error
			_ = db.Where("id = ?", accountID).Delete(&model.MonitoredAccount{}).Error
		}
	}()

	unauthorized, err := http.Get(apiServer.URL + "/api/v1/monitored-accounts/" + publicKey)
	if err != nil {
		t.Fatalf("request protected endpoint: %v", err)
	}
	_ = unauthorized.Body.Close()
	if unauthorized.StatusCode != http.StatusUnauthorized {
		t.Fatalf("protected endpoint status = %d, want 401", unauthorized.StatusCode)
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	ready, err := client.Get(apiServer.URL + "/readyz")
	if err != nil {
		t.Fatalf("query readiness: %v", err)
	}
	_ = ready.Body.Close()
	if ready.StatusCode != http.StatusOK {
		t.Fatalf("readiness status = %d", ready.StatusCode)
	}
	loginBody := bytes.NewBufferString(`{"token":"integration-token"}`)
	login, err := client.Post(apiServer.URL+"/api/v1/auth/session", "application/json", loginBody)
	if err != nil {
		t.Fatalf("create authenticated session: %v", err)
	}
	_ = login.Body.Close()
	if login.StatusCode != http.StatusOK {
		t.Fatalf("session status = %d", login.StatusCode)
	}
	metricsRequest, _ := http.NewRequest(http.MethodGet, apiServer.URL+"/metrics", nil)
	metricsRequest.Header.Set("Authorization", "Bearer integration-token")
	metricsResponse, err := client.Do(metricsRequest)
	if err != nil {
		t.Fatalf("scrape authenticated metrics: %v", err)
	}
	metricsBody, _ := io.ReadAll(metricsResponse.Body)
	_ = metricsResponse.Body.Close()
	if metricsResponse.StatusCode != http.StatusOK || !bytes.Contains(metricsBody, []byte("sentinel_auth_attempts_total")) {
		t.Fatalf("metrics status = %d, body = %s", metricsResponse.StatusCode, metricsBody)
	}

	requestBody, _ := json.Marshal(map[string]string{"publicKey": publicKey})
	response, err := client.Post(apiServer.URL+"/api/v1/monitored-accounts", "application/json", bytes.NewReader(requestBody))
	if err != nil {
		t.Fatalf("register monitored account: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("register status = %d, body = %s", response.StatusCode, body)
	}
	var monitorResponse struct {
		Account model.MonitoredAccount `json:"account"`
	}
	if err := json.NewDecoder(response.Body).Decode(&monitorResponse); err != nil {
		t.Fatalf("decode monitor response: %v", err)
	}
	accountID = monitorResponse.Account.ID
	if accountID == 0 {
		t.Fatal("monitor response did not return a persisted account ID")
	}

	select {
	case <-streamReady:
	case <-time.After(5 * time.Second):
		t.Fatal("Horizon SSE stream did not connect")
	}

	streamContext, stopStream := context.WithTimeout(context.Background(), 8*time.Second)
	defer stopStream()
	streamRequest, _ := http.NewRequestWithContext(
		streamContext,
		http.MethodGet,
		apiServer.URL+"/api/v1/monitored-accounts/"+publicKey+"/events",
		nil,
	)
	streamResponse, err := client.Do(streamRequest)
	if err != nil {
		t.Fatalf("open browser SSE stream: %v", err)
	}
	defer streamResponse.Body.Close()
	if streamResponse.StatusCode != http.StatusOK {
		t.Fatalf("browser SSE status = %d", streamResponse.StatusCode)
	}
	if contentType := streamResponse.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "text/event-stream") {
		t.Fatalf("browser SSE content type = %q", contentType)
	}

	close(releaseOperation)
	eventType, eventData := readSSEEvent(t, streamResponse.Body, "alert")
	if eventType != "alert" {
		t.Fatalf("event type = %q", eventType)
	}
	var alert model.SentinelAlert
	if err := json.Unmarshal(eventData, &alert); err != nil {
		t.Fatalf("decode alert event: %v", err)
	}
	if alert.RuleID != "MASTER_KEY_ZEROED" || alert.OperationID != operationID {
		t.Fatalf("unexpected alert: %#v", alert)
	}
	select {
	case job := <-delivered:
		if job.Alert.ID != alert.ID || job.Channel != "webhook" {
			t.Fatalf("unexpected notification job: %#v", job)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("durable notification was not delivered")
	}

	var storedAccount model.MonitoredAccount
	if err := db.First(&storedAccount, accountID).Error; err != nil {
		t.Fatalf("load persisted account: %v", err)
	}
	if storedAccount.LastCursor != operationID {
		t.Fatalf("persisted cursor = %q, want %q", storedAccount.LastCursor, operationID)
	}
	var alertCount int64
	if err := db.Model(&model.SentinelAlert{}).
		Where("monitored_account_id = ? AND operation_id = ?", accountID, operationID).
		Count(&alertCount).Error; err != nil {
		t.Fatalf("count persisted alerts: %v", err)
	}
	if alertCount != 1 {
		t.Fatalf("persisted alert count = %d, want 1", alertCount)
	}
	assertNotificationSent(t, db, alert.ID)
	assertPaginationAndRetention(t, client, apiServer.URL, repo, db, accountID, publicKey)
}

type integrationSender struct {
	delivered chan<- model.NotificationJob
}

func (sender integrationSender) Send(ctx context.Context, job model.NotificationJob) error {
	select {
	case sender.delivered <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func assertNotificationSent(t *testing.T, db *gorm.DB, alertID uint) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var delivery model.NotificationDelivery
		result := db.Where("sentinel_alert_id = ?", alertID).First(&delivery)
		if result.Error == nil && delivery.Status == "sent" && delivery.Attempts == 1 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("notification delivery did not reach sent status")
}

func assertPaginationAndRetention(t *testing.T, client *http.Client, serverURL string, repo *repository.Repository, db *gorm.DB, accountID uint, publicKey string) {
	t.Helper()
	now := time.Now().UTC()
	newAlerts := []model.SentinelAlert{
		{MonitoredAccountID: accountID, RuleID: "PAGE_ONE", Severity: "low", Message: "page", OperationID: fmt.Sprintf("page-1-%d", now.UnixNano()), CreatedAt: now.Add(time.Second)},
		{MonitoredAccountID: accountID, RuleID: "PAGE_TWO", Severity: "medium", Message: "page", OperationID: fmt.Sprintf("page-2-%d", now.UnixNano()), CreatedAt: now.Add(2 * time.Second)},
	}
	if err := db.Create(&newAlerts).Error; err != nil {
		t.Fatalf("seed paginated alerts: %v", err)
	}
	first := fetchAlertPage(t, client, serverURL+"/api/v1/monitored-accounts/"+publicKey+"/alerts?limit=2")
	if len(first.Alerts) != 2 || first.NextCursor == "" {
		t.Fatalf("first alert page = %#v", first)
	}
	second := fetchAlertPage(t, client, serverURL+"/api/v1/monitored-accounts/"+publicKey+"/alerts?limit=2&cursor="+first.NextCursor)
	if len(second.Alerts) == 0 {
		t.Fatalf("second alert page = %#v", second)
	}

	oldOperation := model.RelevantOperation{MonitoredAccountID: accountID, OperationID: fmt.Sprintf("old-%d", now.UnixNano()), OperationType: "payment", Payload: "{}", CreatedAt: now.Add(-48 * time.Hour)}
	if err := db.Create(&oldOperation).Error; err != nil {
		t.Fatalf("seed expired operation: %v", err)
	}
	oldAlert := model.SentinelAlert{MonitoredAccountID: accountID, RuleID: "OLD", Severity: "low", Message: "old", OperationID: oldOperation.OperationID, CreatedAt: now.Add(-48 * time.Hour)}
	if err := db.Create(&oldAlert).Error; err != nil {
		t.Fatalf("seed expired alert: %v", err)
	}
	oldDelivery := model.NotificationDelivery{SentinelAlertID: oldAlert.ID, Channel: "webhook", Status: "sent", NextAttemptAt: now.Add(-48 * time.Hour)}
	if err := db.Create(&oldDelivery).Error; err != nil {
		t.Fatalf("seed expired delivery: %v", err)
	}
	result, err := repo.Prune(context.Background(), now.Add(-24*time.Hour), now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("prune retention data: %v", err)
	}
	if result.Alerts < 1 || result.Operations < 1 || result.Deliveries < 1 {
		t.Fatalf("retention result = %#v", result)
	}
}

type alertPageResponse struct {
	Alerts     []model.SentinelAlert `json:"alerts"`
	NextCursor string                `json:"nextCursor"`
}

func fetchAlertPage(t *testing.T, client *http.Client, endpoint string) alertPageResponse {
	t.Helper()
	response, err := client.Get(endpoint)
	if err != nil {
		t.Fatalf("fetch alert page: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("alert page status = %d, body = %s", response.StatusCode, body)
	}
	var page alertPageResponse
	if err := json.NewDecoder(response.Body).Decode(&page); err != nil {
		t.Fatalf("decode alert page: %v", err)
	}
	return page
}

func readSSEEvent(t *testing.T, reader io.Reader, wantedType string) (string, []byte) {
	t.Helper()
	scanner := bufio.NewScanner(reader)
	var eventType string
	var data strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event:"):
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			data.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		case line == "":
			if eventType == wantedType {
				return eventType, []byte(data.String())
			}
			eventType = ""
			data.Reset()
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read browser SSE: %v", err)
	}
	t.Fatalf("browser SSE closed before event %q", wantedType)
	return "", nil
}

func integrationPublicKey(name string, now time.Time) string {
	digest := sha512.Sum512([]byte(fmt.Sprintf("%s-%d", name, now.UnixNano())))
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(digest[:])
	return "G" + encoded[:55]
}
