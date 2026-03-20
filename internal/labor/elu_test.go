package labor

import (
	"testing"
)

func TestPointsTrend(t *testing.T) {
	tests := []struct {
		current      float64
		sevenDaysAgo float64
		expected     string
	}{
		{50.0, 40.0, "up"},    // diff > 5
		{40.0, 50.0, "down"},  // diff < -5
		{42.0, 40.0, "stable"}, // diff within ±5
		{0, 0, "stable"},
	}
	for _, tt := range tests {
		got := computePointsTrend(tt.current, tt.sevenDaysAgo)
		if got != tt.expected {
			t.Errorf("trend(%.1f, %.1f) = %q, want %q", tt.current, tt.sevenDaysAgo, got, tt.expected)
		}
	}
}
