package auth

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestHashAPIKey_DeterministicAndHex(t *testing.T) {
	rawKey := "some-super-secret-api-key-value"

	firstHash := HashAPIKey(rawKey)
	secondHash := HashAPIKey(rawKey)

	if firstHash != secondHash {
		t.Errorf("expected HashAPIKey to be deterministic, got %q and %q", firstHash, secondHash)
	}

	if len(firstHash) != 64 { // SHA-256 => 32 bytes => 64 hex characters
		t.Errorf("expected hash length of 64 hex chars, got %d", len(firstHash))
	}

	if _, err := hex.DecodeString(firstHash); err != nil {
		t.Errorf("expected hex-encoded hash, decode failed: %v", err)
	}
}

func TestHashAPIKey_DifferentInputsGiveDifferentHashes(t *testing.T) {
	if HashAPIKey("alpha") == HashAPIKey("beta") {
		t.Error("expected distinct inputs to produce distinct hashes")
	}
}

func TestGenerateRawAPIKey_ReturnsUniqueKeys(t *testing.T) {
	firstKey, err := GenerateRawAPIKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	secondKey, err := GenerateRawAPIKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if firstKey == secondKey {
		t.Error("expected two generated keys to differ")
	}

	// base64url(32 bytes) is 44 chars including padding
	if len(firstKey) < 40 {
		t.Errorf("generated key looks too short: %d chars", len(firstKey))
	}

	// URLEncoding uses '-' and '_' instead of '+' and '/'.
	if strings.ContainsAny(firstKey, "+/") {
		t.Errorf("expected base64url (no '+' or '/'), got %q", firstKey)
	}
}

func TestValidateAPIKeyFormat(t *testing.T) {
	tests := []struct {
		name    string
		rawKey  string
		isValid bool
	}{
		{"empty", "", false},
		{"too short", "short-key", false},
		{"boundary too short", strings.Repeat("a", 19), false},
		{"minimum valid length", strings.Repeat("a", 20), true},
		{"typical base64url 32-byte key", strings.Repeat("a", 43), true},
		{"maximum valid length", strings.Repeat("a", 100), true},
		{"too long", strings.Repeat("a", 101), false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if got := ValidateAPIKeyFormat(testCase.rawKey); got != testCase.isValid {
				t.Errorf("ValidateAPIKeyFormat(len=%d) = %v, want %v",
					len(testCase.rawKey), got, testCase.isValid)
			}
		})
	}
}
