package customer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// --- CLV ---

func TestCalculateCLV_ZeroVisits(t *testing.T) {
	clv := CalculateCLV(1500, 0, time.Now().AddDate(-1, 0, 0), "low")
	assert.Equal(t, 0.0, clv, "zero visits must return 0")
}

func TestCalculateCLV_Basic(t *testing.T) {
	// avgCheck=1500 cents ($15), 12 visits, low risk (24 month lifespan).
	// monthsActive = int(hours/730) — varies slightly with calendar month lengths,
	// so we verify positivity and the formula invariants rather than a hard value.
	firstVisit := time.Now().AddDate(0, -6, 0)
	clv := CalculateCLV(1500, 12, firstVisit, "low")
	assert.Greater(t, clv, 0.0, "CLV must be positive")
	assert.Less(t, clv, 2000.0, "CLV must be a plausible dollar amount")
}

func TestCalculateCLV_CriticalChurn(t *testing.T) {
	// avgCheck=5000 cents ($50), 4 visits, critical risk → lifespan=3 months.
	// CLV is strictly lower than equivalent low-risk guest; positive and bounded.
	firstVisit := time.Now().AddDate(0, -4, 0)
	clv := CalculateCLV(5000, 4, firstVisit, "critical")
	assert.Greater(t, clv, 0.0, "CLV must be positive for critical guest")
	assert.Less(t, clv, 500.0, "CLV must be reduced by critical churn lifespan")

	// Critical lifespan (3) < high (9) < medium (18) < low (24): lower CLV
	clvLow := CalculateCLV(5000, 4, firstVisit, "low")
	assert.Less(t, clv, clvLow, "critical churn CLV must be lower than low-risk CLV")
}

func TestCalculateCLV_HighChurn(t *testing.T) {
	firstVisit := time.Now().AddDate(0, -3, 0)
	clv := CalculateCLV(2000, 6, firstVisit, "high")
	assert.Greater(t, clv, 0.0, "CLV should be positive")
}

func TestCalculateCLV_MediumChurn(t *testing.T) {
	firstVisit := time.Now().AddDate(-1, 0, 0)
	clv := CalculateCLV(3000, 12, firstVisit, "medium")
	// freq = 12/12 = 1, lifespan = 18
	// CLV = 30 * 1 * 18 * 0.65 = 351
	assert.InDelta(t, 351.0, clv, 5.0)
}

func TestCalculateCLV_NewGuest(t *testing.T) {
	// First visit was today — monthsActive floors to 1
	clv := CalculateCLV(2500, 1, time.Now(), "low")
	// freq = 1/1 = 1, lifespan = 24
	// CLV = 25 * 1 * 24 * 0.65 = 390
	assert.InDelta(t, 390.0, clv, 5.0)
}

// --- Segmentation ---

func TestSegmentGuest_Champion(t *testing.T) {
	seg := SegmentGuest(3, 25, 60000) // r=5, f=5, m=5
	assert.Equal(t, "champion", seg)
}

func TestSegmentGuest_Loyal(t *testing.T) {
	seg := SegmentGuest(45, 15, 25000) // r=2, f=4, m=4
	assert.Equal(t, "loyal", seg)
}

func TestSegmentGuest_PotentialLoyalist(t *testing.T) {
	seg := SegmentGuest(5, 3, 3000) // r=5, f=2, m=1
	assert.Equal(t, "potential_loyalist", seg)
}

func TestSegmentGuest_AtRisk(t *testing.T) {
	seg := SegmentGuest(90, 8, 15000) // r=1, f=3, m=3
	assert.Equal(t, "at_risk", seg)
}

func TestSegmentGuest_Lapsed(t *testing.T) {
	seg := SegmentGuest(200, 1, 1000) // r=1 → lapsed
	assert.Equal(t, "lapsed", seg)
}

func TestSegmentGuest_New(t *testing.T) {
	seg := SegmentGuest(5, 1, 1500) // r=5, visits≤2
	assert.Equal(t, "new", seg)
}

func TestSegmentGuest_Regular(t *testing.T) {
	seg := SegmentGuest(20, 4, 8000) // r=3, f=2, m=2, composite=7
	assert.Equal(t, "regular", seg)
}

// --- Churn Prediction ---

func TestPredictChurn_TooFewVisits(t *testing.T) {
	risk, prob := PredictChurn([]time.Time{time.Now()})
	assert.Equal(t, "low", risk)
	assert.Equal(t, 0.1, prob)
}

func TestPredictChurn_Low(t *testing.T) {
	// Visit every 7 days, last visit was yesterday → well within interval
	now := time.Now()
	dates := []time.Time{
		now.AddDate(0, 0, -28),
		now.AddDate(0, 0, -21),
		now.AddDate(0, 0, -14),
		now.AddDate(0, 0, -7),
		now.AddDate(0, 0, -1),
	}
	risk, prob := PredictChurn(dates)
	assert.Equal(t, "low", risk)
	assert.Less(t, prob, 0.2)
}

func TestPredictChurn_Medium(t *testing.T) {
	// Visit every 7 days, last visit 10 days ago (overdue 3 days = 0.43x avg)
	now := time.Now()
	dates := []time.Time{
		now.AddDate(0, 0, -31),
		now.AddDate(0, 0, -24),
		now.AddDate(0, 0, -17),
		now.AddDate(0, 0, -10),
	}
	risk, prob := PredictChurn(dates)
	assert.Equal(t, "medium", risk)
	assert.InDelta(t, 0.25, prob, 0.01)
}

func TestPredictChurn_High(t *testing.T) {
	// Visit every 10 days, last visit 20 days ago (overdue 10 days = 1.0x avg)
	now := time.Now()
	dates := []time.Time{
		now.AddDate(0, 0, -50),
		now.AddDate(0, 0, -40),
		now.AddDate(0, 0, -30),
		now.AddDate(0, 0, -20),
	}
	risk, prob := PredictChurn(dates)
	assert.Equal(t, "high", risk)
	assert.InDelta(t, 0.60, prob, 0.01)
}

func TestPredictChurn_Critical(t *testing.T) {
	// Visit every 7 days, last visit 90 days ago (severely overdue)
	now := time.Now()
	dates := []time.Time{
		now.AddDate(0, 0, -118),
		now.AddDate(0, 0, -111),
		now.AddDate(0, 0, -104),
		now.AddDate(0, 0, -90),
	}
	risk, prob := PredictChurn(dates)
	assert.Equal(t, "critical", risk)
	assert.InDelta(t, 0.90, prob, 0.01)
}
