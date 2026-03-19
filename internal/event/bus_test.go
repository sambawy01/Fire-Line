package event_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/opsnerve/fireline/internal/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBus_PublishSubscribe(t *testing.T) {
	bus := event.New()

	var received event.Envelope
	bus.Subscribe("orders.created", func(ctx context.Context, env event.Envelope) error {
		received = env
		return nil
	})

	env := event.Envelope{
		EventID:   "evt-1",
		EventType: "orders.created",
		OrgID:     "org-1",
		Source:    "test",
		Payload:   map[string]string{"order_id": "123"},
	}

	bus.Publish(context.Background(), env)

	assert.Equal(t, "evt-1", received.EventID)
	assert.Equal(t, "orders.created", received.EventType)
	assert.Equal(t, "org-1", received.OrgID)
}

func TestBus_WildcardStar(t *testing.T) {
	bus := event.New()

	var count atomic.Int32
	bus.Subscribe("orders.*", func(ctx context.Context, env event.Envelope) error {
		count.Add(1)
		return nil
	})

	bus.Publish(context.Background(), event.Envelope{EventType: "orders.created"})
	bus.Publish(context.Background(), event.Envelope{EventType: "orders.updated"})
	bus.Publish(context.Background(), event.Envelope{EventType: "menu.updated"}) // should NOT match

	assert.Equal(t, int32(2), count.Load())
}

func TestBus_WildcardGT(t *testing.T) {
	bus := event.New()

	var count atomic.Int32
	bus.Subscribe("adapter.>", func(ctx context.Context, env event.Envelope) error {
		count.Add(1)
		return nil
	})

	bus.Publish(context.Background(), event.Envelope{EventType: "adapter.toast.connected"})
	bus.Publish(context.Background(), event.Envelope{EventType: "adapter.square.orders.synced"})
	bus.Publish(context.Background(), event.Envelope{EventType: "orders.created"}) // should NOT match

	assert.Equal(t, int32(2), count.Load())
}

func TestBus_MultipleSubscribers(t *testing.T) {
	bus := event.New()

	var count atomic.Int32
	for i := 0; i < 3; i++ {
		bus.Subscribe("orders.created", func(ctx context.Context, env event.Envelope) error {
			count.Add(1)
			return nil
		})
	}

	bus.Publish(context.Background(), event.Envelope{EventType: "orders.created"})

	assert.Equal(t, int32(3), count.Load())
}

func TestBus_DeadLetterQueue(t *testing.T) {
	bus := event.New()

	bus.Subscribe("orders.created", func(ctx context.Context, env event.Envelope) error {
		return fmt.Errorf("handler failed")
	})

	bus.Publish(context.Background(), event.Envelope{
		EventID:   "evt-fail",
		EventType: "orders.created",
	})

	dlq := bus.DLQ()
	require.Len(t, dlq, 1)
	assert.Equal(t, "evt-fail", dlq[0].EventID)
}

func TestBus_DrainDLQ(t *testing.T) {
	bus := event.New()

	bus.Subscribe("x.y", func(ctx context.Context, env event.Envelope) error {
		return fmt.Errorf("fail")
	})

	bus.Publish(context.Background(), event.Envelope{EventType: "x.y", EventID: "e1"})
	bus.Publish(context.Background(), event.Envelope{EventType: "x.y", EventID: "e2"})

	drained := bus.DrainDLQ()
	assert.Len(t, drained, 2)
	assert.Empty(t, bus.DLQ())
}

func TestBus_NoMatch(t *testing.T) {
	bus := event.New()

	called := false
	bus.Subscribe("menu.updated", func(ctx context.Context, env event.Envelope) error {
		called = true
		return nil
	})

	bus.Publish(context.Background(), event.Envelope{EventType: "orders.created"})
	assert.False(t, called)
}

func TestBus_Middleware(t *testing.T) {
	var order []string

	mw := func(next event.Handler) event.Handler {
		return func(ctx context.Context, env event.Envelope) error {
			order = append(order, "before")
			err := next(ctx, env)
			order = append(order, "after")
			return err
		}
	}

	bus := event.New(mw)
	bus.Subscribe("x.y", func(ctx context.Context, env event.Envelope) error {
		order = append(order, "handler")
		return nil
	})

	bus.Publish(context.Background(), event.Envelope{EventType: "x.y"})
	assert.Equal(t, []string{"before", "handler", "after"}, order)
}

func TestBus_ExactSubjectMatch(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		match   bool
	}{
		{"orders.created", "orders.created", true},
		{"orders.created", "orders.updated", false},
		{"orders.*", "orders.created", true},
		{"orders.*", "orders.updated", true},
		{"orders.*", "menu.created", false},
		{"orders.*", "orders.created.detail", false},
		{">", "orders.created", true},
		{">", "a", true},
		{"orders.>", "orders.created", true},
		{"orders.>", "orders.created.detail", true},
		{"orders.>", "menu.created", false},
		{"*.created", "orders.created", true},
		{"*.created", "menu.created", true},
		{"*.created", "orders.updated", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.subject, func(t *testing.T) {
			bus := event.New()
			matched := false
			bus.Subscribe(tt.pattern, func(ctx context.Context, env event.Envelope) error {
				matched = true
				return nil
			})
			bus.Publish(context.Background(), event.Envelope{EventType: tt.subject})
			assert.Equal(t, tt.match, matched, "pattern=%q subject=%q", tt.pattern, tt.subject)
		})
	}
}
