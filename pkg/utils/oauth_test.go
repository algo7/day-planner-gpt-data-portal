package utils

import (
	"net/url"
	"os"
	"strings"
	"testing"

	"golang.org/x/oauth2"
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

func TestGenerateOauthURL(t *testing.T) {
	// Mock OAuth2 Config
	config := &oauth2.Config{
		ClientID:     "testClientID",
		ClientSecret: "testClientSecret",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://localhost:8080/auth",
			TokenURL: "http://localhost:8080/token",
		},
	}

	// Call the function
	resultURL := GenerateOauthURL(config)

	// Parse the URL
	parsedURL, err := url.Parse(resultURL)

	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}

	// Check if the URL starts with the AuthURL
	if !strings.HasPrefix(resultURL, config.Endpoint.AuthURL) {
		t.Errorf("URL does not start with AuthURL, got %s", resultURL)
	}

	// Check if the URL contains the client ID
	if parsedURL.Query().Get("client_id") != config.ClientID {
		t.Errorf("URL does not contain correct client_id, expected %s, got %s", config.ClientID, parsedURL.Query().Get("client_id"))
	}

	// Check if the URL contains the redirect URL
	if parsedURL.Query().Get("redirect_uri") != config.RedirectURL {
		t.Errorf("URL does not contain correct redirect_uri, expected %s, got %s", config.RedirectURL, parsedURL.Query().Get("redirect_uri"))
	}
}
