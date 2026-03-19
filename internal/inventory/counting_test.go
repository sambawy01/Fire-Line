package inventory

import (
	"testing"
)

func TestCountTypeValidation(t *testing.T) {
	valid := []string{"full", "spot_check"}
	invalid := []string{"partial", "", "FULL"}

	for _, v := range valid {
		if !validCountType(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}
	for _, v := range invalid {
		if validCountType(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func TestCountProgress(t *testing.T) {
	lines := []CountLine{
		{IngredientID: "a", CountedQty: ptrFloat(10.0)},
		{IngredientID: "b", CountedQty: nil},
		{IngredientID: "c", CountedQty: ptrFloat(5.0)},
	}
	counted, total := countProgress(lines)
	if counted != 2 || total != 3 {
		t.Errorf("expected 2/3, got %d/%d", counted, total)
	}
}

func ptrFloat(f float64) *float64 { return &f }
