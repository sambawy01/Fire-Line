package customer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opsnerve/fireline/internal/event"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// Service provides customer intelligence capabilities including AI-powered
// segmentation and summary generation via Ollama.
type Service struct {
	pool   *pgxpool.Pool
	bus    *event.Bus
	ollama *OllamaClient
}

// New creates a new customer intelligence service.
func New(pool *pgxpool.Pool, bus *event.Bus, ollama *OllamaClient) *Service {
	return &Service{pool: pool, bus: bus, ollama: ollama}
}

// CustomerDetail holds full customer profile data including AI fields.
type CustomerDetail struct {
	CustomerID         string     `json:"customer_id"`
	Name               string     `json:"name"`
	Email              string     `json:"email"`
	Phone              string     `json:"phone"`
	FirstVisit         *time.Time `json:"first_visit"`
	LastVisit          *time.Time `json:"last_visit"`
	TotalVisits        int        `json:"total_visits"`
	TotalSpend         int64      `json:"total_spend"`
	AvgCheck           int64      `json:"avg_check"`
	Segment            string     `json:"segment"`
	AISummary          string     `json:"ai_summary"`
	AISummaryUpdatedAt *time.Time `json:"ai_summary_updated_at"`
}

// CustomerSummary holds location-wide customer KPI rollups.
type CustomerSummary struct {
	TotalCustomers   int            `json:"total_customers"`
	AvgLifetimeValue int64          `json:"avg_lifetime_value"`
	VIPCount         int            `json:"vip_count"`
	AtRiskCount      int            `json:"at_risk_count"`
	SegmentCounts    map[string]int `json:"segment_counts"`
}

// AnalyzeResult reports the outcome of a batch AI analysis run.
type AnalyzeResult struct {
	Analyzed int    `json:"analyzed"`
	Errors   int    `json:"errors"`
	Message  string `json:"message"`
}

// validSegments is the set of labels Ollama is expected to return.
var validSegments = map[string]bool{
	"new":      true,
	"regular":  true,
	"vip":      true,
	"lapsed":   true,
	"at_risk":  true,
}

// GetCustomers returns all customers for the given location ordered by total
// spend descending.
func (s *Service) GetCustomers(ctx context.Context, orgID, locationID string) ([]CustomerDetail, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var customers []CustomerDetail

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT customer_id, name, COALESCE(email,''), COALESCE(phone,''),
			        first_visit, last_visit, total_visits, total_spend, avg_check,
			        segment, COALESCE(ai_summary,''), ai_summary_updated_at
			 FROM customers
			 WHERE location_id = $1
			 ORDER BY total_spend DESC`,
			locationID,
		)
		if err != nil {
			return fmt.Errorf("query customers: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var c CustomerDetail
			if err := rows.Scan(
				&c.CustomerID,
				&c.Name,
				&c.Email,
				&c.Phone,
				&c.FirstVisit,
				&c.LastVisit,
				&c.TotalVisits,
				&c.TotalSpend,
				&c.AvgCheck,
				&c.Segment,
				&c.AISummary,
				&c.AISummaryUpdatedAt,
			); err != nil {
				return fmt.Errorf("scan customer row: %w", err)
			}
			customers = append(customers, c)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	if customers == nil {
		customers = []CustomerDetail{}
	}
	return customers, nil
}

// GetSummary returns location-wide customer KPI rollups.
func (s *Service) GetSummary(ctx context.Context, orgID, locationID string) (*CustomerSummary, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	summary := &CustomerSummary{
		SegmentCounts: make(map[string]int),
	}

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT segment, COUNT(*)::INT, COALESCE(AVG(total_spend), 0)::BIGINT
			 FROM customers
			 WHERE location_id = $1
			 GROUP BY segment`,
			locationID,
		)
		if err != nil {
			return fmt.Errorf("query customer summary: %w", err)
		}
		defer rows.Close()

		var totalCustomers int
		var totalSpend int64

		for rows.Next() {
			var segment string
			var count int
			var avgSpend int64
			if err := rows.Scan(&segment, &count, &avgSpend); err != nil {
				return fmt.Errorf("scan summary row: %w", err)
			}
			summary.SegmentCounts[segment] = count
			totalCustomers += count
			totalSpend += avgSpend * int64(count)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		summary.TotalCustomers = totalCustomers
		if totalCustomers > 0 {
			summary.AvgLifetimeValue = totalSpend / int64(totalCustomers)
		}

		// VIP count.
		summary.VIPCount = summary.SegmentCounts["vip"]

		// At-risk count = at_risk + lapsed.
		summary.AtRiskCount = summary.SegmentCounts["at_risk"] + summary.SegmentCounts["lapsed"]

		return nil
	})
	if err != nil {
		return nil, err
	}

	return summary, nil
}

