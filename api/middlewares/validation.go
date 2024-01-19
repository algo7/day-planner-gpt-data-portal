package middlewares

import (
	"context"
	"fmt"
	"time"

	redisclient "github.com/algo7/day-planner-gpt-data-portal/internal/redis"
	"github.com/gofiber/fiber/v2"
)

var protectedURL = []string{"/v1/email/outlook", "/v1/email/google"}

// ValidateAPIKey validates the API key
func ValidateAPIKey(c *fiber.Ctx, apiKey string) (bool, error) {
	// Get the API key from Redis
	storedKey, err := redisclient.Rdb.Get(context.Background(), fmt.Sprintf("apikey_%s", apiKey)).Result()
	if err != nil {
		return false, fmt.Errorf("Error getting the API key from Redis: %v", err)
	}

	// Compare the API key from Redis with the API key from the request
	if apiKey != storedKey {
		return false, nil
	}

	// If the key is valid, refresh its TTL to 7 days
	ttl := 7 * 24 * time.Hour

	err = redisclient.Rdb.Expire(context.Background(), fmt.Sprintf("apikey_%s", apiKey), ttl).Err()
	if err != nil {
		return false, fmt.Errorf("error refreshing API key TTL: %v", err)
	}

	return true, nil
}

// AuthFilter allows the request to pass through if the request is not one of the protected URLs.
// False means the request will be blocked. True means the request will be allowed.
func AuthFilter(c *fiber.Ctx) bool {

	// Check if the request is one of the protected URLs
	for _, url := range protectedURL {
		if c.Path() == url {
			// Do not allow the request to pass through directly
			return false
		}
	}

	// Otherwise, allow the request to pass through directly
	return true
}
