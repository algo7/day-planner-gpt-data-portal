package controllers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	redisclient "github.com/algo7/day-planner-gpt-data-portal/internal/redis"
	"github.com/algo7/day-planner-gpt-data-portal/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

/*
* OAuth2 PCKE Flow
 */

// GetOAtuh returns the auth URL for the given OAuth2 provider
// @Summary Get OAuth2 Authentication URL
// @Description This endpoint generates the OAuth2 authentication URL for the specified provider.
// @Tags OAuth2
// @Accept json
// @Produce json
// @Param provider query string true "Name of the OAuth2 provider to generate the authentication URL for"
// @Success 200 {object} map[string]string "Returns a message with the URL to visit to authorize the application"
// @Failure 400 {object} map[string]string "Returns an error message if the provided OAuth2 provider is invalid"
// @Failure 500 {object} map[string]string "Returns an error message if there was an error loading the OAuth2 configuration or generating the OAuth2 URL"
// @Router /v1/auth/oauth [get]
func GetOAtuh(c *fiber.Ctx) error {

	provider := c.Query("provider")

	// Check if the provider is valid
	_, ok := utils.ValidProviders[provider]
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid provider"})
	}

	// Check Access Token Status
	token, err := utils.RetrieveToken(provider)
	if err != nil && err != redis.Nil {
		log.Printf("Error getting token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error Checking Access Token Status"})
	}

	if token != nil {
		// Calculate how many minutes are left until the token expires and round it up to the nearest minute
		minutesLeft := int(token.Expiry.Sub(time.Now()).Minutes() + 1)
		if minutesLeft > 0 {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{"error": fmt.Sprintf("Access Token is still valid for %v miutes", minutesLeft)})
		}
	}

	// Load the OAuth2 config from the JSON file
	config, err := utils.GetOAuth2Config(provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Error loading OAuth2 config: %v", err)})
	}

	// Get the URL to visit to authorize the application
	authURL, _, err := utils.GenerateOauthURL(config, provider, "PCKE")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Error generating OAuth2 URL: %v", err)})
	}

	// Show the user the URL to visit to authorize our application
	return c.Status(200).JSON(fiber.Map{"data": fmt.Sprintf("Please complete the authorization workflow by going to the following URL:\n %s", authURL)})
}

// GetOAuthCallBack handles the redirect from the OAuth2 provider
// @Summary OAuth2 Callback Endpoint
// @Description This endpoint handles the callback from the OAuth2 provider, exchanges the authorization code for an access token, and saves the token.
// @Tags OAuth2
// @Accept json
// @Produce json
// @Param code query string true "Authorization code returned by the OAuth2 provider"
// @Param state query string true "State token for CSRF protection"
// @Success 307 {string} string "Redirects to the OAuth success route on successful token exchange and save"
// @Failure 400 {object} map[string]string "Returns an error message if the authorization code or state token is missing or invalid, or if the OAuth2 provider is invalid"
// @Failure 500 {object} map[string]string "Returns an error message if there was an error getting the OAuth2 configuration, exchanging the code for a token, or saving the token"
// @Router /v1/auth/oauth/callback [get]
func GetOAuthCallBack(c *fiber.Ctx) error {

	// Get the authorization code and the state token from the request
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No authorization code found in the request"})
	}

	// Check if the state token is valid
	if state == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No state token found in the request"})
	}

	stateToken, err := redisclient.Rdb.GetDel(context.Background(), fmt.Sprintf("stateToken_%s", state)).Result()
	if err != nil {
		if err == redis.Nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid state token or state token has expired"})
		}
		log.Printf("Error getting state token from Redis: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error getting state token"})
	}

	if stateToken != state {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid state token"})
	}

	// Parses the state token base on - as the delimiter to get the provider
	provider := strings.Split(state, "-")[0]

	// Check if the provider is valid
	_, ok := utils.ValidProviders[provider]
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid provider"})
	}

	// Empty OAuth2 config to be filled based on the provider
	authConfig, err := utils.GetOAuth2Config(provider)
	if err != nil {
		log.Printf("Error getting OAuth2 config: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error getting OAuth2 config"})
	}

	// Exchange the code for an access token here
	tok, err := utils.ExchangeCodeForToken(authConfig, code)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error exchanging code for access token"})
	}

	// Save the token in Redis
	err = utils.SaveToken(provider, tok)
	if err != nil {
		log.Printf("Error saving token: %v", err)
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error saving access token"})
	}

	// return c.SendString(fmt.Sprintf("Authorization code: %s", code))
	return c.RedirectToRoute("oauth_success", nil, fiber.StatusTemporaryRedirect)
}

