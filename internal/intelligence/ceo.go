package intelligence

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// ─── CEO Briefing Structs ───────────────────────────────────────────────────

// CEOBriefing is a cross-location intelligence summary for executive oversight.
type CEOBriefing struct {
	GeneratedAt          string             `json:"generated_at"`
	FraudRiskScore       float64            `json:"fraud_risk_score"`
	WorkforceHealthScore float64            `json:"workforce_health_score"`
	OpenAnomalies        int                `json:"open_anomalies"`
	CriticalAnomalies    int                `json:"critical_anomalies"`
	LocationScores       []LocationScore    `json:"location_scores"`
	TurnoverRisks        []TurnoverRisk     `json:"turnover_risks"`
	StaffingAlerts       []StaffingAlert    `json:"staffing_alerts"`
	TopPerformers        []PerformerSummary `json:"top_performers"`
	TrainingROI          []TrainingInsight  `json:"training_roi"`
}

// LocationScore aggregates operational health metrics for a single location.
type LocationScore struct {
	LocationID     string  `json:"location_id"`
	LocationName   string  `json:"location_name"`
	AnomalyCount   int     `json:"anomaly_count"`
	TaskCompletion float64 `json:"task_completion_rate"`
	AttendanceRate float64 `json:"attendance_rate"`
	LaborCostPct   float64 `json:"labor_cost_pct"`
	RiskLevel      string  `json:"risk_level"`
}

// TurnoverRisk identifies employees showing signs of potential turnover.
type TurnoverRisk struct {
	EmployeeID  string   `json:"employee_id"`
	DisplayName string   `json:"display_name"`
	LocationID  string   `json:"location_id"`
	RiskScore   float64  `json:"risk_score"`
	Signals     []string `json:"signals"`
}

// StaffingAlert flags locations with insufficient staffing.
type StaffingAlert struct {
	LocationID   string `json:"location_id"`
	LocationName string `json:"location_name"`
	Date         string `json:"date"`
	Scheduled    int    `json:"scheduled_staff"`
	Required     int    `json:"required_staff"`
	Gap          int    `json:"gap"`
}

// PerformerSummary highlights top-performing employees.
type PerformerSummary struct {
	EmployeeID  string  `json:"employee_id"`
	DisplayName string  `json:"display_name"`
	LocationID  string  `json:"location_id"`
	Points      float64 `json:"points"`
	Role        string  `json:"role"`
}

// TrainingInsight compares performance of certified vs uncertified employees.
type TrainingInsight struct {
	Certification    string  `json:"certification"`
	CertifiedCount   int     `json:"certified_count"`
	UncertifiedCount int     `json:"uncertified_count"`
	AvgELUCertified  float64 `json:"avg_elu_certified"`
	AvgELUUncertified float64 `json:"avg_elu_uncertified"`
	Lift             float64 `json:"lift_pct"`
}

// ─── GetCEOBriefing ─────────────────────────────────────────────────────────

