package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/pquerna/otp/totp"
)

type MFAEnrollment struct {
	Secret        string
	QRCodeURL     string
	RecoveryCodes []string
}

func EnrollMFA(email, issuer string) (*MFAEnrollment, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: email,
	})
	if err != nil {
		return nil, fmt.Errorf("generate TOTP key: %w", err)
	}

	codes, err := generateRecoveryCodes(10)
	if err != nil {
		return nil, fmt.Errorf("generate recovery codes: %w", err)
	}

	return &MFAEnrollment{
		Secret:        key.Secret(),
		QRCodeURL:     key.URL(),
		RecoveryCodes: codes,
	}, nil
}

func VerifyTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

// VerifyTOTPWithSkew allows a +/- 1 period window for clock drift.
func VerifyTOTPWithSkew(secret, code string) bool {
	valid, _ := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Skew: 1, // allow 1 period before and after
	})
	return valid
}

func generateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := range codes {
		b := make([]byte, 5) // 5 bytes = 10 hex chars
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		codes[i] = hex.EncodeToString(b)
	}
	return codes, nil
}
