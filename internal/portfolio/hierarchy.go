package portfolio

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opsnerve/fireline/internal/tenant"
	"github.com/opsnerve/fireline/pkg/database"
)

// PortfolioNode represents a node in the portfolio hierarchy.
type PortfolioNode struct {
	NodeID       string    `json:"node_id"`
	OrgID        string    `json:"org_id"`
	ParentNodeID *string   `json:"parent_node_id"`
	Name         string    `json:"name"`
	NodeType     string    `json:"node_type"`
	LocationID   *string   `json:"location_id"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateNode inserts a new portfolio hierarchy node.
func (s *Service) CreateNode(ctx context.Context, orgID string, parentID *string, name, nodeType string, locationID *string) (*PortfolioNode, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if nodeType == "" {
		return nil, errors.New("node_type is required")
	}

	// Normalize empty-string pointers to nil so PostgreSQL receives NULL
	// rather than an empty string, which is invalid for a UUID column.
	if parentID != nil && *parentID == "" {
		parentID = nil
	}
	if locationID != nil && *locationID == "" {
		locationID = nil
	}

	var node PortfolioNode
	err := s.pool.QueryRow(ctx, `
		INSERT INTO portfolio_nodes (org_id, parent_node_id, name, node_type, location_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING node_id, org_id, parent_node_id, name, node_type, location_id, sort_order, created_at, updated_at
	`, orgID, parentID, name, nodeType, locationID).Scan(
		&node.NodeID, &node.OrgID, &node.ParentNodeID, &node.Name, &node.NodeType,
		&node.LocationID, &node.SortOrder, &node.CreatedAt, &node.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// GetHierarchy returns the full flat list of portfolio nodes for an org.
// Clients reconstruct the tree using parent_node_id.
func (s *Service) GetHierarchy(ctx context.Context, orgID string) ([]PortfolioNode, error) {
	tenantCtx := tenant.WithOrgID(ctx, orgID)
	var nodes []PortfolioNode

	err := database.TenantTx(tenantCtx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(tenantCtx, `
			SELECT node_id, org_id, parent_node_id, name, node_type, location_id, sort_order, created_at, updated_at
			FROM portfolio_nodes
			ORDER BY sort_order, name
		`)
		if err != nil {
			return fmt.Errorf("query nodes: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var n PortfolioNode
			if err := rows.Scan(
				&n.NodeID, &n.OrgID, &n.ParentNodeID, &n.Name, &n.NodeType,
				&n.LocationID, &n.SortOrder, &n.CreatedAt, &n.UpdatedAt,
			); err != nil {
				return err
			}
			nodes = append(nodes, n)
		}
		return rows.Err()
	})
	if nodes == nil {
		nodes = []PortfolioNode{}
	}
	return nodes, err
}

// UpdateNode renames a portfolio node.
func (s *Service) UpdateNode(ctx context.Context, orgID, nodeID, name string) error {
	if name == "" {
		return errors.New("name is required")
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE portfolio_nodes
		SET name = $1, updated_at = now()
		WHERE node_id = $2 AND org_id = $3
	`, name, nodeID, orgID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DeleteNode removes a portfolio node and cascades to children via FK.
func (s *Service) DeleteNode(ctx context.Context, orgID, nodeID string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM portfolio_nodes
		WHERE node_id = $1 AND org_id = $2
	`, nodeID, orgID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
