package menu

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// Service provides menu intelligence capabilities.
type Service struct {
	pool *pgxpool.Pool
	bus  *event.Bus
}

// New creates a new menu intelligence service.
func New(pool *pgxpool.Pool, bus *event.Bus) *Service {
	return &Service{pool: pool, bus: bus}
}

// MenuItemAnalysis holds the computed analytics for a single menu item.
type MenuItemAnalysis struct {
	MenuItemID       string          `json:"menu_item_id"`
	Name             string          `json:"name"`
	Category         string          `json:"category"`
	Price            int64           `json:"price"`
	FoodCost         int64           `json:"food_cost"`
	UnitsSold        int             `json:"units_sold"`
	ContribMargin    int64           `json:"contrib_margin"`
	ContribMarginPct float64         `json:"contrib_margin_pct"`
	PopularityPct    float64         `json:"popularity_pct"`
	HealthScore      float64         `json:"health_score"`
	Classification   string          `json:"classification"`
	ByChannel        []ChannelMargin `json:"by_channel"`
}

// ChannelMargin holds per-channel margin details for a menu item.
type ChannelMargin struct {
	Channel    string  `json:"channel"`
	Revenue    int64   `json:"revenue"`
	Commission int64   `json:"commission"`
	FoodCost   int64   `json:"food_cost"`
	Margin     int64   `json:"margin"`
	MarginPct  float64 `json:"margin_pct"`
	UnitsSold  int     `json:"units_sold"`
}

// MenuSummary holds location-wide rollup KPIs.
type MenuSummary struct {
	TotalItems        int               `json:"total_items"`
	AvgMarginPct      float64           `json:"avg_margin_pct"`
	PowerhouseCount   int               `json:"powerhouse_count"`
	UnderperformCount int               `json:"underperform_count"`
	Categories        []CategorySummary `json:"categories"`
}

// CategorySummary holds rollup KPIs for a single menu category.
type CategorySummary struct {
	Category     string  `json:"category"`
	ItemCount    int     `json:"item_count"`
	AvgMarginPct float64 `json:"avg_margin_pct"`
	TopItem      string  `json:"top_item"`
}

// rawItem is an internal accumulator used while building analysis data.
type rawItem struct {
	menuItemID string
	name       string
	category   string
	price      int64
	foodCost   int64
	// channelSales maps channel -> units sold
	channelSales map[string]int
}

const deliveryCommissionRate = 0.30

