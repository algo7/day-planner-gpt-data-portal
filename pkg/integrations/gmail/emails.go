package gmail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
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

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	// authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	// fmt.Printf("Go to the following link in your browser then type the "+
	// 	"authorization code: \n%v\n", authURL)

	// var authCode string
	// if _, err := fmt.Scan(&authCode); err != nil {
	// 	log.Fatalf("Unable to read authorization code: %v", err)
	// }

	tok, err := config.Exchange(context.TODO(), "")
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// GetEmails calls the Gmail API to get the user's emails.
func GetEmails() ([]Email, error) {

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(config)

	srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
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
		log.Fatalf("Unable to retrieve messages: %v", err)
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
			log.Fatalf("Unable to retrieve message: %v", err)
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
