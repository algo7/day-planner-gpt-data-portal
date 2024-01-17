package controllers

import (
	"fmt"

	"github.com/algo7/day-planner-gpt-data-portal/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

// GetAuth returns the auth page
func GetAuth(c *fiber.Ctx) error {

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

// GetOauthRedirect handles the redirect from the OAuth2 provider
func GetOauthRedirect(c *fiber.Ctx) error {
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

	return c.SendString(fmt.Sprintf("Authorization code: %s", code))
}
