package gmail

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/algo7/day-planner-gpt-data-portal/pkg/integrations"
	"github.com/algo7/day-planner-gpt-data-portal/pkg/utils"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// GetEmails calls the Gmail API to get the user's emails.
func GetEmails() ([]integrations.Email, error) {

	// Get the OAuth2 config
	config, err := utils.GetOAuth2Config("google")
	if err != nil {
		return nil, err
	}

	// Get the token from redis
	token, err := utils.RetrieveToken("google")
	if err != nil {
		return nil, err
	}

	// Create a new HTTP client and bind it to the token
	client := config.Client(context.Background(), token)

	// Create a new Gmail service client using the HTTP client
	srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve Gmail client: %w", err)
	}

	// The current logged in user
	user := "me"

	// Get the current time
	// now := time.Now()

	// Subtract 2 days from the current time
	// twoDaysAgo := now.AddDate(0, 0, -2)

	// Format the time in ISO 8601 format
	// twoDaysAgoStr := twoDaysAgo.Format("2006-01-02")

	// m, err := srv.Users.Messages.List(user).Q("is:unread").Q(fmt.Sprintf("after:%s", twoDaysAgoStr)).Do()
	m, err := srv.Users.Messages.List(user).Q("is:unread").Do()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve messages: %w", err)
	}

	if len(m.Messages) == 0 {
		fmt.Println("No messages found.")
		return nil, err
	}

	// Get the content of each email
	gmailEmails := []integrations.Email{}
	for _, msg := range m.Messages {
		c, err := srv.Users.Messages.Get(user, msg.Id).Do()
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve message: %w", err)
		}

		gmailEmails = append(gmailEmails, integrations.Email{
			Subject:          getHeader("Subject", c.Payload.Headers),
			Body:             getMessageBody(c.Payload),
			Sender:           getHeader("From", c.Payload.Headers),
			RecievedDateTime: getHeader("Date", c.Payload.Headers),
		})

	}

	return gmailEmails, nil
}

func getHeader(name string, headers []*gmail.MessagePartHeader) string {
	for _, header := range headers {
		if header.Name == name {
			return header.Value
		}
	}
	return ""
}

func getMessageBody(payload *gmail.MessagePart) string {
	if payload.MimeType == "text/plain" {
		data, _ := base64.URLEncoding.DecodeString(payload.Body.Data)
		return string(data)
	}

	for _, part := range payload.Parts {
		if part.MimeType == "text/plain" {
			data, _ := base64.URLEncoding.DecodeString(part.Body.Data)
			return string(data)
		}
		// Recursively check in nested parts
		if len(part.Parts) > 0 {
			return getMessageBody(part)
		}
	}
	return ""
}
