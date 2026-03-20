package inventory

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// PurchaseOrder represents a purchase order header.
type PurchaseOrder struct {
	PurchaseOrderID string     `json:"purchase_order_id"`
	OrgID           string     `json:"org_id"`
	LocationID      string     `json:"location_id"`
	VendorName      string     `json:"vendor_name"`
	Status          string     `json:"status"`
	Source          string     `json:"source"`
	SuggestedAt     *time.Time `json:"suggested_at,omitempty"`
	ApprovedBy      *string    `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	ReceivedBy      *string    `json:"received_by,omitempty"`
	ReceivedAt      *time.Time `json:"received_at,omitempty"`
	TotalEstimated  int64      `json:"total_estimated"`
	TotalActual     int64      `json:"total_actual"`
	Notes           string     `json:"notes"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LineCount       int        `json:"line_count"`
}

// POLine represents a single line within a purchase order.
type POLine struct {
	POLineID         string     `json:"po_line_id"`
	OrgID            string     `json:"org_id"`
	PurchaseOrderID  string     `json:"purchase_order_id"`
	IngredientID     string     `json:"ingredient_id"`
	IngredientName   string     `json:"ingredient_name"`
	OrderedQty       float64    `json:"ordered_qty"`
	OrderedUnit      string     `json:"ordered_unit"`
	EstimatedUnitCost int       `json:"estimated_unit_cost"`
	ReceivedQty      *float64   `json:"received_qty,omitempty"`
	ReceivedUnitCost *int       `json:"received_unit_cost,omitempty"`
	VarianceQty      *float64   `json:"variance_qty,omitempty"`
	VarianceFlag     *string    `json:"variance_flag,omitempty"`
	ReceivedAt       *time.Time `json:"received_at,omitempty"`
	Note             string     `json:"note"`
	CreatedAt        time.Time  `json:"created_at"`
}

// POWithLines combines a PurchaseOrder header with its lines.
type POWithLines struct {
	PurchaseOrder
	Lines []POLine `json:"lines"`
}

// POLineInput holds data for creating a purchase order line.
type POLineInput struct {
	IngredientID      string  `json:"ingredient_id"`
	OrderedQty        float64 `json:"ordered_qty"`
	OrderedUnit       string  `json:"ordered_unit"`
	EstimatedUnitCost int     `json:"estimated_unit_cost"`
}

// ReceiveLineInput holds receiving data for a single line.
type ReceiveLineInput struct {
	POLineID         string  `json:"po_line_id"`
	ReceivedQty      float64 `json:"received_qty"`
	ReceivedUnitCost int     `json:"received_unit_cost"`
	Note             string  `json:"note"`
}

// Discrepancy records a variance between ordered and received quantities.
type Discrepancy struct {
	IngredientName string  `json:"ingredient_name"`
	Ordered        float64 `json:"ordered"`
	Received       float64 `json:"received"`
	Flag           string  `json:"flag"`
}

// PARBreach represents an ingredient that has breached its reorder point.
type PARBreach struct {
	IngredientID          string  `json:"ingredient_id"`
	Name                  string  `json:"name"`
	CurrentLevel          float64 `json:"current_level"`
	ReorderPoint          float64 `json:"reorder_point"`
	PARLevel              float64 `json:"par_level"`
	AvgDailyUsage         float64 `json:"avg_daily_usage"`
	ProjectedStockoutDays float64 `json:"projected_stockout_days"`
	VendorName            string  `json:"vendor_name"`
	HasPendingPO          bool    `json:"has_pending_po"`
}

// computeVarianceFlag determines the variance category with a 2% tolerance.
// not_received if received == 0 and ordered > 0.
// short if received < ordered * 0.98.
// over if received > ordered * 1.02.
// exact otherwise.
func computeVarianceFlag(ordered, received float64) string {
	if ordered == 0 {
		return "exact"
	}
	if received == 0 {
		return "not_received"
	}
	ratio := received / ordered
	if ratio < 0.98 {
		return "short"
	}
	if ratio > 1.02 {
		return "over"
	}
	return "exact"
}

