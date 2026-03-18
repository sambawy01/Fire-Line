package auth

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID string `json:"user_id"`
	OrgID  string `json:"org_id"`
	Role   string `json:"role"`
	Email  string `json:"email"`
}

type accessTokenClaims struct {
	jwt.RegisteredClaims
	UserClaims
}

type TokenIssuer struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	ttl        time.Duration
}

func NewTokenIssuer(privKey *rsa.PrivateKey, pubKey *rsa.PublicKey, ttl time.Duration) *TokenIssuer {
	return &TokenIssuer{
		privateKey: privKey,
		publicKey:  pubKey,
		ttl:        ttl,
	}
}

func (ti *TokenIssuer) GenerateAccessToken(claims UserClaims) (string, error) {
	now := time.Now()
	tokenClaims := accessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "fireline",
			Subject:   claims.UserID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ti.ttl)),
		},
		UserClaims: claims,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, tokenClaims)
	signed, err := token.SignedString(ti.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (ti *TokenIssuer) ValidateAccessToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &accessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return ti.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}

	claims, ok := token.Claims.(*accessTokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &claims.UserClaims, nil
}
