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
// @Summary Get Outlook Emails
// @Description Retrieves emails from the user's Outlook account.
// @Tags Emails
// @Accept json
// @Produce json
// @Success 200 {array} integrations.Email "List of Outlook emails"
// @Failure 302 {string} string "Redirect to Outlook authentication if the access token is missing or invalid"
// @Failure 500 {string} string "Unable to retrieve emails due to server error or token retrieval issue"
// @Router /v1/email/outlook [get]
func GetOutlookEmails(c *fiber.Ctx) error {

	emails, err := outlook.GetEmails()

	if err != nil {

		// Redis related errors that are not due to the token key not being found
		if strings.Contains(err.Error(), "redis") && err != redis.Nil {
			log.Printf("Error getting emails due to redis connection: %v", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		// Redis related errors that are due to the token key not being found
		if err == redis.Nil {
			log.Println("Outlook Access token not found in redis")
			return c.RedirectToRoute("outlook_auth", nil, 302)
		}

		// Non-redis related errors
		log.Printf("Error getting emails: %v", err)
		return c.RedirectToRoute("outlook_auth", nil, 302)
	}
	return c.JSON(emails)
}

// GetGmailEmails returns the user's Gmail emails.
// @Summary Get Gmail Emails
// @Description Retrieves emails from the user's Gmail account.
// @Tags Emails
// @Accept json
// @Produce json
// @Success 200 {array} integrations.Email "List of Gmail emails"
// @Failure 302 {string} string "Redirect to Google authentication if the access token is missing or invalid"
// @Failure 500 {string} string "Unable to retrieve emails due to server error or token retrieval issue"
// @Router /v1/email/google [get]
func GetGmailEmails(c *fiber.Ctx) error {

	emails, err := gmail.GetEmails()

	if err != nil {

		// Redis related errors that are not due to the token key not being found
		if strings.Contains(err.Error(), "redis") && err != redis.Nil {
			log.Printf("Error getting emails due to redis connection: %v", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		// Redis related errors that are due to the token key not being found
		if err == redis.Nil {
			log.Println("Gmail Access token not found in redis")
			return c.RedirectToRoute("google_auth", nil, 302)
		}

		// Non-redis related errors
		log.Printf("Error getting emails: %v", err)
		return c.RedirectToRoute("google_auth", nil, 302)
	}

	return c.JSON(emails)
}
