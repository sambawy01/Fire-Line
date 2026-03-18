package tenant

import (
	"context"
	"errors"
)

type contextKey string

const tenantKey contextKey = "tenant_org_id"

var ErrNoTenant = errors.New("no tenant in context")

func WithOrgID(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, tenantKey, orgID)
}

func OrgIDFrom(ctx context.Context) (string, error) {
	orgID, ok := ctx.Value(tenantKey).(string)
	if !ok || orgID == "" {
		return "", ErrNoTenant
	}
	return orgID, nil
}
