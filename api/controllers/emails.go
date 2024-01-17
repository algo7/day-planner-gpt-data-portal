package controllers

import (
	"github.com/algo7/day-planner-gpt-data-portal/pkg/integrations/gmail"
	"github.com/algo7/day-planner-gpt-data-portal/pkg/integrations/outlook"
	"github.com/gofiber/fiber/v2"
)

// GetOutlookEmails returns the user's outlook emails.
func GetOutlookEmails(c *fiber.Ctx) error {
	emails, err := outlook.GetEmails()
	if err != nil {
		return c.SendString(err.Error())
	}
	return c.JSON(emails)
}

// GetGmailEmails returns the user's outlook emails.
func GetGmailEmails(c *fiber.Ctx) error {
	emails, err := gmail.GetEmails()
	if err != nil {
		return c.SendString(err.Error())
	}
	return c.JSON(emails)
}
