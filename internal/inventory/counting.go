package inventory

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

type CountSession struct {
	CountID     string     `json:"count_id"`
	OrgID       string     `json:"org_id"`
	LocationID  string     `json:"location_id"`
	CountedBy   string     `json:"counted_by"`
	CountType   string     `json:"count_type"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	ApprovedBy  *string    `json:"approved_by,omitempty"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
}

type CountLine struct {
	CountLineID  string   `json:"count_line_id"`
	IngredientID string   `json:"ingredient_id"`
	Name         string   `json:"name"`
	Category     string   `json:"category"`
	ExpectedQty  *float64 `json:"expected_qty"`
	CountedQty   *float64 `json:"counted_qty"`
	Unit         string   `json:"unit"`
	Note         string   `json:"note"`
}

type CountWithLines struct {
	CountSession
	Lines    []CountLine `json:"lines"`
	Progress Progress    `json:"progress"`
}

type Progress struct {
	Counted int `json:"counted"`
	Total   int `json:"total"`
}

type CountLineInput struct {
	IngredientID string  `json:"ingredient_id"`
	CountedQty   float64 `json:"counted_qty"`
	Unit         string  `json:"unit"`
	Note         string  `json:"note"`
}

func validCountType(ct string) bool {
	return ct == "full" || ct == "spot_check"
}

func countProgress(lines []CountLine) (counted, total int) {
	total = len(lines)
	for _, l := range lines {
		if l.CountedQty != nil {
			counted++
		}
	}
	return
}

func (s *Service) CreateCount(ctx context.Context, orgID, locationID, countedBy, countType string, category string) (*CountSession, error) {
	if !validCountType(countType) {
		return nil, fmt.Errorf("invalid count_type: %s", countType)
	}

	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var cs CountSession

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(tenantCtx,
			`INSERT INTO inventory_counts (org_id, location_id, counted_by, count_type)
			 VALUES ($1, $2, $3, $4)
			 RETURNING count_id, org_id, location_id, counted_by, count_type, status, started_at`,
			orgID, locationID, countedBy, countType,
		).Scan(&cs.CountID, &cs.OrgID, &cs.LocationID, &cs.CountedBy, &cs.CountType, &cs.Status, &cs.StartedAt)
		if err != nil {
			return fmt.Errorf("insert count: %w", err)
		}

		categoryFilter := ""
		args := []any{orgID, cs.CountID, locationID}
		if category != "" && countType == "spot_check" {
			categoryFilter = " AND i.category = $4"
			args = append(args, category)
		}

		_, err = tx.Exec(tenantCtx,
			`INSERT INTO inventory_count_lines (org_id, count_id, location_id, ingredient_id, unit)
			 SELECT i.org_id, $2, $3, i.ingredient_id, i.unit
			 FROM ingredients i
			 WHERE i.org_id = $1 AND i.status = 'active'`+categoryFilter,
			args...,
		)
		if err != nil {
			return fmt.Errorf("populate lines: %w", err)
		}

		return nil
	})

	return &cs, err
}

func (s *Service) GetCount(ctx context.Context, orgID, countID string) (*CountWithLines, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var result CountWithLines

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(tenantCtx,
			`SELECT count_id, org_id, location_id, counted_by, count_type, status, started_at, submitted_at, approved_by, approved_at
			 FROM inventory_counts WHERE count_id = $1`,
			countID,
		).Scan(&result.CountID, &result.OrgID, &result.LocationID, &result.CountedBy,
			&result.CountType, &result.Status, &result.StartedAt, &result.SubmittedAt,
			&result.ApprovedBy, &result.ApprovedAt)
		if err != nil {
			return fmt.Errorf("get count: %w", err)
		}

		rows, err := tx.Query(tenantCtx,
			`SELECT cl.count_line_id, cl.ingredient_id, i.name, i.category,
			        cl.expected_qty, cl.counted_qty, cl.unit, COALESCE(cl.note, '')
			 FROM inventory_count_lines cl
			 JOIN ingredients i ON i.ingredient_id = cl.ingredient_id
			 WHERE cl.count_id = $1
			 ORDER BY i.category, i.name`,
			countID,
		)
		if err != nil {
			return fmt.Errorf("get lines: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var l CountLine
			if err := rows.Scan(&l.CountLineID, &l.IngredientID, &l.Name, &l.Category,
				&l.ExpectedQty, &l.CountedQty, &l.Unit, &l.Note); err != nil {
				return fmt.Errorf("scan line: %w", err)
			}
			result.Lines = append(result.Lines, l)
		}
		return rows.Err()
	})

	if err == nil {
		c, t := countProgress(result.Lines)
		result.Progress = Progress{Counted: c, Total: t}
	}
	return &result, err
}

func (s *Service) UpsertCountLines(ctx context.Context, orgID, countID string, lines []CountLineInput) (int, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var updated int

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		for _, l := range lines {
			tag, err := tx.Exec(tenantCtx,
				`UPDATE inventory_count_lines
				 SET counted_qty = $1, note = $2, updated_at = now()
				 WHERE count_id = $3 AND ingredient_id = $4`,
				l.CountedQty, l.Note, countID, l.IngredientID,
			)
			if err != nil {
				return fmt.Errorf("upsert line %s: %w", l.IngredientID, err)
			}
			updated += int(tag.RowsAffected())
		}
		return nil
	})
	return updated, err
}

func (s *Service) SubmitCount(ctx context.Context, orgID, countID string) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(tenantCtx,
			`UPDATE inventory_counts SET status = 'submitted', submitted_at = now(), updated_at = now()
			 WHERE count_id = $1 AND status = 'in_progress'`,
			countID,
		)
		if err != nil {
			return fmt.Errorf("submit count: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("count not found or already submitted")
		}
		return nil
	})
}

func (s *Service) ApproveCount(ctx context.Context, orgID, countID, approvedBy string) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(tenantCtx,
			`UPDATE inventory_counts SET status = 'approved', approved_by = $2, approved_at = now(), updated_at = now()
			 WHERE count_id = $1 AND status = 'submitted'`,
			countID, approvedBy,
		)
		if err != nil {
			return fmt.Errorf("approve count: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("count not found or not submitted")
		}
		return nil
	})
}
