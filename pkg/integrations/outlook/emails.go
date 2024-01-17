package outlook

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// GetEmails calls the Microsoft Graph API to get the user's emails.
func GetEmails(accessToken string) error {
	url := "https://graph.microsoft.com/v1.0/me/messages"

	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)

	// add authorization header to the req
	req.Header.Add("Authorization", "Bearer "+accessToken)

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response JSON
	var data map[string]interface{}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Printf("Error while decoding the response: %v", err)
		return err
	}

	// Here you can process the data
	// For now, we just print it
	// fmt.Println(data)

	// Loop through the emails and gets the body.content and the subject
	for _, email := range data["value"].([]interface{}) {
		fmt.Println(email)
		// fmt.Println("Body:", email.(map[string]interface{})["bodyPreview"])
	}

	return nil
}
