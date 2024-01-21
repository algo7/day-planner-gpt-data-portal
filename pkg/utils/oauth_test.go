package utils

import (
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	redisclient "github.com/algo7/day-planner-gpt-data-portal/internal/redis"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
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
	db, mock := redismock.NewClientMock()
	assert := assert.New(t)

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

	// Mock Redis and Setoperation before calling the function
	redisclient.Rdb = db
	keyPattern := `^stateToken_[a-fA-F0-9]{64}$`
	valuePattern := `^[a-fA-F0-9]{64}$`
	mock.Regexp().ExpectSet(keyPattern, valuePattern, 2*time.Minute).SetVal("OK") // Simulate successful SET operation

	// Call the function
	resultURL, _, err := GenerateOauthURL(config, "google", "PCKE")
	assert.NoError(err, "GenerateOauthURL returned an error")

	// Close the mock database connection
	defer db.Close()

	// Parse the URL
	parsedURL, err := url.Parse(resultURL)
	assert.NoError(err, "Failed to parse URL")

	// Check if the URL starts with the AuthURL
	assert.True(strings.HasPrefix(resultURL, config.Endpoint.AuthURL), "URL does not start with AuthURL")

	// Check if the URL contains the client ID
	assert.Equal(config.ClientID, parsedURL.Query().Get("client_id"), "URL does not contain correct client_id")

	// Check if the URL contains the redirect URL
	assert.Equal(config.RedirectURL, parsedURL.Query().Get("redirect_uri"), "URL does not contain correct redirect_uri")

	// Check if the URL contains the state token
	assert.Regexp(`^[a-fA-F0-9]{64}$`, parsedURL.Query().Get("state"), "URL does not contain correct state token")

	// Check if Redis expectations were met
	assert.NoError(mock.ExpectationsWereMet(), "there were unfulfilled expectations")
}