// computeAvgDailyUsage calculates average daily usage from total usage over a period.
func computeAvgDailyUsage(totalUsage float64, days int) float64 {
	if days == 0 {
		return 0.0
	}
	return totalUsage / float64(days)
}

// effectiveReorderPoint returns the greater of the manual reorder point and the
// dynamic reorder point computed from lead time and average daily usage.
func effectiveReorderPoint(manualReorderPoint float64, leadTimeDays int, avgDailyUsage float64) float64 {
	dynamic := float64(leadTimeDays) * avgDailyUsage
	if dynamic > manualReorderPoint {
		return dynamic
	}
	return manualReorderPoint
}

// CreatePO creates a new purchase order with the given lines.
func (s *Service) CreatePO(ctx context.Context, orgID, locationID, vendorName, notes string, lines []POLineInput) (*PurchaseOrder, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var po PurchaseOrder

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(tenantCtx,
			`INSERT INTO purchase_orders (org_id, location_id, vendor_name, notes, source)
			 VALUES ($1, $2, $3, $4, 'manual')
			 RETURNING purchase_order_id, org_id, location_id, vendor_name, status, source,
			           suggested_at, approved_by, approved_at, received_by, received_at,
			           total_estimated, total_actual, COALESCE(notes, ''), created_at, updated_at`,
			orgID, locationID, vendorName, notes,
		).Scan(
			&po.PurchaseOrderID, &po.OrgID, &po.LocationID, &po.VendorName, &po.Status, &po.Source,
			&po.SuggestedAt, &po.ApprovedBy, &po.ApprovedAt, &po.ReceivedBy, &po.ReceivedAt,
			&po.TotalEstimated, &po.TotalActual, &po.Notes, &po.CreatedAt, &po.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert purchase order: %w", err)
		}

		var totalEstimated int64
		for _, l := range lines {
			_, err := tx.Exec(tenantCtx,
				`INSERT INTO purchase_order_lines (org_id, purchase_order_id, ingredient_id, ordered_qty, ordered_unit, estimated_unit_cost)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				orgID, po.PurchaseOrderID, l.IngredientID, l.OrderedQty, l.OrderedUnit, l.EstimatedUnitCost,
			)
			if err != nil {
				return fmt.Errorf("insert po line for ingredient %s: %w", l.IngredientID, err)
			}
			totalEstimated += int64(l.OrderedQty * float64(l.EstimatedUnitCost))
		}

		_, err = tx.Exec(tenantCtx,
			`UPDATE purchase_orders SET total_estimated = $1, updated_at = now() WHERE purchase_order_id = $2`,
			totalEstimated, po.PurchaseOrderID,
		)
		if err != nil {
			return fmt.Errorf("update total_estimated: %w", err)
		}
		po.TotalEstimated = totalEstimated

		return nil
	})

	return &po, err
}

// ListPOs returns purchase orders for a location with optional status filter.
func (s *Service) ListPOs(ctx context.Context, orgID, locationID, status string) ([]PurchaseOrder, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []PurchaseOrder

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		query := `SELECT po.purchase_order_id, po.org_id, po.location_id, po.vendor_name,
		                 po.status, po.source, po.suggested_at, po.approved_by, po.approved_at,
		                 po.received_by, po.received_at, po.total_estimated, po.total_actual,
		                 COALESCE(po.notes, ''), po.created_at, po.updated_at,
		                 (SELECT COUNT(*) FROM purchase_order_lines WHERE purchase_order_id = po.purchase_order_id) AS line_count
		          FROM purchase_orders po
		          WHERE po.org_id = $1`
		args := []any{orgID}
		argIdx := 2

		if locationID != "" {
			query += fmt.Sprintf(" AND po.location_id = $%d", argIdx)
			args = append(args, locationID)
			argIdx++
		}
		if status != "" {
			query += fmt.Sprintf(" AND po.status = $%d", argIdx)
			args = append(args, status)
		}
		query += " ORDER BY po.created_at DESC"

		rows, err := tx.Query(tenantCtx, query, args...)
		if err != nil {
			return fmt.Errorf("list purchase orders: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var po PurchaseOrder
			if err := rows.Scan(
				&po.PurchaseOrderID, &po.OrgID, &po.LocationID, &po.VendorName,
				&po.Status, &po.Source, &po.SuggestedAt, &po.ApprovedBy, &po.ApprovedAt,
				&po.ReceivedBy, &po.ReceivedAt, &po.TotalEstimated, &po.TotalActual,
				&po.Notes, &po.CreatedAt, &po.UpdatedAt, &po.LineCount,
			); err != nil {
				return fmt.Errorf("scan purchase order: %w", err)
			}
			results = append(results, po)
		}
		return rows.Err()
	})

	return results, err
}

// GetPO returns a purchase order with its lines.
func (s *Service) GetPO(ctx context.Context, orgID, poID string) (*POWithLines, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var result POWithLines

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(tenantCtx,
			`SELECT purchase_order_id, org_id, location_id, vendor_name, status, source,
			        suggested_at, approved_by, approved_at, received_by, received_at,
			        total_estimated, total_actual, COALESCE(notes, ''), created_at, updated_at
			 FROM purchase_orders WHERE purchase_order_id = $1`,
			poID,
		).Scan(
			&result.PurchaseOrderID, &result.OrgID, &result.LocationID, &result.VendorName,
			&result.Status, &result.Source, &result.SuggestedAt, &result.ApprovedBy, &result.ApprovedAt,
			&result.ReceivedBy, &result.ReceivedAt, &result.TotalEstimated, &result.TotalActual,
			&result.Notes, &result.CreatedAt, &result.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("get purchase order: %w", err)
		}

		rows, err := tx.Query(tenantCtx,
			`SELECT pol.po_line_id, pol.org_id, pol.purchase_order_id, pol.ingredient_id,
			        i.name, pol.ordered_qty, pol.ordered_unit, pol.estimated_unit_cost,
			        pol.received_qty, pol.received_unit_cost, pol.variance_qty, pol.variance_flag,
			        pol.received_at, COALESCE(pol.note, ''), pol.created_at
			 FROM purchase_order_lines pol
			 JOIN ingredients i ON i.ingredient_id = pol.ingredient_id
			 WHERE pol.purchase_order_id = $1
			 ORDER BY pol.created_at`,
			poID,
		)
		if err != nil {
			return fmt.Errorf("get po lines: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var l POLine
			if err := rows.Scan(
				&l.POLineID, &l.OrgID, &l.PurchaseOrderID, &l.IngredientID,
				&l.IngredientName, &l.OrderedQty, &l.OrderedUnit, &l.EstimatedUnitCost,
				&l.ReceivedQty, &l.ReceivedUnitCost, &l.VarianceQty, &l.VarianceFlag,
				&l.ReceivedAt, &l.Note, &l.CreatedAt,
			); err != nil {
				return fmt.Errorf("scan po line: %w", err)
			}
			result.Lines = append(result.Lines, l)
		}
		return rows.Err()
	})

	return &result, err
}

// UpdatePOStatus transitions a purchase order to a new status.
// Valid transitions: draft->approved, draft->cancelled, approved->cancelled.
func (s *Service) UpdatePOStatus(ctx context.Context, orgID, poID, newStatus, userID string) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		var currentStatus string
		err := tx.QueryRow(tenantCtx,
			`SELECT status FROM purchase_orders WHERE purchase_order_id = $1`,
			poID,
		).Scan(&currentStatus)
		if err != nil {
			return fmt.Errorf("get po status: %w", err)
		}

		// Validate transition
		switch {
		case currentStatus == "draft" && newStatus == "approved":
			// allowed
		case currentStatus == "draft" && newStatus == "cancelled":
			// allowed
		case currentStatus == "approved" && newStatus == "cancelled":
			// allowed
		default:
			return fmt.Errorf("invalid status transition: %s -> %s", currentStatus, newStatus)
		}

		if newStatus == "approved" {
			_, err = tx.Exec(tenantCtx,
				`UPDATE purchase_orders
				 SET status = $1, approved_by = $2, approved_at = now(), updated_at = now()
				 WHERE purchase_order_id = $3`,
				newStatus, userID, poID,
			)
		} else {
			_, err = tx.Exec(tenantCtx,
				`UPDATE purchase_orders SET status = $1, updated_at = now() WHERE purchase_order_id = $2`,
				newStatus, poID,
			)
		}
		if err != nil {
			return fmt.Errorf("update po status: %w", err)
		}
		return nil
	})
}

// UpdatePODraft replaces the lines of a draft purchase order.
func (s *Service) UpdatePODraft(ctx context.Context, orgID, poID, notes string, lines []POLineInput) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		var currentStatus string
		err := tx.QueryRow(tenantCtx,
			`SELECT status FROM purchase_orders WHERE purchase_order_id = $1`,
			poID,
		).Scan(&currentStatus)
		if err != nil {
			return fmt.Errorf("get po status: %w", err)
		}
		if currentStatus != "draft" {
			return fmt.Errorf("can only edit draft purchase orders, current status: %s", currentStatus)
		}

		_, err = tx.Exec(tenantCtx,
			`DELETE FROM purchase_order_lines WHERE purchase_order_id = $1`,
			poID,
		)
		if err != nil {
			return fmt.Errorf("delete existing lines: %w", err)
		}

		var totalEstimated int64
		for _, l := range lines {
			_, err := tx.Exec(tenantCtx,
				`INSERT INTO purchase_order_lines (org_id, purchase_order_id, ingredient_id, ordered_qty, ordered_unit, estimated_unit_cost)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				orgID, poID, l.IngredientID, l.OrderedQty, l.OrderedUnit, l.EstimatedUnitCost,
			)
			if err != nil {
				return fmt.Errorf("insert po line for ingredient %s: %w", l.IngredientID, err)
			}
			totalEstimated += int64(l.OrderedQty * float64(l.EstimatedUnitCost))
		}

		_, err = tx.Exec(tenantCtx,
			`UPDATE purchase_orders SET notes = $1, total_estimated = $2, updated_at = now() WHERE purchase_order_id = $3`,
			notes, totalEstimated, poID,
		)
		if err != nil {
			return fmt.Errorf("update po draft: %w", err)
		}
		return nil
	})
}

