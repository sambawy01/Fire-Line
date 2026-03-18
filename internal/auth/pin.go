package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Memory      = 64 * 1024 // 64MB
	argon2Iterations  = 3
	argon2Parallelism = 1
	argon2SaltLength  = 16
	argon2KeyLength   = 32
)

func HashPIN(pin string) (string, error) {
	salt := make([]byte, argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(pin), salt, argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2Memory,
		argon2Iterations,
		argon2Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func VerifyPIN(encoded, pin string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, fmt.Errorf("invalid hash format")
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, fmt.Errorf("parse params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}

	computedHash := argon2.IDKey([]byte(pin), salt, iterations, memory, parallelism, uint32(len(expectedHash)))

	return subtle.ConstantTimeCompare(computedHash, expectedHash) == 1, nil
}

func ValidatePINPolicy(pin string) error {
	if len(pin) < 4 {
		return fmt.Errorf("PIN must be at least 4 digits")
	}
	if len(pin) > 8 {
		return fmt.Errorf("PIN must be at most 8 digits")
	}
	for _, r := range pin {
		if r < '0' || r > '9' {
			return fmt.Errorf("PIN must contain only digits")
		}
	}
	if isSequential(pin) {
		return fmt.Errorf("sequential PINs are not allowed")
	}
	if isRepeated(pin) {
		return fmt.Errorf("repeated-digit PINs are not allowed")
	}
	return nil
}

func isSequential(pin string) bool {
	ascending, descending := true, true
	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[i-1]+1 {
			ascending = false
		}
		if pin[i] != pin[i-1]-1 {
			descending = false
		}
	}
	return ascending || descending
}

func isRepeated(pin string) bool {
	for i := 1; i < len(pin); i++ {
		if pin[i] != pin[0] {
			return false
		}
	}
	return true
}
