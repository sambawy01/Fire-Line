package inventory

import (
	"math"
	"testing"
)

func TestCategorizeCauses_UnrecordedWaste(t *testing.T) {
	// Shortage that is largely explained by logged waste → high unrecorded_waste probability
	sig := VarianceSignals{
		VarianceQty:      -10.0,
		TheoreticalUsage: 50.0,
		LoggedWasteQty:   8.0, // explains 80% of the shortage
	}
	causes := categorizeCauses(sig)

	prob, ok := causes["unrecorded_waste"]
	if !ok {
		t.Fatal("expected unrecorded_waste cause to be present")
	}
	// 8/10 = 0.8
	if math.Abs(prob-0.8) > 0.01 {
		t.Errorf("expected unrecorded_waste ~0.8, got %f", prob)
	}
}

func TestCategorizeCauses_PortioningSignal(t *testing.T) {
	// Shortage with portioning flag and no waste logged → over_portioning should appear
	sig := VarianceSignals{
		VarianceQty:      -6.0,
		TheoreticalUsage: 40.0,
		LoggedWasteQty:   0.0,
		PortioningFlag:   true,
	}
	causes := categorizeCauses(sig)

	_, ok := causes["over_portioning"]
	if !ok {
		t.Fatal("expected over_portioning cause to be present")
	}
}

func TestCategorizeCauses_SmallVarianceMeasurementError(t *testing.T) {
	// Small variance (< 5%) should flag measurement_error
	sig := VarianceSignals{
		VarianceQty:      1.0,
		TheoreticalUsage: 100.0, // 1% variance
		LoggedWasteQty:   0.0,
	}
	causes := categorizeCauses(sig)

	prob, ok := causes["measurement_error"]
	if !ok {
		t.Fatal("expected measurement_error cause to be present")
	}
	if prob < 0.5 {
		t.Errorf("expected measurement_error probability >= 0.5, got %f", prob)
	}
}

func TestCategorizeCauses_ZeroTheoretical(t *testing.T) {
	sig := VarianceSignals{
		VarianceQty:      -5.0,
		TheoreticalUsage: 0.0,
	}
	causes := categorizeCauses(sig)

	if _, ok := causes["unknown"]; !ok {
		t.Fatal("expected unknown cause when theoretical usage is zero")
	}
}

func TestClassifySeverity_Boundaries(t *testing.T) {
	cases := []struct {
		pct      float64
		expected string
	}{
		{0.0, "info"},
		{0.04, "info"},
		{-0.04, "info"},
		{0.05, "warning"},
		{-0.05, "warning"},
		{0.10, "warning"},
		{0.14, "warning"},
		{0.15, "critical"},
		{-0.15, "critical"},
		{0.50, "critical"},
	}

	for _, tc := range cases {
		got := classifySeverity(tc.pct)
		if got != tc.expected {
			t.Errorf("classifySeverity(%f): expected %q, got %q", tc.pct, tc.expected, got)
		}
	}
}

func TestNormalize(t *testing.T) {
	if got := normalize(10, 100); math.Abs(got-0.1) > 0.001 {
		t.Errorf("expected 0.1, got %f", got)
	}
	if got := normalize(5, 0); got != 0 {
		t.Errorf("expected 0 for zero theoretical, got %f", got)
	}
	if got := normalize(-5, 50); math.Abs(got-(-0.1)) > 0.001 {
		t.Errorf("expected -0.1, got %f", got)
	}
}
