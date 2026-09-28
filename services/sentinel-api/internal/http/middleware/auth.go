package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stellar-account-inspector/sentinel-api/internal/observability"
)

const sessionCookieName = "sentinel_session"

type AuthOptions struct {
	Tokens        []string
	SessionSecret string
	SessionTTL    time.Duration
	SecureCookie  bool
	Metrics       *observability.Metrics
}

type Authenticator struct {
	tokenHashes  map[string][sha256.Size]byte
	secret       []byte
	sessionTTL   time.Duration
	secureCookie bool
	metrics      *observability.Metrics
	now          func() time.Time
}

func NewAuthenticator(options AuthOptions) *Authenticator {
	hashes := make(map[string][sha256.Size]byte, len(options.Tokens))
	for _, token := range options.Tokens {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			continue
		}
		digest := sha256.Sum256([]byte(trimmed))
		hashes[hex.EncodeToString(digest[:8])] = digest
	}
	ttl := options.SessionTTL
	if ttl <= 0 {
		ttl = 12 * time.Hour
	}
	return &Authenticator{
		tokenHashes:  hashes,
		secret:       []byte(options.SessionSecret),
		sessionTTL:   ttl,
		secureCookie: options.SecureCookie,
		metrics:      options.Metrics,
		now:          time.Now,
	}
}

func (auth *Authenticator) Enabled() bool {
	return len(auth.tokenHashes) > 0
}

func (auth *Authenticator) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !auth.Enabled() || request.URL.Path == "/healthz" || request.URL.Path == "/readyz" || request.URL.Path == "/api/v1/auth/session" {
			next.ServeHTTP(writer, request)
			return
		}
		if auth.requestAuthenticated(request) {
			next.ServeHTTP(writer, request)
			return
		}
		writeAuthJSON(writer, http.StatusUnauthorized, map[string]any{
			"error": "Autenticación requerida. Inicia una sesión Sentinel o usa un Bearer token válido.",
		})
		auth.recordAttempt("unauthorized")
	})
}

func (auth *Authenticator) Session(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Cache-Control", "no-store")
	switch request.Method {
	case http.MethodGet:
		writeAuthJSON(writer, http.StatusOK, map[string]bool{
			"enabled":       auth.Enabled(),
			"authenticated": !auth.Enabled() || auth.requestAuthenticated(request),
		})
	case http.MethodPost:
		if !auth.Enabled() {
			writeAuthJSON(writer, http.StatusOK, map[string]bool{"enabled": false, "authenticated": true})
			return
		}
		var input struct {
			Token string `json:"token"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(writer, request.Body, 8*1024))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			auth.recordAttempt("malformed")
			writeAuthJSON(writer, http.StatusBadRequest, map[string]string{"error": "El token de acceso no es válido."})
			return
		}
		keyID, valid := auth.validateToken(input.Token)
		if !valid {
			auth.recordAttempt("invalid")
			writeAuthJSON(writer, http.StatusUnauthorized, map[string]string{"error": "El token de acceso no coincide con la configuración de Sentinel."})
			return
		}
		expiresAt := auth.now().UTC().Add(auth.sessionTTL)
		http.SetCookie(writer, &http.Cookie{
			Name:     sessionCookieName,
			Value:    auth.signSession(keyID, expiresAt),
			Path:     "/api/v1",
			HttpOnly: true,
			Secure:   auth.secureCookie,
			SameSite: http.SameSiteStrictMode,
			Expires:  expiresAt,
			MaxAge:   int(auth.sessionTTL.Seconds()),
		})
		auth.recordAttempt("success")
		writeAuthJSON(writer, http.StatusOK, map[string]any{
			"enabled":       true,
			"authenticated": true,
			"expiresAt":     expiresAt,
		})
	case http.MethodDelete:
		http.SetCookie(writer, &http.Cookie{
			Name:     sessionCookieName,
			Value:    "",
			Path:     "/api/v1",
			HttpOnly: true,
			Secure:   auth.secureCookie,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   -1,
			Expires:  time.Unix(1, 0),
		})
		writer.WriteHeader(http.StatusNoContent)
	default:
		writer.Header().Set("Allow", "GET, POST, DELETE")
		writeAuthJSON(writer, http.StatusMethodNotAllowed, map[string]string{"error": "Método no permitido."})
	}
}

func (auth *Authenticator) recordAttempt(result string) {
	if auth.metrics != nil {
		auth.metrics.AuthAttempt(result)
	}
}

func (auth *Authenticator) requestAuthenticated(request *http.Request) bool {
	if header := strings.TrimSpace(request.Header.Get("Authorization")); strings.HasPrefix(header, "Bearer ") {
		_, valid := auth.validateToken(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if valid {
			return true
		}
	}
	cookie, err := request.Cookie(sessionCookieName)
	return err == nil && auth.verifySession(cookie.Value)
}

func (auth *Authenticator) validateToken(token string) (string, bool) {
	candidate := sha256.Sum256([]byte(strings.TrimSpace(token)))
	for keyID, expected := range auth.tokenHashes {
		if subtle.ConstantTimeCompare(candidate[:], expected[:]) == 1 {
			return keyID, true
		}
	}
	return "", false
}

func (auth *Authenticator) signSession(keyID string, expiresAt time.Time) string {
	payload := keyID + "." + strconv.FormatInt(expiresAt.Unix(), 10)
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
	mac := hmac.New(sha256.New, auth.secret)
	_, _ = mac.Write([]byte(encodedPayload))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encodedPayload + "." + signature
}

func (auth *Authenticator) verifySession(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return false
	}
	providedSignature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, auth.secret)
	_, _ = mac.Write([]byte(parts[0]))
	if !hmac.Equal(providedSignature, mac.Sum(nil)) {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	fields := strings.Split(string(payload), ".")
	if len(fields) != 2 {
		return false
	}
	if _, exists := auth.tokenHashes[fields[0]]; !exists {
		return false
	}
	expiresAt, err := strconv.ParseInt(fields[1], 10, 64)
	return err == nil && auth.now().UTC().Unix() < expiresAt
}

func writeAuthJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
