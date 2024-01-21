package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	redisclient "github.com/algo7/day-planner-gpt-data-portal/internal/redis"
	"github.com/algo7/day-planner-gpt-data-portal/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

/*
* OAuth2 PCKE Flow
 */

// GetAuthOutlook returns the auth page for Outlook
// @Summary Get Outlook Auth Page
// @Description Redirects to the Outlook OAuth2 authentication page.
// @Tags OAuth2
// @Accept json
// @Produce json
// @Success 302 {string} string "Redirect to the Outlook OAuth2 authentication page."
// @Failure 500 {string} string "Error loading OAuth2 config"
// @Router /v1/auth/oauth/outlook/auth [get]
func GetAuthOutlook(c *fiber.Ctx) error {

	// Load the OAuth2 config from the JSON file
	config, err := utils.GetOAuth2Config("outlook")
	if err != nil {
		return c.SendString(fmt.Sprintf("Error loading OAuth2 config: %v", err))
	}

	// Get the URL to visit to authorize the application
	authURL, _, err := utils.GenerateOauthURL(config, "outlook", "PCKE")
	if err != nil {
		return c.SendString(fmt.Sprintf("Error generating OAuth2 URL: %v", err))
	}

	// Show the user the URL to visit to authorize our application
	return c.Status(fiber.StatusContinue).SendString(fmt.Sprintf("Please complete the authorization workflow by going to the following URL %s", authURL))
}

// GetAuthGoogle returns the auth page for Google
// @Summary Get Google Auth Page
// @Description Redirects to the Google OAuth2 authentication page.
// @Tags OAuth2
// @Accept json
// @Produce json
// @Param code query string true "Authorization code from Google OAuth2 provider"
// @Success 302 {string} string "Redirect to a predefined route after successful authorization"
// @Failure 400 {string} string "No authorization code found in the request"
// @Failure 500 {string} string "Error loading OAuth2 config or exchanging code for token"
// @Router /v1/auth/oauth/google/auth [get]
func GetAuthGoogle(c *fiber.Ctx) error {

	b, err := os.ReadFile("./credentials/google_credentials.json")
	if err != nil {
		c.SendString(fmt.Sprintf("Unable to read client secret file: %v", err))
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		c.SendString(fmt.Sprintf("Unable to parse client secret file to config: %v", err))
	}

	// Get the URL to visit to authorize the application
	authURL, _, err := utils.GenerateOauthURL(config, "google", "PCKE")
	if err != nil {
		return c.SendString(fmt.Sprintf("Error generating OAuth2 URL: %v", err))
	}

	// Show the user the URL to visit to authorize our application
	return c.Status(fiber.StatusContinue).SendString(fmt.Sprintf("Please complete the authorization workflow by going to the following URL %s", authURL))
}

// GetOAuthCallBack handles the redirect from the OAuth2 provider
// @Summary Handles OAuth2 Redirect
// @Description Handles the callback from the give OAuth2 provider, exchanging the authorization code for an access token.
// @Tags OAuth2
// @Accept json
// @Produce json
// @Param code query string true "Authorization code from Google OAuth2 provider"
// @Success 302 {string} string "Redirect to a predefined route after successful authorization"
// @Failure 400 {string} string "No authorization code found in the request"
// @Failure 400 {string} string "No state token found in the request"
// @Failure 500 {string} string "Internal server error"
// @Router /v1/auth/oauth/callback [get]
func GetOAuthCallBack(c *fiber.Ctx) error {

	// Get the authorization code and the state token from the request
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		return c.Status(fiber.StatusBadRequest).SendString("No authorization code found in the request")
	}

	// Check if the state token is valid
	if state == "" {
		return c.Status(fiber.StatusBadRequest).SendString("No state token found in the request")
	}

	stateToken, err := redisclient.Rdb.GetDel(context.Background(), fmt.Sprintf("stateToken_%s", state)).Result()
	if err != nil {
		if err == redis.Nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid state token")
		}
		log.Printf("Error getting state token from Redis: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error getting state token")
	}

	if stateToken != state {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid state token")
	}

	// Parses the state token base on - as the delimiter to get the provider
	provider := strings.Split(state, "-")[0]

	// Check if the provider is valid
	_, ok := utils.ValidProviders[provider]
	if !ok {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid provider")
	}

	// Empty OAuth2 config to be filled based on the provider
	authConfig, err := utils.GetOAuth2Config(provider)
	if err != nil {
		log.Printf("Error getting OAuth2 config: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error getting OAuth2 config")
	}

	// Exchange the code for an access token here
	tok, err := utils.ExchangeCodeForToken(authConfig, code)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		c.Status(fiber.StatusInternalServerError).SendString("Error exchanging code for token")
	}

	// Save the token in Redis
	err = utils.SaveToken(provider, tok)
	if err != nil {
		log.Printf("Error saving token: %v", err)
		c.Status(fiber.StatusInternalServerError).SendString("Error saving token")
	}

	// return c.SendString(fmt.Sprintf("Authorization code: %s", code))
	return c.RedirectToRoute("oauth_success", nil, 302)
}

