package routes

import (
	"github.com/algo7/day-planner-gpt-data-portal/api/controllers"
	"github.com/gofiber/fiber/v2"
)

// AuthRoutes is the route handler for the calendars API.
func AuthRoutes(app *fiber.App) {
	app.Get("/auth", controllers.GetAuth)
	app.Get("/oauth_redirect", controllers.GetOauthRedirect)
}
