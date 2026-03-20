package operations

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// KitchenStation represents a kitchen production station.
type KitchenStation struct {
	StationID     string    `json:"station_id"`
	OrgID         string    `json:"org_id"`
	LocationID    string    `json:"location_id"`
	Name          string    `json:"name"`
	StationType   string    `json:"station_type"`
	MaxConcurrent int       `json:"max_concurrent"`
	Status        string    `json:"status"`
	CurrentLoad   int       `json:"current_load"`
	CreatedAt     time.Time `json:"created_at"`
}

// ResourceProfile represents a single task a station must perform for a menu item.
type ResourceProfile struct {
	ProfileID    string    `json:"profile_id"`
	OrgID        string    `json:"org_id"`
	MenuItemID   string    `json:"menu_item_id"`
	StationType  string    `json:"station_type"`
	TaskSequence int       `json:"task_sequence"`
	DurationSecs int       `json:"duration_secs"`
	ELURequired  float64   `json:"elu_required"`
	BatchSize    int       `json:"batch_size"`
	CreatedAt    time.Time `json:"created_at"`
}

// KitchenCapacity holds per-station load and an overall capacity summary.
type KitchenCapacity struct {
	LocationID      string           `json:"location_id"`
	OverallLoadPct  float64          `json:"overall_load_pct"`
	StationCapacity []StationLoad    `json:"station_capacity"`
	ComputedAt      time.Time        `json:"computed_at"`
}

// StationLoad holds capacity data for a single station.
type StationLoad struct {
	StationID     string  `json:"station_id"`
	Name          string  `json:"name"`
	StationType   string  `json:"station_type"`
	MaxConcurrent int     `json:"max_concurrent"`
	CurrentLoad   int     `json:"current_load"`
	LoadPct       float64 `json:"load_pct"`
}

// TicketTimeEstimate holds the estimated preparation time for a set of menu items.
type TicketTimeEstimate struct {
	MenuItemIDs       []string       `json:"menu_item_ids"`
	EstimatedSecs     int            `json:"estimated_secs"`
	PerStationBreakdown []StationTime `json:"per_station_breakdown"`
}

// StationTime holds the time contribution of one station type.
type StationTime struct {
	StationType  string `json:"station_type"`
	DurationSecs int    `json:"duration_secs"`
}

// defaultStationTimes returns fallback duration in seconds per station type.
func defaultStationTimes() map[string]int {
	return map[string]int{
		"grill":  420,
		"fryer":  300,
		"saute":  360,
		"prep":   180,
		"expo":   60,
		"dish":   120,
	}
}

// loadPct computes load percentage given current and max values.
// Returns 0 if max is 0.
func loadPct(current, max int) float64 {
	if max == 0 {
		return 0.0
	}
	return float64(current) / float64(max) * 100
}

// SetupDefaultStations inserts 6 default stations for a location.
// Existing stations for the location are not affected.
func (s *Service) SetupDefaultStations(ctx context.Context, orgID, locationID string) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	type stationDef struct {
		name        string
		stationType string
		maxConcurrent int
	}

	defaults := []stationDef{
		{"Grill", "grill", 4},
		{"Fryer", "fryer", 3},
		{"Saute", "saute", 3},
		{"Prep", "prep", 4},
		{"Expo", "expo", 2},
		{"Dish", "dish", 2},
	}

	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		for _, d := range defaults {
			if _, err := tx.Exec(tenantCtx,
				`INSERT INTO kitchen_stations (org_id, location_id, name, station_type, max_concurrent)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT DO NOTHING`,
				orgID, locationID, d.name, d.stationType, d.maxConcurrent,
			); err != nil {
				return fmt.Errorf("insert station %s: %w", d.name, err)
			}
		}
		return nil
	})
}

