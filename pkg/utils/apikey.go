package utils

import (
	"crypto/rand"
	"encoding/hex"
)

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

	// Return the random bytes as a hexadecimal string
	return apiKey, nil
}

// Unused
// // CreateHMACSigniture creates HMAC signiture for the given string
// func CreateHMACSigniture(secret string, data string) string {

// 	// Create a new HMAC by defining the hash type and the key (as byte array)
// 	h := hmac.New(sha256.New, []byte(secret))

// 	// Write Data to it
// 	h.Write([]byte(data))

// 	// Get result and encode as hexadecimal string
// 	return hex.EncodeToString(h.Sum(nil))
// }
