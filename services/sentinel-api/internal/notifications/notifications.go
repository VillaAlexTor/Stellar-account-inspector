package notifications

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/stellar-account-inspector/sentinel-api/internal/model"
	"github.com/stellar-account-inspector/sentinel-api/internal/observability"
	"github.com/stellar-account-inspector/sentinel-api/internal/repository"
)

type Sender interface {
	Send(context.Context, model.NotificationJob) error
}

type Dispatcher struct {
	repository  *repository.Repository
	senders     map[string]Sender
	channels    []string
	poll        time.Duration
	maxAttempts int
	logger      *slog.Logger
	metrics     *observability.Metrics
	wake        chan struct{}
	now         func() time.Time
}

func NewDispatcher(repo *repository.Repository, senders map[string]Sender, poll time.Duration, maxAttempts int, logger *slog.Logger, metrics *observability.Metrics) *Dispatcher {
	channels := make([]string, 0, len(senders))
	for channel := range senders {
		channels = append(channels, channel)
	}
	sort.Strings(channels)
	return &Dispatcher{
		repository: repo, senders: senders, channels: channels, poll: poll, maxAttempts: maxAttempts,
		logger: logger, metrics: metrics, wake: make(chan struct{}, 1), now: time.Now,
	}
}

func (dispatcher *Dispatcher) Channels() []string {
	return append([]string(nil), dispatcher.channels...)
}

func (dispatcher *Dispatcher) Wake() {
	select {
	case dispatcher.wake <- struct{}{}:
	default:
	}
}

func (dispatcher *Dispatcher) Run(ctx context.Context) {
	if len(dispatcher.channels) == 0 {
		dispatcher.logger.InfoContext(ctx, "notifications_disabled")
		return
	}
	dispatcher.logger.InfoContext(ctx, "notifications_started", "channels", dispatcher.channels)
	ticker := time.NewTicker(dispatcher.poll)
	defer ticker.Stop()
	for {
		dispatcher.process(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-dispatcher.wake:
		}
	}
}

func (dispatcher *Dispatcher) process(ctx context.Context) {
	jobs, err := dispatcher.repository.ClaimNotificationJobs(dispatcher.now().UTC(), 20)
	if err != nil {
		dispatcher.logger.ErrorContext(ctx, "notification_claim_failed", "error", err)
		return
	}
	for _, job := range jobs {
		sender, exists := dispatcher.senders[job.Channel]
		if !exists {
			_ = dispatcher.repository.MarkNotificationFailed(job.DeliveryID, true, dispatcher.now(), "canal no configurado")
			continue
		}
		err := sender.Send(ctx, job)
		if err == nil {
			if markErr := dispatcher.repository.MarkNotificationSent(job.DeliveryID, dispatcher.now()); markErr != nil {
				dispatcher.logger.ErrorContext(ctx, "notification_mark_sent_failed", "channel", job.Channel, "delivery_id", job.DeliveryID, "error", markErr)
				continue
			}
			dispatcher.recordMetric(job.Channel, "sent")
			dispatcher.logger.InfoContext(ctx, "notification_sent", "channel", job.Channel, "delivery_id", job.DeliveryID, "alert_id", job.Alert.ID)
			continue
		}
		terminal := job.Attempts >= dispatcher.maxAttempts
		nextAttempt := dispatcher.now().Add(retryDelay(job.Attempts))
		if markErr := dispatcher.repository.MarkNotificationFailed(job.DeliveryID, terminal, nextAttempt, err.Error()); markErr != nil {
			dispatcher.logger.ErrorContext(ctx, "notification_mark_failed_error", "channel", job.Channel, "delivery_id", job.DeliveryID, "error", markErr)
		}
		result := "retry"
		if terminal {
			result = "failed"
		}
		dispatcher.recordMetric(job.Channel, result)
		dispatcher.logger.WarnContext(ctx, "notification_delivery_failed", "channel", job.Channel, "delivery_id", job.DeliveryID, "attempt", job.Attempts, "terminal", terminal, "error", err)
	}
}

func (dispatcher *Dispatcher) recordMetric(channel, result string) {
	if dispatcher.metrics != nil {
		dispatcher.metrics.Notification(channel, result)
	}
}

func retryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 8 {
		attempt = 8
	}
	return time.Duration(1<<uint(attempt-1)) * 5 * time.Second
}

type WebhookSender struct {
	endpoint string
	secret   []byte
	client   *http.Client
}

func NewWebhookSender(endpoint, secret string, timeout time.Duration) *WebhookSender {
	return &WebhookSender{endpoint: endpoint, secret: []byte(secret), client: notificationHTTPClient(timeout)}
}

