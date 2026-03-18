package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/opsnerve/fireline/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key, &key.PublicKey
}

func TestGenerateAccessToken_Valid(t *testing.T) {
	privKey, pubKey := testKeyPair(t)
	issuer := auth.NewTokenIssuer(privKey, pubKey, 15*time.Minute)

	claims := auth.UserClaims{
		UserID: "user-123",
		OrgID:  "org-456",
		Role:   "gm",
		Email:  "gm@restaurant.com",
	}

	tokenStr, err := issuer.GenerateAccessToken(claims)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	parsed, err := issuer.ValidateAccessToken(tokenStr)
	require.NoError(t, err)
	assert.Equal(t, "user-123", parsed.UserID)
	assert.Equal(t, "org-456", parsed.OrgID)
	assert.Equal(t, "gm", parsed.Role)
	assert.Equal(t, "gm@restaurant.com", parsed.Email)
}

func TestValidateAccessToken_Expired(t *testing.T) {
	privKey, pubKey := testKeyPair(t)
	issuer := auth.NewTokenIssuer(privKey, pubKey, -1*time.Minute) // already expired

	claims := auth.UserClaims{
		UserID: "user-123",
		OrgID:  "org-456",
		Role:   "staff",
		Email:  "staff@test.com",
	}

	tokenStr, err := issuer.GenerateAccessToken(claims)
	require.NoError(t, err)

	_, err = issuer.ValidateAccessToken(tokenStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestValidateAccessToken_WrongKey(t *testing.T) {
	privKey1, _ := testKeyPair(t)
	_, pubKey2 := testKeyPair(t) // different key pair
	issuer1 := auth.NewTokenIssuer(privKey1, pubKey2, 15*time.Minute)

	claims := auth.UserClaims{
		UserID: "user-123",
		OrgID:  "org-456",
		Role:   "staff",
		Email:  "staff@test.com",
	}

	tokenStr, err := issuer1.GenerateAccessToken(claims)
	require.NoError(t, err)

	_, err = issuer1.ValidateAccessToken(tokenStr)
	assert.Error(t, err)
}
