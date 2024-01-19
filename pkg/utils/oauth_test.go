package utils

import (
	"os"
	"testing"
)

func TestOAuth2ConfigFromJSON(t *testing.T) {
	// Create a temporary JSON file with OAuth2 config
	tempFile, err := os.CreateTemp("./", "prefix-*.json")
	if err != nil {
		t.Fatalf("Cannot create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Ensure file clean-up

	// Define the JSON data
	data := `{
        "client_id": "testClientID",
        "client_secret": "testClientSecret",
        "redirect_url": "http://localhost:8080",
        "scopes": ["openid", "profile", "email"],
        "auth_url": "http://localhost:8080/auth",
        "token_url": "http://localhost:8080/token"
    }`

	// Write and close the file
	if _, err := tempFile.Write([]byte(data)); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tempFile.Close()

	// Test the function
	config, err := OAuth2ConfigFromJSON(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to get OAuth2 config from JSON: %v", err)
	}

	// Assertions to verify the config
	assertEqual(t, "ClientID", "testClientID", config.ClientID)
	assertEqual(t, "ClientSecret", "testClientSecret", config.ClientSecret)
	assertEqual(t, "RedirectURL", "http://localhost:8080", config.RedirectURL)
	assertEqual(t, "AuthURL", "http://localhost:8080/auth", config.Endpoint.AuthURL)
	assertEqual(t, "TokenURL", "http://localhost:8080/token", config.Endpoint.TokenURL)
}

// Helper function for asserting equality
func assertEqual(t *testing.T, name, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected %s to be '%s', got '%s'", name, expected, actual)
	}
}
