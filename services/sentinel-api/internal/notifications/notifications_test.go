package notifications

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stellar-account-inspector/sentinel-api/internal/model"
)

func TestWebhookSenderSignsPayload(t *testing.T) {
	var receivedBody []byte
	var signature, deliveryID string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		receivedBody, _ = io.ReadAll(request.Body)
		signature = request.Header.Get("X-Sentinel-Signature")
		deliveryID = request.Header.Get("X-Sentinel-Delivery-ID")
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	job := testJob()
	if err := NewWebhookSender(server.URL, "secret", time.Second).Send(context.Background(), job); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	mac := hmac.New(sha256.New, []byte("secret"))
	_, _ = mac.Write(receivedBody)
	if signature != "sha256="+hex.EncodeToString(mac.Sum(nil)) || deliveryID != "42" {
		t.Fatalf("signature=%q deliveryID=%q", signature, deliveryID)
	}
	var payload map[string]any
	if err := json.Unmarshal(receivedBody, &payload); err != nil || payload["event"] != "sentinel.alert.created" {
		t.Fatalf("payload = %s, error = %v", receivedBody, err)
	}
}

func TestTelegramSenderEscapesHTML(t *testing.T) {
	var form url.Values
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_ = request.ParseForm()
		form = request.Form
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	job := testJob()
	job.Alert.Message = "riesgo <crítico> & revisar"
	if err := NewTelegramSender(server.URL, "token", "chat", time.Second).Send(context.Background(), job); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if form.Get("chat_id") != "chat" || !strings.Contains(form.Get("text"), "&lt;crítico&gt; &amp; revisar") {
		t.Fatalf("Telegram form = %#v", form)
	}
}

func TestEmailMessageSanitizesHeaders(t *testing.T) {
	message := emailMessage("alerts@example.com\r\nBcc: bad@example.com", []string{"owner@example.com"}, testJob())
	if strings.Contains(message, "\r\nBcc:") {
		t.Fatalf("header injection remained in message: %s", message)
	}
}

func testJob() model.NotificationJob {
	return model.NotificationJob{
		DeliveryID: 42,
		PublicKey:  "GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		Network:    "testnet",
		Alert: model.SentinelAlert{
			ID: 7, RuleID: "MASTER_KEY_ZEROED", Severity: "critical", Message: "La master key cambió.",
			OperationID: "123", CreatedAt: time.Date(2026, time.September, 27, 12, 0, 0, 0, time.UTC),
		},
	}
}
