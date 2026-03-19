package event

import (
	"context"
	"log/slog"
	"strings"
	"sync"
)

// Envelope wraps every event with metadata for tracing and tenant scoping.
type Envelope struct {
	EventID       string `json:"event_id"`
	EventType     string `json:"event_type"`     // NATS-compatible subject: "orders.created"
	OrgID         string `json:"org_id"`          // tenant scope
	LocationID    string `json:"location_id"`     // optional location scope
	Source        string `json:"source"`          // originating module/adapter
	SchemaVersion int    `json:"schema_version"`  // payload schema version
	Payload       any    `json:"payload"`         // event-specific data
}

// Handler processes an event envelope. Return error to send to dead letter queue.
type Handler func(ctx context.Context, env Envelope) error

// subscription holds a handler and its subject pattern.
type subscription struct {
	subject string
	handler Handler
}

// Bus is an in-process event bus with NATS-compatible subject naming.
// Subjects use dot-separated tokens: "orders.created", "menu.updated".
// Wildcard "*" matches a single token, ">" matches one or more trailing tokens.
type Bus struct {
	mu          sync.RWMutex
	subs        []subscription
	dlq         []Envelope // dead letter queue for failed events
	dlqMu       sync.Mutex
	middlewares []Middleware
}

// Middleware wraps a handler with cross-cutting concerns.
type Middleware func(Handler) Handler

// New creates a new event bus.
func New(middlewares ...Middleware) *Bus {
	return &Bus{
		middlewares: middlewares,
	}
}

// Subscribe registers a handler for a subject pattern.
// Patterns support NATS-style wildcards:
//   - "*" matches exactly one token
//   - ">" matches one or more trailing tokens
func (b *Bus) Subscribe(subject string, h Handler) {
	wrapped := h
	for i := len(b.middlewares) - 1; i >= 0; i-- {
		wrapped = b.middlewares[i](wrapped)
	}
	b.mu.Lock()
	b.subs = append(b.subs, subscription{subject: subject, handler: wrapped})
	b.mu.Unlock()
}

// Publish dispatches an event to all matching subscribers synchronously.
// Failed handlers send the event to the dead letter queue.
func (b *Bus) Publish(ctx context.Context, env Envelope) {
	b.mu.RLock()
	matches := make([]subscription, 0, len(b.subs))
	for _, s := range b.subs {
		if matchSubject(s.subject, env.EventType) {
			matches = append(matches, s)
		}
	}
	b.mu.RUnlock()

	for _, s := range matches {
		if err := s.handler(ctx, env); err != nil {
			slog.Error("event handler failed, sending to DLQ",
				"event_type", env.EventType,
				"event_id", env.EventID,
				"error", err,
			)
			b.sendToDLQ(env)
		}
	}
}

// DLQ returns a copy of the dead letter queue contents.
func (b *Bus) DLQ() []Envelope {
	b.dlqMu.Lock()
	defer b.dlqMu.Unlock()
	out := make([]Envelope, len(b.dlq))
	copy(out, b.dlq)
	return out
}

// DrainDLQ removes and returns all dead letter queue entries.
func (b *Bus) DrainDLQ() []Envelope {
	b.dlqMu.Lock()
	defer b.dlqMu.Unlock()
	out := b.dlq
	b.dlq = nil
	return out
}

func (b *Bus) sendToDLQ(env Envelope) {
	b.dlqMu.Lock()
	b.dlq = append(b.dlq, env)
	b.dlqMu.Unlock()
}

// matchSubject checks if a NATS-style pattern matches a subject.
func matchSubject(pattern, subject string) bool {
	patTokens := strings.Split(pattern, ".")
	subTokens := strings.Split(subject, ".")

	for i, pt := range patTokens {
		if pt == ">" {
			return i < len(subTokens) // ">" matches one or more remaining tokens
		}
		if i >= len(subTokens) {
			return false
		}
		if pt != "*" && pt != subTokens[i] {
			return false
		}
	}
	return len(patTokens) == len(subTokens)
}
