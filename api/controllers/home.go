package controllers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// GetHome returns the home page.
func GetHome(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.SendString("No authorization code found in the request")
	}

	// Exchange the code for an access token here
	// ...

	return c.SendString(fmt.Sprintf("Authorization code: %s", code))
}
