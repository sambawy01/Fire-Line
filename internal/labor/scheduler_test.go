package labor

import "testing"

func TestOvertimeRiskThresholds(t *testing.T) {
	tests := []struct {
		hours    float64
		expected string
	}{
		{37.0, ""},        // no risk
		{38.5, "warning"},
		{40.0, "warning"}, // exactly 40 is warning
		{40.5, "critical"},
	}
	for _, tt := range tests {
		got := classifyOvertimeRisk(tt.hours)
		if got != tt.expected {
			t.Errorf("classifyOvertimeRisk(%.1f) = %q, want %q", tt.hours, got, tt.expected)
		}
	}
}

func TestEmployeeAvailable(t *testing.T) {
	tests := []struct {
		name         string
		availability map[string]any
		dowName      string
		blockTime    string
		want         bool
	}{
		{
			name:         "no availability constraint",
			availability: map[string]any{},
			dowName:      "Monday",
			blockTime:    "09:00",
			want:         true,
		},
		{
			name:         "bool true means available",
			availability: map[string]any{"Monday": true},
			dowName:      "Monday",
			blockTime:    "09:00",
			want:         true,
		},
		{
			name:         "bool false means unavailable",
			availability: map[string]any{"Monday": false},
			dowName:      "Monday",
			blockTime:    "09:00",
			want:         false,
		},
		{
			name: "time window — inside",
			availability: map[string]any{
				"Tuesday": map[string]any{"start": "09:00", "end": "17:00"},
			},
			dowName:   "Tuesday",
			blockTime: "12:00",
			want:      true,
		},
		{
			name: "time window — before start",
			availability: map[string]any{
				"Tuesday": map[string]any{"start": "09:00", "end": "17:00"},
			},
			dowName:   "Tuesday",
			blockTime: "08:30",
			want:      false,
		},
		{
			name: "time window — at end (exclusive)",
			availability: map[string]any{
				"Tuesday": map[string]any{"start": "09:00", "end": "17:00"},
			},
			dowName:   "Tuesday",
			blockTime: "17:00",
			want:      false,
		},
		{
			name: "different day not constrained",
			availability: map[string]any{
				"Tuesday": map[string]any{"start": "09:00", "end": "17:00"},
			},
			dowName:   "Wednesday",
			blockTime: "09:00",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := employeeAvailable(tt.availability, tt.dowName, tt.blockTime)
			if got != tt.want {
				t.Errorf("employeeAvailable(%v, %q, %q) = %v, want %v",
					tt.availability, tt.dowName, tt.blockTime, got, tt.want)
			}
		})
	}
}
