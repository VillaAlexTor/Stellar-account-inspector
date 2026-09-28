package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthenticatorSessionAndBearer(t *testing.T) {
	auth := NewAuthenticator(AuthOptions{
		Tokens:        []string{"test-token"},
		SessionSecret: "0123456789abcdef0123456789abcdef",
		SessionTTL:    time.Hour,
	})
	fixedNow := time.Date(2026, time.September, 27, 12, 0, 0, 0, time.UTC)
	auth.now = func() time.Time { return fixedNow }

	login := httptest.NewRecorder()
	auth.Session(login, httptest.NewRequest(http.MethodPost, "/api/v1/auth/session", bytes.NewBufferString(`{"token":"test-token"}`)))
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", login.Code, login.Body.String())
	}
	response := login.Result()
	cookies := response.Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode {
		t.Fatalf("unexpected session cookie: %#v", cookies)
	}

	protected := auth.Require(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))

	withoutCredentials := httptest.NewRecorder()
	protected.ServeHTTP(withoutCredentials, httptest.NewRequest(http.MethodGet, "/api/v1/monitored-accounts/G", nil))
	if withoutCredentials.Code != http.StatusUnauthorized {
		t.Fatalf("missing credentials status = %d", withoutCredentials.Code)
	}

	withCookieRequest := httptest.NewRequest(http.MethodGet, "/api/v1/monitored-accounts/G", nil)
	withCookieRequest.AddCookie(cookies[0])
	withCookie := httptest.NewRecorder()
	protected.ServeHTTP(withCookie, withCookieRequest)
	if withCookie.Code != http.StatusNoContent {
		t.Fatalf("cookie credentials status = %d", withCookie.Code)
	}

	withBearerRequest := httptest.NewRequest(http.MethodGet, "/api/v1/monitored-accounts/G", nil)
	withBearerRequest.Header.Set("Authorization", "Bearer test-token")
	withBearer := httptest.NewRecorder()
	protected.ServeHTTP(withBearer, withBearerRequest)
	if withBearer.Code != http.StatusNoContent {
		t.Fatalf("bearer credentials status = %d", withBearer.Code)
	}

	auth.now = func() time.Time { return fixedNow.Add(2 * time.Hour) }
	expiredRequest := httptest.NewRequest(http.MethodGet, "/api/v1/monitored-accounts/G", nil)
	expiredRequest.AddCookie(cookies[0])
	expired := httptest.NewRecorder()
	protected.ServeHTTP(expired, expiredRequest)
	if expired.Code != http.StatusUnauthorized {
		t.Fatalf("expired session status = %d", expired.Code)
	}
}

func TestAuthenticatorRejectsInvalidToken(t *testing.T) {
	auth := NewAuthenticator(AuthOptions{
		Tokens:        []string{"expected"},
		SessionSecret: "0123456789abcdef0123456789abcdef",
	})
	recorder := httptest.NewRecorder()
	auth.Session(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/auth/session", bytes.NewBufferString(`{"token":"wrong"}`)))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("invalid token status = %d", recorder.Code)
	}
}
