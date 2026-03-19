package inventory

import (
	"testing"
)

func TestComputeVarianceFlag(t *testing.T) {
	tests := []struct {
		ordered  float64
		received float64
		expected string
	}{
		{49.0, 48.02, "exact"},
		{49.0, 48.01, "short"},
		{49.0, 49.98, "exact"},
		{49.0, 49.99, "over"},
		{49.0, 0.0, "not_received"},
		{10.0, 10.0, "exact"},
		{10.0, 9.8, "exact"},
		{10.0, 9.79, "short"},
		{10.0, 10.2, "exact"},
		{10.0, 10.21, "over"},
		{0.0, 0.0, "exact"},
	}
	for _, tt := range tests {
		got := computeVarianceFlag(tt.ordered, tt.received)
		if got != tt.expected {
			t.Errorf("computeVarianceFlag(%.2f, %.2f) = %q, want %q", tt.ordered, tt.received, got, tt.expected)
		}
	}
}

func TestComputeAvgDailyUsage(t *testing.T) {
	got := computeAvgDailyUsage(100.0, 10)
	if got != 10.0 {
		t.Errorf("expected 10.0, got %.2f", got)
	}
	got = computeAvgDailyUsage(0, 0)
	if got != 0.0 {
		t.Errorf("expected 0.0, got %.2f", got)
	}
}

func TestEffectiveReorderPoint(t *testing.T) {
	got := effectiveReorderPoint(25.0, 1, 8.0)
	if got != 25.0 {
		t.Errorf("expected 25.0, got %.2f", got)
	}
	got = effectiveReorderPoint(25.0, 3, 10.0)
	if got != 30.0 {
		t.Errorf("expected 30.0, got %.2f", got)
	}
}
