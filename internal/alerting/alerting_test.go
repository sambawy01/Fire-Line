package alerting_test

import (
	"context"
	"testing"

	"github.com/opsnerve/fireline/internal/alerting"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_DefaultRules(t *testing.T) {
	bus := event.New()
	svc := alerting.New(bus)
	svc.RegisterDefaultRules()

	// Fire inventory event
	bus.Publish(context.Background(), event.Envelope{
		EventID:    "evt-1",
		EventType:  "inventory.usage.updated",
		OrgID:      "org-1",
		LocationID: "loc-1",
		Source:     "inventory",
	})

	queue := svc.GetQueue("org-1", "loc-1")
	require.Len(t, queue, 1)
	assert.Equal(t, "Inventory usage recalculated", queue[0].Title)
	assert.Equal(t, alerting.SeverityInfo, queue[0].Severity)
	assert.Equal(t, "active", queue[0].Status)
}

func TestService_CriticalAlert(t *testing.T) {
	bus := event.New()
	svc := alerting.New(bus)
	svc.RegisterDefaultRules()

	bus.Publish(context.Background(), event.Envelope{
		EventID:    "evt-err",
		EventType:  "adapter.error",
		OrgID:      "org-1",
		LocationID: "loc-1",
		Source:     "toast",
	})

	queue := svc.GetQueue("org-1", "loc-1")
	require.Len(t, queue, 1)
	assert.Equal(t, alerting.SeverityCritical, queue[0].Severity)
	assert.Equal(t, "POS adapter error", queue[0].Title)
}

func TestService_Acknowledge(t *testing.T) {
	bus := event.New()
	svc := alerting.New(bus)
	svc.RegisterDefaultRules()

	bus.Publish(context.Background(), event.Envelope{
		EventID: "evt-1", EventType: "inventory.usage.updated",
		OrgID: "org-1", LocationID: "loc-1",
	})

	queue := svc.GetQueue("org-1", "loc-1")
	require.Len(t, queue, 1)

	ok := svc.Acknowledge(queue[0].AlertID)
	assert.True(t, ok)

	// Active queue should be empty now
	activeQueue := svc.GetQueue("org-1", "loc-1")
	assert.Empty(t, activeQueue)
}

func TestService_Resolve(t *testing.T) {
	bus := event.New()
	svc := alerting.New(bus)
	svc.RegisterDefaultRules()

	bus.Publish(context.Background(), event.Envelope{
		EventID: "evt-1", EventType: "financial.metrics.updated",
		OrgID: "org-1", LocationID: "loc-1",
	})

	queue := svc.GetQueue("org-1", "loc-1")
	require.Len(t, queue, 1)

	ok := svc.Resolve(queue[0].AlertID)
	assert.True(t, ok)

	assert.Equal(t, 0, svc.ActiveCount("org-1"))
}

func TestService_SeveritySorting(t *testing.T) {
	bus := event.New()
	svc := alerting.New(bus)
	svc.RegisterDefaultRules()

	// Fire info event first
	bus.Publish(context.Background(), event.Envelope{
		EventID: "evt-info", EventType: "inventory.usage.updated",
		OrgID: "org-1", LocationID: "loc-1",
	})
	// Then critical
	bus.Publish(context.Background(), event.Envelope{
		EventID: "evt-crit", EventType: "adapter.error",
		OrgID: "org-1", LocationID: "loc-1",
	})

	queue := svc.GetQueue("org-1", "loc-1")
	require.Len(t, queue, 2)
	assert.Equal(t, alerting.SeverityCritical, queue[0].Severity, "critical should sort first")
	assert.Equal(t, alerting.SeverityInfo, queue[1].Severity)
}

func TestService_ActiveCount(t *testing.T) {
	bus := event.New()
	svc := alerting.New(bus)
	svc.RegisterDefaultRules()

	assert.Equal(t, 0, svc.ActiveCount("org-1"))

	bus.Publish(context.Background(), event.Envelope{
		EventID: "e1", EventType: "inventory.usage.updated",
		OrgID: "org-1", LocationID: "loc-1",
	})
	bus.Publish(context.Background(), event.Envelope{
		EventID: "e2", EventType: "financial.metrics.updated",
		OrgID: "org-1", LocationID: "loc-1",
	})

	assert.Equal(t, 2, svc.ActiveCount("org-1"))
	assert.Equal(t, 0, svc.ActiveCount("org-2"), "different org should have 0")
}

func TestService_LocationFilter(t *testing.T) {
	bus := event.New()
	svc := alerting.New(bus)
	svc.RegisterDefaultRules()

	bus.Publish(context.Background(), event.Envelope{
		EventID: "e1", EventType: "inventory.usage.updated",
		OrgID: "org-1", LocationID: "loc-1",
	})
	bus.Publish(context.Background(), event.Envelope{
		EventID: "e2", EventType: "inventory.usage.updated",
		OrgID: "org-1", LocationID: "loc-2",
	})

	all := svc.GetQueue("org-1", "")
	assert.Len(t, all, 2)

	loc1 := svc.GetQueue("org-1", "loc-1")
	assert.Len(t, loc1, 1)
}

func TestService_CustomRule(t *testing.T) {
	bus := event.New()
	svc := alerting.New(bus)

	svc.AddRule(alerting.Rule{
		RuleID:    "custom-high-variance",
		Name:      "High Inventory Variance",
		Module:    "inventory",
		EventType: "inventory.variance.detected",
		Severity:  alerting.SeverityWarning,
		Enabled:   true,
		Evaluate: func(ctx context.Context, env event.Envelope) *alerting.Alert {
			return &alerting.Alert{
				OrgID:       env.OrgID,
				LocationID:  env.LocationID,
				Title:       "High variance detected: Ground Beef at 15%",
				Description: "Variance exceeds 10% threshold",
				Status:      "active",
			}
		},
	})

	bus.Publish(context.Background(), event.Envelope{
		EventID: "var-1", EventType: "inventory.variance.detected",
		OrgID: "org-1", LocationID: "loc-1",
	})

	queue := svc.GetQueue("org-1", "loc-1")
	require.Len(t, queue, 1)
	assert.Equal(t, "High variance detected: Ground Beef at 15%", queue[0].Title)
	assert.Equal(t, alerting.SeverityWarning, queue[0].Severity)
}

func TestService_AlertCreatedEvent(t *testing.T) {
	bus := event.New()
	svc := alerting.New(bus)

	var alertCreated bool
	bus.Subscribe("alerting.alert.created", func(ctx context.Context, env event.Envelope) error {
		alertCreated = true
		return nil
	})

	svc.RegisterDefaultRules()

	bus.Publish(context.Background(), event.Envelope{
		EventID: "e1", EventType: "adapter.error",
		OrgID: "org-1", LocationID: "loc-1",
	})

	assert.True(t, alertCreated, "should fire alerting.alert.created event")
}
