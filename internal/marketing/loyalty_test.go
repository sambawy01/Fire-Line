package marketing

import (
	"testing"
)

func TestCalculateTier(t *testing.T) {
	tests := []struct {
		points float64
		want   string
	}{
		{0, "bronze"},
		{499, "bronze"},
		{500, "silver"},
		{1999, "silver"},
		{2000, "gold"},
		{4999, "gold"},
		{5000, "platinum"},
	}
	for _, tt := range tests {
		if got := calculateTier(tt.points); got != tt.want {
			t.Errorf("calculateTier(%.0f) = %q, want %q", tt.points, got, tt.want)
		}
	}
}
