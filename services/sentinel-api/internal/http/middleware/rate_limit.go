package middleware

import (
	"encoding/json"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RateLimitOptions struct {
	Window            time.Duration
	IPLimit           int
	AccountLimit      int
	TrustedProxyCIDRs []string
}

type rateEntry struct {
	startedAt time.Time
	count     int
}

type RateLimiter struct {
	mu             sync.Mutex
	window         time.Duration
	ipLimit        int
	accountLimit   int
	trustedProxies []*net.IPNet
	entries        map[string]rateEntry
	now            func() time.Time
}

func NewRateLimiter(options RateLimitOptions) *RateLimiter {
	window := options.Window
	if window <= 0 {
		window = time.Minute
	}
	ipLimit := options.IPLimit
	if ipLimit <= 0 {
		ipLimit = 120
	}
	accountLimit := options.AccountLimit
	if accountLimit <= 0 {
		accountLimit = 30
	}
	trusted := make([]*net.IPNet, 0, len(options.TrustedProxyCIDRs))
	for _, value := range options.TrustedProxyCIDRs {
		if _, network, err := net.ParseCIDR(value); err == nil {
			trusted = append(trusted, network)
		}
	}
	return &RateLimiter{
		window:         window,
		ipLimit:        ipLimit,
		accountLimit:   accountLimit,
		trustedProxies: trusted,
		entries:        make(map[string]rateEntry),
		now:            time.Now,
	}
}

func (limiter *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/healthz" {
			next.ServeHTTP(writer, request)
			return
		}
		if !limiter.allow(writer, "ip:"+limiter.clientIP(request), limiter.ipLimit) {
			return
		}
		if account := accountFromPath(request.URL.Path); account != "" &&
			!limiter.allow(writer, "account:"+account, limiter.accountLimit) {
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func (limiter *RateLimiter) AllowAccount(writer http.ResponseWriter, publicKey string) bool {
	return limiter.allow(writer, "account:"+strings.ToUpper(publicKey), limiter.accountLimit)
}

func (limiter *RateLimiter) allow(writer http.ResponseWriter, key string, limit int) bool {
	now := limiter.now().UTC()
	limiter.mu.Lock()
	entry := limiter.entries[key]
	if entry.startedAt.IsZero() || now.Sub(entry.startedAt) >= limiter.window {
		entry = rateEntry{startedAt: now}
	}
	entry.count++
	limiter.entries[key] = entry
	if len(limiter.entries) > 4096 {
		for entryKey, candidate := range limiter.entries {
			if now.Sub(candidate.startedAt) >= limiter.window {
				delete(limiter.entries, entryKey)
			}
		}
	}
	remaining := max(0, limit-entry.count)
	remainingWindow := limiter.window - now.Sub(entry.startedAt)
	retryAfter := max(1, int(math.Ceil(remainingWindow.Seconds())))
	allowed := entry.count <= limit
	limiter.mu.Unlock()

	writer.Header().Set("RateLimit-Limit", strconv.Itoa(limit))
	writer.Header().Set("RateLimit-Remaining", strconv.Itoa(remaining))
	if allowed {
		return true
	}
	writer.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(writer).Encode(map[string]string{
		"error": "Límite temporal alcanzado. Espera antes de volver a intentarlo.",
	})
	return false
}

func (limiter *RateLimiter) clientIP(request *http.Request) string {
	remote := parseIP(request.RemoteAddr)
	if remote == nil {
		return "unknown"
	}
	if !limiter.isTrustedProxy(remote) {
		return remote.String()
	}

	forwarded := strings.Split(request.Header.Get("X-Forwarded-For"), ",")
	for index := len(forwarded) - 1; index >= 0; index-- {
		candidate := net.ParseIP(strings.TrimSpace(forwarded[index]))
		if candidate != nil && !limiter.isTrustedProxy(candidate) {
			return candidate.String()
		}
	}
	return remote.String()
}

func (limiter *RateLimiter) isTrustedProxy(ip net.IP) bool {
	for _, network := range limiter.trustedProxies {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func parseIP(address string) net.IP {
	host, _, err := net.SplitHostPort(address)
	if err == nil {
		return net.ParseIP(host)
	}
	return net.ParseIP(strings.Trim(address, "[]"))
}

func accountFromPath(path string) string {
	const prefix = "/api/v1/monitored-accounts/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	account := strings.Split(strings.TrimPrefix(path, prefix), "/")[0]
	if len(account) != 56 || !strings.HasPrefix(strings.ToUpper(account), "G") {
		return ""
	}
	return strings.ToUpper(account)
}
