package utils

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	redisclient "github.com/algo7/day-planner-gpt-data-portal/internal/redis"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// OAuth2Config is a struct to hold the OAuth2 configuration
type OAuth2Config struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
	AuthURL      string   `json:"auth_url"`
	TokenURL     string   `json:"token_url"`
}

// ValidProviders is a slice of valid OAuth2 providers
var ValidProviders = map[string]bool{
	"google":  true,
	"outlook": true,
}

// GenerateStateToken generates a random state token for OAuth2 authorization
func generateStateToken(provider string) (string, error) {
	b := make([]byte, 16) // 16 bytes equals 128 bits
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Calculate the SHA-256 hash sum of the 'b' byte slice
	sha256Hash := sha256.New()
	_, err = io.WriteString(sha256Hash, string(b))
	if err != nil {
		return "", err
	}

	// Convert the hash sum to a hexadecimal string
	hashSum := fmt.Sprintf("%s-%x", provider, sha256Hash.Sum(nil))

	return hashSum, nil
}

// GetOAuth2Config returns the OAuth2 config for the specified provider
func GetOAuth2Config(provider string) (*oauth2.Config, error) {

	// Initialize the OAuth2 config variable
	authConfig := &oauth2.Config{}

	switch provider {

	case "google":

		// Load google credentials from JSON file
		b, err := os.ReadFile("./credentials/google_credentials.json")
		if err != nil {
			return nil, fmt.Errorf("Unable to read client secret file for %s: %v", provider, err)
		}

		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse client secret file to config for %s: %v", provider, err)
		}

		authConfig = config

	case "outlook":

		// Load outlook credentials from JSON file
		config, err := oauth2ConfigFromJSON("./credentials/outlook_credentials.json")
		if err != nil {
			return nil, fmt.Errorf("Unable to read client secret file for %s: %v", provider, err)
		}

		authConfig = config

	default:
		return nil, fmt.Errorf("Invalid provider: %s", provider)
	}

	return authConfig, nil
}

// oauth2ConfigFromJSON reads a JSON file and returns an OAuth2 config
func oauth2ConfigFromJSON(fileName string) (*oauth2.Config, error) {

	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %v", err)
	}

	var cfg OAuth2Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %v", err)
	}

	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
	}, nil
}

// GenerateOauthURL prints the URL to visit to authorize the application
func GenerateOauthURL(config *oauth2.Config, provider string, flowType string) (string, string, error) {

	switch flowType {
	case "PCKE":
		// Generate a random state token
		stateToken, err := generateStateToken(provider)
		if err != nil {
			return "", "", fmt.Errorf("unable to generate state token: %w", err)
		}

		// Save the state token to redis and set the time to live to 2 minutes
		err = redisclient.Rdb.Set(context.Background(), fmt.Sprintf("stateToken_%s", stateToken), stateToken, 2*time.Minute).Err()
		if err != nil {
			return "", "", fmt.Errorf("unable to save state token to redis: %w", err)
		}

		authURL := config.AuthCodeURL(stateToken, oauth2.AccessTypeOffline)
		// fmt.Printf("Go to the following link in your browser then type the "+
		// 	"authorization code: \n%v\n", authURL)

		return authURL, "", nil

	case "Device":
		config.Endpoint.DeviceAuthURL = google.Endpoint.DeviceAuthURL
		resp, err := config.DeviceAuth(context.Background(), oauth2.AccessTypeOffline)
		if err != nil {
			return "", "", fmt.Errorf("unable to retrieve device auth: %w", err)
		}

		if resp == nil {
			return "", "", fmt.Errorf("device auth response is nil")
		}

		return fmt.Sprintf("Please go to %s and enter the following code %s", resp.VerificationURI, resp.UserCode), resp.DeviceCode, nil
	}

	return "", "", fmt.Errorf("invalid flow type: %s", flowType)
}