/*
* Refresh Token
 */

// GetNewTokenFromRefreshToken handles the redirect from the OAuth2 provider
// @Summary Get New Token From Refresh Token
// @Description This endpoint retrieves a new access token using the refresh token for the specified provider.
// @Tags OAuth2
// @Accept json
// @Produce json
// @Param provider query string true "Name of the OAuth2 provider to get the new access token for"
// @Success 307 {string} string "Redirects to the OAuth success route on successful token retrieval and update"
// @Failure 400 {object} map[string]string "Returns an error message if the provided OAuth2 provider is invalid"
// @Failure 500 {object} map[string]string "Returns an error message if there was an error getting the OAuth2 configuration, retrieving the token, getting the new token from the refresh token, or updating the token"
// @Router /v1/auth/oauth/refresh [get]
func GetNewTokenFromRefreshToken(c *fiber.Ctx) error {

	provider := c.Query("provider")

	// Check if the provider is valid
	_, ok := utils.ValidProviders[provider]
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid provider"})
	}

	// Get the OAuth2 config for the provider
	providerConfig, err := utils.GetOAuth2Config(provider)
	if err != nil {
		log.Printf("Error getting OAuth2 config: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error getting OAuth2 config"})
	}

	tok, err := utils.RetrieveToken(provider)
	if err != nil {
		log.Printf("Error getting token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Access token not found or has expired"})
	}

	// Get the token from the refresh token
	newTok, err := utils.GetTokenFromRefreshToken(providerConfig, tok.RefreshToken)
	if err != nil {
		log.Printf("Error getting token from refresh token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error getting access token from refresh token"})
	}

	// Update the token in Redis
	err = utils.UpdateToken(provider, newTok)
	if err != nil {
		log.Printf("Error updating token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error updating access token"})
	}

	return c.RedirectToRoute("oauth_success", nil, fiber.StatusTemporaryRedirect)
}

/*
* Internal Auth
 */

// GetAPIKey returns a page to get the initial API key
// @Summary Get API Key
// @Description This endpoint checks if the initial password exists in Redis and if it does, renders the API key form. If the initial password does not exist or has been used, it redirects to the home page or prompts the user to restart the server.
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} string "Renders the API key form if the initial password exists and has not been used"
// @Success 307 {string} string "Redirects to the home page if the initial password has been used"
// @Failure 500 {object} map[string]string "Returns an error message if the initial password does not exist or there was an error getting the initial password"
// @Router /v1/auth/internal/apikey [get]
func GetAPIKey(c *fiber.Ctx) error {

	// Check if the initial password exists in Redis
	initialPassword, err := redisclient.Rdb.Get(context.Background(), "initial_password").Result()
	if err != nil {
		// If the initial password does not exist, warn the user to restart the server to generate a new password.
		if err == redis.Nil {
			return c.
				Status(fiber.StatusInternalServerError).
				JSON(fiber.Map{"error": "Initial password does not exist. Please restart the server to generate a new password"})
		}
		// If there is an error getting the initial password, log the error and return a 500 status code.
		log.Printf("Error getting initial password: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error getting initial password"})
	}

	// If the initial password is an empty string, redirect to the home page. It means that the initial password has been used.
	if initialPassword == "" {
		return c.RedirectToRoute("home", nil, fiber.StatusTemporaryRedirect)
	}

	return c.Render("apikey_form", fiber.Map{})
}