// DeletePO deletes a purchase order if it is in draft status.
func (s *Service) DeletePO(ctx context.Context, orgID, poID string) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		var currentStatus string
		err := tx.QueryRow(tenantCtx,
			`SELECT status FROM purchase_orders WHERE purchase_order_id = $1`,
			poID,
		).Scan(&currentStatus)
		if err != nil {
			return fmt.Errorf("get po status: %w", err)
		}
		if currentStatus != "draft" {
			return fmt.Errorf("can only delete draft purchase orders, current status: %s", currentStatus)
		}

		_, err = tx.Exec(tenantCtx,
			`DELETE FROM purchase_order_lines WHERE purchase_order_id = $1`,
			poID,
		)
		if err != nil {
			return fmt.Errorf("delete po lines: %w", err)
		}

		_, err = tx.Exec(tenantCtx,
			`DELETE FROM purchase_orders WHERE purchase_order_id = $1`,
			poID,
		)
		if err != nil {
			return fmt.Errorf("delete purchase order: %w", err)
		}
		return nil
	})
}

// ListPendingPOs returns approved purchase orders awaiting delivery.
func (s *Service) ListPendingPOs(ctx context.Context, orgID, locationID string) ([]PurchaseOrder, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []PurchaseOrder

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		query := `SELECT po.purchase_order_id, po.org_id, po.location_id, po.vendor_name,
		                 po.status, po.source, po.suggested_at, po.approved_by, po.approved_at,
		                 po.received_by, po.received_at, po.total_estimated, po.total_actual,
		                 COALESCE(po.notes, ''), po.created_at, po.updated_at,
		                 (SELECT COUNT(*) FROM purchase_order_lines WHERE purchase_order_id = po.purchase_order_id) AS line_count
		          FROM purchase_orders po
		          WHERE po.org_id = $1 AND po.status = 'approved'`
		args := []any{orgID}

		if locationID != "" {
			query += " AND po.location_id = $2"
			args = append(args, locationID)
		}
		query += " ORDER BY po.approved_at ASC"

		rows, err := tx.Query(tenantCtx, query, args...)
		if err != nil {
			return fmt.Errorf("list pending pos: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var po PurchaseOrder
			if err := rows.Scan(
				&po.PurchaseOrderID, &po.OrgID, &po.LocationID, &po.VendorName,
				&po.Status, &po.Source, &po.SuggestedAt, &po.ApprovedBy, &po.ApprovedAt,
				&po.ReceivedBy, &po.ReceivedAt, &po.TotalEstimated, &po.TotalActual,
				&po.Notes, &po.CreatedAt, &po.UpdatedAt, &po.LineCount,
			); err != nil {
				return fmt.Errorf("scan pending po: %w", err)
			}
			results = append(results, po)
		}
		return rows.Err()
	})

	if err != nil {
		return nil, err
	}

	// Compute days_since_approved in Go
	now := time.Now()
	for i := range results {
		if results[i].ApprovedAt != nil {
			_ = now.Sub(*results[i].ApprovedAt) // available for callers via ApprovedAt field
		}
	}

	return results, nil
}

