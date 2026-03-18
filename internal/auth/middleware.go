package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/opsnerve/fireline/internal/tenant"
)

type contextKeyType string

const (
	userIDKey contextKeyType = "auth_user_id"
	roleKey   contextKeyType = "auth_role"
)

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFrom(ctx context.Context) string {
	v, _ := ctx.Value(userIDKey).(string)
	return v
}

func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey, role)
}

func RoleFrom(ctx context.Context) string {
	v, _ := ctx.Value(roleKey).(string)
	return v
}

func AuthMiddleware(issuer *TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "AUTH_TOKEN_MISSING", "missing or invalid Authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := issuer.ValidateAccessToken(tokenStr)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "invalid or expired token")
				return
			}

			// Inject tenant context (for TenantTx)
			ctx := r.Context()
			ctx = tenant.WithOrgID(ctx, claims.OrgID)
			ctx = WithUserID(ctx, claims.UserID)
			ctx = WithRole(ctx, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := RoleFrom(r.Context())
			if !HasPermission(role, permission) {
				writeError(w, http.StatusForbidden, "AUTH_PERMISSION_DENIED", "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
