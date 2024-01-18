package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/algo7/day-planner-gpt-data-portal/pkg/utils"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Email is a struct to hold the email data
type Email struct {
	Subject          string `json:"subject"`
	Body             string `json:"body"`
	Sender           string `json:"sender"`
	RecievedDateTime string `json:"recievedDateTime"`
}

// GetEmails calls the Gmail API to get the user's emails.
func GetEmails() ([]Email, error) {

	b, err := os.ReadFile("google_credentials.json")
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file: %w", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file to config: %w", err)
	}

	client, err := utils.GetClient(config, "google_token.json")
	if err != nil {
		return nil, fmt.Errorf("unable to get OAuth client: %w", err)
	}

	srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve Gmail client: %w", err)
	}

	user := "me"

	// Get the current time
	now := time.Now()

	// Subtract 2 days from the current time
	twoDaysAgo := now.AddDate(0, 0, -2)

	// Format the time in ISO 8601 format
	twoDaysAgoStr := twoDaysAgo.Format("2006-01-02")

	fmt.Println("Two days ago was:", twoDaysAgoStr)

	m, err := srv.Users.Messages.List(user).Q("is:unread").Q(fmt.Sprintf("after:%s", twoDaysAgoStr)).Do()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve messages: %w", err)
	}

	if len(m.Messages) == 0 {
		fmt.Println("No messages found.")
		return nil, err
	}

	// Get the content of each email
	gmailEmails := []Email{}
	for _, msg := range m.Messages {
		c, err := srv.Users.Messages.Get(user, msg.Id).Do()
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve message: %w", err)
		}

		gmailEmails = append(gmailEmails, Email{
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