// GetCEOBriefing computes a cross-location intelligence summary for executive review.
func (s *Service) GetCEOBriefing(ctx context.Context, orgID string) (*CEOBriefing, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	briefing := &CEOBriefing{
		GeneratedAt:    time.Now().UTC().Format(time.RFC3339),
		LocationScores: []LocationScore{},
		TurnoverRisks:  []TurnoverRisk{},
		StaffingAlerts: []StaffingAlert{},
		TopPerformers:  []PerformerSummary{},
		TrainingROI:    []TrainingInsight{},
	}

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		// ── Anomaly counts ──────────────────────────────────────────────
		if err := tx.QueryRow(tenantCtx,
			`SELECT
				COUNT(*) FILTER (WHERE status = 'open')::INT,
				COUNT(*) FILTER (WHERE status = 'open' AND severity = 'critical')::INT
			FROM anomalies`,
		).Scan(&briefing.OpenAnomalies, &briefing.CriticalAnomalies); err != nil {
			return fmt.Errorf("anomaly counts: %w", err)
		}

		// Fraud risk score: 100 - (open_anomalies * 10), min 0.
		briefing.FraudRiskScore = math.Max(0, 100-float64(briefing.OpenAnomalies)*10)

		// ── Location scores ─────────────────────────────────────────────
		locRows, err := tx.Query(tenantCtx,
			`SELECT l.location_id, l.name,
				-- Anomaly count at this location
				(SELECT COUNT(*)::INT FROM anomalies a
				 WHERE a.location_id = l.location_id AND a.status = 'open'),
				-- Task completion rate (last 30 days)
				COALESCE((
					SELECT CASE WHEN COUNT(*) > 0
						THEN COUNT(*) FILTER (WHERE t.status = 'completed')::FLOAT / COUNT(*)::FLOAT
						ELSE 0 END
					FROM tasks t WHERE t.location_id = l.location_id
						AND t.created_at >= NOW() - INTERVAL '30 days'
				), 0),
				-- Attendance rate: completed shifts / total shifts (last 30 days)
				COALESCE((
					SELECT CASE WHEN COUNT(*) > 0
						THEN COUNT(*) FILTER (WHERE s.status = 'completed')::FLOAT / COUNT(*)::FLOAT
						ELSE 0 END
					FROM shifts s WHERE s.location_id = l.location_id
						AND s.clock_in >= NOW() - INTERVAL '30 days'
				), 0),
				-- Labor cost % vs revenue (last 30 days)
				COALESCE((
					SELECT CASE WHEN COALESCE(SUM(c.total), 0) > 0
						THEN (
							SELECT COALESCE(SUM(EXTRACT(EPOCH FROM (sh.clock_out - sh.clock_in)) / 3600.0 * sh.hourly_rate), 0)
							FROM shifts sh
							WHERE sh.location_id = l.location_id
								AND sh.clock_out IS NOT NULL
								AND sh.status = 'completed'
								AND sh.clock_in >= NOW() - INTERVAL '30 days'
						)::FLOAT / SUM(c.total)::FLOAT * 100
						ELSE 0 END
					FROM checks c
					WHERE c.location_id = l.location_id
						AND c.status IN ('closed', 'paid')
						AND c.created_at >= NOW() - INTERVAL '30 days'
				), 0)
			FROM locations l
			WHERE l.org_id = $1 AND l.status = 'active'
			ORDER BY l.name`,
			orgID,
		)
		if err != nil {
			return fmt.Errorf("location scores query: %w", err)
		}
		defer locRows.Close()

		var healthScoreSum float64
		var healthScoreCount int

		for locRows.Next() {
			var ls LocationScore
			if err := locRows.Scan(
				&ls.LocationID, &ls.LocationName,
				&ls.AnomalyCount, &ls.TaskCompletion, &ls.AttendanceRate, &ls.LaborCostPct,
			); err != nil {
				return err
			}

			ls.TaskCompletion = math.Round(ls.TaskCompletion*10000) / 100
			ls.AttendanceRate = math.Round(ls.AttendanceRate*10000) / 100
			ls.LaborCostPct = math.Round(ls.LaborCostPct*100) / 100

			// Determine risk level based on anomaly count and task completion.
			switch {
			case ls.AnomalyCount >= 5 || ls.TaskCompletion < 50:
				ls.RiskLevel = "high"
			case ls.AnomalyCount >= 2 || ls.TaskCompletion < 75:
				ls.RiskLevel = "medium"
			default:
				ls.RiskLevel = "low"
			}

			// Accumulate for workforce health score.
			locHealth := (ls.AttendanceRate + ls.TaskCompletion) / 2
			healthScoreSum += locHealth
			healthScoreCount++

			briefing.LocationScores = append(briefing.LocationScores, ls)
		}
		if err := locRows.Err(); err != nil {
			return err
		}

		if healthScoreCount > 0 {
			briefing.WorkforceHealthScore = math.Round(healthScoreSum/float64(healthScoreCount)*100) / 100
		}

		// ── Turnover risks ──────────────────────────────────────────────
		turnoverRows, err := tx.Query(tenantCtx,
			`SELECT e.employee_id, COALESCE(e.display_name, ''),
				COALESCE(e.location_id::TEXT, ''),
				COALESCE((
					SELECT SUM(spe.points)
					FROM staff_point_events spe
					WHERE spe.employee_id = e.employee_id
						AND spe.created_at >= NOW() - INTERVAL '30 days'
				), 0) AS points_delta,
				COALESCE((
					SELECT COUNT(*)::INT FROM shifts s
					WHERE s.employee_id = e.employee_id
						AND s.clock_in >= NOW() - INTERVAL '30 days'
						AND s.status = 'no_show'
				), 0) AS late_count,
				0 AS swap_count
			FROM employees e
			WHERE e.org_id = $1 AND e.status = 'active'
			ORDER BY points_delta ASC
			LIMIT 20`,
			orgID,
		)
		if err != nil {
			return fmt.Errorf("turnover risks query: %w", err)
		}
		defer turnoverRows.Close()

		for turnoverRows.Next() {
			var tr TurnoverRisk
			var pointsDelta float64
			var lateCount, swapCount int

			if err := turnoverRows.Scan(
				&tr.EmployeeID, &tr.DisplayName, &tr.LocationID,
				&pointsDelta, &lateCount, &swapCount,
			); err != nil {
				return err
			}

			tr.Signals = []string{}
			var riskScore float64

			if pointsDelta < 0 {
				tr.Signals = append(tr.Signals, "declining_points")
				riskScore += math.Min(40, math.Abs(pointsDelta)*4)
			}
			if lateCount >= 3 {
				tr.Signals = append(tr.Signals, "late_clock_ins")
				riskScore += math.Min(30, float64(lateCount)*5)
			}
			if swapCount >= 3 {
				tr.Signals = append(tr.Signals, "frequent_swaps")
				riskScore += math.Min(30, float64(swapCount)*5)
			}

			tr.RiskScore = math.Min(100, math.Round(riskScore*100)/100)
			briefing.TurnoverRisks = append(briefing.TurnoverRisks, tr)
		}
		if err := turnoverRows.Err(); err != nil {
			return err
		}

		// ── Staffing alerts ─────────────────────────────────────────────
		// Check locations with < 3 scheduled shifts in the next 7 days.
		{
			saRows, err := tx.Query(tenantCtx,
				`SELECT l.location_id::TEXT, l.name,
					COUNT(DISTINCT s.employee_id)::INT AS scheduled
				FROM locations l
				LEFT JOIN shifts s ON s.location_id = l.location_id
					AND s.clock_in >= CURRENT_DATE
					AND s.clock_in < CURRENT_DATE + INTERVAL '7 days'
				WHERE l.org_id = $1 AND l.status = 'active'
				GROUP BY l.location_id, l.name
				HAVING COUNT(DISTINCT s.employee_id) < 5
				ORDER BY scheduled ASC
				LIMIT 20`,
				orgID,
			)
			if err == nil {
				defer saRows.Close()
				for saRows.Next() {
					var sa StaffingAlert
					if err := saRows.Scan(&sa.LocationID, &sa.LocationName, &sa.Scheduled); err != nil {
						break
					}
					sa.Date = "next 7 days"
					sa.Required = 5
					sa.Gap = sa.Required - sa.Scheduled
					briefing.StaffingAlerts = append(briefing.StaffingAlerts, sa)
				}
			}
		}

		// ── Top performers ──────────────────────────────────────────────
		perfRows, err := tx.Query(tenantCtx,
			`SELECT e.employee_id, COALESCE(e.display_name, ''),
				COALESCE(e.location_id::TEXT, ''),
				COALESCE(e.staff_points, 0),
				COALESCE(e.role, '')
			FROM employees e
			WHERE e.org_id = $1 AND e.status = 'active'
			ORDER BY e.staff_points DESC NULLS LAST
			LIMIT 5`,
			orgID,
		)
		if err != nil {
			return fmt.Errorf("top performers query: %w", err)
		}
		defer perfRows.Close()

		for perfRows.Next() {
			var ps PerformerSummary
			if err := perfRows.Scan(&ps.EmployeeID, &ps.DisplayName, &ps.LocationID, &ps.Points, &ps.Role); err != nil {
				return err
			}
			briefing.TopPerformers = append(briefing.TopPerformers, ps)
		}
		if err := perfRows.Err(); err != nil {
			return err
		}

		// ── Training ROI ────────────────────────────────────────────────
		// Compare average ELU rating of certified vs uncertified employees.
		// certifications is TEXT[], elu_ratings is JSONB with station keys.
		roiRows, err := tx.Query(tenantCtx,
			`WITH cert_list AS (
				SELECT DISTINCT unnest(certifications) AS cert
				FROM employees
				WHERE org_id = $1 AND certifications IS NOT NULL AND array_length(certifications, 1) > 0
			)
			SELECT c.cert,
				(SELECT COUNT(*)::INT FROM employees e
				 WHERE e.org_id = $1 AND e.status = 'active'
					AND c.cert = ANY(e.certifications)
				) AS certified_count,
				(SELECT COUNT(*)::INT FROM employees e
				 WHERE e.org_id = $1 AND e.status = 'active'
					AND (e.certifications IS NULL OR NOT c.cert = ANY(e.certifications))
				) AS uncertified_count,
				COALESCE((
					SELECT AVG(v::NUMERIC) FROM employees e,
						jsonb_each_text(e.elu_ratings) AS kv(k, v)
					WHERE e.org_id = $1 AND e.status = 'active'
						AND c.cert = ANY(e.certifications)
						AND e.elu_ratings IS NOT NULL
				), 0) AS avg_elu_cert,
				COALESCE((
					SELECT AVG(v::NUMERIC) FROM employees e,
						jsonb_each_text(e.elu_ratings) AS kv(k, v)
					WHERE e.org_id = $1 AND e.status = 'active'
						AND (e.certifications IS NULL OR NOT c.cert = ANY(e.certifications))
						AND e.elu_ratings IS NOT NULL
				), 0) AS avg_elu_uncert
			FROM cert_list c
			ORDER BY c.cert`,
			orgID,
		)
		if err != nil {
			return fmt.Errorf("training ROI query: %w", err)
		}
		defer roiRows.Close()

		for roiRows.Next() {
			var ti TrainingInsight
			if err := roiRows.Scan(
				&ti.Certification,
				&ti.CertifiedCount, &ti.UncertifiedCount,
				&ti.AvgELUCertified, &ti.AvgELUUncertified,
			); err != nil {
				return err
			}
			ti.AvgELUCertified = math.Round(ti.AvgELUCertified*100) / 100
			ti.AvgELUUncertified = math.Round(ti.AvgELUUncertified*100) / 100
			if ti.AvgELUUncertified > 0 {
				ti.Lift = math.Round((ti.AvgELUCertified-ti.AvgELUUncertified)/ti.AvgELUUncertified*10000) / 100
			}
			briefing.TrainingROI = append(briefing.TrainingROI, ti)
		}
		if err := roiRows.Err(); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get CEO briefing: %w", err)
	}
	return briefing, nil
}
