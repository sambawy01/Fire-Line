package menu

import (
	"testing"
)

func TestClassifyItem(t *testing.T) {
	tests := []struct {
		margin, velocity, complexity, satisfaction, strategic float64
		expected                                              string
	}{
		{80, 80, 60, 70, 50, "powerhouse"},
		{80, 30, 40, 70, 50, "hidden_gem"},
		{30, 80, 60, 50, 50, "crowd_pleaser"},
		{60, 60, 80, 50, 50, "workhorse"},
		{80, 50, 20, 50, 50, "complex_star"},
		{70, 20, 50, 50, 50, "declining_star"},
		{30, 30, 50, 50, 50, "underperformer"},
		{30, 30, 50, 50, 80, "strategic_anchor"},
	}
	for _, tt := range tests {
		got := classifyItem(tt.margin, tt.velocity, tt.complexity, tt.satisfaction, tt.strategic)
		if got != tt.expected {
			t.Errorf("classify(%.0f,%.0f,%.0f,%.0f,%.0f) = %q, want %q",
				tt.margin, tt.velocity, tt.complexity, tt.satisfaction, tt.strategic, got, tt.expected)
		}
	}
}

func TestNormalizeScore(t *testing.T) {
	if s := normalizeScore(50, 100); s != 50.0 {
		t.Errorf("expected 50, got %.1f", s)
	}
	if s := normalizeScore(150, 100); s != 100.0 {
		t.Errorf("expected 100 (capped), got %.1f", s)
	}
	if s := normalizeScore(0, 100); s != 0.0 {
		t.Errorf("expected 0, got %.1f", s)
	}
	if s := normalizeScore(50, 0); s != 0.0 {
		t.Errorf("expected 0 (div by zero), got %.1f", s)
	}
}
