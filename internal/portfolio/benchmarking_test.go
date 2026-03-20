package portfolio

import (
	"math"
	"testing"
)

func TestPercentileRank(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		value  float64
		want   float64
	}{
		{
			name:   "lowest value returns 0",
			values: []float64{10, 20, 30, 40, 50},
			value:  10,
			want:   0,
		},
		{
			name:   "highest value returns 80",
			values: []float64{10, 20, 30, 40, 50},
			value:  50,
			want:   80,
		},
		{
			name:   "middle value",
			values: []float64{10, 20, 30, 40, 50},
			value:  30,
			want:   40,
		},
		{
			name:   "empty slice returns 0",
			values: []float64{},
			value:  100,
			want:   0,
		},
		{
			name:   "single element returns 0",
			values: []float64{42},
			value:  42,
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PercentileRank(tt.values, tt.value)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("PercentileRank(%v, %v) = %v, want %v", tt.values, tt.value, got, tt.want)
			}
		})
	}
}

func TestQuartiles(t *testing.T) {
	tests := []struct {
		name       string
		values     []float64
		wantMedian float64
		wantQ1     float64
		wantQ3     float64
	}{
		{
			name:       "even count",
			values:     []float64{1, 2, 3, 4, 5, 6, 7, 8},
			wantMedian: 4.5,
			wantQ1:     2.75,
			wantQ3:     6.25,
		},
		{
			name:       "odd count",
			values:     []float64{1, 3, 5, 7, 9},
			wantMedian: 5,
			wantQ1:     3, // linear interpolation: index 1.0 → sorted[1] = 3
			wantQ3:     7, // linear interpolation: index 3.0 → sorted[3] = 7
		},
		{
			name:       "empty returns zeros",
			values:     []float64{},
			wantMedian: 0,
			wantQ1:     0,
			wantQ3:     0,
		},
		{
			name:       "all equal values",
			values:     []float64{5, 5, 5, 5},
			wantMedian: 5,
			wantQ1:     5,
			wantQ3:     5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			median, q1, q3 := Quartiles(tt.values)
			if math.Abs(median-tt.wantMedian) > 0.01 {
				t.Errorf("median = %v, want %v", median, tt.wantMedian)
			}
			if math.Abs(q1-tt.wantQ1) > 0.01 {
				t.Errorf("q1 = %v, want %v", q1, tt.wantQ1)
			}
			if math.Abs(q3-tt.wantQ3) > 0.01 {
				t.Errorf("q3 = %v, want %v", q3, tt.wantQ3)
			}
		})
	}
}

func TestDetectOutliersLogic(t *testing.T) {
	// Test the pure IQR outlier detection math.
	// Benchmarks: 4 normal locations + 1 extreme outlier.
	benchmarks := []LocationBenchmark{
		{LocationID: "loc1", LocationName: "Normal A", Revenue: 100_000},
		{LocationID: "loc2", LocationName: "Normal B", Revenue: 105_000},
		{LocationID: "loc3", LocationName: "Normal C", Revenue: 110_000},
		{LocationID: "loc4", LocationName: "Normal D", Revenue: 108_000},
		{LocationID: "loc5", LocationName: "Outlier", Revenue: 300_000}, // way above
	}

	revenues := make([]float64, len(benchmarks))
	for i, b := range benchmarks {
		revenues[i] = float64(b.Revenue)
	}

	_, q1, q3 := Quartiles(revenues)
	iqr := q3 - q1
	upper := q3 + 1.5*iqr

	var found bool
	for _, b := range benchmarks {
		v := float64(b.Revenue)
		if v > upper {
			found = true
			if b.LocationID != "loc5" {
				t.Errorf("expected loc5 as outlier, got %s", b.LocationID)
			}
		}
	}

	if !found {
		t.Error("expected to find at least one outlier above upper fence")
	}
}

func TestPercentileRankDistribution(t *testing.T) {
	// Verify that percentile ranks span the expected range for a uniform distribution.
	values := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

	first := PercentileRank(values, 10)
	last := PercentileRank(values, 100)

	if first != 0 {
		t.Errorf("lowest value should have percentile 0, got %v", first)
	}
	if last < 80 || last > 100 {
		t.Errorf("highest value should have percentile near 90, got %v", last)
	}

	// Verify monotonic ordering
	prev := -1.0
	for _, v := range values {
		p := PercentileRank(values, v)
		if p < prev {
			t.Errorf("percentile ranks should be non-decreasing: %v < %v at value %v", p, prev, v)
		}
		prev = p
	}
}