/*
* OAuth2 Device Flow
 */

// GetAuthGoogleDevice gets the information for the device flow for Google
// @Summary Gets the link and user code for the device flow for Google
// @Description Gets the link and user code for the device flow for Google
// @Tags OAuth2
// @Accept json
// @Produce json
// @Success 200 {string} string "Please go to https://www.google.com/device and enter the following code xxx-xxx-xxx"
// @Failure 500 {string} string "Error loading OAuth2 config"
// @Router /v1/auth/oauth/google/device [get]
func GetAuthGoogleDevice(c *fiber.Ctx) error {

	b, err := os.ReadFile("./credentials/google_device_credentials.json")
	if err != nil {
		log.Printf("Unable to read client secret file: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "email")
	if err != nil {
		log.Printf("Unable to parse client secret file to config: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Get the URL to visit to authorize the application
	url, deviceCode, err := utils.GenerateOauthURL(config, "google", "Device")
	if err != nil {
		log.Printf("Error getting device flow info: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Start polling for the token in a non-blocking way
	go func() {
		tok, err := utils.PollToken(config, deviceCode)
		if err != nil {
			log.Println(fmt.Errorf("unable to poll token: %v", err))
			return
		}

		// Marshals the token into a JSON object
		tokenJSON, err := json.Marshal(tok)
		if err != nil {
			log.Println(fmt.Errorf("Unable to marshal token: %v", err))
		}
		ttl := 7 * 24 * time.Hour
		err = redisclient.Rdb.Set(context.Background(), "google", tokenJSON, ttl).Err()
		if err != nil {
			log.Println(fmt.Errorf("unable to save the polled token to redis: %w", err))
			return
		}
	}()

	// Redirect the user to the authURL
	return c.SendString(url)
}

/*
* Refresh Token
 */

// GetNewTokenFromRefreshToken handles the redirect from the OAuth2 provider
// @Summary OAuth2 Redirect for Google
// @Description Handles the callback from Google OAuth2 authentication, exchanging the authorization code for an access token.
// @Tags OAuth2
// @Accept json
// @Produce json
// @Param provider query string true "Provider to get the new token from the refresh token"
// @Failure 500 {string} string "Internal server error"
// @Success 302 {string} string "Redirect to the auth success page"
// @Router /v1/auth/oauth/refresh [get]
func GetNewTokenFromRefreshToken(c *fiber.Ctx) error {

	provider := c.Query("provider")

	// Check if the provider is valid
	_, ok := utils.ValidProviders[provider]
	if !ok {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid provider")
	}

	// Get the OAuth2 config for the provider
	providerConfig, err := utils.GetOAuth2Config(provider)
	if err != nil {
		log.Printf("Error getting OAuth2 config: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error getting OAuth2 config")
	}

	tok, err := utils.RetrieveToken(provider)
	if err != nil {
		log.Printf("Error getting token: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Token not found or has fully expired")
	}

	// Get the token from the refresh token
	newTok, err := utils.GetTokenFromRefreshToken(providerConfig, tok.RefreshToken)
	if err != nil {
		log.Printf("Error getting token from refresh token: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error getting token from refresh token")
	}

	// Update the token in Redis
	err = utils.UpdateToken(provider, newTok)
	if err != nil {
		log.Printf("Error updating token: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error updating token")
	}

	return c.RedirectToRoute("oauth_success", nil, 302)
}

/*
* Internal Auth
 */

// GetAPIKey returns a page to get the initial API key
// @Summary Get API Key Page
// @Description Returns a page to get the initial API key. If the initial password has expired, it redirects to the home page.
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Render the API key form page"
// @Failure 302 {string} string "Redirect to the home page if the initial password has expired"
// @Failure 500 {string} string "Error getting initial password from Redis"
// @Router /v1/auth/internal/apikey [get]
func GetAPIKey(c *fiber.Ctx) error {

	// Check if the initial password exists in Redis
	initialPassword, err := redisclient.Rdb.Get(context.Background(), "initial_password").Result()
	if err != nil {
		// If the initial password does not exist, warn the user to restart the server to generate a new password.
		if err == redis.Nil {
			return c.
				Status(fiber.StatusInternalServerError).
				SendString("Initial password does not exist. Please restart the server to generate a new password.")
		}
		// If there is an error getting the initial password, log the error and return a 500 status code.
		log.Printf("Error getting initial password: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// If the initial password is an empty string, redirect to the home page. It means that the initial password has been used.
	if initialPassword == "" {
		return c.RedirectToRoute("home", nil, 302)
	}

	return c.Render("apikey_form", fiber.Map{})
}

// PostAPIKey generates and returns a new API key
// @Summary Generate API Key
// @Description Generates a new API key and stores it in Redis with a TTL of 7 days. If the initial password has expired, it redirects to the home page.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param password formData string true "Password for API key generation"
// @Success 200 {string} string "API key: {apiKey}"
// @Failure 302 {string} string "Redirect to the home page if the initial password has expired"
// @Failure 400 {string} string "Incorrect password"
// @Failure 500 {string} string "Error getting initial password or generating API key"
// @Router /v1/auth/internal/apikey [post]
func PostAPIKey(c *fiber.Ctx) error {

	// Check if the initial password is still in Redis
	initialPassword, err := redisclient.Rdb.Get(context.Background(), "initial_password").Result()
	if err != nil {
		// If the initial password does not exist, warn the user to restart the server to generate a new password.
		if err == redis.Nil {
			return c.SendString("Initial password does not exist. Please restart the server to generate a new password.")
		}
		// If there is an error getting the initial password, log the error and return a 500 status code.
		log.Printf("Error getting initial password: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// If the initial password is an empty string, redirect to the home page. It means that the initial password has been used.
	if initialPassword == "" {
		return c.RedirectToRoute("home", nil, 302)
	}

	// Get the password from the form
	password := c.FormValue("password")

	// Compare the password with the initial password
	if password != initialPassword {
		return c.SendString("Incorrect password")
	}

	// Generate an API key.
	apiKey, err := utils.GenerateAPIKey()
	if err != nil {
		log.Printf("Error generating API key: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Set the API key in Redis with a TTL of 7 days.
	ttl := 7 * 24 * time.Hour // 7 days in hours

	// Save the key in the database
	err = redisclient.Rdb.Set(context.Background(), fmt.Sprintf("apikey_%s", apiKey), apiKey, ttl).Err()
	if err != nil {
		log.Printf("Error saving API key: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Set the initial password to an empty string
	err = redisclient.Rdb.Set(context.Background(), "initial_password", "", 0).Err()
	if err != nil {
		log.Printf("Error Deleting the initial password: %v", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Return the API key.
	return c.SendString(fmt.Sprintf("API key: %s", apiKey))
}

/*
* Success
 */

// GetAuthSuccess returns a page to show that the oauth authentication was successful
// @Summary OAuth2 Success Page
// @Description Returns a page to show that the oauth authentication was successful.
// @Tags OAuth2
// @Accept */*
// @Produce plain
// @Success 200 {string} string "Auth Success"
// @Router /success [get]
func GetAuthSuccess(c *fiber.Ctx) error {
	return c.Status(200).SendString("Auth Success")
}
