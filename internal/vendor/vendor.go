package vendor

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// Service provides vendor intelligence capabilities.
type Service struct {
	pool *pgxpool.Pool
	bus  *event.Bus
}

// New creates a new vendor intelligence service.
func New(pool *pgxpool.Pool, bus *event.Bus) *Service {
	return &Service{pool: pool, bus: bus}
}

// VendorAnalysis holds computed analytics for a single vendor.
type VendorAnalysis struct {
	VendorName     string  `json:"vendor_name"`
	ItemsSupplied  int     `json:"items_supplied"`
	TotalSpend     int64   `json:"total_spend"`
	SpendPct       float64 `json:"spend_pct"`
	AvgCostPerItem int64   `json:"avg_cost_per_item"`
	Score          int     `json:"score"`
}

// VendorSummary holds location-wide rollup KPIs across all vendors.
type VendorSummary struct {
	TotalVendors      int     `json:"total_vendors"`
	TotalSpend        int64   `json:"total_spend"`
	TopVendorName     string  `json:"top_vendor_name"`
	TopVendorPct      float64 `json:"top_vendor_pct"`
	AvgItemsPerVendor float64 `json:"avg_items_per_vendor"`
}

// rawVendor is an internal accumulator for query results before scoring.
type rawVendor struct {
	vendorName    string
	itemsSupplied int
	totalSpend    int64
	deviationPct  float64 // avg price deviation from org baseline
}

