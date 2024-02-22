package outlook

import (
	"context"
	"fmt"
	"time"

	"github.com/algo7/day-planner-gpt-data-portal/pkg/integrations"
	"github.com/algo7/day-planner-gpt-data-portal/pkg/utils"
	abstractions "github.com/microsoft/kiota-abstractions-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"
)

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
func GetEmails() ([]integrations.Email, error) {

	accessToken, err := utils.RetrieveToken("outlook")
	if err != nil {
		return nil, err
	}

	// Create an instance of CustomAuthenticationProvider with the access token
	customAuthProvider := &CustomAuthenticationProvider{accessToken: accessToken.AccessToken}

	// Create a new Graph service client with the custom authentication provider
	adapter, err := msgraphsdk.NewGraphRequestAdapter(customAuthProvider)
	if err != nil {
		return nil, fmt.Errorf("Could not create request adapter: %v", err)
	}

	graphClient := msgraphsdk.NewGraphServiceClient(adapter)

	// Get the current time
	now := time.Now()

	// Subtract 3 days from the current time
	dateDiff := now.AddDate(0, 0, -2)

	// Format the time in ISO 8601 format
	dateDiffStr := dateDiff.Format("2006-01-02T15:04:05Z")

	requestFilter := fmt.Sprintf("singleValueExtendedProperties/Any(ep: ep/id eq 'String 0x001A' and contains(ep/value, 'IPM.Note')) and receivedDateTime ge %s ", dateDiffStr)

	requestParameters := &graphusers.ItemMessagesRequestBuilderGetQueryParameters{
		Select:  []string{"sender", "subject", "bodyPreview", "receivedDateTime"},
		Orderby: []string{"receivedDateTime DESC"},
		Filter:  &requestFilter,
	}
	configuration := &graphusers.ItemMessagesRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParameters,
	}

	messages, err := graphClient.Me().Messages().Get(context.Background(), configuration)

	if err != nil {
		return nil, fmt.Errorf("Error getting messages: %w", err)
	}

	// Initialize iterator
	pageIterator, _ := msgraphcore.NewPageIterator[*models.Message](messages, graphClient.GetAdapter(), models.CreateMessageCollectionResponseFromDiscriminatorValue)

	OutlookEmails := []integrations.Email{}

	// Iterate over all pages
	err = pageIterator.Iterate(context.Background(), func(message *models.Message) bool {

		OutlookEmails = append(OutlookEmails, integrations.Email{
			Subject:          *message.GetSubject(),
			Body:             *message.GetBodyPreview(),
			Sender:           *message.GetSender().GetEmailAddress().GetAddress(),
			RecievedDateTime: message.GetReceivedDateTime().Format("2006-01-02T15:04:05Z"),
		})

		// Return true to continue the iteration
		return true
	})

	// Check for errors
	if err != nil {
		return nil, fmt.Errorf("Error iterating over messages: %w", err)
	}

	return OutlookEmails, nil
}
