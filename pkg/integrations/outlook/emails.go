package outlook

import (
	"context"
	"fmt"
	"log"
	"time"

	abstractions "github.com/microsoft/kiota-abstractions-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"
)

type OutlookEmail struct {
	Subject          string `json:"subject"`
	Body             string `json:"body"`
	Sender           string `json:"sender"`
	RecievedDateTime string `json:"recievedDateTime"`
}

// CustomAuthenticationProvider implements the AuthenticationProvider interface
type CustomAuthenticationProvider struct {
	accessToken string
}

// AuthenticateRequest adds the Authorization header to the request
func (c *CustomAuthenticationProvider) AuthenticateRequest(ctx context.Context, requestInfo *abstractions.RequestInformation, additionalAuthenticationContext map[string]interface{}) error {
	if requestInfo != nil {
		requestInfo.Headers.Add("Authorization", "Bearer "+c.accessToken)
	}
	return nil
}

// GetEmails calls the Microsoft Graph API to get the user's emails.
func GetEmails(accessToken string) error {

	// Create an instance of CustomAuthenticationProvider with the access token
	customAuthProvider := &CustomAuthenticationProvider{accessToken: accessToken}

	// Create a new Graph service client with the custom authentication provider
	adapter, err := msgraphsdk.NewGraphRequestAdapter(customAuthProvider)
	if err != nil {
		log.Fatalf("Could not create request adapter: %v", err)
	}

	graphClient := msgraphsdk.NewGraphServiceClient(adapter)

	// Use the graphClient to make API calls
	user, err := graphClient.Me().Get(context.Background(), nil)
	if err != nil {
		log.Fatalf("Error getting user: %v", err)
	}

	log.Printf("User: %v", user)

	// Get the current time
	now := time.Now()

	// Subtract 2 days from the current time
	twoDaysAgo := now.AddDate(0, 0, -2)

	// Format the time in ISO 8601 format
	twoDaysAgoStr := twoDaysAgo.Format("2006-01-02T15:04:05Z")
	fmt.Println(twoDaysAgoStr)

	requestFilter := fmt.Sprintf("singleValueExtendedProperties/Any(ep: ep/id eq 'String 0x001A' and contains(ep/value, 'IPM.Note')) and receivedDateTime ge %s ", twoDaysAgoStr)

	requestParameters := &graphusers.ItemMessagesRequestBuilderGetQueryParameters{
		Select:  []string{"sender", "subject", "receivedDateTime"},
		Orderby: []string{"receivedDateTime DESC"},
		Filter:  &requestFilter,
	}
	configuration := &graphusers.ItemMessagesRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParameters,
	}

	messages, err := graphClient.Me().Messages().Get(context.Background(), configuration)

	if err != nil {
		return fmt.Errorf("Error getting messages: %w", err)
	}

	// Initialize iterator
	pageIterator, _ := msgraphcore.NewPageIterator[*models.Message](messages, graphClient.GetAdapter(), models.CreateMessageCollectionResponseFromDiscriminatorValue)

	// Iterate over all pages
	err = pageIterator.Iterate(context.Background(), func(message *models.Message) bool {
		fmt.Printf("%s\n", *message.GetSubject())
		// Return true to continue the iteration
		return true
	})

	// Check for errors
	if err != nil {
		return fmt.Errorf("Error iterating over messages: %w", err)
	}

	return nil
}
