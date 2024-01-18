package utils

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	redisclient "github.com/algo7/day-planner-gpt-data-portal/internal/redis"
)

// CreateHMACSigniture creates HMAC signiture for the given string
func CreateHMACSigniture(secret string, data string) string {

	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(secret))

	// Write Data to it
	h.Write([]byte(data))

	// Get result and encode as hexadecimal string
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateAPIKey generates an 32 byte API key in Hex format
func GenerateAPIKey() (string, error) {

	// Allocate a slice of 32 bytes.
	bytes := make([]byte, 32)

	// Fill the byte slice with cryptographically secure random bytes
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Hex encode the bytes
	apiKey := hex.EncodeToString(bytes)

	// Set the API key in Redis with a TTL of 7 days.
	ttl := 7 * 24 * time.Hour // 7 days in hours

	// Save the key in the database
	err = redisclient.Rdb.Set(context.Background(), apiKey, apiKey, ttl).Err()
	if err != nil {
		return "", fmt.Errorf("Error saving API key to database: %v", err)
	}

	// Return the random bytes as a hexadecimal string
	return hex.EncodeToString(bytes), nil
}