// AnalyzeAll runs AI segmentation and summary generation for every customer in
// the location. Customers are processed sequentially to avoid overloading
// Ollama. If Ollama is unreachable on the very first call, the function returns
// immediately with an error.
func (s *Service) AnalyzeAll(ctx context.Context, orgID, locationID string) (*AnalyzeResult, error) {
	customers, err := s.GetCustomers(ctx, orgID, locationID)
	if err != nil {
		return nil, fmt.Errorf("fetch customers: %w", err)
	}

	result := &AnalyzeResult{}

	for i, c := range customers {
		// Compute time-based metrics.
		now := time.Now()
		var daysSinceFirst, daysSinceLast int
		if c.FirstVisit != nil {
			daysSinceFirst = int(now.Sub(*c.FirstVisit).Hours() / 24)
		}
		if c.LastVisit != nil {
			daysSinceLast = int(now.Sub(*c.LastVisit).Hours() / 24)
		}

		spendDollars := float64(c.TotalSpend) / 100.0
		avgCheckDollars := float64(c.AvgCheck) / 100.0

		// --- Step 1: segmentation ---
		segPrompt := buildSegmentPrompt(c.TotalVisits, spendDollars, avgCheckDollars, daysSinceFirst, daysSinceLast)
		segResponse, err := s.ollama.Generate(ctx, segPrompt)
		if err != nil {
			// On the first customer, treat an unreachable Ollama as a fatal error.
			if i == 0 {
				return nil, fmt.Errorf("ollama unreachable: %w", err)
			}
			result.Errors++
			continue
		}

		newSegment := strings.ToLower(strings.TrimSpace(segResponse))
		if !validSegments[newSegment] {
			// Invalid label — keep the existing segment.
			newSegment = c.Segment
		}

		// --- Step 2: actionable summary ---
		sumPrompt := buildSummaryPrompt(c.Name, newSegment, c.TotalVisits, spendDollars, avgCheckDollars, daysSinceLast)
		aiSummary, err := s.ollama.Generate(ctx, sumPrompt)
		if err != nil {
			result.Errors++
			continue
		}

		// --- Step 3: persist ---
		tenantCtx := tenant.WithOrgID(ctx, orgID)
		updateErr := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
			_, err := tx.Exec(tenantCtx,
				`UPDATE customers
				 SET segment = $1, ai_summary = $2, ai_summary_updated_at = now()
				 WHERE customer_id = $3`,
				newSegment, aiSummary, c.CustomerID,
			)
			return err
		})
		if updateErr != nil {
			result.Errors++
			continue
		}

		result.Analyzed++
	}

	result.Message = fmt.Sprintf("analyzed %d customers, %d errors", result.Analyzed, result.Errors)
	return result, nil
}

func buildSegmentPrompt(visits int, spendDollars, avgCheckDollars float64, daysSinceFirst, daysSinceLast int) string {
	return fmt.Sprintf(
		`You are a restaurant customer analyst. Given this customer data:
- Total visits: %d
- Total spend: $%.2f
- Average check: $%.2f
- Days since first visit: %d
- Days since last visit: %d

Classify as exactly ONE of: new, regular, vip, at_risk, lapsed
Reply with ONLY the label.`,
		visits, spendDollars, avgCheckDollars, daysSinceFirst, daysSinceLast,
	)
}

func buildSummaryPrompt(name, segment string, visits int, spendDollars, avgCheckDollars float64, daysSinceLast int) string {
	return fmt.Sprintf(
		`You are a restaurant manager's AI assistant. Write a 1-2 sentence actionable insight:
- Name: %s
- Segment: %s
- Total visits: %d
- Total spend: $%.2f
- Average check: $%.2f
- Days since last visit: %d

Be specific and actionable.`,
		name, segment, visits, spendDollars, avgCheckDollars, daysSinceLast,
	)
}
