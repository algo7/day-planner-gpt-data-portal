package routes

import (
	"github.com/algo7/day-planner-gpt-data-portal/api/controllers"
	"github.com/gofiber/fiber/v2"
)

// AuthRoutes is the route handler for the calendars API.
func AuthRoutes(app *fiber.App) {
	app.Get("/v1/auth/oauth/google/device", controllers.GetAuthGoogleDevice).Name("google_auth_device")
	app.Get("/v1/auth/oauth", controllers.GetOAtuh).Name("oauth_auth")
	app.Get("/v1/auth/oauth/callback", controllers.GetOAuthCallBack).Name("oauth_callback")
	app.Get("/v1/auth/oauth/refresh", controllers.GetNewTokenFromRefreshToken).Name("oauth_refresh")
	app.Get("/v1/auth/success", controllers.GetAuthSuccess).Name("oauth_success")
	app.Get("/v1/auth/internal/apikey", controllers.GetAPIKey).Name("get_api_key")
	app.Post("/v1/auth/internal/apikey", controllers.PostAPIKey).Name("post_api_key")
}
