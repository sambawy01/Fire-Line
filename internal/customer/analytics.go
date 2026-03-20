package customer

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// SegmentBucket holds the count of guests in a named segment.
type SegmentBucket struct {
	Segment string `json:"segment"`
	Count   int    `json:"count"`
}

// ChurnBucket holds the count of guests at a churn risk level.
type ChurnBucket struct {
	Risk  string `json:"risk"`
	Count int    `json:"count"`
}

// CLVBucket holds the count of guests in a CLV dollar-range bucket.
type CLVBucket struct {
	Label string `json:"label"`
	Min   int    `json:"min"`
	Max   int    `json:"max"` // -1 = unbounded
	Count int    `json:"count"`
}

// GetSegmentDistribution returns per-segment guest counts for the org.
func (s *Service) GetSegmentDistribution(ctx context.Context, orgID string) ([]SegmentBucket, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var buckets []SegmentBucket

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT COALESCE(segment, 'unknown') AS segment, COUNT(*)::INT
			 FROM guest_profiles
			 WHERE org_id = $1
			 GROUP BY segment
			 ORDER BY COUNT(*) DESC`,
			orgID,
		)
		if err != nil {
			return fmt.Errorf("query segment distribution: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var b SegmentBucket
			if err := rows.Scan(&b.Segment, &b.Count); err != nil {
				return fmt.Errorf("scan segment bucket: %w", err)
			}
			buckets = append(buckets, b)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if buckets == nil {
		buckets = []SegmentBucket{}
	}
	return buckets, nil
}

// GetChurnDistribution returns per-risk-tier guest counts for the org.
func (s *Service) GetChurnDistribution(ctx context.Context, orgID string) ([]ChurnBucket, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var buckets []ChurnBucket

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT COALESCE(churn_risk, 'unknown') AS churn_risk, COUNT(*)::INT
			 FROM guest_profiles
			 WHERE org_id = $1
			 GROUP BY churn_risk
			 ORDER BY CASE churn_risk
			     WHEN 'low'      THEN 1
			     WHEN 'medium'   THEN 2
			     WHEN 'high'     THEN 3
			     WHEN 'critical' THEN 4
			     ELSE 5
			 END`,
			orgID,
		)
		if err != nil {
			return fmt.Errorf("query churn distribution: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var b ChurnBucket
			if err := rows.Scan(&b.Risk, &b.Count); err != nil {
				return fmt.Errorf("scan churn bucket: %w", err)
			}
			buckets = append(buckets, b)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if buckets == nil {
		buckets = []ChurnBucket{}
	}
	return buckets, nil
}

// clvBuckets defines the fixed CLV histogram breaks (dollar values).
var clvBuckets = []CLVBucket{
	{Label: "$0–50", Min: 0, Max: 50},
	{Label: "$50–200", Min: 50, Max: 200},
	{Label: "$200–500", Min: 200, Max: 500},
	{Label: "$500–1000", Min: 500, Max: 1000},
	{Label: "$1000+", Min: 1000, Max: -1},
}

// GetCLVDistribution returns a histogram of guest CLV scores across fixed
// dollar-range buckets.
func (s *Service) GetCLVDistribution(ctx context.Context, orgID string) ([]CLVBucket, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	result := make([]CLVBucket, len(clvBuckets))
	copy(result, clvBuckets)

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT
			     COUNT(*) FILTER (WHERE clv_score >= 0   AND clv_score < 50)   AS b0,
			     COUNT(*) FILTER (WHERE clv_score >= 50  AND clv_score < 200)  AS b1,
			     COUNT(*) FILTER (WHERE clv_score >= 200 AND clv_score < 500)  AS b2,
			     COUNT(*) FILTER (WHERE clv_score >= 500 AND clv_score < 1000) AS b3,
			     COUNT(*) FILTER (WHERE clv_score >= 1000)                     AS b4
			 FROM guest_profiles
			 WHERE org_id = $1`,
			orgID,
		)
		if err != nil {
			return fmt.Errorf("query CLV distribution: %w", err)
		}
		defer rows.Close()

		if rows.Next() {
			var b0, b1, b2, b3, b4 int
			if err := rows.Scan(&b0, &b1, &b2, &b3, &b4); err != nil {
				return fmt.Errorf("scan CLV row: %w", err)
			}
			result[0].Count = b0
			result[1].Count = b1
			result[2].Count = b2
			result[3].Count = b3
			result[4].Count = b4
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