// GetVendors queries vendor spend and item coverage derived from existing
// ingredient_location_configs, recipe_explosion, and check_items tables.
func (s *Service) GetVendors(ctx context.Context, orgID, locationID string, from, to time.Time) ([]VendorAnalysis, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	vendorMap := make(map[string]*rawVendor)

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// --- Query 1: vendor spend and item coverage from sales data ---
		rows, err := tx.Query(tenantCtx,
			`SELECT ilc.vendor_name,
			        COUNT(DISTINCT ilc.ingredient_id) AS items_supplied,
			        COALESCE(SUM(
			            ci_qty.total_qty * re.quantity_per_unit * COALESCE(ilc.local_cost_per_unit, i.cost_per_unit)
			        )::BIGINT, 0) AS total_spend
			 FROM ingredient_location_configs ilc
			 JOIN ingredients i ON i.ingredient_id = ilc.ingredient_id AND i.org_id = ilc.org_id
			 LEFT JOIN recipe_explosion re ON re.ingredient_id = ilc.ingredient_id AND re.org_id = ilc.org_id
			 LEFT JOIN (
			     SELECT ci.menu_item_id, SUM(ci.quantity) AS total_qty
			     FROM check_items ci
			     JOIN checks c ON c.check_id = ci.check_id AND c.org_id = ci.org_id
			     WHERE c.location_id = $1 AND c.status = 'closed' AND ci.voided_at IS NULL
			       AND c.closed_at >= $2 AND c.closed_at < $3 AND ci.menu_item_id IS NOT NULL
			     GROUP BY ci.menu_item_id
			 ) ci_qty ON ci_qty.menu_item_id = re.menu_item_id
			 WHERE ilc.location_id = $1 AND ilc.vendor_name IS NOT NULL AND ilc.vendor_name != ''
			 GROUP BY ilc.vendor_name
			 ORDER BY total_spend DESC`,
			locationID, from, to,
		)
		if err != nil {
			return fmt.Errorf("query vendor spend: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var vendorName string
			var itemsSupplied int
			var totalSpend int64
			if err := rows.Scan(&vendorName, &itemsSupplied, &totalSpend); err != nil {
				return fmt.Errorf("scan vendor spend row: %w", err)
			}
			vendorMap[vendorName] = &rawVendor{
				vendorName:    vendorName,
				itemsSupplied: itemsSupplied,
				totalSpend:    totalSpend,
			}
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate vendor spend rows: %w", err)
		}

		// --- Query 2: per-vendor price deviation from org baseline ---
		devRows, err := tx.Query(tenantCtx,
			`SELECT ilc.vendor_name,
			        AVG(CASE WHEN i.cost_per_unit > 0
			            THEN ABS(COALESCE(ilc.local_cost_per_unit, i.cost_per_unit) - i.cost_per_unit)::FLOAT / i.cost_per_unit * 100
			            ELSE 0 END) AS avg_deviation_pct
			 FROM ingredient_location_configs ilc
			 JOIN ingredients i ON i.ingredient_id = ilc.ingredient_id AND i.org_id = ilc.org_id
			 WHERE ilc.location_id = $1 AND ilc.vendor_name IS NOT NULL AND ilc.vendor_name != ''
			 GROUP BY ilc.vendor_name`,
			locationID,
		)
		if err != nil {
			return fmt.Errorf("query vendor deviation: %w", err)
		}
		defer devRows.Close()

		for devRows.Next() {
			var vendorName string
			var avgDeviationPct float64
			if err := devRows.Scan(&vendorName, &avgDeviationPct); err != nil {
				return fmt.Errorf("scan vendor deviation row: %w", err)
			}
			if v, ok := vendorMap[vendorName]; ok {
				v.deviationPct = avgDeviationPct
			}
		}
		if err := devRows.Err(); err != nil {
			return fmt.Errorf("iterate vendor deviation rows: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(vendorMap) == 0 {
		return []VendorAnalysis{}, nil
	}

	// Collect raw vendors in the deterministic order from query 1 (by total_spend DESC).
	// Rebuild ordered slice from the map preserving spend order requires re-sorting here
	// since map iteration is non-deterministic.
	vendors := make([]*rawVendor, 0, len(vendorMap))
	for _, v := range vendorMap {
		vendors = append(vendors, v)
	}

	// Sort deterministically: highest spend first, then vendor name for ties.
	sortVendors(vendors)

	return computeVendorAnalysis(vendors), nil
}

// sortVendors sorts vendors by total_spend DESC, then vendor_name ASC for ties.
func sortVendors(vendors []*rawVendor) {
	n := len(vendors)
	for i := 1; i < n; i++ {
		for j := i; j > 0; j-- {
			a, b := vendors[j-1], vendors[j]
			if a.totalSpend < b.totalSpend || (a.totalSpend == b.totalSpend && a.vendorName > b.vendorName) {
				vendors[j-1], vendors[j] = vendors[j], vendors[j-1]
			} else {
				break
			}
		}
	}
}

// computeVendorAnalysis performs all in-memory calculations: spend pcts, scoring.
// It is pure (no I/O) to keep it easily testable.
func computeVendorAnalysis(vendors []*rawVendor) []VendorAnalysis {
	// Compute total spend across all vendors for SpendPct.
	var grandTotal int64
	for _, v := range vendors {
		grandTotal += v.totalSpend
	}

	// Find the vendor with the most items for coverage scoring.
	var maxItems int
	for _, v := range vendors {
		if v.itemsSupplied > maxItems {
			maxItems = v.itemsSupplied
		}
	}

	results := make([]VendorAnalysis, len(vendors))
	for i, v := range vendors {
		// SpendPct: guard against zero total.
		var spendPct float64
		if grandTotal > 0 {
			spendPct = float64(v.totalSpend) / float64(grandTotal) * 100
		}

		// AvgCostPerItem: guard against zero items.
		var avgCostPerItem int64
		if v.itemsSupplied > 0 {
			avgCostPerItem = v.totalSpend / int64(v.itemsSupplied)
		}

		// Score = price_score * 0.5 + coverage_score * 0.5, clamped 0-100.
		priceScore := 100.0 - v.deviationPct
		if priceScore < 0 {
			priceScore = 0
		}

		var coverageScore float64
		if maxItems > 0 {
			coverageScore = float64(v.itemsSupplied) / float64(maxItems) * 100
		}

		rawScore := priceScore*0.5 + coverageScore*0.5
		score := int(rawScore)
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}

		results[i] = VendorAnalysis{
			VendorName:     v.vendorName,
			ItemsSupplied:  v.itemsSupplied,
			TotalSpend:     v.totalSpend,
			SpendPct:       spendPct,
			AvgCostPerItem: avgCostPerItem,
			Score:          score,
		}
	}

	return results
}

// GetSummary calls GetVendors and aggregates location-wide vendor KPIs.
func (s *Service) GetSummary(ctx context.Context, orgID, locationID string, from, to time.Time) (*VendorSummary, error) {
	vendors, err := s.GetVendors(ctx, orgID, locationID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get vendor summary: %w", err)
	}

	summary := &VendorSummary{
		TotalVendors: len(vendors),
	}

	if len(vendors) == 0 {
		return summary, nil
	}

	var totalSpend int64
	var totalItems int
	var topVendor VendorAnalysis

	for i, v := range vendors {
		totalSpend += v.TotalSpend
		totalItems += v.ItemsSupplied
		// Vendors are already sorted by spend DESC; first entry is the top vendor.
		if i == 0 {
			topVendor = v
		}
	}

	summary.TotalSpend = totalSpend
	summary.TopVendorName = topVendor.VendorName

	// TopVendorPct: guard against zero total.
	if totalSpend > 0 {
		summary.TopVendorPct = float64(topVendor.TotalSpend) / float64(totalSpend) * 100
	}

	// AvgItemsPerVendor: guard against zero vendors (already guarded above).
	summary.AvgItemsPerVendor = float64(totalItems) / float64(len(vendors))

	return summary, nil
}
