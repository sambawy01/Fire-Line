package auth_test

import (
	"testing"

	"github.com/opsnerve/fireline/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword_ReturnsHash(t *testing.T) {
	hash, err := auth.HashPassword("SecureP@ss123!")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "SecureP@ss123!", hash)
}

func TestVerifyPassword_CorrectPassword(t *testing.T) {
	hash, err := auth.HashPassword("SecureP@ss123!")
	require.NoError(t, err)
	assert.True(t, auth.VerifyPassword(hash, "SecureP@ss123!"))
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	hash, err := auth.HashPassword("SecureP@ss123!")
	require.NoError(t, err)
	assert.False(t, auth.VerifyPassword(hash, "WrongPassword1!"))
}

func TestValidatePasswordPolicy_Valid(t *testing.T) {
	err := auth.ValidatePasswordPolicy("SecureP@ss123!")
	assert.NoError(t, err)
}

func TestValidatePasswordPolicy_TooShort(t *testing.T) {
	err := auth.ValidatePasswordPolicy("Short1!")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "12 characters")
}

func TestValidatePasswordPolicy_MissingCategories(t *testing.T) {
	err := auth.ValidatePasswordPolicy("alllowercaselong")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "3 of 4")
}
