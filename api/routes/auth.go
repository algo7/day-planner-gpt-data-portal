package routes

import (
	"github.com/algo7/day-planner-gpt-data-portal/api/controllers"
	"github.com/gofiber/fiber/v2"
)

// AuthRoutes is the route handler for the calendars API.
func AuthRoutes(app *fiber.App) {
	app.Get("/outlook/auth", controllers.GetAuthOutlook).Name("outlook_auth")
	app.Get("/outlook/oauth_redirect", controllers.GetOauthRedirectOutlook).Name("outlook_oauth_redirect")
	app.Get("/google/auth", controllers.GetAuthGoogle).Name("google_auth")
	app.Get("/google/oauth_redirect", controllers.GetOauthRedirectGoogle).Name("google_oauth_redirect")
	app.Get("/success", controllers.GetAuthSuccess).Name("success")
	app.Get("/apikey", controllers.GetAPIKey).Name("get_api_key")
	app.Post("/apikey", controllers.PostAPIKey).Name("post_api_key")
}
