package onboarding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ─── Concept Inference ───────────────────────────────────────────────────────

func TestInferConceptFromAvgCheck_QuickService(t *testing.T) {
	assert.Equal(t, "quick_service", inferConceptFromAvgCheck(800))   // $8.00
	assert.Equal(t, "quick_service", inferConceptFromAvgCheck(0))     // $0 — no data
	assert.Equal(t, "quick_service", inferConceptFromAvgCheck(1199))  // just under $12
}

func TestInferConceptFromAvgCheck_FastCasual(t *testing.T) {
	assert.Equal(t, "fast_casual", inferConceptFromAvgCheck(1200))  // exactly $12
	assert.Equal(t, "fast_casual", inferConceptFromAvgCheck(1800))  // $18
	assert.Equal(t, "fast_casual", inferConceptFromAvgCheck(2499))  // just under $25
}

func TestInferConceptFromAvgCheck_CasualDining(t *testing.T) {
	assert.Equal(t, "casual_dining", inferConceptFromAvgCheck(2500)) // exactly $25
	assert.Equal(t, "casual_dining", inferConceptFromAvgCheck(3500)) // $35
	assert.Equal(t, "casual_dining", inferConceptFromAvgCheck(4999)) // just under $50
}

func TestInferConceptFromAvgCheck_UpscaleCasual(t *testing.T) {
	assert.Equal(t, "upscale_casual", inferConceptFromAvgCheck(5000))  // exactly $50
	assert.Equal(t, "upscale_casual", inferConceptFromAvgCheck(7500))  // $75
	assert.Equal(t, "upscale_casual", inferConceptFromAvgCheck(9999))  // just under $100
}

func TestInferConceptFromAvgCheck_FineDining(t *testing.T) {
	assert.Equal(t, "fine_dining", inferConceptFromAvgCheck(10000)) // exactly $100
	assert.Equal(t, "fine_dining", inferConceptFromAvgCheck(25000)) // $250 tasting menu
}

// ─── Module Recommendations ──────────────────────────────────────────────────

func TestRecommendModules_ReduceWaste(t *testing.T) {
	mods := recommendModules([]string{"reduce_waste"})
	assert.Contains(t, mods, "inventory")
	assert.Contains(t, mods, "menu_scoring")
}

func TestRecommendModules_BoostRevenue(t *testing.T) {
	mods := recommendModules([]string{"boost_revenue"})
	assert.Contains(t, mods, "financial")
	assert.Contains(t, mods, "marketing")
	assert.Contains(t, mods, "menu_scoring")
}

func TestRecommendModules_LaborEfficiency(t *testing.T) {
	mods := recommendModules([]string{"labor_efficiency"})
	assert.Contains(t, mods, "labor")
	assert.Contains(t, mods, "scheduling")
}

func TestRecommendModules_FoodCostControl(t *testing.T) {
	mods := recommendModules([]string{"food_cost_control"})
	assert.Contains(t, mods, "inventory")
	assert.Contains(t, mods, "financial")
	assert.Contains(t, mods, "vendor")
}

func TestRecommendModules_GuestExperience(t *testing.T) {
	mods := recommendModules([]string{"guest_experience"})
	assert.Contains(t, mods, "customers")
	assert.Contains(t, mods, "operations")
}

func TestRecommendModules_GrowthInsights(t *testing.T) {
	mods := recommendModules([]string{"growth_insights"})
	assert.Contains(t, mods, "reporting")
	assert.Contains(t, mods, "portfolio")
}

func TestRecommendModules_Deduplication(t *testing.T) {
	// Both reduce_waste and food_cost_control recommend "inventory" — should appear once
	mods := recommendModules([]string{"reduce_waste", "food_cost_control"})
	count := 0
	for _, m := range mods {
		if m == "inventory" {
			count++
		}
	}
	assert.Equal(t, 1, count, "inventory should appear exactly once")
}

func TestRecommendModules_Empty(t *testing.T) {
	mods := recommendModules([]string{})
	assert.NotNil(t, mods)
	assert.Len(t, mods, 0)
}

func TestRecommendModules_MultiplePriorities(t *testing.T) {
	mods := recommendModules([]string{"reduce_waste", "boost_revenue", "labor_efficiency"})
	assert.Contains(t, mods, "inventory")
	assert.Contains(t, mods, "financial")
	assert.Contains(t, mods, "labor")
	assert.Contains(t, mods, "marketing")
}

// ─── Checklist Generation ────────────────────────────────────────────────────

func TestGenerateChecklistItems_AlwaysHasBaseItems(t *testing.T) {
	items := generateChecklistItems("casual_dining", []string{})
	titles := make([]string, len(items))
	for i, it := range items {
		titles[i] = it.Title
	}
	assert.Contains(t, titles, "Complete your restaurant profile")
	assert.Contains(t, titles, "Invite your management team")
	assert.Contains(t, titles, "Connect your POS system")
	assert.Contains(t, titles, "Import your menu")
}

func TestGenerateChecklistItems_ConceptSpecific_FineDining(t *testing.T) {
	items := generateChecklistItems("fine_dining", []string{})
	titles := make([]string, len(items))
	for i, it := range items {
		titles[i] = it.Title
	}
	assert.Contains(t, titles, "Configure wine list and pairings")
	assert.Contains(t, titles, "Enable chef's tasting menu management")
}

func TestGenerateChecklistItems_ConceptSpecific_QuickService(t *testing.T) {
	items := generateChecklistItems("quick_service", []string{})
	titles := make([]string, len(items))
	for i, it := range items {
		titles[i] = it.Title
	}
	assert.Contains(t, titles, "Set up combo/value meal pricing")
}

func TestGenerateChecklistItems_ModuleItems_Inventory(t *testing.T) {
	items := generateChecklistItems("casual_dining", []string{"inventory"})
	titles := make([]string, len(items))
	for i, it := range items {
		titles[i] = it.Title
	}
	assert.Contains(t, titles, "Set PAR levels for key ingredients")
	assert.Contains(t, titles, "Run your first inventory count")
}

func TestGenerateChecklistItems_ModuleItems_Labor(t *testing.T) {
	items := generateChecklistItems("casual_dining", []string{"labor", "scheduling"})
	titles := make([]string, len(items))
	for i, it := range items {
		titles[i] = it.Title
	}
	assert.Contains(t, titles, "Import your staff roster")
	assert.Contains(t, titles, "Build your first schedule")
}

func TestGenerateChecklistItems_AllHaveRequiredFields(t *testing.T) {
	items := generateChecklistItems("casual_dining", []string{"inventory", "financial", "labor"})
	for _, it := range items {
		assert.NotEmpty(t, it.Title, "every item must have a title")
		assert.NotEmpty(t, it.Category, "every item must have a category")
	}
}

func TestGenerateChecklistItems_MinimumCount(t *testing.T) {
	// Base (5) + concept (2 for casual_dining) + module items (2 for inventory)
	items := generateChecklistItems("casual_dining", []string{"inventory"})
	assert.GreaterOrEqual(t, len(items), 7)
}
