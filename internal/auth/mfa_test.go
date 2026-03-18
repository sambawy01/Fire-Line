package auth_test

import (
	"testing"
	"time"

	"github.com/opsnerve/fireline/internal/auth"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMFA_EnrollAndVerify(t *testing.T) {
	enrollment, err := auth.EnrollMFA("test@restaurant.com", "FireLine")
	require.NoError(t, err)
	assert.NotEmpty(t, enrollment.Secret)
	assert.NotEmpty(t, enrollment.QRCodeURL)
	assert.Len(t, enrollment.RecoveryCodes, 10)

	// Generate a valid TOTP code from the secret
	code, err := totp.GenerateCode(enrollment.Secret, time.Now())
	require.NoError(t, err)

	// Verify should succeed
	assert.True(t, auth.VerifyTOTP(enrollment.Secret, code))
}

func TestMFA_WrongCode(t *testing.T) {
	enrollment, err := auth.EnrollMFA("test@restaurant.com", "FireLine")
	require.NoError(t, err)

	assert.False(t, auth.VerifyTOTP(enrollment.Secret, "000000"))
}

func TestMFA_RecoveryCodes_Unique(t *testing.T) {
	enrollment, err := auth.EnrollMFA("test@restaurant.com", "FireLine")
	require.NoError(t, err)

	seen := make(map[string]bool)
	for _, code := range enrollment.RecoveryCodes {
		assert.False(t, seen[code], "duplicate recovery code")
		seen[code] = true
		assert.Len(t, code, 10) // 5 bytes hex = 10 chars
	}
}
