package operations

import "testing"

func TestDefaultTicketTimes(t *testing.T) {
	defaults := defaultStationTimes()
	if defaults["grill"] != 420 {
		t.Error("grill should be 420")
	}
	if defaults["fryer"] != 300 {
		t.Error("fryer should be 300")
	}
}

func TestLoadPercentage(t *testing.T) {
	pct := loadPct(3, 4)
	if pct != 75.0 {
		t.Errorf("expected 75%%, got %.1f", pct)
	}
	pct = loadPct(0, 4)
	if pct != 0.0 {
		t.Errorf("expected 0%%, got %.1f", pct)
	}
}
