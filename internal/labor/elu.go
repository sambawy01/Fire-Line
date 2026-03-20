package labor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// EmployeeProfile holds ELU ratings, staff points, certifications, and
// availability for a single employee.
type EmployeeProfile struct {
	EmployeeID     string             `json:"employee_id"`
	DisplayName    string             `json:"display_name"`
	Role           string             `json:"role"`
	Status         string             `json:"status"`
	ELURatings     map[string]float64 `json:"elu_ratings"`
	StaffPoints    float64            `json:"staff_points"`
	PointsTrend    string             `json:"points_trend"`
	Certifications []string           `json:"certifications"`
	Availability   map[string]any     `json:"availability"`
}

// computePointsTrend returns "up", "down", or "stable" based on the
// difference between current staff_points and the total points awarded
// more than 7 days ago (baseline).
//
// up:     diff > 5
// down:   diff < -5
// stable: diff within ±5
func computePointsTrend(current, sevenDaysAgo float64) string {
	diff := current - sevenDaysAgo
	switch {
	case diff > 5:
		return "up"
	case diff < -5:
		return "down"
	default:
		return "stable"
	}
}

// GetEmployeeProfile returns the full ELU profile for a single employee.
// Points trend is derived by comparing current staff_points to the sum of
// point events older than 7 days (i.e., the baseline before the last week).
func (s *Service) GetEmployeeProfile(ctx context.Context, orgID, employeeID string) (*EmployeeProfile, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var profile EmployeeProfile
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		var eluJSON, availJSON []byte
		var certsJSON []byte
		var staffPoints float64

		err := tx.QueryRow(tenantCtx,
			`SELECT
			    e.employee_id::TEXT,
			    e.display_name,
			    e.role,
			    e.status,
			    e.elu_ratings,
			    e.staff_points,
			    COALESCE(array_to_json(e.certifications), '[]')::TEXT,
			    e.availability
			FROM employees e
			WHERE e.employee_id = $1`,
			employeeID,
		).Scan(
			&profile.EmployeeID,
			&profile.DisplayName,
			&profile.Role,
			&profile.Status,
			&eluJSON,
			&staffPoints,
			&certsJSON,
			&availJSON,
		)
		if err != nil {
			return fmt.Errorf("query employee profile: %w", err)
		}

		profile.StaffPoints = staffPoints

		if err := json.Unmarshal(eluJSON, &profile.ELURatings); err != nil {
			profile.ELURatings = map[string]float64{}
		}
		if profile.ELURatings == nil {
			profile.ELURatings = map[string]float64{}
		}

		if err := json.Unmarshal(availJSON, &profile.Availability); err != nil {
			profile.Availability = map[string]any{}
		}
		if profile.Availability == nil {
			profile.Availability = map[string]any{}
		}

		// certifications is a TEXT[] — pgx scans it as []string
		if certsJSON != nil {
			var certs []string
			if err := json.Unmarshal(certsJSON, &certs); err == nil {
				profile.Certifications = certs
			}
		}
		if profile.Certifications == nil {
			profile.Certifications = []string{}
		}

		// Compute points trend: sum of events from before 7 days ago = baseline
		var sevenDaysAgo float64
		err = tx.QueryRow(tenantCtx,
			`SELECT COALESCE(SUM(points), 0)
			 FROM staff_point_events
			 WHERE employee_id = $1
			   AND created_at < now() - INTERVAL '7 days'`,
			employeeID,
		).Scan(&sevenDaysAgo)
		if err != nil {
			return fmt.Errorf("query points baseline: %w", err)
		}

		profile.PointsTrend = computePointsTrend(staffPoints, sevenDaysAgo)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// ListEmployeeProfiles returns ELU profiles for all employees at a location.
func (s *Service) ListEmployeeProfiles(ctx context.Context, orgID, locationID string) ([]EmployeeProfile, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var profiles []EmployeeProfile
	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT
			    e.employee_id::TEXT,
			    e.display_name,
			    e.role,
			    e.status,
			    e.elu_ratings,
			    e.staff_points,
			    COALESCE(array_to_json(e.certifications), '[]')::TEXT,
			    e.availability
			FROM employees e
			WHERE e.location_id = $1
			ORDER BY e.display_name`,
			locationID,
		)
		if err != nil {
			return fmt.Errorf("query employee profiles: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var p EmployeeProfile
			var eluJSON, availJSON []byte
			var certsJSON []byte
			var staffPoints float64

			if err := rows.Scan(
				&p.EmployeeID,
				&p.DisplayName,
				&p.Role,
				&p.Status,
				&eluJSON,
				&staffPoints,
				&certsJSON,
				&availJSON,
			); err != nil {
				return fmt.Errorf("scan profile row: %w", err)
			}

			p.StaffPoints = staffPoints

			if err := json.Unmarshal(eluJSON, &p.ELURatings); err != nil {
				p.ELURatings = map[string]float64{}
			}
			if p.ELURatings == nil {
				p.ELURatings = map[string]float64{}
			}

			if err := json.Unmarshal(availJSON, &p.Availability); err != nil {
				p.Availability = map[string]any{}
			}
			if p.Availability == nil {
				p.Availability = map[string]any{}
			}

			if certsJSON != nil {
				var certs []string
				if err := json.Unmarshal(certsJSON, &certs); err == nil {
					p.Certifications = certs
				}
			}
			if p.Certifications == nil {
				p.Certifications = []string{}
			}

			// Compute trend per employee
			var sevenDaysAgo float64
			if err := tx.QueryRow(tenantCtx,
				`SELECT COALESCE(SUM(points), 0)
				 FROM staff_point_events
				 WHERE employee_id = $1
				   AND created_at < now() - INTERVAL '7 days'`,
				p.EmployeeID,
			).Scan(&sevenDaysAgo); err != nil {
				return fmt.Errorf("query points baseline for %s: %w", p.EmployeeID, err)
			}

			p.PointsTrend = computePointsTrend(staffPoints, sevenDaysAgo)
			profiles = append(profiles, p)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate profile rows: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if profiles == nil {
		profiles = []EmployeeProfile{}
	}
	return profiles, nil
}

// UpdateELURatings overwrites the ELU ratings for an employee.
func (s *Service) UpdateELURatings(ctx context.Context, orgID, employeeID string, ratings map[string]float64) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	data, err := json.Marshal(ratings)
	if err != nil {
		return fmt.Errorf("marshal elu ratings: %w", err)
	}
	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(tenantCtx,
			`UPDATE employees SET elu_ratings = $1 WHERE employee_id = $2`,
			data, employeeID,
		)
		return err
	})
}

// UpdateAvailability overwrites the availability schedule for an employee.
func (s *Service) UpdateAvailability(ctx context.Context, orgID, employeeID string, availability map[string]any) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	data, err := json.Marshal(availability)
	if err != nil {
		return fmt.Errorf("marshal availability: %w", err)
	}
	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(tenantCtx,
			`UPDATE employees SET availability = $1 WHERE employee_id = $2`,
			data, employeeID,
		)
		return err
	})
}

// UpdateCertifications overwrites the certifications list for an employee.
func (s *Service) UpdateCertifications(ctx context.Context, orgID, employeeID string, certs []string) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(tenantCtx,
			`UPDATE employees SET certifications = $1 WHERE employee_id = $2`,
			certs, employeeID,
		)
		return err
	})
}
