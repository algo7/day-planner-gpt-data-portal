package controllers

import (
	"log"
	"strings"

	"github.com/algo7/day-planner-gpt-data-portal/pkg/integrations/gmail"
	"github.com/algo7/day-planner-gpt-data-portal/pkg/integrations/outlook"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// GetOutlookEmails returns the user's outlook emails.
func GetOutlookEmails(c *fiber.Ctx) error {

	emails, err := outlook.GetEmails()

	if err != nil {

		// Redis related errors that are not due to the token key not being found
		if strings.Contains(err.Error(), "redis") && !strings.Contains(err.Error(), redis.Nil.Error()) {
			log.Printf("Error getting emails due to redis connection: %v", err)
			return c.SendString("Unable to get emails due to token retrieval error. Please check the server logs.")
		}

		// Redis related errors that are due to the token key not being found
		if strings.Contains(err.Error(), redis.Nil.Error()) {
			log.Println("Access token not found in redis")
			return c.RedirectToRoute("outlook_auth", nil, 302)
		}

		// Non-redis related errors
		log.Printf("Error getting emails: %v", err)
		return c.RedirectToRoute("outlook_auth", nil, 302)
	}
	return c.JSON(emails)
}

// GetGmailEmails returns the user's outlook emails.
func GetGmailEmails(c *fiber.Ctx) error {
	emails, err := gmail.GetEmails()

	if err != nil {

		// Redis related errors that are not due to the token key not being found
		if strings.Contains(err.Error(), "redis") && !strings.Contains(err.Error(), redis.Nil.Error()) {
			log.Printf("Error getting emails due to redis connection: %v", err)
			return c.SendString("Unable to get emails due to token retrieval error. Please check the server logs.")
		}

		// Redis related errors that are due to the token key not being found
		if strings.Contains(err.Error(), redis.Nil.Error()) {
			log.Printf("Token not found in redis")
			return c.RedirectToRoute("google_auth", nil, 302)
		}

		// Non-redis related errors
		log.Printf("Error getting emails: %v", err)
		return c.RedirectToRoute("google_auth", nil, 302)
	}

	return c.JSON(emails)
}
