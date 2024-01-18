package controllers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	redisclient "github.com/algo7/day-planner-gpt-data-portal/internal/redis"
	"github.com/algo7/day-planner-gpt-data-portal/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// GetAuthOutlook returns the auth page for Outlook
func GetAuthOutlook(c *fiber.Ctx) error {

	// Load the OAuth2 config from the JSON file
	config, err := utils.OAuth2ConfigFromJSON("outlook_credentials.json")
	if err != nil {
		return c.SendString(fmt.Sprintf("Error loading Oauth2 config: %v", err))
	}

	// Get the URL to visit to authorize the application
	authURL := utils.GetTokenFromWeb(config)

	// Redirect the user to the authURL
	return c.Redirect(authURL, 302)
}

// GetOauthRedirectOutlook handles the redirect from the OAuth2 provider
func GetOauthRedirectOutlook(c *fiber.Ctx) error {

	code := c.Query("code")
	if code == "" {
		return c.SendString("No authorization code found in the request")
	}

	// Load the OAuth2 config from the JSON file
	config, err := utils.OAuth2ConfigFromJSON("outlook_credentials.json")
	if err != nil {
		return c.SendString(fmt.Sprintf("Error loading Oauth2 config: %v", err))
	}

	// Exchange the code for an access token here
	utils.ExchangeCodeForToken(config, code, "outlook")

	// return c.SendString(fmt.Sprintf("Authorization code: %s", code))
	return c.RedirectToRoute("outlook", nil, 302)
}

// GetAuthGoogle returns the auth page for Google
func GetAuthGoogle(c *fiber.Ctx) error {

	b, err := os.ReadFile("google_credentials.json")
	if err != nil {
		c.SendString(fmt.Sprintf("Unable to read client secret file: %v", err))
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		c.SendString(fmt.Sprintf("Unable to parse client secret file to config: %v", err))
	}

	// Get the URL to visit to authorize the application
	authURL := utils.GetTokenFromWeb(config)

	// Redirect the user to the authURL
	return c.Redirect(authURL, 302)
}

// GetOauthRedirectGoogle handles the redirect from the OAuth2 provider
func GetOauthRedirectGoogle(c *fiber.Ctx) error {

	code := c.Query("code")
	if code == "" {
		return c.SendString("No authorization code found in the request")
	}

	b, err := os.ReadFile("google_credentials.json")
	if err != nil {
		c.SendString(fmt.Sprintf("Unable to read client secret file: %v", err))
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		c.SendString(fmt.Sprintf("Unable to parse client secret file to config: %v", err))
	}

	// Exchange the code for an access token here
	utils.ExchangeCodeForToken(config, code, "google")

	// return c.SendString(fmt.Sprintf("Authorization code: %s", code))
	return c.RedirectToRoute("google", nil, 302)
}

// GetAPIKey returns a page to get the initial API key
func GetAPIKey(c *fiber.Ctx) error {
	err := redisclient.Rdb.Get(context.Background(), "initial_password").Err()
	if err != nil {
		// If the initial password has expired, redirect to the home page.
		if err == redis.Nil {
			return c.RedirectToRoute("home", nil, 302)
		}
		// If there is an error getting the initial password, log the error and return a 500 status code.
		log.Printf("Error getting initial password: %v", err)
		return c.SendStatus(500)
	}
	return c.Render("apikey_form", fiber.Map{})
}

// PostAPIKey returns a new API key
func PostAPIKey(c *fiber.Ctx) error {

	// Check if the initial password is still in Redis
	err := redisclient.Rdb.Get(context.Background(), "initial_password").Err()
	if err != nil {
		// If the initial password has expired, redirect to the home page.
		if err == redis.Nil {
			return c.RedirectToRoute("home", nil, 302)
		}
		// If there is an error getting the initial password, log the error and return a 500 status code.
		log.Printf("Error getting initial password: %v", err)
		return c.SendStatus(500)
	}

	// Get the password from the form
	password := c.FormValue("password")

	// Get the initial API key from Redis
	initialAPIKey, err := redisclient.Rdb.Get(context.Background(), "initial_password").Result()
	if err != nil {
		if err == redis.Nil {
			return c.SendString("Initial password has expired. Please restart the server to generate a new password.")
		}
		return c.SendString(fmt.Sprintf("Error getting initial password: %v", err))
	}

	// Compare the password with the initial API key
	if password != initialAPIKey {
		return c.SendString("Incorrect password")
	}

	// Generate an API key.
	apiKey, err := utils.GenerateAPIKey()
	if err != nil {
		return c.SendString(fmt.Sprintf("Error generating API key: %v", err))
	}

	// Set the API key in Redis with a TTL of 7 days.
	ttl := 7 * 24 * time.Hour // 7 days in hours

	// Save the key in the database
	err = redisclient.Rdb.Set(context.Background(), "api_key", apiKey, ttl).Err()
	if err != nil {
		return c.SendString(fmt.Sprintf("Error Generating API key: %v", err))
	}

	// Expire the initial password
	err = redisclient.Rdb.Del(context.Background(), "initial_password").Err()
	if err != nil {
		return c.SendString(fmt.Sprintf("Error Deleting the initial password: %v", err))
	}

	// Return the API key.
	return c.SendString(fmt.Sprintf("API key: %s", apiKey))
}
