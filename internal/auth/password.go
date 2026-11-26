package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	saltBytes = 16
	keyBytes  = 32
	timeCost  = 2
	memoryKiB = 64 * 1024
	threads   = 2
)

// NewSalt returns a per-user random salt, hex encoded.
func NewSalt() (string, error) {
	buf := make([]byte, saltBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// HashPassword hashes password with argon2id and the user's salt.
func HashPassword(password, saltHex string) (string, error) {
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return "", fmt.Errorf("decode password salt: %w", err)
	}
	sum := argon2.IDKey([]byte(password), salt, timeCost, memoryKiB, threads, keyBytes)
	return hex.EncodeToString(sum), nil
}

// CheckPassword compares password to a stored argon2id hash using the user's salt.
func CheckPassword(password, saltHex, hashHex string) bool {
	sum, err := HashPassword(password, saltHex)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(sum), []byte(hashHex)) == 1
}
