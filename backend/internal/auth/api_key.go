package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// HashAPIKey generates a SHA-256 hash of the raw API key and returns it as hex encoded string
func HashAPIKey(rawKey string) string {
	hash := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(hash[:])
}

// GenerateRawAPIKey generates a cryptographically random 32-byte API key
// Returns the key as a base64url encoded string
func GenerateRawAPIKey() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	encodedKey := base64.URLEncoding.EncodeToString(randomBytes)
	return encodedKey, nil
}

// ValidateAPIKeyFormat checks if the raw API key has a valid format
// Returns true if the key is non-empty and has reasonable length
func ValidateAPIKeyFormat(rawKey string) bool {
	// Check non-empty
	if rawKey == "" {
		return false
	}

	// Check reasonable length (base64url encoded 32 bytes is typically 43 characters)
	// Allow some flexibility for different encodings/formats
	if len(rawKey) < 20 || len(rawKey) > 100 {
		return false
	}

	return true
}