// ReceivePO processes delivery receiving for an approved purchase order.
// Returns discrepancies, total_actual in cents, and any error.
func (s *Service) ReceivePO(ctx context.Context, orgID, poID, receivedBy string, inputLines []ReceiveLineInput) ([]Discrepancy, int64, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var discrepancies []Discrepancy
	var totalActual int64

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		var currentStatus string
		err := tx.QueryRow(tenantCtx,
			`SELECT status FROM purchase_orders WHERE purchase_order_id = $1`,
			poID,
		).Scan(&currentStatus)
		if err != nil {
			return fmt.Errorf("get po status: %w", err)
		}
		if currentStatus != "approved" {
			return fmt.Errorf("can only receive approved purchase orders, current status: %s", currentStatus)
		}

		// Build map of inputs by po_line_id
		inputMap := make(map[string]ReceiveLineInput, len(inputLines))
		for _, il := range inputLines {
			inputMap[il.POLineID] = il
		}

		// Query all lines for this PO
		rows, err := tx.Query(tenantCtx,
			`SELECT pol.po_line_id, pol.ingredient_id, i.name, pol.ordered_qty, pol.ordered_unit
			 FROM purchase_order_lines pol
			 JOIN ingredients i ON i.ingredient_id = pol.ingredient_id
			 WHERE pol.purchase_order_id = $1`,
			poID,
		)
		if err != nil {
			return fmt.Errorf("query po lines: %w", err)
		}

		type lineRecord struct {
			poLineID     string
			ingredientID string
			name         string
			orderedQty   float64
			orderedUnit  string
		}
		var allLines []lineRecord
		for rows.Next() {
			var lr lineRecord
			if err := rows.Scan(&lr.poLineID, &lr.ingredientID, &lr.name, &lr.orderedQty, &lr.orderedUnit); err != nil {
				rows.Close()
				return fmt.Errorf("scan po line: %w", err)
			}
			allLines = append(allLines, lr)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}

		for _, lr := range allLines {
			var receivedQty float64
			var receivedUnitCost int
			var note string

			if input, ok := inputMap[lr.poLineID]; ok {
				receivedQty = input.ReceivedQty
				receivedUnitCost = input.ReceivedUnitCost
				note = input.Note
			} else {
				receivedQty = 0
				receivedUnitCost = 0
				note = ""
			}

			varianceQty := receivedQty - lr.orderedQty
			flag := computeVarianceFlag(lr.orderedQty, receivedQty)

			_, err = tx.Exec(tenantCtx,
				`UPDATE purchase_order_lines
				 SET received_qty = $1, received_unit_cost = $2, variance_qty = $3,
				     variance_flag = $4, received_at = now(), note = $5
				 WHERE po_line_id = $6`,
				receivedQty, receivedUnitCost, varianceQty, flag, note, lr.poLineID,
			)
			if err != nil {
				return fmt.Errorf("update po line %s: %w", lr.poLineID, err)
			}

			totalActual += int64(receivedQty * float64(receivedUnitCost))

			if flag != "exact" {
				discrepancies = append(discrepancies, Discrepancy{
					IngredientName: lr.name,
					Ordered:        lr.orderedQty,
					Received:       receivedQty,
					Flag:           flag,
				})
			}
		}

		_, err = tx.Exec(tenantCtx,
			`UPDATE purchase_orders
			 SET status = 'received', received_by = $1, received_at = now(),
			     total_actual = $2, updated_at = now()
			 WHERE purchase_order_id = $3`,
			receivedBy, totalActual, poID,
		)
		if err != nil {
			return fmt.Errorf("update po received: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	if len(discrepancies) > 0 {
		s.bus.Publish(ctx, event.Envelope{
			EventID:   fmt.Sprintf("%s.discrepancy", poID),
			EventType: "inventory.delivery.discrepancy",
			OrgID:     orgID,
			Source:    "inventory",
			Payload: map[string]any{
				"purchase_order_id": poID,
				"discrepancies":     discrepancies,
				"total_actual":      totalActual,
			},
		})
	}

	return discrepancies, totalActual, nil
}

// GenerateSuggestedPOs creates or updates draft POs based on PAR breaches from a count.
func (s *Service) GenerateSuggestedPOs(ctx context.Context, orgID, locationID, countID string) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	type countedIngredient struct {
		ingredientID    string
		countedQty      float64
		parLevel        float64
		reorderPoint    float64
		leadTimeDays    int
		avgDailyUsage   float64
		vendorName      string
		localCostPerUnit int
	}

	var ingredients []countedIngredient

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT cl.ingredient_id,
			        COALESCE(cl.counted_qty, 0) AS counted_qty,
			        COALESCE(ilc.par_level, 0) AS par_level,
			        COALESCE(ilc.reorder_point, 0) AS reorder_point,
			        COALESCE(ilc.lead_time_days, 1) AS lead_time_days,
			        COALESCE(ilc.avg_daily_usage, 0) AS avg_daily_usage,
			        COALESCE(ilc.vendor_name, '') AS vendor_name,
			        COALESCE(ilc.local_cost_per_unit, 0) AS local_cost_per_unit
			 FROM inventory_count_lines cl
			 LEFT JOIN ingredient_location_configs ilc
			   ON ilc.ingredient_id = cl.ingredient_id AND ilc.location_id = $1
			 WHERE cl.count_id = $2 AND cl.counted_qty IS NOT NULL`,
			locationID, countID,
		)
		if err != nil {
			return fmt.Errorf("query count lines for suggestions: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var ci countedIngredient
			if err := rows.Scan(
				&ci.ingredientID, &ci.countedQty, &ci.parLevel, &ci.reorderPoint,
				&ci.leadTimeDays, &ci.avgDailyUsage, &ci.vendorName, &ci.localCostPerUnit,
			); err != nil {
				return fmt.Errorf("scan count line: %w", err)
			}
			ingredients = append(ingredients, ci)
		}
		return rows.Err()
	})
	if err != nil {
		return err
	}

	// Group breaching ingredients by vendor_name
	vendorIngredients := make(map[string][]countedIngredient)
	for _, ci := range ingredients {
		eff := effectiveReorderPoint(ci.reorderPoint, ci.leadTimeDays, ci.avgDailyUsage)
		if ci.countedQty < eff {
			vendor := ci.vendorName
			if vendor == "" {
				vendor = "Unknown Vendor"
			}
			vendorIngredients[vendor] = append(vendorIngredients[vendor], ci)
		}
	}

	now := time.Now()

	for vendor, vendorItems := range vendorIngredients {
		err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
			// Check if a draft PO already exists for this vendor+location
			var existingPOID string
			err := tx.QueryRow(tenantCtx,
				`SELECT purchase_order_id FROM purchase_orders
				 WHERE org_id = $1 AND location_id = $2 AND vendor_name = $3 AND status = 'draft'
				 LIMIT 1`,
				orgID, locationID, vendor,
			).Scan(&existingPOID)

			var poID string
			if err == pgx.ErrNoRows {
				// Create new draft PO
				err = tx.QueryRow(tenantCtx,
					`INSERT INTO purchase_orders (org_id, location_id, vendor_name, source, suggested_at)
					 VALUES ($1, $2, $3, 'system_recommended', $4)
					 RETURNING purchase_order_id`,
					orgID, locationID, vendor, now,
				).Scan(&poID)
				if err != nil {
					return fmt.Errorf("create suggested po for vendor %s: %w", vendor, err)
				}
			} else if err != nil {
				return fmt.Errorf("check existing draft po: %w", err)
			} else {
				poID = existingPOID
			}

			// Get existing line ingredient IDs for this PO
			existingIngredients := make(map[string]bool)
			lineRows, err := tx.Query(tenantCtx,
				`SELECT ingredient_id FROM purchase_order_lines WHERE purchase_order_id = $1`,
				poID,
			)
			if err != nil {
				return fmt.Errorf("query existing po lines: %w", err)
			}
			for lineRows.Next() {
				var ingID string
				if err := lineRows.Scan(&ingID); err != nil {
					lineRows.Close()
					return fmt.Errorf("scan existing ingredient: %w", err)
				}
				existingIngredients[ingID] = true
			}
			lineRows.Close()
			if err := lineRows.Err(); err != nil {
				return err
			}

			var totalEstimated int64
			// Get current total_estimated
			err = tx.QueryRow(tenantCtx,
				`SELECT total_estimated FROM purchase_orders WHERE purchase_order_id = $1`,
				poID,
			).Scan(&totalEstimated)
			if err != nil {
				return fmt.Errorf("get current total_estimated: %w", err)
			}

			for _, ci := range vendorItems {
				if existingIngredients[ci.ingredientID] {
					continue // skip duplicate
				}
				suggestedQty := ci.parLevel - ci.countedQty
				if suggestedQty <= 0 {
					suggestedQty = 1
				}
				estimatedCost := int64(suggestedQty * float64(ci.localCostPerUnit))

				_, err := tx.Exec(tenantCtx,
					`INSERT INTO purchase_order_lines (org_id, purchase_order_id, ingredient_id, ordered_qty, ordered_unit, estimated_unit_cost)
					 VALUES ($1, $2, $3, $4, '', $5)`,
					orgID, poID, ci.ingredientID, suggestedQty, ci.localCostPerUnit,
				)
				if err != nil {
					return fmt.Errorf("insert suggested line for ingredient %s: %w", ci.ingredientID, err)
				}
				totalEstimated += estimatedCost
			}

			_, err = tx.Exec(tenantCtx,
				`UPDATE purchase_orders SET total_estimated = $1, updated_at = now() WHERE purchase_order_id = $2`,
				totalEstimated, poID,
			)
			if err != nil {
				return fmt.Errorf("update total_estimated for suggested po: %w", err)
			}

			// Determine severity
			severity := "warning"
			for _, ci := range vendorItems {
				if ci.parLevel > 0 && ci.countedQty < ci.parLevel*0.20 {
					severity = "critical"
					break
				}
			}

			s.bus.Publish(ctx, event.Envelope{
				EventID:    fmt.Sprintf("%s.%s.po.suggested", countID, vendor),
				EventType:  "inventory.po.suggested",
				OrgID:      orgID,
				LocationID: locationID,
				Source:     "inventory",
				Payload: map[string]any{
					"purchase_order_id": poID,
					"vendor_name":       vendor,
					"count_id":          countID,
					"severity":          severity,
					"ingredient_count":  len(vendorItems),
				},
			})

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// GetPARBreaches returns ingredients that have breached their effective reorder point.
func (s *Service) GetPARBreaches(ctx context.Context, orgID, locationID string) ([]PARBreach, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var results []PARBreach

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT
			     i.ingredient_id,
			     i.name,
			     COALESCE(latest.counted_qty, 0) AS current_level,
			     COALESCE(ilc.reorder_point, 0) AS reorder_point,
			     COALESCE(ilc.par_level, 0) AS par_level,
			     COALESCE(ilc.avg_daily_usage, 0) AS avg_daily_usage,
			     COALESCE(ilc.lead_time_days, 1) AS lead_time_days,
			     COALESCE(ilc.vendor_name, '') AS vendor_name,
			     EXISTS (
			         SELECT 1
			         FROM purchase_order_lines pol
			         JOIN purchase_orders po ON po.purchase_order_id = pol.purchase_order_id
			         WHERE pol.ingredient_id = i.ingredient_id
			           AND po.location_id = $1
			           AND po.status IN ('draft', 'approved')
			     ) AS has_pending_po
			 FROM ingredients i
			 LEFT JOIN ingredient_location_configs ilc
			   ON ilc.ingredient_id = i.ingredient_id AND ilc.location_id = $1
			 LEFT JOIN LATERAL (
			     SELECT cl.counted_qty
			     FROM inventory_count_lines cl
			     JOIN inventory_counts ic ON ic.count_id = cl.count_id
			     WHERE cl.ingredient_id = i.ingredient_id
			       AND ic.location_id = $1
			       AND ic.status IN ('submitted', 'approved')
			     ORDER BY ic.started_at DESC
			     LIMIT 1
			 ) latest ON true
			 WHERE i.org_id = $2
			   AND i.status = 'active'
			   AND COALESCE(ilc.reorder_point, 0) > 0`,
			locationID, orgID,
		)
		if err != nil {
			return fmt.Errorf("query par breaches: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var pb PARBreach
			var leadTimeDays int
			if err := rows.Scan(
				&pb.IngredientID, &pb.Name, &pb.CurrentLevel,
				&pb.ReorderPoint, &pb.PARLevel, &pb.AvgDailyUsage, &leadTimeDays,
				&pb.VendorName, &pb.HasPendingPO,
			); err != nil {
				return fmt.Errorf("scan par breach: %w", err)
			}

			eff := effectiveReorderPoint(pb.ReorderPoint, leadTimeDays, pb.AvgDailyUsage)
			if pb.CurrentLevel >= eff {
				continue // not a breach
			}

			if pb.AvgDailyUsage > 0 {
				pb.ProjectedStockoutDays = pb.CurrentLevel / pb.AvgDailyUsage
			} else {
				pb.ProjectedStockoutDays = 0
			}

			results = append(results, pb)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort by projected_stockout_days ASC (most urgent first) — in Go
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].ProjectedStockoutDays < results[j-1].ProjectedStockoutDays; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}

	return results, nil
}
