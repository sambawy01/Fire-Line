package operations

import "testing"

func TestClassifyOverload(t *testing.T) {
	tests := []struct {
		pct  float64
		want string
	}{
		{80, "normal"},
		{85, "elevated"},
		{90, "elevated"},
		{95, "elevated"},
		{96, "critical"},
		{100, "critical"},
	}
	for _, tt := range tests {
		if got := classifyOverload(tt.pct); got != tt.want {
			t.Errorf("classifyOverload(%.0f) = %q, want %q", tt.pct, got, tt.want)
		}
	}
}

func TestComputeHealthScore(t *testing.T) {
	// 80*0.25 + 90*0.25 + 70*0.20 + 60*0.15 + 50*0.15 = 20+22.5+14+9+7.5 = 73
	got := computeHealthScore(80, 90, 70, 60, 50)
	if got != 73.0 {
		t.Errorf("expected 73, got %.1f", got)
	}
}

func TestComputePriority(t *testing.T) {
	got := computePriority(0.9, 0.8, 1.0, 0.5)
	// 0.9*0.35 + 0.8*0.25 + 1.0*0.20 + 0.5*0.20 = 0.315+0.2+0.2+0.1 = 0.815
	if got < 0.81 || got > 0.82 {
		t.Errorf("expected ~0.815, got %.3f", got)
	}
}

func TestClassifyHealth(t *testing.T) {
	tests := []struct {
		score float64
		want  string
	}{
		{95, "excellent"},
		{80, "good"},
		{65, "fair"},
		{45, "poor"},
		{30, "critical"},
	}
	for _, tt := range tests {
		if got := classifyHealth(tt.score); got != tt.want {
			t.Errorf("classifyHealth(%.0f) = %q, want %q", tt.score, got, tt.want)
		}
	}
}