func (sender *WebhookSender) Send(ctx context.Context, job model.NotificationJob) error {
	payload, err := json.Marshal(notificationPayload(job))
	if err != nil {
		return fmt.Errorf("serializar webhook: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("crear webhook: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "stellar-sentinel/1")
	request.Header.Set("X-Sentinel-Delivery-ID", fmt.Sprintf("%d", job.DeliveryID))
	if len(sender.secret) > 0 {
		mac := hmac.New(sha256.New, sender.secret)
		_, _ = mac.Write(payload)
		request.Header.Set("X-Sentinel-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	}
	response, err := sender.client.Do(request)
	if err != nil {
		return fmt.Errorf("enviar webhook: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
		return fmt.Errorf("webhook respondió %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

type TelegramSender struct {
	endpoint string
	chatID   string
	client   *http.Client
}

func NewTelegramSender(apiURL, token, chatID string, timeout time.Duration) *TelegramSender {
	endpoint := strings.TrimRight(apiURL, "/") + "/bot" + token + "/sendMessage"
	return &TelegramSender{endpoint: endpoint, chatID: chatID, client: notificationHTTPClient(timeout)}
}

func notificationHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func (sender *TelegramSender) Send(ctx context.Context, job model.NotificationJob) error {
	values := url.Values{
		"chat_id":                  {sender.chatID},
		"parse_mode":               {"HTML"},
		"disable_web_page_preview": {"true"},
		"text":                     {telegramMessage(job)},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("crear mensaje Telegram: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := sender.client.Do(request)
	if err != nil {
		return fmt.Errorf("enviar Telegram: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
		return fmt.Errorf("Telegram respondió %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

type SMTPSender struct {
	host, address, username, password, from, tlsMode string
	to                                               []string
	timeout                                          time.Duration
}

func NewSMTPSender(host string, port int, username, password, from string, to []string, tlsMode string, timeout time.Duration) *SMTPSender {
	return &SMTPSender{host: host, address: net.JoinHostPort(host, fmt.Sprintf("%d", port)), username: username, password: password, from: from, to: append([]string(nil), to...), tlsMode: tlsMode, timeout: timeout}
}

func (sender *SMTPSender) Send(ctx context.Context, job model.NotificationJob) error {
	client, connection, err := sender.connect(ctx)
	if err != nil {
		return err
	}
	defer connection.Close()
	defer client.Close()
	if sender.username != "" {
		if err := client.Auth(smtp.PlainAuth("", sender.username, sender.password, sender.host)); err != nil {
			return fmt.Errorf("autenticar SMTP: %w", err)
		}
	}
	if err := client.Mail(sender.from); err != nil {
		return fmt.Errorf("remitente SMTP: %w", err)
	}
	for _, recipient := range sender.to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("destinatario SMTP: %w", err)
		}
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("abrir cuerpo SMTP: %w", err)
	}
	if _, err := writer.Write([]byte(emailMessage(sender.from, sender.to, job))); err != nil {
		_ = writer.Close()
		return fmt.Errorf("enviar cuerpo SMTP: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("cerrar cuerpo SMTP: %w", err)
	}
	return client.Quit()
}

func (sender *SMTPSender) connect(ctx context.Context) (*smtp.Client, net.Conn, error) {
	dialer := &net.Dialer{Timeout: sender.timeout}
	var connection net.Conn
	var err error
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: sender.host}
	if sender.tlsMode == "implicit" {
		connection, err = tls.DialWithDialer(dialer, "tcp", sender.address, tlsConfig)
	} else {
		connection, err = dialer.DialContext(ctx, "tcp", sender.address)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("conectar SMTP: %w", err)
	}
	client, err := smtp.NewClient(connection, sender.host)
	if err != nil {
		connection.Close()
		return nil, nil, fmt.Errorf("iniciar SMTP: %w", err)
	}
	if sender.tlsMode == "starttls" {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			client.Close()
			connection.Close()
			return nil, nil, fmt.Errorf("el servidor SMTP no ofrece STARTTLS")
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			client.Close()
			connection.Close()
			return nil, nil, fmt.Errorf("activar STARTTLS: %w", err)
		}
	}
	return client, connection, nil
}

func notificationPayload(job model.NotificationJob) map[string]any {
	return map[string]any{
		"event": "sentinel.alert.created", "deliveryId": job.DeliveryID,
		"account": map[string]string{"publicKey": job.PublicKey, "network": job.Network},
		"alert":   job.Alert,
	}
}

func telegramMessage(job model.NotificationJob) string {
	return fmt.Sprintf("<b>Sentinel · %s</b>\n%s\n\n%s\n\nCuenta: <code>%s</code>\nRed: %s\nOperación: <code>%s</code>",
		html.EscapeString(strings.ToUpper(job.Alert.Severity)), html.EscapeString(job.Alert.RuleID), html.EscapeString(job.Alert.Message),
		html.EscapeString(job.PublicKey), html.EscapeString(job.Network), html.EscapeString(job.Alert.OperationID))
}

func emailMessage(from string, to []string, job model.NotificationJob) string {
	subject := sanitizeHeader(fmt.Sprintf("[Sentinel][%s] %s", strings.ToUpper(job.Alert.Severity), job.Alert.RuleID))
	body := fmt.Sprintf("Sentinel detectó una alerta.\r\n\r\nRegla: %s\r\nSeveridad: %s\r\nCuenta: %s\r\nRed: %s\r\nOperación: %s\r\nFecha: %s\r\n\r\n%s\r\n",
		job.Alert.RuleID, job.Alert.Severity, job.PublicKey, job.Network, job.Alert.OperationID, job.Alert.CreatedAt.UTC().Format(time.RFC3339), job.Alert.Message)
	return "From: " + sanitizeHeader(from) + "\r\n" +
		"To: " + sanitizeHeader(strings.Join(to, ", ")) + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + body
}

func sanitizeHeader(value string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(value)
}
