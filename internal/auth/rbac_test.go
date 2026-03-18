package auth_test

import (
	"testing"

	"github.com/opsnerve/fireline/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestRole_HasPermission(t *testing.T) {
	tests := []struct {
		role       string
		permission string
		expected   bool
	}{
		{"owner", "financial:read", true},
		{"owner", "system:admin", true},
		{"gm", "financial:read", true},
		{"gm", "system:admin", false},
		{"shift_manager", "staff:read", true},
		{"shift_manager", "financial:read", false},
		{"staff", "schedule:read_own", true},
		{"staff", "financial:read", false},
		{"read_only", "reporting:read", true},
		{"read_only", "inventory:write", false},
	}

	for _, tt := range tests {
		t.Run(tt.role+"_"+tt.permission, func(t *testing.T) {
			result := auth.HasPermission(tt.role, tt.permission)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRole_Hierarchy(t *testing.T) {
	// Owner has all permissions that GM has
	gmPerms := auth.PermissionsForRole("gm")
	for _, p := range gmPerms {
		assert.True(t, auth.HasPermission("owner", p), "owner should have gm permission: %s", p)
	}

	// GM has all permissions that shift_manager has
	smPerms := auth.PermissionsForRole("shift_manager")
	for _, p := range smPerms {
		assert.True(t, auth.HasPermission("gm", p), "gm should have shift_manager permission: %s", p)
	}
}
