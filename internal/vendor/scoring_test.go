package vendor

import (
	"testing"
)

func TestOverallScore(t *testing.T) {
	score := calculateOverallScore(80, 70, 60, 90)
	// 80*0.30 + 70*0.25 + 60*0.25 + 90*0.20 = 24+17.5+15+18 = 74.5
	if score != 74.5 {
		t.Errorf("expected 74.5, got %.1f", score)
	}
}

func TestOverallScoreAllPerfect(t *testing.T) {
	score := calculateOverallScore(100, 100, 100, 100)
	if score != 100.0 {
		t.Errorf("expected 100.0, got %.2f", score)
	}
}

func TestOverallScoreAllZero(t *testing.T) {
	score := calculateOverallScore(0, 0, 0, 0)
	if score != 0.0 {
		t.Errorf("expected 0.0, got %.2f", score)
	}
}

func TestOverallScoreWeightsSum(t *testing.T) {
	// Verify weights sum to 1.0: 0.30+0.25+0.25+0.20 = 1.0
	// If only one dimension is 100, the overall should equal that weight * 100.
	priceOnly := calculateOverallScore(100, 0, 0, 0)
	if priceOnly != 30.0 {
		t.Errorf("price weight expected 30.0, got %.2f", priceOnly)
	}

	deliveryOnly := calculateOverallScore(0, 100, 0, 0)
	if deliveryOnly != 25.0 {
		t.Errorf("delivery weight expected 25.0, got %.2f", deliveryOnly)
	}

	qualityOnly := calculateOverallScore(0, 0, 100, 0)
	if qualityOnly != 25.0 {
		t.Errorf("quality weight expected 25.0, got %.2f", qualityOnly)
	}

	accuracyOnly := calculateOverallScore(0, 0, 0, 100)
	if accuracyOnly != 20.0 {
		t.Errorf("accuracy weight expected 20.0, got %.2f", accuracyOnly)
	}
}
