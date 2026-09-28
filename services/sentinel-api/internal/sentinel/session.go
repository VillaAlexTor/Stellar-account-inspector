package sentinel

import (
	"sync"
	"time"

	"github.com/stellar-account-inspector/sentinel-api/internal/model"
)

type Status struct {
	State  string    `json:"state"`
	Detail string    `json:"detail,omitempty"`
	At     time.Time `json:"at"`
}

type BrowserEvent struct {
	Type string
	Data any
}

type Session struct {
	mu          sync.RWMutex
	status      Status
	subscribers map[chan BrowserEvent]struct{}
}

func newSession() *Session {
	return &Session{
		status:      Status{State: "connecting", At: time.Now().UTC()},
		subscribers: make(map[chan BrowserEvent]struct{}),
	}
}

func (session *Session) Subscribe() (<-chan BrowserEvent, func()) {
	channel := make(chan BrowserEvent, 64)
	session.mu.Lock()
	session.subscribers[channel] = struct{}{}
	current := session.status
	session.mu.Unlock()
	channel <- BrowserEvent{Type: "status", Data: current}

	return channel, func() {
		session.mu.Lock()
		if _, exists := session.subscribers[channel]; exists {
			delete(session.subscribers, channel)
			close(channel)
		}
		session.mu.Unlock()
	}
}

func (session *Session) SetStatus(state, detail string) {
	status := Status{State: state, Detail: detail, At: time.Now().UTC()}
	session.mu.Lock()
	session.status = status
	session.broadcastLocked(BrowserEvent{Type: "status", Data: status})
	session.mu.Unlock()
}

func (session *Session) PublishAlert(alert model.SentinelAlert) {
	session.mu.Lock()
	session.broadcastLocked(BrowserEvent{Type: "alert", Data: alert})
	session.mu.Unlock()
}

func (session *Session) CurrentStatus() Status {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.status
}

func (session *Session) broadcastLocked(event BrowserEvent) {
	for subscriber := range session.subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
}
