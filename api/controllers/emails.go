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
// @ID getOutlookEmails
// @Description This endpoint retrieves emails from Outlook. If there is an error, it redirects to the Outlook authentication route or returns a server error.
// @Tags Email
// @Accept json
// @Produce json
// @Success 200 {array} integrations.Email "Returns the retrieved emails"
// @Failure 307 {string} string "Redirects to the Outlook authentication route if the access token is not found in Redis or there is a non-Redis related error"
// @Success 200 {Object} Response "Returns a message if the outlook session has expired"
// @Router /v1/email/outlook [get]
func GetOutlookEmails(c *fiber.Ctx) error {

	emails, err := outlook.GetEmails()

	if err != nil {

		// Redis related errors that are not due to the token key not being found
		if strings.Contains(err.Error(), "redis") && err != redis.Nil {
			log.Printf("Error getting emails due to redis connection: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(Response{Error: "Unable to retrieve emails due to server error or token retrieval issue"})
		}

		// Redis related errors that are due to the token key not being found
		if err == redis.Nil {
			log.Println("Outlook Access token not found in redis")
			return c.RedirectToRoute("outlook_auth", nil, fiber.StatusTemporaryRedirect)
		}

		// Non-redis related errors
		log.Printf("Error getting emails: %v", err)

		// return c.RedirectToRoute("outlook_auth", nil, fiber.StatusTemporaryRedirect)
		return c.Status(fiber.StatusOK).JSON(Response{
			Data: "You outlook session has expired, please re-authenticate using provider=outlook",
		})
	}
	return c.Status(fiber.StatusOK).JSON(emails)
}

// GetGmailEmails returns the user's Gmail emails.
// @Summary Get Gmail Emails
// @ID getGmailEmails
// @Description This endpoint retrieves emails from Gmail. If there is an error, it redirects to the Google authentication route or returns a server error.
// @Tags Email
// @Accept json
// @Produce json
// @Success 200 {array} integrations.Email "Returns the retrieved emails"
// @Success 200 {Object} Response "Returns a message if the Gmail session has expired"
// @Failure 500 {object} Response "Returns an error message if there is a Redis related error that is not due to the token key not being found"
// @Router /v1/email/google [get]
func GetGmailEmails(c *fiber.Ctx) error {

	emails, err := gmail.GetEmails()

	if err != nil {

		// Redis related errors that are not due to the token key not being found
		if strings.Contains(err.Error(), "redis") && err != redis.Nil {
			log.Printf("Error getting emails due to redis connection: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(Response{Error: "Unable to retrieve emails due to server error or token retrieval issue"})
		}

		// Redis related errors that are due to the token key not being found
		if err == redis.Nil {
			log.Println("Gmail Access token not found in redis")
			return c.RedirectToRoute("google_auth", nil, fiber.StatusTemporaryRedirect)
		}

		// Non-redis related errors
		log.Printf("Error getting emails: %v", err)
		// return c.RedirectToRoute("google_auth", nil, fiber.StatusTemporaryRedirect)
		return c.Status(fiber.StatusOK).JSON(Response{
			Data: "You gmail session has expired, please re-authenticate using provider=google",
		})
	}

	return c.Status(fiber.StatusOK).JSON(emails)
}
