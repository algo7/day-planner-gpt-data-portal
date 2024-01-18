package controllers

import (
	"strings"

	"github.com/algo7/day-planner-gpt-data-portal/pkg/integrations/gmail"
	"github.com/algo7/day-planner-gpt-data-portal/pkg/integrations/outlook"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// GetOutlookEmails returns the user's outlook emails.
func GetOutlookEmails(c *fiber.Ctx) error {
	emails, err := outlook.GetEmails()

	// Check if the error is due to redis connection
	if strings.Contains(err.Error(), "redis") {
		log.Errorf("Error getting emails due to redis connection: %v", err)
		return c.SendString("Unable to get emails due to token retrieval error. Please check the server logs.")
	}

	// Other errors
	if err != nil {
		log.Errorf("Error getting emails: %v", err)
		return c.RedirectToRoute("outlook_auth", nil, 302)
	}

	return c.JSON(emails)
}

// GetGmailEmails returns the user's outlook emails.
func GetGmailEmails(c *fiber.Ctx) error {
	emails, err := gmail.GetEmails()

	// Check if the error is due to redis connection
	if strings.Contains(err.Error(), "redis") {
		log.Errorf("Error getting emails due to redis connection: %v", err)
		return c.SendString("Unable to get emails due to token retrieval error. Please check the server logs.")
	}

	if err != nil {
		log.Errorf("Error getting emails: %v", err)
		return c.RedirectToRoute("google_auth", nil, 302)
	}
	return c.JSON(emails)
}
