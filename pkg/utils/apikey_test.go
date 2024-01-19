package utils

import "testing"

func TestGenerateAPIKey(t *testing.T) {
	apiKey, err := GenerateAPIKey()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedLength := 64 // 32 bytes = 64 hexadecimal characters
	if len(apiKey) != expectedLength {
		t.Errorf("Expected length: %d, got: %d", expectedLength, len(apiKey))
	}
}
