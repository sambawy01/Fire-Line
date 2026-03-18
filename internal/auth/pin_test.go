package auth_test

import (
	"testing"

	"github.com/opsnerve/fireline/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPIN_ReturnsHash(t *testing.T) {
	hash, err := auth.HashPIN("1234")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Contains(t, hash, "$argon2id$")
}

func TestVerifyPIN_Correct(t *testing.T) {
	hash, err := auth.HashPIN("5678")
	require.NoError(t, err)
	ok, err := auth.VerifyPIN(hash, "5678")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestVerifyPIN_Wrong(t *testing.T) {
	hash, err := auth.HashPIN("5678")
	require.NoError(t, err)
	ok, err := auth.VerifyPIN(hash, "9999")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestValidatePINPolicy_Valid(t *testing.T) {
	assert.NoError(t, auth.ValidatePINPolicy("4829"))
}

func TestValidatePINPolicy_TooShort(t *testing.T) {
	err := auth.ValidatePINPolicy("123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "4")
}

func TestValidatePINPolicy_Sequential(t *testing.T) {
	assert.Error(t, auth.ValidatePINPolicy("1234"))
	assert.Error(t, auth.ValidatePINPolicy("4321"))
}

func TestValidatePINPolicy_Repeated(t *testing.T) {
	assert.Error(t, auth.ValidatePINPolicy("1111"))
	assert.Error(t, auth.ValidatePINPolicy("0000"))
}