// PollToken Poll Google's authorization server to retrieve the token
func PollToken(config *oauth2.Config, deviceCode string) (*oauth2.Token, error) {

	// Prepare the form data
	// See: https://developers.google.com/identity/protocols/oauth2/limited-input-device#step-4:-poll-googles-authorization-server
	form := url.Values{}
	form.Add("client_id", config.ClientID)
	form.Add("client_secret", config.ClientSecret)
	form.Add("device_code", deviceCode)
	form.Add("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	// Start polling
	for {

		// Prepare the request
		req, err := http.NewRequest("POST", config.Endpoint.TokenURL, strings.NewReader(form.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		// Send the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		// Read the response body
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unable to read response body when polling: %w", err)
		}

		// If the response contains an access token, return it
		if strings.Contains(string(data), "access_token") {
			var tok oauth2.Token
			err = json.Unmarshal(data, &tok)
			if err != nil {
				return nil, fmt.Errorf("unable to unmarshal token from the polled result: %w", err)
			}

			return &tok, nil
		}

		time.Sleep(5 * time.Second) // Polling interval
	}
}

// ExchangeCodeForToken handles the redirect from the OAuth2 provider and exchanges the code for a token
func ExchangeCodeForToken(config *oauth2.Config, authCode string) (*oauth2.Token, error) {

	// Converts authorization code into a token
	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return tok, fmt.Errorf("Unable to retrieve token from web: %w", err)
	}

	return tok, nil
}

// RetrieveToken retrieves the OAuth token from redis.
func RetrieveToken(provider string) (*oauth2.Token, error) {

	// Retrieves the token from redis
	token, err := redisclient.Rdb.HGetAll(context.Background(), provider).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, err
		}

		return nil, fmt.Errorf("Unable to retrieve token from redis: %w", err)
	}

	// If the token is not found in redis, return an error
	if len(token) == 0 {
		return nil, redis.Nil
	}

	// Marshals the token into a JSON object in order to unmarshal it into an oauth2.Token struct
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("Unable to marshal token: %w", err)
	}

	// Unmarshals the token into an oauth2.Token struct
	tok := &oauth2.Token{}

	err = json.Unmarshal(tokenJSON, tok)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal token: %w", err)
	}

	return tok, nil
}

// SaveToken saves the token to redis.
func SaveToken(provider string, token *oauth2.Token) error {

	// Calculates the time to live for the token
	ttl := token.Expiry.Sub(time.Now().UTC())

	// Marshals the token into a JSON object
	var tokenMap map[string]interface{}

	// Marshals the token into a JSON object in order to unmarshal it into an oauth2.Token struct
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("Unable to marshal token: %w", err)
	}

	// Unmarshals the token into a map
	err = json.Unmarshal(tokenJSON, &tokenMap)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal token: %w", err)
	}

	// Saves the token to redis
	err = redisclient.Rdb.HSet(context.Background(), provider, tokenMap).Err()
	if err != nil {
		return fmt.Errorf("Unable to save token to redis: %w", err)
	}

	// Sets the time to live for the token
	err = redisclient.Rdb.Expire(context.Background(), provider, ttl).Err()

	return nil
}

// UpdateToken updates the token in redis.
func UpdateToken(provider string, token *oauth2.Token) error {

	// Calculates the time to live for the token
	ttl := token.Expiry.Sub(time.Now().UTC())

	// Marshals the token into a JSON object
	var tokenMap map[string]interface{}

	// Marshals the token into a JSON object in order to unmarshal it into an oauth2.Token struct
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("Unable to marshal token: %w", err)
	}

	// Unmarshals the token into a map
	err = json.Unmarshal(tokenJSON, &tokenMap)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal token: %w", err)
	}

	// Saves the token to redis
	err = redisclient.Rdb.HSet(context.Background(), provider, tokenMap).Err()
	if err != nil {
		return fmt.Errorf("Unable to save token to redis: %w", err)
	}

	// Sets the time to live for the token
	err = redisclient.Rdb.Expire(context.Background(), provider, ttl).Err()

	return nil
}

// GetTokenFromRefreshToken retrieves a token from a refresh token
func GetTokenFromRefreshToken(config *oauth2.Config, refreshToken string) (*oauth2.Token, error) {

	// Get the new token from the refresh token
	tok, err := config.TokenSource(context.Background(), &oauth2.Token{
		RefreshToken: refreshToken,
	}).Token()

	if err != nil {
		return nil, fmt.Errorf("Unable to get token from refresh token: %w", err)
	}
	return tok, nil
}
