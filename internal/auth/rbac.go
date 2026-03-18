package auth

var rolePermissions = map[string][]string{
	"read_only": {
		"reporting:read",
	},
	"staff": {
		"schedule:read_own",
		"tasks:read",
		"tasks:write",
		"menu:read",
		"reporting:read",
	},
	"shift_manager": {
		"schedule:read_own",
		"tasks:read",
		"tasks:write",
		"menu:read",
		"reporting:read",
		"staff:read",
		"schedule:read",
		"schedule:write",
		"inventory:read",
	},
	"gm": {
		"schedule:read_own",
		"tasks:read",
		"tasks:write",
		"menu:read",
		"reporting:read",
		"staff:read",
		"schedule:read",
		"schedule:write",
		"inventory:read",
		"staff:write",
		"inventory:write",
		"menu:write",
		"financial:read",
		"vendor:read",
		"vendor:write",
		"customer:read",
	},
	"owner": {
		"schedule:read_own",
		"tasks:read",
		"tasks:write",
		"menu:read",
		"reporting:read",
		"staff:read",
		"schedule:read",
		"schedule:write",
		"inventory:read",
		"staff:write",
		"inventory:write",
		"menu:write",
		"financial:read",
		"vendor:read",
		"vendor:write",
		"customer:read",
		"financial:write",
		"system:admin",
		"audit:read",
		"integrations:manage",
		"billing:manage",
	},
}

func HasPermission(role, permission string) bool {
	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

func PermissionsForRole(role string) []string {
	perms, ok := rolePermissions[role]
	if !ok {
		return nil
	}
	result := make([]string, len(perms))
	copy(result, perms)
	return result
}

func ValidRole(role string) bool {
	_, ok := rolePermissions[role]
	return ok
}