// GetStations returns all stations for a location with current active load.
func (s *Service) GetStations(ctx context.Context, orgID, locationID string) ([]KitchenStation, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var stations []KitchenStation

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT ks.station_id, ks.org_id, ks.location_id, ks.name, ks.station_type,
			        ks.max_concurrent, ks.status, ks.created_at,
			        COALESCE(
			            (SELECT COUNT(*)
			             FROM kds_ticket_items kti
			             JOIN kds_tickets kt ON kt.ticket_id = kti.ticket_id
			             WHERE kt.location_id = ks.location_id
			               AND kti.station_type = ks.station_type
			               AND kti.status IN ('fired', 'cooking')
			               AND kti.org_id = ks.org_id
			            ), 0
			        ) AS current_load
			 FROM kitchen_stations ks
			 WHERE ks.org_id = $1 AND ks.location_id = $2
			 ORDER BY ks.name`,
			orgID, locationID,
		)
		if err != nil {
			return fmt.Errorf("query stations: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var st KitchenStation
			if err := rows.Scan(
				&st.StationID, &st.OrgID, &st.LocationID, &st.Name, &st.StationType,
				&st.MaxConcurrent, &st.Status, &st.CreatedAt, &st.CurrentLoad,
			); err != nil {
				return fmt.Errorf("scan station: %w", err)
			}
			stations = append(stations, st)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if stations == nil {
		stations = []KitchenStation{}
	}
	return stations, nil
}

// GetResourceProfiles returns all resource profile task sequences for a menu item.
func (s *Service) GetResourceProfiles(ctx context.Context, orgID, menuItemID string) ([]ResourceProfile, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	var profiles []ResourceProfile

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx,
			`SELECT profile_id, org_id, menu_item_id, station_type, task_sequence,
			        duration_secs, elu_required, batch_size, created_at
			 FROM menu_item_resource_profiles
			 WHERE org_id = $1 AND menu_item_id = $2
			 ORDER BY task_sequence`,
			orgID, menuItemID,
		)
		if err != nil {
			return fmt.Errorf("query resource profiles: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var p ResourceProfile
			if err := rows.Scan(
				&p.ProfileID, &p.OrgID, &p.MenuItemID, &p.StationType, &p.TaskSequence,
				&p.DurationSecs, &p.ELURequired, &p.BatchSize, &p.CreatedAt,
			); err != nil {
				return fmt.Errorf("scan resource profile: %w", err)
			}
			profiles = append(profiles, p)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if profiles == nil {
		profiles = []ResourceProfile{}
	}
	return profiles, nil
}

// SetResourceProfile replaces all resource profiles for a menu item.
func (s *Service) SetResourceProfile(ctx context.Context, orgID, menuItemID string, profiles []ResourceProfile) error {
	tenantCtx := tenant.WithOrgID(ctx, orgID)

	return database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(tenantCtx,
			`DELETE FROM menu_item_resource_profiles WHERE org_id = $1 AND menu_item_id = $2`,
			orgID, menuItemID,
		); err != nil {
			return fmt.Errorf("delete existing profiles: %w", err)
		}

		for _, p := range profiles {
			if _, err := tx.Exec(tenantCtx,
				`INSERT INTO menu_item_resource_profiles
				 (org_id, menu_item_id, station_type, task_sequence, duration_secs, elu_required, batch_size)
				 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				orgID, menuItemID, p.StationType, p.TaskSequence, p.DurationSecs, p.ELURequired, p.BatchSize,
			); err != nil {
				return fmt.Errorf("insert profile seq %d: %w", p.TaskSequence, err)
			}
		}
		return nil
	})
}

// CalculateCapacity computes per-station load and overall capacity for a location.
func (s *Service) CalculateCapacity(ctx context.Context, orgID, locationID string) (*KitchenCapacity, error) {
	stations, err := s.GetStations(ctx, orgID, locationID)
	if err != nil {
		return nil, fmt.Errorf("get stations: %w", err)
	}

	var totalLoad, totalMax int
	stationLoads := make([]StationLoad, 0, len(stations))

	for _, st := range stations {
		if st.Status != "active" {
			continue
		}
		sl := StationLoad{
			StationID:     st.StationID,
			Name:          st.Name,
			StationType:   st.StationType,
			MaxConcurrent: st.MaxConcurrent,
			CurrentLoad:   st.CurrentLoad,
			LoadPct:       loadPct(st.CurrentLoad, st.MaxConcurrent),
		}
		stationLoads = append(stationLoads, sl)
		totalLoad += st.CurrentLoad
		totalMax += st.MaxConcurrent
	}

	overallPct := loadPct(totalLoad, totalMax)

	return &KitchenCapacity{
		LocationID:      locationID,
		OverallLoadPct:  overallPct,
		StationCapacity: stationLoads,
		ComputedAt:      time.Now().UTC(),
	}, nil
}

// EstimateTicketTime estimates total preparation time for the given menu item IDs.
// It sums the maximum duration per station type across all items.
func (s *Service) EstimateTicketTime(ctx context.Context, orgID, locationID string, menuItemIDs []string) (*TicketTimeEstimate, error) {
	defaults := defaultStationTimes()

	// stationMax tracks the longest task per station (critical path per station).
	stationMax := map[string]int{}

	for _, itemID := range menuItemIDs {
		profiles, err := s.GetResourceProfiles(ctx, orgID, itemID)
		if err != nil {
			return nil, fmt.Errorf("get profiles for item %s: %w", itemID, err)
		}

		if len(profiles) == 0 {
			// No profile: use defaults — treat item as needing all default stations?
			// Per spec, fallback to defaults. We apply defaults for known station types.
			for stType, dur := range defaults {
				if cur, ok := stationMax[stType]; !ok || dur > cur {
					stationMax[stType] = dur
				}
			}
			continue
		}

		// Sum durations per station type for this item.
		itemStationSum := map[string]int{}
		for _, p := range profiles {
			itemStationSum[p.StationType] += p.DurationSecs
		}

		for stType, dur := range itemStationSum {
			if cur, ok := stationMax[stType]; !ok || dur > cur {
				stationMax[stType] = dur
			}
		}
	}

	// Build breakdown and total (critical path = max across all station maxes, simplified as sum).
	breakdown := make([]StationTime, 0, len(stationMax))
	totalSecs := 0
	for stType, dur := range stationMax {
		breakdown = append(breakdown, StationTime{StationType: stType, DurationSecs: dur})
		totalSecs += dur
	}

	return &TicketTimeEstimate{
		MenuItemIDs:         menuItemIDs,
		EstimatedSecs:       totalSecs,
		PerStationBreakdown: breakdown,
	}, nil
}