// PostAPIKey generates and returns a new API key
// @Summary Post API Key
// @Description This endpoint checks if the initial password exists in Redis, compares it with the password from the form, generates an API key if the passwords match, saves the API key in Redis with a TTL of 7 days, and sets the initial password to an empty string.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param password formData string true "Password from the form"
// @Success 200 {object} map[string]string "Returns the generated API key"
// @Failure 400 {object} map[string]string "Returns an error message if the password from the form does not match the initial password"
// @Failure 500 {object} map[string]string "Returns an error message if the initial password does not exist, there was an error getting the initial password, generating the API key, saving the API key, or deleting the initial password"
// @Router /v1/auth/internal/apikey [post]
func PostAPIKey(c *fiber.Ctx) error {

	// Check if the initial password is still in Redis
	initialPassword, err := redisclient.Rdb.Get(context.Background(), "initial_password").Result()
	if err != nil {
		// If the initial password does not exist, warn the user to restart the server to generate a new password.
		if err == redis.Nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Initial password does not exist. Please restart the server to generate a new password"})
		}
		// If there is an error getting the initial password, log the error and return a 500 status code.
		log.Printf("Error getting initial password: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error getting initial password"})
	}

	// If the initial password is an empty string, redirect to the home page. It means that the initial password has been used.
	if initialPassword == "" {
		return c.RedirectToRoute("home", nil, fiber.StatusTemporaryRedirect)
	}

	// Get the password from the form
	password := c.FormValue("password")

	// Compare the password with the initial password
	if password != initialPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Incorrect password"})
	}

	// Generate an API key.
	apiKey, err := utils.GenerateAPIKey()
	if err != nil {
		log.Printf("Error generating API key: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error generating API key"})
	}

	// Set the API key in Redis with a TTL of 7 days.
	ttl := 7 * 24 * time.Hour // 7 days in hours

	// Save the key in the database
	err = redisclient.Rdb.Set(context.Background(), fmt.Sprintf("apikey_%s", apiKey), apiKey, ttl).Err()
	if err != nil {
		log.Printf("Error saving API key: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error saving API key"})
	}

	// Set the initial password to an empty string
	err = redisclient.Rdb.Set(context.Background(), "initial_password", "", 0).Err()
	if err != nil {
		log.Printf("Error Deleting the initial password: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error Deleting the initial password"})
	}

	// Return the API key.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": fmt.Sprintf("API key: %s", apiKey)})
}

/*
* Success
 */

// GetAuthSuccess returns a page to show that the oauth authentication was successful
// @Summary OAuth2 Success Page
// @Description This endpoint returns a success message after successful authentication.
// @Tags OAuth2
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Returns a success message indicating successful authentication"
// @Router /v1/auth/success [get]
func GetAuthSuccess(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": "Auth Success"})
}

/*
// * OAuth2 Device Flow
//  */

// // GetAuthGoogleDevice gets the information for the device flow for Google
// // @Summary Gets the link and user code for the device flow for Google
// // @Description Gets the link and user code for the device flow for Google
// // @Tags OAuth2
// // @Accept json
// // @Produce json
// // @Success 200 {string} string "Please go to https://www.google.com/device and enter the following code xxx-xxx-xxx"
// // @Failure 500 {string} string "Error loading OAuth2 config"
// // @Router /v1/auth/oauth/google/device [get]
// func GetAuthGoogleDevice(c *fiber.Ctx) error {

// 	config, err := utils.GetOAuth2Config("google")
// 	if err != nil {
// 		log.Printf("Error getting OAuth2 config: %v", err)
// 		return c.SendStatus(fiber.StatusInternalServerError)
// 	}

// 	// Get the URL to visit to authorize the application
// 	url, deviceCode, err := utils.GenerateOauthURL(config, "google", "Device")
// 	if err != nil {
// 		log.Printf("Error getting device flow info: %v", err)
// 		return c.SendStatus(fiber.StatusInternalServerError)
// 	}

// 	// Start polling for the token in a non-blocking way
// 	go func() {
// 		tok, err := utils.PollToken(config, deviceCode)
// 		if err != nil {
// 			log.Println(fmt.Errorf("unable to poll token: %v", err))
// 			return
// 		}

// 		// Marshals the token into a JSON object
// 		tokenJSON, err := json.Marshal(tok)
// 		if err != nil {
// 			log.Println(fmt.Errorf("Unable to marshal token: %v", err))
// 		}
// 		ttl := 7 * 24 * time.Hour
// 		err = redisclient.Rdb.Set(context.Background(), "google", tokenJSON, ttl).Err()
// 		if err != nil {
// 			log.Println(fmt.Errorf("unable to save the polled token to redis: %w", err))
// 			return
// 		}
// 	}()

// 	// Redirect the user to the authURL
// 	return c.SendString(url)
// }
