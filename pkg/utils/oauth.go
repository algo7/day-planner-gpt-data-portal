package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	redisclient "github.com/algo7/day-planner-gpt-data-portal/internal/redis"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
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

// OAuth2ConfigFromJSON reads a JSON file and returns an OAuth2 config
func OAuth2ConfigFromJSON(fileName string) (*oauth2.Config, error) {

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

// GetTokenFromWeb prints the URL to visit to authorize the application
func GetTokenFromWeb(config *oauth2.Config) string {

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	// fmt.Printf("Go to the following link in your browser then type the "+
	// 	"authorization code: \n%v\n", authURL)

	return authURL
}

// GetClient Retrieve a token, saves the token, then returns the generated client.
func GetClient(config *oauth2.Config, tokenFileName string) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := TokenFromFile(tokenFileName)
	if err != nil {
		return nil, fmt.Errorf("unable to get token from file: %w", err)
	}
	return config.Client(context.Background(), tok), nil
}

// ExchangeCodeForToken handles the redirect from the OAuth2 provider and exchanges the code for a token
func ExchangeCodeForToken(config *oauth2.Config, authCode string, redisKey string) (*oauth2.Token, error) {

	// Converts authorization code into a token
	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token from web: %w", err)
	}

	// Marshals the token into a JSON object
	tokenJSON, err := json.Marshal(tok)
	if err != nil {
		return nil, fmt.Errorf("Unable to marshal token: %w", err)
	}

	// Calculates the time to live for the token
	ttl := tok.Expiry.Sub(time.Now())

	// Saves the token to redis
	err = redisclient.Rdb.Set(context.Background(), redisKey, tokenJSON, ttl).Err()
	if err != nil {
		return nil, fmt.Errorf("Unable to save token to redis: %w", err)
	}

	return tok, nil
}

// RetrieveToken retrieves the OAuth token from redis.
func RetrieveToken(redisKey string) (*oauth2.Token, error) {

	// Retrieves the token from redis
	tokenJSON, err := redisclient.Rdb.Get(context.Background(), redisKey).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("Token does not exist in redis: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token from redis: %w", err)
	}

	// Unmarshals the token
	var tok oauth2.Token
	err = json.Unmarshal([]byte(tokenJSON), &tok)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal token: %w", err)
	}

	return &tok, nil
}

// TokenFromFile retrieves a Token from a given file path.
func TokenFromFile(fileName string) (*oauth2.Token, error) {

	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read token file: %w", err)
	}

	var tok oauth2.Token
	err = json.Unmarshal(data, &tok)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal token: %w", err)
	}

	return &tok, nil
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {

	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Unable to cache oauth token: %w", err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return fmt.Errorf("Unable to encode token: %w", err)
	}

	return nil
}