// AnalyzeMenuItems queries menu items, recipe costs, and sales data to produce
// per-item margin, popularity, classification, and health scores.
func (s *Service) AnalyzeMenuItems(ctx context.Context, orgID, locationID string, from, to time.Time) ([]MenuItemAnalysis, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	// itemMap collects data from both queries inside TenantTx, keyed by menu_item_id.
	itemMap := make(map[string]*rawItem)

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// --- Query 1: food cost per menu item ---
		rows, err := tx.Query(tenantCtx,
			`SELECT mi.menu_item_id, mi.name, mi.category, mi.price,
			        COALESCE(SUM(re.quantity_per_unit * COALESCE(ilc.local_cost_per_unit, i.cost_per_unit)), 0)::BIGINT AS food_cost
			 FROM menu_items mi
			 LEFT JOIN recipe_explosion re ON re.menu_item_id = mi.menu_item_id AND re.org_id = mi.org_id
			 LEFT JOIN ingredients i ON i.ingredient_id = re.ingredient_id AND i.org_id = mi.org_id
			 LEFT JOIN ingredient_location_configs ilc
			     ON ilc.ingredient_id = i.ingredient_id AND ilc.location_id = mi.location_id AND ilc.org_id = mi.org_id
			 WHERE mi.location_id = $1 AND mi.available = true
			 GROUP BY mi.menu_item_id, mi.name, mi.category, mi.price
			 ORDER BY mi.name`,
			locationID,
		)
		if err != nil {
			return fmt.Errorf("query food cost: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id, name, category string
			var price, foodCost int64
			if err := rows.Scan(&id, &name, &category, &price, &foodCost); err != nil {
				return fmt.Errorf("scan food cost row: %w", err)
			}
			itemMap[id] = &rawItem{
				menuItemID:   id,
				name:         name,
				category:     category,
				price:        price,
				foodCost:     foodCost,
				channelSales: make(map[string]int),
			}
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate food cost rows: %w", err)
		}

		// --- Query 2: units sold per item per channel ---
		salesRows, err := tx.Query(tenantCtx,
			`SELECT ci.menu_item_id, c.channel, SUM(ci.quantity)::INT AS units_sold
			 FROM check_items ci
			 JOIN checks c ON c.check_id = ci.check_id AND c.org_id = ci.org_id
			 WHERE c.location_id = $1 AND c.status = 'closed' AND ci.voided_at IS NULL
			   AND c.closed_at >= $2 AND c.closed_at < $3 AND ci.menu_item_id IS NOT NULL
			 GROUP BY ci.menu_item_id, c.channel`,
			locationID, from, to,
		)
		if err != nil {
			return fmt.Errorf("query sales: %w", err)
		}
		defer salesRows.Close()

		for salesRows.Next() {
			var menuItemID, channel string
			var unitsSold int
			if err := salesRows.Scan(&menuItemID, &channel, &unitsSold); err != nil {
				return fmt.Errorf("scan sales row: %w", err)
			}
			// Sales may reference items that are no longer available; skip those.
			item, ok := itemMap[menuItemID]
			if !ok {
				continue
			}
			item.channelSales[channel] += unitsSold
		}
		if err := salesRows.Err(); err != nil {
			return fmt.Errorf("iterate sales rows: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Build a deterministic slice from the map AFTER the transaction completes.
	items := make([]*rawItem, 0, len(itemMap))
	for _, v := range itemMap {
		items = append(items, v)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].name < items[j].name
	})

	return computeAnalysis(items), nil
}

// computeAnalysis performs all in-memory calculations: channel margins, classification,
// and health scores. It is pure (no I/O) to keep it easily testable.
func computeAnalysis(items []*rawItem) []MenuItemAnalysis {
	if len(items) == 0 {
		return []MenuItemAnalysis{}
	}

	// --- Step 1: build per-item channel margins and aggregate totals ---
	type intermediate struct {
		raw           *rawItem
		unitsSold     int
		avgMargin     int64   // blended contribution margin (cents)
		avgMarginPct  float64 // blended margin pct
		byChannel     []ChannelMargin
	}

	intermediates := make([]intermediate, len(items))
	for idx, it := range items {
		var totalUnits int
		var weightedMargin int64 // sum of (margin * units) for blending
		var channels []ChannelMargin

		for channel, units := range it.channelSales {
			revenue := it.price * int64(units)
			var commission int64
			if channel == "delivery" {
				commission = int64(float64(revenue) * deliveryCommissionRate)
			}
			netRevenue := revenue - commission
			fc := it.foodCost * int64(units)
			margin := netRevenue - fc
			var marginPct float64
			if netRevenue > 0 {
				marginPct = float64(margin) / float64(netRevenue) * 100
			}
			channels = append(channels, ChannelMargin{
				Channel:    channel,
				Revenue:    revenue,
				Commission: commission,
				FoodCost:   fc,
				Margin:     margin,
				MarginPct:  marginPct,
				UnitsSold:  units,
			})
			totalUnits += units
			weightedMargin += margin
		}

		// Sort channels for deterministic output.
		sort.Slice(channels, func(a, b int) bool {
			return channels[a].Channel < channels[b].Channel
		})

		// Blended contribution margin per unit (uses actual sold mix).
		var contribMargin int64
		var contribMarginPct float64
		if totalUnits > 0 {
			contribMargin = weightedMargin / int64(totalUnits)
		} else {
			// No sales: use dine_in (no commission) as proxy.
			netRevenue := it.price
			contribMargin = netRevenue - it.foodCost
		}
		// For margin pct, use price as denominator when no sales (conservative).
		netRevPerUnit := it.price // default (dine_in / takeout)
		if totalUnits > 0 {
			// Compute blended net revenue per unit from channel data.
			var totalNetRevenue int64
			for _, ch := range channels {
				totalNetRevenue += ch.Revenue - ch.Commission
			}
			netRevPerUnit = totalNetRevenue / int64(totalUnits)
		}
		if netRevPerUnit > 0 {
			contribMarginPct = float64(contribMargin) / float64(netRevPerUnit) * 100
		}

		intermediates[idx] = intermediate{
			raw:          it,
			unitsSold:    totalUnits,
			avgMargin:    contribMargin,
			avgMarginPct: contribMarginPct,
			byChannel:    channels,
		}
	}

	// --- Step 2: group by category for classification thresholds ---
	type catStats struct {
		indices []int
		margins []float64
	}
	catMap := make(map[string]*catStats)
	for i, im := range intermediates {
		cat := im.raw.category
		if catMap[cat] == nil {
			catMap[cat] = &catStats{}
		}
		catMap[cat].indices = append(catMap[cat].indices, i)
		catMap[cat].margins = append(catMap[cat].margins, im.avgMarginPct)
	}

	// Compute per-category median margin and popularity threshold.
	type catThreshold struct {
		medianMargin      float64
		popularityThresh  float64 // Kasavana-Smith: (1/N)*0.7*100
		totalUnitsSold    int
	}
	catThresholds := make(map[string]catThreshold)
	for cat, cs := range catMap {
		n := len(cs.indices)
		// Popularity threshold: (1/N)*0.7*100 — each item's popularity is expressed
		// as a percent-of-category-total, threshold is 70% of the even share.
		popularityThresh := (1.0 / float64(n)) * 0.7 * 100

		// Compute total units sold in this category.
		var totalUnits int
		for _, i := range cs.indices {
			totalUnits += intermediates[i].unitsSold
		}

		// Median margin.
		sorted := make([]float64, len(cs.margins))
		copy(sorted, cs.margins)
		sort.Float64s(sorted)
		var median float64
		if n%2 == 0 {
			median = (sorted[n/2-1] + sorted[n/2]) / 2
		} else {
			median = sorted[n/2]
		}
		catThresholds[cat] = catThreshold{
			medianMargin:     median,
			popularityThresh: popularityThresh,
			totalUnitsSold:   totalUnits,
		}
	}

	// --- Step 3: health score — percentile rank across ALL items in location ---
	// Collect all margin pcts and units sold for location-wide ranking.
	allMargins := make([]float64, len(intermediates))
	allUnits := make([]int, len(intermediates))
	for i, im := range intermediates {
		allMargins[i] = im.avgMarginPct
		allUnits[i] = im.unitsSold
	}

	// Build results.
	results := make([]MenuItemAnalysis, len(intermediates))
	for idx, im := range intermediates {
		cat := im.raw.category
		ct := catThresholds[cat]

		// Popularity: this item's share of category units sold (as %).
		var popularityPct float64
		if ct.totalUnitsSold > 0 {
			popularityPct = float64(im.unitsSold) / float64(ct.totalUnitsSold) * 100
		}

		// Classification.
		highMargin := im.avgMarginPct >= ct.medianMargin
		highPopularity := popularityPct >= ct.popularityThresh
		var classification string
		switch {
		case highMargin && highPopularity:
			classification = "powerhouse"
		case highMargin && !highPopularity:
			classification = "hidden_gem"
		case !highMargin && highPopularity:
			classification = "crowd_pleaser"
		default:
			classification = "underperformer"
		}

		// Health score: 50% margin percentile + 50% popularity percentile, scaled 0-100.
		marginRank := percentileRankFloat(allMargins, im.avgMarginPct)
		unitsRank := percentileRankInt(allUnits, im.unitsSold)
		healthScore := (marginRank*0.5 + unitsRank*0.5) * 100

		results[idx] = MenuItemAnalysis{
			MenuItemID:       im.raw.menuItemID,
			Name:             im.raw.name,
			Category:         cat,
			Price:            im.raw.price,
			FoodCost:         im.raw.foodCost,
			UnitsSold:        im.unitsSold,
			ContribMargin:    im.avgMargin,
			ContribMarginPct: im.avgMarginPct,
			PopularityPct:    popularityPct,
			HealthScore:      healthScore,
			Classification:   classification,
			ByChannel:        im.byChannel,
		}
	}

	return results
}

// percentileRankFloat returns the fraction [0,1] of values in data that are
// strictly less than target. Ties count as 0.5 of their occurrences.
func percentileRankFloat(data []float64, target float64) float64 {
	n := len(data)
	if n == 0 {
		return 0
	}
	var below, equal int
	for _, v := range data {
		if v < target {
			below++
		} else if v == target {
			equal++
		}
	}
	return (float64(below) + float64(equal)*0.5) / float64(n)
}

// percentileRankInt returns the fraction [0,1] of values in data that are
// strictly less than target. Ties count as 0.5 of their occurrences.
func percentileRankInt(data []int, target int) float64 {
	n := len(data)
	if n == 0 {
		return 0
	}
	var below, equal int
	for _, v := range data {
		if v < target {
			below++
		} else if v == target {
			equal++
		}
	}
	return (float64(below) + float64(equal)*0.5) / float64(n)
}

// GetSummary calls AnalyzeMenuItems and computes location-wide rollup KPIs.
func (s *Service) GetSummary(ctx context.Context, orgID, locationID string, from, to time.Time) (*MenuSummary, error) {
	items, err := s.AnalyzeMenuItems(ctx, orgID, locationID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get summary: %w", err)
	}

	summary := &MenuSummary{
		TotalItems: len(items),
	}

	if len(items) == 0 {
		summary.Categories = []CategorySummary{}
		return summary, nil
	}

	// Location-wide avg margin pct.
	var totalMarginPct float64
	for _, it := range items {
		if it.Classification == "powerhouse" {
			summary.PowerhouseCount++
		} else if it.Classification == "underperformer" {
			summary.UnderperformCount++
		}
		totalMarginPct += it.ContribMarginPct
	}
	summary.AvgMarginPct = totalMarginPct / float64(len(items))

	// Build per-category summaries.
	type catAcc struct {
		items       []MenuItemAnalysis
		totalMargin float64
	}
	catAccMap := make(map[string]*catAcc)
	for _, it := range items {
		if catAccMap[it.Category] == nil {
			catAccMap[it.Category] = &catAcc{}
		}
		acc := catAccMap[it.Category]
		acc.items = append(acc.items, it)
		acc.totalMargin += it.ContribMarginPct
	}

	// Sort categories for deterministic output.
	catNames := make([]string, 0, len(catAccMap))
	for cat := range catAccMap {
		catNames = append(catNames, cat)
	}
	sort.Strings(catNames)

	for _, cat := range catNames {
		acc := catAccMap[cat]
		n := len(acc.items)
		avgMargin := acc.totalMargin / float64(n)

		// Top item = highest contrib margin item in the category.
		topItem := acc.items[0]
		for _, it := range acc.items[1:] {
			if it.ContribMargin > topItem.ContribMargin {
				topItem = it
			}
		}

		summary.Categories = append(summary.Categories, CategorySummary{
			Category:     cat,
			ItemCount:    n,
			AvgMarginPct: avgMargin,
			TopItem:      topItem.Name,
		})
	}

	return summary, nil
}
