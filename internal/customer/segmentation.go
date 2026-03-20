package customer

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// SegmentGuest computes an RFM-based segment label for a guest using absolute
// thresholds. recencyDays is days since last visit, visits is total visit
// count, totalSpendCents is lifetime spend in cents.
//
// Returned labels: "champion", "loyal", "potential_loyalist", "at_risk",
// "new", "lapsed", "regular".
func SegmentGuest(recencyDays, visits, totalSpendCents int) string {
	// Score each RFM dimension 1–5.
	r := recencyScore(recencyDays)
	f := frequencyScore(visits)
	m := monetaryScore(totalSpendCents)

	composite := r + f + m // 3–15

	switch {
	case r >= 4 && f >= 4 && m >= 4:
		return "champion"
	case f >= 4 && m >= 4:
		return "loyal"
	case r >= 4 && f >= 2:
		return "potential_loyalist"
	case r <= 2 && f >= 3:
		return "at_risk"
	case visits <= 2 && r >= 4:
		return "new"
	case r == 1:
		return "lapsed"
	case composite >= 6:
		return "regular"
	default:
		return "new"
	}
}

func recencyScore(days int) int {
	switch {
	case days < 7:
		return 5
	case days < 14:
		return 4
	case days < 30:
		return 3
	case days < 60:
		return 2
	default:
		return 1
	}
}

func frequencyScore(visits int) int {
	switch {
	case visits > 20:
		return 5
	case visits > 10:
		return 4
	case visits > 5:
		return 3
	case visits > 2:
		return 2
	default:
		return 1
	}
}

func monetaryScore(spendCents int) int {
	switch {
	case spendCents > 50000:
		return 5
	case spendCents > 20000:
		return 4
	case spendCents > 10000:
		return 3
	case spendCents > 5000:
		return 2
	default:
		return 1
	}
}

// RunSegmentation batches all guests in the org through SegmentGuest and
// writes the result to guest_profiles.segment. Returns count of updated rows.
func (s *Service) RunSegmentation(ctx context.Context, orgID string) (int, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	type guestRow struct {
		guestID     string
		totalVisits int
		totalSpend  int64
		lastVisitAt *time.Time
	}

	var guests []guestRow

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT guest_id, total_visits, total_spend, last_visit_at
			 FROM guest_profiles
			 WHERE org_id = $1`,
			orgID,
		)
		if err != nil {
			return fmt.Errorf("query guests: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var g guestRow
			if err := rows.Scan(&g.guestID, &g.totalVisits, &g.totalSpend, &g.lastVisitAt); err != nil {
				return fmt.Errorf("scan guest: %w", err)
			}
			guests = append(guests, g)
		}
		return rows.Err()
	})
	if err != nil {
		return 0, err
	}

	updated := 0
	now := time.Now()
	for _, g := range guests {
		recencyDays := 9999
		if g.lastVisitAt != nil {
			recencyDays = int(now.Sub(*g.lastVisitAt).Hours() / 24)
		}
		seg := SegmentGuest(recencyDays, g.totalVisits, int(g.totalSpend))

		err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
			_, err := tx.Exec(tenantCtx,
				`UPDATE guest_profiles SET segment = $1, updated_at = now() WHERE guest_id = $2`,
				seg, g.guestID,
			)
			return err
		})
		if err != nil {
			continue
		}
		updated++
	}

	return updated, nil
}
