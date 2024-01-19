package routes

import (
	"github.com/algo7/day-planner-gpt-data-portal/api/controllers"
	"github.com/gofiber/fiber/v2"
)

// AuthRoutes is the route handler for the calendars API.
func AuthRoutes(app *fiber.App) {
	app.Get("/v1/auth/oauth/outlook", controllers.GetAuthOutlook).Name("outlook_auth")
	app.Get("/v1/auth/oauth/outlook/oauth_redirect", controllers.GetOauthRedirectOutlook).Name("outlook_oauth_redirect")
	app.Get("/v1/auth/oauth/google/auth", controllers.GetAuthGoogle).Name("google_auth")
	app.Get("/v1/auth/oauth/google/oauth_redirect", controllers.GetOauthRedirectGoogle).Name("google_oauth_redirect")
	app.Get("/v1/auth/success", controllers.GetAuthSuccess).Name("oauth_success")
	app.Get("/v1/auth/internal/apikey", controllers.GetAPIKey).Name("get_api_key")
	app.Post("/v1/auth/internal/apikey", controllers.PostAPIKey).Name("post_api_key")
}
