package controllers

import (
	"github.com/algo7/day-planner-gpt-data-portal/pkg/integrations/gmail"
	"github.com/algo7/day-planner-gpt-data-portal/pkg/integrations/outlook"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// GetOutlookEmails returns the user's outlook emails.
func GetOutlookEmails(c *fiber.Ctx) error {
	emails, err := outlook.GetEmails()
	if err != nil {
		log.Errorf("Error getting emails: %v", err)
		return c.RedirectToRoute("outlook_auth", nil, 302)
	}
	return c.JSON(emails)
}

// GetGmailEmails returns the user's outlook emails.
func GetGmailEmails(c *fiber.Ctx) error {
	emails, err := gmail.GetEmails()
	if err != nil {
		log.Errorf("Error getting emails: %v", err)
		return c.RedirectToRoute("google_auth", nil, 302)
	}
	return c.JSON(emails)
}
